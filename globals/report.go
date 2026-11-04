package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

/*
"kundenname": "Herzapfelhof Luehs",
"Reportdatum": "07.10.2025",
"uvhost": "RLU",
"sps_hostname": "RLU-UV4",
"version": "RLU.05_PJ",
"docker_name": "Rolf Lühs, Herzapefelhof",

	"show": {
	  "RT": true,
	  "KT": true,
	  "TW": true,
	  "O2": true,
	  "CO2": true
	},

"data": []
*/
type ReportGeneral struct {
	Kundenname  string          `json:"kundenname"`
	Uvhost      string          `json:"uvhost"`
	SpsHostname string          `json:"sps_hostname"`
	Version     string          `json:"version"`
	DockerName  string          `json:"docker_name"`
	Show        map[string]bool `json:"show"`
}
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

type DailyReportExtension struct {
	RTavg           string
	KTavg           string
	Kzeit           string
	Ktakt           string
	ZT              string
	TW              string
	CO2takt         string
	O2takt          string
	RTsoll          string
	RTmax           string
	RTmin           string
	KTmax           string
	KTmin           string
	O2max           string
	O2min           string
	CO2max          string
	CO2min          string
	LM              string
	Frost           string
	Lzeit           string
	Index           int
	Raum            string
	LagerungDatum   string // Datum der Einlagerung
	Sorte           string // Sorte, die Eingleaget wurde
	Schaltdifferenz string // Schaltdifferenz
	TwSoll          string // Tauwasser Sollwert
	CO2Soll         string // CO2 Sollwert
	O2Soll          string // O2 Sollwert
	Menge           string // Menge Ware im Lager

}

// // Extension to the csv logData_R01.csv, this data is gatered from victoriametrics
// type MetricsReportData struct {
// 	OutTempAvg      string // Average outside temperature
// 	OutTempMin      string // Minimum outside temperature
// 	OutTempMax      string // Maximum outside temperature
// 	AbtauTkt        string // Abtau Takte
// 	AbtauZt         string // Abtau Zeit
// 	AbtElTkt        string // Abtau Elektro Takte
// 	AbtElzt         string // Abtau Elektro Zeit
// 	LagerungDatum   string // Datum der Einlagerung
// 	Sorte           string // Sorte, die Einglagert wurde
// 	Schaltdifferenz string // Schaltdifferenz
// }

// DailyReportGatherer acquires yesterdays date für the report and data from the Plc from the global data structure.
func (gs *GlobalState) DailyReportGatherer() (err error) {
	time.Sleep(7 * time.Second) // Wait a bit to ensure that PLC values are updated

	gs.Report.mu.Lock()

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
		e.TW = unMapArrayString(mapTW, i)
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

	gs.Report.mu.Unlock()

	return nil

}

func unMapArrayString(m map[string]string, idx int) string {
	key := strconv.Itoa(idx)
	val, ok := m[key]
	if !ok {
		log.Printf("📃 Error: key %s not found in map", key)
		return "---"
	}
	return val
}

// GetArrayString gets an array of strings from the Flat keys map. It returns a map where the keys are the indices as strings and the values are the corresponding array elements.
func (ps *PlcState) GetArrayString(key string) (sval map[string]string) {
	v, ok := ps.Flat[key]
	if !ok {
		log.Printf("📃 Error findVal. %s is not found in Plc.Flat...", key)
	}
	//check if there are enough elemets in the values map
	if len(v.Values.ArrayString) != v.Length {
		log.Printf("📃 Error findVal. %s has not enough elements in Values.ArrayString... expected: %d, got: %d", key, v.Length, len(v.Values.ArrayString))
	}
	return v.Values.ArrayString
}

func (gs *GlobalState) findVal(key string) (sval map[string]string) {
	v, ok := gs.Plc.Flat[key]
	if !ok {
		log.Printf("📃 Error findVal. %s is not found in Plc.Flat...", key)
	}

	//check if there are enough elemets in the values map
	if len(v.Values.ArrayString) != v.Length {
		log.Printf("📃 Error findVal. %s has not enough elements in Values.ArrayString... expected: %d, got: %d", key, v.Length, len(v.Values.ArrayString))
	}

	return v.Values.ArrayString
}

