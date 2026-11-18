package web

import (
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/go-chi/chi/v5"
)

func reportSubRouter(gs *globals.GlobalState) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", PathReportDaily(gs))
		r.Get("/ort", PathReportOrt(gs))
		r.Get("/saison", PathReportSaison(gs))
	}
}

func PathReportDaily(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startDate := r.URL.Query().Get("start")

		jsonb := gs.ToDailyReportJson(startDate)

		if jsonb == nil {
			http.Error(w, "Error gathering data report", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(jsonb); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

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

func PathReportOrt(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the "ort" URL parameter
		ort := r.URL.Query().Get("ort")
		query := r.URL.Query().Get("query")
		startDate := r.URL.Query().Get("start")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		blob := gs.OrtToJson(ort, query, startDate)
		if _, err := w.Write(blob); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}
