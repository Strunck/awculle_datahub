package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

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

type ReportOrtResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	Raumnummer  string          `json:"raumnummer"`
	Kundenname  string          `json:"kundenname"`
	Reportdatum string          `json:"Reportdatum"`
	Uvhost      string          `json:"uvhost"`
	SpsHostname string          `json:"sps_hostname"`
	Version     string          `json:"version"`
	DockerName  string          `json:"docker_name"`
	Data        []ReportCsvLine `json:"data"`
}

func (gs *GlobalState) ToJson(ort string) []byte {
	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	ortPlus1, err := strconv.Atoi(ort)
	if err != nil {
		log.Println("Invalid ort parameter:", ort)
		return []byte("{\"success\":false,\"message\":\"invalid ort parameter\", \"data\":[]}")
	}
	log.Println("Requested Report for Ort:", ort, "-> ortplus1:", ortPlus1)

	var repRes ReportOrtResponse
	repRes.Data = make([]ReportCsvLine, 0)

	dlen := gs.Report.LenAll()

	log.Println("Report Data Length:", dlen)
	if dlen == 0 {
		repRes.Success = false
		repRes.Message = "No report data available"
		repRes.Data = make([]ReportCsvLine, 0)

	} else { // Report data should be available

		if ortPlus1 > 0 && ortPlus1 <= dlen {
			anz := gs.Report.LenAllOrt(ortPlus1 - 1)
			repRes.Success = true
			repRes.Message = "OK Anzahl Datensätze " + strconv.Itoa(anz)
			repRes.Raumnummer = gs.Cfg.Plc.SWSettings.Raumnummer[ortPlus1-1]
			repRes.Kundenname = gs.Cfg.Plc.SWSettings.Kunde
			repRes.Reportdatum = fmt.Sprintf("%s", time.Now().Format("2.Jan 2006 15:04"))
			repRes.Uvhost = gs.Cfg.Plc.Base
			repRes.SpsHostname = "" // TODO herausfinden, wie man den Hostnamen der SPS bekommt
			repRes.Version = gs.Stats.CodesysVersion
			repRes.DockerName = "" // TODO herausfinden, wie man den Dockernamen bekommt
			repRes.Data = gs.Report.GetAllOrt(ortPlus1 - 1)
		} else {
			repRes.Success = false
			repRes.Message = "Parameter Ort " + ort + " out of range, must be between 1 and " + strconv.Itoa(dlen)
			repRes.Data = gs.Report.GetAllOrt(ortPlus1 - 1)
		}
	}

	res, err := json.MarshalIndent(repRes, "", "  ")
	if err != nil {
		log.Println("Error marshalling ReportOrtResponse to JSON:", err)
		return []byte("{\"success\":false,\"message\":\"error marshalling ReportOrtResponse to JSON\", \"data\":[]}")
	}
	return res
}
