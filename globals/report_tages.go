package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"time"
)

type DailyReportData struct {
	Reportdatum string                 `json:"Reportdatum"`
	AussenTemp  string                 `json:"Aussen"`
	Data        []DailyReportExtension `json:"data"`
}
type DailyReportResponse struct {
	ReportGeneral
	DailyReportData
}

type ReportCsvLine struct {
	Datum   string `csv:"Datum"`
	RTavg   string `csv:"RTavg"`
	KTavg   string `csv:"KTavg"`
	Kzeit   string `csv:"Kzeit"`
	Ktakt   string `csv:"Ktakt"`
	ZT      string `csv:"ZT"`
	TW      string `csv:"TW"`
	CO2takt string `csv:"CO2takt"`
	O2takt  string `csv:"O2takt"`
	RTsoll  string `csv:"RTsoll"`
	RTmax   string `csv:"RTmax"`
	RTmin   string `csv:"RTmin"`
	KTmax   string `csv:"KTmax"`
	KTmin   string `csv:"KTmin"`
	O2max   string `csv:"O2max"`
	O2min   string `csv:"O2min"`
	CO2max  string `csv:"CO2max"`
	CO2min  string `csv:"CO2min"`
	LM      string `csv:"LM"`
	Frost   string `csv:"Frost"`
	Lzeit   string `csv:"Lzeit"`
}

// DailyReportExtension holds data needed for the daily report
// reflect package is used to compose the VictoriaMetrics MetricQL query string
// metric and short must be a find a metric, see middleware/metrics
// the method tag has to be a valid VictoriaMetrics MetricQL rollup function
// Additional Info: only gauges are used, even for counters
type DailyReportExtension struct {
	RTavg           string `metric:"uvr_prozess_celsius" short:"Temp" method:"avg_over_time" scale:"1"`
	KTavg           string `metric:"uvr_prozess_celsius" short:"TempKontroll" method:"avg_over_time" scale:"1"`
	Kzeit           string `metric:"uvr_seconds_total" short:"KaelteVentil" method:"delta" scale:"0"`
	Ktakt           string `metric:"uvr_takte_total" short:"KaelteTakte" method:"delta" scale:"0"`
	ZT              string `method:"ZeitPerTakt" scale:"1"`
	TW              string `metric:"uvr_menge_liter" short:"SaiTau" method:"delta" scale:"1"`
	CO2takt         string `metric:"uvr_takte_saison" short:"SaiCO2" method:"delta" scale:"0"`
	O2takt          string `metric:"uvr_takte_saison" short:"SaiO2Lft" method:"delta" scale:"0"`
	RTsoll          string `metric:"uvr_prozess_celsius" short:"SollTempEff" method:"avg_over_time" scale:"1"`
	RTmax           string `metric:"uvr_prozess_celsius" short:"Temp" method:"max_over_time" scale:"1"`
	RTmin           string `metric:"uvr_prozess_celsius" short:"Temp" method:"min_over_time" scale:"1"`
	KTmax           string `metric:"uvr_prozess_celsius" short:"TempKontroll" method:"max_over_time" scale:"1"`
	KTmin           string `metric:"uvr_prozess_celsius" short:"TempKontroll" method:"min_over_time" scale:"1"`
	O2max           string `metric:"uvr_prozess_ratio" short:"O2" method:"max_over_time" scale:"2"`
	O2min           string `metric:"uvr_prozess_ratio" short:"O2" method:"min_over_time" scale:"2"`
	CO2max          string `metric:"uvr_prozess_ratio" short:"CO2" method:"max_over_time" scale:"2"`
	CO2min          string `metric:"uvr_prozess_ratio" short:"CO2" method:"min_over_time" scale:"2"`
	LM              string `metric:"uvr_prozess_status" short:"LMsperr" method:"count" scale:"0"`
	Frost           string `metric:"uvr_prozess_status" short:"AlFrost" method:"count" scale:"0"`
	Lzeit           string `metric:"uvr_takte_total" short:"LuefterZeit" method:"delta" scale:"0"`
	Index           int    `method:"Index"`
	Raum            string `method:"Raum"`
	LagerungDatum   string `method:"LagerungDatum"`             // Datum der Einlagerung
	Sorte           string `method:"Sorte"`                     // Sorte, die Einglagert wurde
	Schaltdifferenz string `method:"Schaltdifferenz" scale:"1"` // Schaltdifferenz
	TwSoll          string `method:"TwSoll" scale:"1"`          // Tauwasser Sollwert
	CO2Soll         string `method:"CO2Soll" scale:"2"`         // CO2 Sollwert
	O2Soll          string `method:"O2Soll" scale:"2"`          // O2 Sollwert
	Menge           string `method:"Menge" scale:"0"`           // Menge Ware im Lager
}