func divideFloatStrings(z string, n string) (res string) {
	fz, err := strconv.ParseFloat(z, 64)
	if err != nil {
		log.Printf("📃 Error converting string to float: %v", err)
		return "---"
	}
	fn, err := strconv.ParseFloat(n, 64)
	if err != nil {
		log.Printf("📃 Error converting string to float: %v", err)
		return "---"
	}
	if fn == 0 {
		return "0"
	}
	fr := fz / fn
	return strconv.FormatFloat(fr, 'f', 1, 64)
}

func (gs *GlobalState) getGeneralReportInfo() ReportGeneral {

	show := make(map[string]bool)
	show["RT"] = getKonfigBooleans("Gar_RaumFuehl", gs)
	show["KT"] = getKonfigBooleans("Gar_Kontroll", gs)
	show["TW"] = getKonfigBooleans("Gar_Tauwasser", gs)
	show["O2"] = getKonfigBooleans("Gar_O2Luefter", gs)
	show["CO2"] = getKonfigBooleans("Gar_CO2_Aktion", gs)

	return ReportGeneral{
		Kundenname:  gs.Cfg.Plc.SWSettings.Kunde,
		Uvhost:      gs.Cfg.Plc.Base,
		SpsHostname: "",                      // TODO herausfinden, wie man den Hostnamen der SPS bekommt
		Version:     gs.Stats.CodesysVersion, // Version der Codesys SPS
		DockerName:  "",                      // TODO herausfinden, wie man den Dockernamen bekommt
		Show:        show,
	}
}

func (gs *GlobalState) ToDailyReportJson() []byte {
	var d DailyReportResponse

	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	// Data for Daily Report
	d.Data = make([]DailyReportExtension, gs.Cfg.Plc.SWSettings.Raumanzahl)
	for _, val := range gs.Report.Daily {
		d.Data[val.Index] = val
	}

	d.ReportGeneral = gs.getGeneralReportInfo()
	d.Reportdatum = gs.Stats.Reportdatum

	json, _ := json.MarshalIndent(d, "", "  ")

	return json
}

func getKonfigBooleans(key string, gs *GlobalState) bool {

	e, ok := gs.Cfg.Plc.Konfig[key]
	if !ok {
		log.Printf("Warning: key %s not found in Konfig map", key)
		return true
	}

	kar, ok := e.([]bool)
	if !ok {
		log.Printf("Warning: key %s has wrong type in Konfig map", key)
		return true
	}

	for _, b := range kar {
		if b {
			return true
		}
	}

	return false
}

func (report *ReportData) LenAll() int {
	report.mu.Lock()
	defer report.mu.Unlock()
	return len(report.All)
}

func (report *ReportData) LenAllOrt(idx int) int {
	report.mu.Lock()
	defer report.mu.Unlock()

	return len(report.All[idx])
}

func (report *ReportData) GetAllOrt(idx int) []ReportCsvLine {
	report.mu.Lock()
	defer report.mu.Unlock()
	return report.All[idx]
}

type SaisonReportRow struct {
	Raumnummer     string `json:"raumnummer"`
	Raumtemperatur string `json:"raumtemperatur"`
	O2_Wert        string `json:"o2_wert"`
	CO2_Wert       string `json:"co2_wert"`
	Tauwasser      string `json:"tauwasser"`
	Kuehlzeit      string `json:"kuehlzeit"`
	Kuehltakte     string `json:"kuehltakte"`
}

type SaisonReportResponse struct {
	ReportGeneral    ReportGeneral              `json:"report_general"`
	SaisonReportRows map[string]SaisonReportRow `json:"saison_report_rows"`
	StartDates       map[string]string          `json:"start_date"`
	StartTimesOfDay  map[string]string          `json:"start_times_of_day"`
	EndDate          string                     `json:"end_date"`
	EndTimeOfDay     string                     `json:"end_time_of_day"`
	TypstData        []string                   `json:"typst_data"`
	colCount         int
}

// reportStringMap stores a string key and a fixed-length array of two strings.
var reportStringMap = map[string][2]string{
	"Raumtemperatur": {"uvr_prozess_celsius", "Temp"},
	"O2_Wert":        {"uvr_prozess_ratio", "O2ist"},
	"CO2_Wert":       {"uvr_prozess_ratio", "CO2ist"},
	"Tauwasser":      {"uvr_menge_liter", "SaiTau"},
	"Kuehlzeit":      {"uvr_seconds_total", "KaelteVentil"},
	"Kuehltakte":     {"uvr_takte_total", "KaelteTakte"},
}

