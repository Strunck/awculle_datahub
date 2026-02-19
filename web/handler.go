package web

import (
	"encoding/json"
	"net/http"
	"strunck/uv4-middleware/globals"
	"strunck/uv4-middleware/view"
	"time"
)

// Path /clock
func PathClock() http.HandlerFunc {
	// Get Time Now as String
	t := time.Now().Local().Format(time.RFC3339)

	return func(w http.ResponseWriter, r *http.Request) {
		view.ServerTime(t).Render(r.Context(), w)
	}
}

// Path /
func PathRoot(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view.Layout(gs).Render(r.Context(), w)
	}
}

// Path /stats
func PathStats(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(gs.Stats.ToJson())
	}
}

// Path /setup
func PathSetup(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(gs.Cfg.ToJson())
	}
}

func PathSWSettings(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		j, err := json.Marshal(gs.Cfg.Plc.SWSettings)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(j)
	}
}

func PathKonfig(gs *globals.GlobalState) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		j, err := json.Marshal(gs.Cfg.Plc.Konfig)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(j)
	}
}

func PathHub(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		view.ShowHub(gs).Render(r.Context(), w)
	}
}

// Path /healthcheck
func PathHealthcheck(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(gs.Stats.OpcHealth.ToJson())
	}
}