// // Extension to the csv logData_R01.csv, this data is gatered from victoriametrics
// type MetricsReportData struct {
// 	AbtauTkt        string // Abtau Takte
// 	AbtauZt         string // Abtau Zeit
// 	AbtElTkt        string // Abtau Elektro Takte
// 	AbtElzt         string // Abtau Elektro Zeit
// }

// DailyReportGatherer acquires yesterdays date für the report and data from the Plc from the global data structure.
func (gs *GlobalState) DailyReportGatherer() (err error) {
	time.Sleep(7 * time.Second) // Wait a bit to ensure that PLC values are updated

	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	log.Println("📃 DailyReportGatherer started...")

	// init data if empty or nil
	if gs.Report.Daily == nil {
		gs.Report.Daily = make(map[string]DailyReportExtension)
	}

	dt := time.Now().AddDate(0, 0, -1).Format("02.01.2006")

	gs.Stats.Reportdatum = dt
	gs.Cfg.Plc.SWSettings.Reportdatum = dt

	var e DailyReportExtension

	// // Get data from Flat.Values
	mapRTavg := gs.Plc.GetArrayString("GVL.gi_RaumTemp_24h_AVG")
	mapKTavg := gs.Plc.GetArrayString("GVL.gi_KontTemp_24h_AVG")
	mapKzeit := gs.Plc.GetArrayString("Z.ZV_Rmm06_24hKaltVZeit_arw")
	mapKtakt := gs.Plc.GetArrayString("Z.ZV_Rmm07_24hKaltVTakt_arw")
	mapTW := gs.Plc.GetArrayString("Z.ZV_Rmm05_24hTau_arw")
	mapCO2takt := gs.Plc.GetArrayString("Z.ZV_Rmm02_24hCO2_arw")
	mapO2takt := gs.Plc.GetArrayString("Z.ZV_Rmm03_24hO2_arw")
	mapRTsoll := gs.Plc.GetArrayString("BD.BA_Rmm14_SollTemp_ari")
	mapRTmax := gs.Plc.GetArrayString("Z.Max_Temp_24h_ari")
	mapRTmin := gs.Plc.GetArrayString("Z.Min_Temp_24h_ari")
	mapKTmax := gs.Plc.GetArrayString("Z.Max_KontTemp_24h_ari")
	mapKTmin := gs.Plc.GetArrayString("Z.Min_KontTemp_24h_ari")
	mapO2max := gs.Plc.GetArrayString("Z.Max_O2_24h_arw")
	mapO2min := gs.Plc.GetArrayString("Z.Min_O2_24h_arw")
	mapCO2max := gs.Plc.GetArrayString("Z.Max_CO2_24h_arw")
	mapCO2min := gs.Plc.GetArrayString("Z.Min_CO2_24h_arw")
	mapLM := gs.Plc.GetArrayString("Z.ZV_Rmm09_24hKaltSperInt_arw")
	mapFrost := gs.Plc.GetArrayString("Z.ZV_Rmm10_24hFrostTakt_arw")
	mapLzeit := gs.Plc.GetArrayString("Z.ZV_Rmm11_24hVerdaZeit_arw")

	//mapJahr := gs.Plc.GetArrayString("GVL.garw_Jahr")
	mapMonat := gs.Plc.GetArrayString("GVL.garw_Monat")
	mapTag := gs.Plc.GetArrayString("GVL.garw_Tag")
	mapDiffRaumTemp := gs.Plc.GetArrayString("PR.PA_Rmm12_DiffRaumTemp_arb")
	mapTwSoll := gs.Plc.GetArrayString("BD.BA_Rmm08_TauMax_arw")
	mapCO2Soll := gs.Plc.GetArrayString("PR.PA_Rmm33_CO2Soll_arw")
	mapO2Soll := gs.Plc.GetArrayString("PR.PA_Rmm31_O2Soll_arw")
	mapMenge := gs.Plc.GetArrayString("GVL.gw_E_Menge")

	for i := 0; i < gs.Cfg.Plc.SWSettings.Raumanzahl; i++ {
		e.Index = i
		e.Raum = gs.Cfg.Plc.SWSettings.Raumnummer[i]

		e.RTavg = unMapArrayString(mapRTavg, i)
		e.KTavg = unMapArrayString(mapKTavg, i)
		e.Kzeit = unMapArrayString(mapKzeit, i)
		e.Ktakt = unMapArrayString(mapKtakt, i)
		e.ZT = divideFloatStrings(e.Ktakt, e.Kzeit)
		e.TW = formatTauwasser(unMapArrayString(mapTW, i))
		e.CO2takt = unMapArrayString(mapCO2takt, i)
		e.O2takt = unMapArrayString(mapO2takt, i)
		e.RTsoll = unMapArrayString(mapRTsoll, i)
		e.RTmax = unMapArrayString(mapRTmax, i)
		e.RTmin = unMapArrayString(mapRTmin, i)
		e.KTmax = unMapArrayString(mapKTmax, i)
		e.KTmin = unMapArrayString(mapKTmin, i)
		e.O2max = unMapArrayString(mapO2max, i)
		e.O2min = unMapArrayString(mapO2min, i)
		e.CO2max = unMapArrayString(mapCO2max, i)
		e.CO2min = unMapArrayString(mapCO2min, i)
		e.LM = unMapArrayString(mapLM, i)
		e.Frost = unMapArrayString(mapFrost, i)
		e.Lzeit = unMapArrayString(mapLzeit, i)

		//Einlagerungsdatum mit Jahrsangabe, ist aber zu lang
		//e.LagerungDatum = fmt.Sprintf("%s.%s.%s", unMapArrayString(mapTag, i), unMapArrayString(mapMonat, i), unMapArrayString(mapJahr, i))
		e.LagerungDatum = fmt.Sprintf("%s.%s", unMapArrayString(mapTag, i), unMapArrayString(mapMonat, i))
		e.Schaltdifferenz = unMapArrayString(mapDiffRaumTemp, i)
		e.TwSoll = unMapArrayString(mapTwSoll, i)
		e.CO2Soll = unMapArrayString(mapCO2Soll, i)
		e.O2Soll = unMapArrayString(mapO2Soll, i)
		e.Menge = unMapArrayString(mapMenge, i)

		/*
			Not used in the moment, but usefull for future extensions
				Z.ZV_Rmm01_24hAbtauEl_arw
				Z.ZV_Rmm04_24hN2_arw
				Z.ZV_Rmm08_24hKaltFreiEx_arw
				Z.ZV_Rmm12_24hAbtauTakt_arw
		*/
		//log.Printf("📃 Daily Report idx: %d\n%#v\n----------", i, e)

		gs.Report.Daily[strconv.Itoa(i+1)] = e
	}

	log.Println("📃 DailyReportGatherer finished...")

	return nil

}

