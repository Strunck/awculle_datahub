package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (report *ReportData) lenAllOrt(idx int) int {
	return len(report.All[idx])
}

func (report *ReportData) getAllOrt(idx int) []ReportCsvLine {
	return report.All[idx]
}

type ReportOrtResponse struct {
	Success     bool            `json:"success"`
	Message     string          `json:"message"`
	Raumnummer  string          `json:"raumnummer"`
	Reportdatum string          `json:"Reportdatum"`
	General     ReportGeneral   `json:"general"`
	Data        []ReportCsvLine `json:"data"`
}

func (gs *GlobalState) OrtToJson(ort, query, startDate string) []byte {
	gs.Report.mu.Lock()
	defer gs.Report.mu.Unlock()

	log.Printf("DEBUG OrtToJson called with ort: %s, query: %s, startDate: %s", ort, query, startDate)

	var err error
	var endDate string
	var ortPlus1 int
	var days time.Duration
	var repRes ReportOrtResponse

	repRes.General = gs.getGeneralReportInfo()

	switch query {
	case "wochenreport":
		days = 7 * 24 * time.Hour
	case "monatsreport":
		days = 30 * 24 * time.Hour
	case "zweiwochendiagramm":
		days = 14 * 24 * time.Hour
	case "zweiwochenreport":
		days = 14 * 24 * time.Hour
	}

	// Default to yesterday for the endDate String
	enddt := time.Now().Add(-24 * time.Hour)

	endDate = fmt.Sprintf("%02d.%02d.%4d", enddt.Day(), enddt.Month(), enddt.Year())
	log.Printf("DEBUG OrtToJson: default endDate: %s\n", endDate)

	startdt, err := time.ParseInLocation(time.DateOnly, startDate, time.Local)
	if err != nil {
		log.Printf("Invalid startDate: '%s' parameter. Using yesterday as endDate: %s", startDate, endDate)
	} else {
		enddt = startdt.Add(days)
		endDate = fmt.Sprintf("%02d.%02d.%4d", enddt.Day(), enddt.Month(), enddt.Year())
	}
	log.Printf("DEBUG OrtToJson: startDate: %s, endDate: %s\n", startDate, endDate)

	ortPlus1, err = strconv.Atoi(ort)
	if err != nil {
		log.Println("Invalid ort parameter:", ort)
		return []byte("{\"success\":false,\"message\":\"invalid ort parameter\", \"data\":[]}")
	}
	log.Println("Requested Report for Ort:", ort, "-> ortplus1:", ortPlus1)

	dlen := gs.Report.lenAllOrt(ortPlus1 - 1)
	log.Println("Total Report Tage:", dlen)

	repRes.Data = make([]ReportCsvLine, dlen)

	log.Println("Report Data Length:", dlen)
	if dlen == 0 {
		repRes.Success = false
		repRes.Message = "No report data available"
		repRes.Data = make([]ReportCsvLine, 0)

	} else { // Report data should be available

		if ortPlus1 > 0 && ortPlus1 <= dlen {
			anz := gs.Report.lenAllOrt(ortPlus1 - 1)

			repRes.Success = true
			repRes.Raumnummer = gs.Cfg.Plc.SWSettings.Raumnummer[ortPlus1-1]
			repRes.Message = "OK Anzahl Datensätze " + strconv.Itoa(anz)
			repRes.Reportdatum = fmt.Sprintf("%s", time.Now().Format("2.Jan 2006 15:04"))

			repRes.Data = gs.Report.getAllOrt(ortPlus1 - 1)

			repRes.Data = findReportData(repRes.Data, endDate)

		} else {
			repRes.Success = false
			repRes.Message = "Parameter Ort " + ort + " out of range, must be between 1 and " + strconv.Itoa(dlen)
			repRes.Data = gs.Report.getAllOrt(ortPlus1 - 1)
		}

		// Reformat response, O2/CO2-Takte no digits,  O2/CO2 Takte 2-digits
		for i, el := range repRes.Data {
			repRes.Data[i].O2takt = formatNoDigit(el.O2takt)
			repRes.Data[i].CO2takt = formatNoDigit(el.CO2takt)
			repRes.Data[i].CO2max = formatFlint2(el.CO2max)
			repRes.Data[i].CO2min = formatFlint2(el.CO2min)
			repRes.Data[i].O2max = formatFlint2(el.O2max)
			repRes.Data[i].O2min = formatFlint2(el.O2min)
		}

	}

	res, err := json.MarshalIndent(repRes, "", "  ")
	if err != nil {
		log.Println("Error marshalling ReportOrtResponse to JSON:", err)
		return []byte("{\"success\":false,\"message\":\"error marshalling ReportOrtResponse to JSON\", \"data\":[]}")
	}
	return res
}

func findReportData(lines []ReportCsvLine, endDate string) []ReportCsvLine {
	log.Printf("DEBUG findReportData: len: %d, endDate: %s\n", len(lines), endDate)

	var foundi int
	for i, l := range lines {
		if endDate == l.Datum {
			foundi = i
			break
		}
	}
	r := lines[:foundi]
	log.Printf("DEBUG findReportData: data found before endDate: %s, result length: %d\n", endDate, len(r))

	if len(r) > 0 {
		return r
	} else {
		return lines
	}

}
