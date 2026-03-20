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

	if gs.MetricRegistry == nil {
		gs.MetricRegistry = prometheus.NewRegistry()

		// Register Metrics from konfig.makeArrayNodes()
		for _, nodeid := range gs.Cfg.Plc.Publish {
			if len(nodeid.Metric) > 0 { // only the ones, that have a metric name
				log.Printf("Registering metric for topic: %s", nodeid.Topic)

				gs.MetricRegistry.MustRegister(nodeid.Gauge)
			}
		}
	}
	return promhttp.HandlerFor(
		gs.MetricRegistry,
		promhttp.HandlerOpts{
			Registry: gs.MetricRegistry})
}