func (gs *GlobalState) ToDailyReportJson(startDate string) []byte {
	var d DailyReportResponse

	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	// Data for Daily Report
	d.Data = make([]DailyReportExtension, gs.Cfg.Plc.SWSettings.Raumanzahl)

	d.ReportGeneral = gs.getGeneralReportInfo()
	fmt.Println("Fetching daily report data for start date:", startDate)
	if startDate != "" {
		d.Reportdatum = startDate
		gs.getDailyReportDataFromDB(d.Data, startDate)

	} else {
		for _, val := range gs.Report.Daily {
			d.Data[val.Index] = val
		}
		d.Reportdatum = gs.Stats.Reportdatum
	}

	json, _ := json.MarshalIndent(d, "", "  ")

	return json
}

/*
	{
		"metric": {
			"base": "EOA",
			"descr": "Saison Tauwassermenge",
			"instance": "nodered:1881",
			"job": "ulovisor",
			"ort": "1",
			"short": "SaiTau",
			"symbol": "ZS_Rmm05_SaiTau_ardw"
		},
		"value": [
			1793573999,
			"0"
		]
	},
*/
type RangeQueryResult struct {
	Metric string `json:"metric"`
	Value  []any  `json:"value"`
}

type RangeQueryMetricMap struct {
	Base     string `json:"base"`
	Descr    string `json:"descr"`
	Instance string `json:"instance"`
	Job      string `json:"job"`
	Ort      string `json:"ort"`
	Short    string `json:"short"`
	Symbol   string `json:"symbol"`
}

