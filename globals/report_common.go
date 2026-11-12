package globals

import (
	"fmt"
	"log"
	"strconv"
)

type ReportGeneral struct {
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
	Kundenname  string          `json:"kundenname"`
	Uvhost      string          `json:"uvhost"`
	SpsHostname string          `json:"sps_hostname"`
	Version     string          `json:"version"`
	DockerName  string          `json:"docker_name"`
	Show        map[string]bool `json:"show"`
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
	ps.mu.Lock()
	defer ps.mu.Unlock()
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
	return fmt.Sprintf("%d", int(val/10))
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
