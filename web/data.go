package web

import (
	"fmt"
	"log"
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/go-chi/chi/v5"
)

// GET /middleware/data, see expectations/values_hmi_flat_processed.json
func PathData(gs *globals.GlobalState) http.HandlerFunc {
	log.Println("Called PathData")
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.Plc.ToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

// Get /middleware/data/{flatkey}, values from a single key, see expectations/values_hmi_flat_processed.json
func PathDataKey(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flatkey := chi.URLParam(r, "flatkey")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		flat, ok := gs.Plc.Flat[flatkey]
		if !ok {
			http.Error(w, fmt.Errorf("key %q not found", flatkey).Error(), http.StatusNotFound)
			return
		}
		if _, err := w.Write(flat.ToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataAlarms(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.Alarms.ActualToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataAlarmHistory(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.Alarms.HistoryToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataV2(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(gs.DataV2.ToJson()); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataV2Scope(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := chi.URLParam(r, "scope")

		fmt.Println("Scope:", scope)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		scopeData := gs.DataV2.GetScopeAll(scope)

		if _, err := w.Write(scopeData); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataV2Group(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := chi.URLParam(r, "scope")
		group := chi.URLParam(r, "group")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		groupData := gs.DataV2.GetGroupAll(scope, group)

		if _, err := w.Write(groupData); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}

func PathDataV2Short(gs *globals.GlobalState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scope := chi.URLParam(r, "scope")
		group := chi.URLParam(r, "group")
		short := chi.URLParam(r, "short")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		shortData := gs.DataV2.GetShort(scope, group, short)

		if _, err := w.Write(shortData); err != nil {
			http.Error(w, "Error writing response", http.StatusInternalServerError)
			return
		}
	}
}