// Historischer oder Aktueller Tagesreport aus SPS oder historischer aus der Datenbank
func (gs *GlobalState) getDailyReportDataFromDB(data []DailyReportExtension, startDate string) {
	refl := reflect.TypeOf(data[0])
	fields := reflect.VisibleFields(refl)

	ortPart := ortString(len(data))

	// First Run for metrics
	for _, field := range fields {
		// Ignore special fields without metric tag
		if _, ok := field.Tag.Lookup("metric"); !ok {
			continue
		}
		timeUnix, err := time.ParseInLocation("2006-01-02T15:04:05", startDate+"T23:59:59", time.Local)
		if err != nil {
			fmt.Printf("Error parsing time '%sT23:59:59': %v\n", startDate, err)
			continue
		}
		query := fmt.Sprintf("query?query=%s(%s{short=\"%s\",%s})&step=1d&time=%d", field.Tag.Get("method"), field.Tag.Get("metric"), field.Tag.Get("short"), ortPart, timeUnix.Unix())
		res, err := gs.QueryMetrics(query)
		if err != nil {
			fmt.Printf("Error querying metrics for '%s': %s\n%s\n", field.Name, query, err)
			continue
		}

		for _, r := range res {
			ortStr, ok := r["metric"].(map[string]any)["ort"].(string)
			if !ok {
				fmt.Printf("Error: 'ort' field not found in metric\n")
				continue
			}
			ortNr, err := strconv.Atoi(ortStr)
			if err != nil {
				fmt.Printf("Error converting 'ort' to int: %v\n", err)
				continue
			}
			valarray, ok := r["value"].([]any)
			if !ok {
				fmt.Printf("Error: 'value' field not found or invalid\n")
				continue
			}
			var valstr string
			if len(valarray) >= 2 {
				valstr = valarray[1].(string)
			}
			valFloat, err := strconv.ParseFloat(valstr, 64)

			scale := field.Tag.Get("scale")
			switch scale {
			case "0":
				valstr = fmt.Sprintf("%d", int(valFloat))
			case "1":
				valstr = fmt.Sprintf("%.1f", valFloat)
			case "2":
				valstr = fmt.Sprintf("%.2f", valFloat)
			}

			ortIdx := ortNr - 1
			if ortIdx < 0 || ortIdx >= len(data) {
				fmt.Printf("Error: 'ort' index %d out of range for data length %d\n", ortIdx, len(data))
				continue
			}
			fieldValue := reflect.ValueOf(&data[ortIdx]).Elem().FieldByName(field.Name)
			if !fieldValue.IsValid() || !fieldValue.CanSet() || fieldValue.Kind() != reflect.String {
				fmt.Printf("Error: cannot set field '%s' on DailyReportExtension\n", field.Name)
				continue
			}
			fieldValue.SetString(valstr)
		}
	}

	// Second Run for hacky stuff

}

func ortString(anz int) (s string) {
	s = "ort=~\""
	for i := range anz {
		s += strconv.Itoa(i + 1)
		if i < anz-1 {
			s += "|"
		}
	}
	return s + "\""
}
