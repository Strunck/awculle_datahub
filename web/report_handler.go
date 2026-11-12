package web

import (
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/go-chi/chi/v5"
)

func reportSubRouter(gs *globals.GlobalState) func(r chi.Router) {
	return func(r chi.Router) {
		r.Get("/", PathReportDaily(gs))
		r.Get("/{ort}", PathReportOrt(gs))
		r.Get("/saison", PathReportSaison(gs))
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
		ort := chi.URLParam(r, "ort")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.OrtToJson(ort, startDate, endDate)); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}
