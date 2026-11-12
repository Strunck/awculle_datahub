package globals

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

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
				queryStr = fmt.Sprintf("query?query=%s{short='%s',ort='%s'}&time=%d", ar[0], ar[1], ort, endTime.Unix())
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
				row.Tauwasser = formatTauwasser(valStr)
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
