package web

import (
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ServeHttp set up all routes of the app
// here I only want a list of my routes
// file handler.go then holds all handler functions beginning with Path
func ServeHttp(gs *globals.GlobalState) (r *chi.Mux, err error) {
	r = chi.NewRouter()

	//r.Use(middleware.Timeout(60 * time.Second)) //TODO this is making problems with SSE connections

	r.Group(func(r chi.Router) {
		// Middleware Loggs to stdout
		r.Use(middleware.Logger)
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)

		// Telegram route to send a message via POST
		r.Post("/middleware/telegram", globals.PostTelegramMsg(gs))
		r.Get("/middleware/telegram", globals.PathTelegramState(gs))

		r.Get("/middleware/", PathRoot(gs))
		r.Get("/middleware/healthcheck", PathHealthcheck(gs))
		r.Get("/middleware/setup", PathSetup(gs))
		r.Get("/middleware/clock", PathClock())
		r.Get("/middleware/hub", PathHub(gs))
		r.Get("/middleware/stats", PathStats(gs))

		// Setup Info
		r.Handle("/middleware/konfig", PathKonfig(gs))
		r.Handle("/middleware/Konfig", PathKonfig(gs))
		r.Handle("/middleware/SWSettings", PathSWSettings(gs))
		r.Handle("/middleware/swsettings", PathSWSettings(gs))

		// Helper for Debugging
		r.Route("/middleware/data", func(r chi.Router) {

			r.Get("/", PathData(gs))
			r.Get("/{flatkey}", PathDataKey(gs)) //see values_hmi_flat_processed.json
			r.Get("/alarms", PathDataAlarms(gs))
			r.Get("/v2", PathDataV2(gs))
			r.Get("/v2/{scope}", PathDataV2Scope(gs))
			r.Get("/v2/{scope}/{group}", PathDataV2Group(gs))
			r.Get("/v2/{scope}/{group}/{short}", PathDataV2Short(gs))

		})

		// Report routes
		r.Route("/middleware/Report", reportSubRouter(gs))
		r.Route("/middleware/report", reportSubRouter(gs))

		// Serve static files from the "assets" directory
		r.Get("/middleware/assets/{file}", PathAssets())

	})

	// Serve prometheus metrics
	r.Handle("/middleware/metrics", MetricsEndpoint(gs)) // No logging for metrics endpoint

	// Websockets -------------
	// Echo incoming messages, for debug
	r.Handle("/middleware/echo", http.HandlerFunc(WSPathEcho))
	// Serve AlarmsKonfiguration, every 2 seconds
	r.Handle("/middleware/ALARMS", http.HandlerFunc(WSPathAlarms(gs)))
	// Serve HMI values, every 2 seconds
	r.Handle("/middleware/DATA", http.HandlerFunc(WSPathDATA(gs)))

	r.Handle("/middleware/DATAv2", http.HandlerFunc(WSPathDATAv2(gs)))

	// Serve HMI values, every 2 seconds, Servers SWSettings as Websocket
	r.Handle("/middleware/STRINGS", http.HandlerFunc(WSPathSTRINGS(gs)))
	// Hilfstring für Grafana Dashboards
	r.Handle("/middleware/RAUM", http.HandlerFunc(WSPathRaum(gs)))

	// Listen to write requests via Websocket to OPCUA
	r.Handle("/middleware/UV4", http.HandlerFunc(WSListenWrite(gs)))

	// ------------------------

	// ----------

	// Data-Star Server-Sent Events
	r.Route("/middleware/ds", func(r chi.Router) {
		if gs.Stats.Developer {
			r.Use(middleware.Logger)
			r.Use(middleware.RequestID)
			r.Use(middleware.RealIP)
		}
		//r.Use(SymbolCtx(gs))
		r.Get("/{symbol}", DataStarHandler(gs))
		r.Get("/latestmsg", DataStarLatestMsgHandler(gs))

	})

	return r, nil
}
