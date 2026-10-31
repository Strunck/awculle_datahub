package web

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strunck/uv4-middleware/globals"
	"time"

	"github.com/go-chi/chi/v5"
)

type QueryParamForReport struct {
	Username   string `in:"path=username"`
	Visibility string `in:"query=visibility"`
	Fork       bool   `in:"query=fork"`
	Token      string `in:"header=Authorization"`
}

func reportSubRouter(gs *globals.GlobalState) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", PathReportDaily(gs))
		r.Get("/{ort}", PathReportOrt(gs))
		r.Get("/saison", PathReportSaison(gs))
	}
}

type ReportOrtResponse struct {
	Success     bool                    `json:"success"`
	Message     string                  `json:"message"`
	Raumnummer  string                  `json:"raumnummer"`
	Kundenname  string                  `json:"kundenname"`
	Reportdatum string                  `json:"Reportdatum"`
	Uvhost      string                  `json:"uvhost"`
	SpsHostname string                  `json:"sps_hostname"`
	Version     string                  `json:"version"`
	DockerName  string                  `json:"docker_name"`
	Data        []globals.ReportCsvLine `json:"data"`
}

func (r ReportOrtResponse) ToJson() []byte {
	res, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		log.Println("Error marshalling ReportOrtResponse to JSON:", err)
		return []byte("{\"success\":false,\"message\":\"error marshalling ReportOrtResponse to JSON\", \"data\":[]}")
	}
	return res
}

func PathReportOrt(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the "ort" URL parameter
		ort := chi.URLParam(r, "ort")
		ortPlus1, err := strconv.Atoi(ort)
		if err != nil {
			http.Error(w, "Invalid ort parameter", http.StatusBadRequest)
			return
		}
		log.Println("Requested Report for Ort:", ort, "-> ortplus1:", ortPlus1)

		var repRes ReportOrtResponse
		repRes.Data = make([]globals.ReportCsvLine, 0)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		dlen := gs.Report.LenAll()

		log.Println("Report Data Length:", dlen)
		if dlen == 0 {
			repRes.Success = false
			repRes.Message = "No report data available"
			repRes.Data = make([]globals.ReportCsvLine, 0)

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

		if _, err := w.Write(repRes.ToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}
func PathReportDaily(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.ToDailyReportJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

/*

Example Response for /middleware/report?query=json
{
  "data": [
    {
      "RTavg": "14.6",
      "KTavg": "13.6",
      "Kzeit": "0",
      "Ktakt": "0",
      "ZT": "",
      "TW": "0.0",
      "O2takt": "0",
      "CO2takt": "0",
      "RTsoll": "2.0",
      "RTmax": "14.6",
      "RTmin": "14.6",
      "KTmax": "13.7",
      "KTmin": "13.7",
      "O2max": "20.46",
      "O2min": "20.46",
      "CO2max": "0.09",
      "CO2min": "0.09",
      "LM": "0",
      "Frost": "0",
      "Index": 0,
      "Raum": "R1"
    },
    {
      "RTavg": "13.5",
      "KTavg": "13.0",
      "Kzeit": "0",
      "Ktakt": "0",
      "ZT": "",
      "TW": "0.0",
      "O2takt": "0",
      "CO2takt": "0",
      "RTsoll": "2.0",
      "RTmax": "13.6",
      "RTmin": "13.5",
      "KTmax": "13.1",
      "KTmin": "13.1",
      "O2max": "20.37",
      "O2min": "20.37",
      "CO2max": "0.09",
      "CO2min": "0.09",
      "LM": "0",
      "Frost": "0",
      "Index": 1,
      "Raum": "R2"
    },
  ],
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
  }
}

*/

func PathReportSaison(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		endDate := r.URL.Query().Get("end")
		startDate := r.URL.Query().Get("start")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.ToSaisonReportJson(startDate, endDate)); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}
