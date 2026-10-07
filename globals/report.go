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

type DailyReportResponse struct {
	Kundenname  string                 `json:"kundenname"`
	Reportdatum string                 `json:"Reportdatum"`
	Uvhost      string                 `json:"uvhost"`
	SpsHostname string                 `json:"sps_hostname"`
	Version     string                 `json:"version"`
	DockerName  string                 `json:"docker_name"`
	Show        map[string]bool        `json:"show"`
	AussenTemp  string                 `json:"Aussen"`
	Data        []DailyReportExtension `json:"data"`
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
	mapTwSoll := gs.Plc.GetArrayString("GVL.gw_Sollentfeuchtung")
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

func (gs *GlobalState) ToDailyReportJson() []byte {
	var d DailyReportResponse

	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()
	d.Data = make([]DailyReportExtension, gs.Cfg.Plc.SWSettings.Raumanzahl)
	for _, val := range gs.Report.Daily {
		//log.Printf("📃 ToDailyReportJson key: %s, val: %#v", key, val)
		d.Data[val.Index] = val
	}

	d.Kundenname = gs.Cfg.Plc.SWSettings.Kunde
	d.Reportdatum = gs.Stats.Reportdatum
	d.Uvhost = gs.Cfg.Plc.Base
	d.SpsHostname = "" // TODO herausfinden, wie man den Hostnamen der SPS bekommt
	d.Version = ""     // TODO herausfinden, wie man die Version der SPS bekommt
	d.DockerName = ""  // TODO herausfinden, wie man den Dockernamen bekommt

	/* // Add keys if report section shoud be shown. e.g. if there is no o2 measurement, do not show that section
	const g = global.get("Konfig");
	const anz = g.Gb_Anzahl_Raeume;
	const show_RT = g.Gar_RaumFuehl.slice(0, anz).includes(true);
	const show_KT = g.Gar_Kontroll.slice(0, anz).includes(true);
	const show_TW = g.Gar_Tauwasser.slice(0,anz).includes(true);
	const show_O2 = g.Gar_O2Luefter.slice(0, anz).includes(true);
	const show_CO2 = g.Gar_CO2_Aktion.slice(0, anz).includes(true);

	msg.payload.show = {
		RT: show_RT,
		KT: show_KT,
		TW: show_TW,
		O2: show_O2,
		CO2: show_CO2
	}*/
	d.Show = make(map[string]bool)

	d.Show["RT"] = getKonfigBooleans("Gar_RaumFuehl", gs)
	d.Show["KT"] = getKonfigBooleans("Gar_Kontroll", gs)
	d.Show["TW"] = getKonfigBooleans("Gar_Tauwasser", gs)
	d.Show["O2"] = getKonfigBooleans("Gar_O2Luefter", gs)
	d.Show["CO2"] = getKonfigBooleans("Gar_CO2_Aktion", gs)

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