// ToSaisonReportJson gathers data for saisonreport and generates JSON
func (gs *GlobalState) ToSaisonReportJson(startDate string, endDate string) []byte {
	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	sr := SaisonReportResponse{
		ReportGeneral:    gs.getGeneralReportInfo(),
		SaisonReportRows: make(map[string]SaisonReportRow),
		StartDates:       make(map[string]string),
		colCount:         7 + 1, // SaisonReportRow + Date field
	}
	// We need startData and endDate to compose the query. However, dates are optional. Thus followin default behavior:
	// endDate defaults to now if not provided
	var endTime time.Time

	if !validateDateString(endDate) {
		endTime = time.Now()
		sr.EndDate = endTime.Format("2006-01-02")
		sr.EndTimeOfDay = endTime.Format("15:04:05")

	} else {
		sr.EndDate = endDate
		sr.EndTimeOfDay = "23:59:59"
		endTime, _ = time.ParseInLocation("2006-01-02T15:04:05", sr.EndDate+"T"+sr.EndTimeOfDay, time.Local)
	}
	// startDate defaults to what is set in the PLC
	mStartDates := gs.getLagerdatenMap()
	if validateDateString(startDate) {
		for k := range mStartDates {
			mStartDates[k] = startDate
		}
	} else { // no startDate provided, use default from PLC
		log.Printf("📃 Info: Using default startDate with 'Einlagerungsdatum'")
	}

	for ort := range mStartDates {
		i, err := strconv.Atoi(ort)
		if err != nil {
			log.Printf("Error converting ort to int: %v", err)
			continue
		}
		if i-1 < 0 || i-1 >= len(gs.Cfg.Plc.SWSettings.Raumnummer) {
			log.Printf("Error: Raumnummer index out of range for ort %s", ort)
			continue
		}
		sr.SaisonReportRows[ort] = SaisonReportRow{Raumnummer: gs.Cfg.Plc.SWSettings.Raumnummer[i-1]}
	}

	deltaQueryList := []string{"uvr_menge_liter", "uvr_seconds_total", "uvr_takte_total"}
	var queryStr string

	for kind, ar := range reportStringMap {

		doDeltaQuery := contains(deltaQueryList, ar[0])

		for ort, startDateStr := range mStartDates {
			startDateUTC, err := time.ParseInLocation("2006-01-02T15:04:05", startDateStr+"T00:00:00", time.Local)
			if err != nil {
				log.Printf("Error in SaisonReport parsing startDate '%s' for ort %s : %v", startDateStr, ort, err)
				continue
			}

			if doDeltaQuery {
				queryStr = composeDeltaQueryString(ar[0], ar[1], ort, startDateUTC.Unix(), endTime.Unix())
			} else {
				queryStr = fmt.Sprintf("%s{short='%s',ort='%s'}&time=%d", ar[0], ar[1], ort, endTime.Unix())
			}

			res, err := gs.QueryMetrics(queryStr)
			if err != nil || len(res) == 0 {
				log.Printf("Error in SaisonReport querying metrics: %v, query: %s", err, queryStr)
				log.Printf("SaisonReport Query result for ort %s: %v", ort, res)
			}
			//log.Printf("DEBUG SaisonReport Query:\n%s\nRes\n%v", queryStr, res)

			var valStr string

			lastResultIndex := len(res) - 1
			if len(res) <= 0 {
				log.Printf("Error: empty result for ort %s", ort)
				valStr = "---"
			} else {
				valArray, ok := res[lastResultIndex]["value"].([]any)
				if !ok || len(valArray) < 2 {
					log.Printf("Error extracting value from result for ort %s: %v", ort, res)
					continue
				}
				valStr, ok = valArray[1].(string)
				if !ok {
					log.Printf("Error converting value to string for ort %s: %v", ort, valArray[1])
					continue
				}
			}

			// TODO Umrechnung anwenden
			row := sr.SaisonReportRows[ort]
			switch kind {
			case "Raumtemperatur":
				row.Raumtemperatur = formatFlint1(valStr)
			case "O2_Wert":
				row.O2_Wert = formatFlint2(valStr)
			case "CO2_Wert":
				row.CO2_Wert = formatFlint2(valStr)
			case "Kuehltakte":
				row.Kuehltakte = valStr
			case "Kuehlzeit":
				row.Kuehlzeit = formatSekundenToDayHoursMinutes(valStr)
			case "Tauwasser":
				row.Tauwasser = formatFlint1(valStr)
			}
			sr.SaisonReportRows[ort] = row
		}
	}

	typst_data := make([]string, sr.colCount*gs.Cfg.Plc.SWSettings.Raumanzahl)

	for rk, rObj := range sr.SaisonReportRows {
		// log.Printf("DEBUG Report Row for rk %s: %v", rk, rObj)
		i, err := strconv.Atoi(rk)
		if err != nil {
			log.Printf("Error converting rk to int: %v", err)
			continue
		}
		i = (i - 1) * sr.colCount

		typst_data[i] = rObj.Raumnummer
		typst_data[i+1] = mStartDates[rk]
		typst_data[i+2] = rObj.Raumtemperatur
		typst_data[i+3] = rObj.O2_Wert
		typst_data[i+4] = rObj.CO2_Wert
		typst_data[i+5] = rObj.Tauwasser
		typst_data[i+6] = rObj.Kuehlzeit
		typst_data[i+7] = rObj.Kuehltakte
	}
	sr.TypstData = typst_data

	json, _ := json.MarshalIndent(sr, "", "  ")

	return json
}

