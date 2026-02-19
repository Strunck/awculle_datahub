package web

import (
	"log"
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsEndpoint(gs *globals.GlobalState) http.Handler {
	log.Println("Init Metrics")
	reg := prometheus.NewRegistry()

	// Register Metrics from konfig.makeArrayNodes()
	for _, nodeid := range gs.Cfg.Plc.Publish {
		if len(nodeid.Metric) > 0 { // only the ones, that have a metric name
			reg.MustRegister(nodeid.Gauge)
		}
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}