// getLagedatenMap retrieves the start dates for each room from the PLC and returns them as a map with keys that represent the room index begining with 1.
func (gs *GlobalState) getLagerdatenMap() map[string]string {

	m := make(map[string]string)

	mapJahr := gs.Plc.GetArrayString("GVL.garw_Jahr")
	mapMonat := gs.Plc.GetArrayString("GVL.garw_Monat")
	mapTag := gs.Plc.GetArrayString("GVL.garw_Tag")

	// k is index as string beginning with 0
	for k, jahr := range mapJahr {

		monat, err := strconv.Atoi(mapMonat[k])
		if err != nil {
			log.Printf("📃 Error converting month to int: %v", err)
			continue
		}
		tag, err := strconv.Atoi(mapTag[k])
		if err != nil {
			log.Printf("📃 Error converting tag to int: %v", err)
			continue
		}

		// convert k to int and add 1 and convert back to string
		kInt, err := strconv.Atoi(k)
		if err != nil {
			log.Printf("📃 Error converting key to int: %v", err)
			continue
		}
		kInt++
		kfrom1 := strconv.Itoa(kInt)

		m[kfrom1] = fmt.Sprintf("%s-%02d-%02d", jahr, monat, tag)
	}
	return m
}

/*
	example for the query

https://eoa-pi5.uv4.sw/vm/api/v1/query

	?query=delta(uvr_takte_total{short='KaelteTakte', ort='1'})
	&step=30d
	&time=2026-10-28

time=2026-10-28 looking start date, if omitted now() is used
step=30d means 30 days lookbehind. If step is larger than the availabe date range, we get 2 results!

	{
	  "status": "success",
	  "data": {
	    "resultType": "vector",
	    "result": [
	      {
	        "metric": {
	          "base": "EOA",
	          "descr": "Saison Tauwassermenge",
	          "instance": "nodered:1881",
	          "job": "ulovisor",
	          "ort": "7",
	          "short": "SaiTau",
	          "symbol": "ZS_Rmm05_SaiTau_ardw"
	        },
	        "value": [
	          1793265827,
	          "7150"
	        ]
	      }
	    ]
	  },
	  "stats": {
	    "seriesFetched": "1",
	    "executionTimeMsec": 11
	  }
	}
*/
func formatSekundenToDayHoursMinutes(sekStr string) string {
	sek, err := strconv.Atoi(sekStr)
	if err != nil {
		return "---"
	}
	days := sek / 86400
	hours := (sek % 86400) / 3600
	minutes := (sek % 3600) / 60
	if days == 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
}

// Achtung: Hacky, because the Tauwasser value is not set correctly in the PLC, scale=1 is missing
func formatTauwasser(valStr string) string {
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "---"
	}
	return fmt.Sprintf("%.1f", val/10)
}

func formatFlint1(valStr string) string {
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "---"
	}
	return fmt.Sprintf("%.1f", val)
}

func formatFlint2(valStr string) string {
	val, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return "---"
	}
	return fmt.Sprintf("%.2f", val)
}
