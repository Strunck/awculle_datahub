package web

import (
	"log"
	"net/http"
	"strunck/awculle_datahub/initialize"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsEndpoint(cfg *initialize.Config) http.Handler {
	log.Println("Init Metrics")
	reg := prometheus.NewRegistry()

	// Register Metrics from konfig.makeArrayNodes()
	for _, nodeid := range cfg.Plc.Publish {
		if len(nodeid.Metric) > 0 { // only the ones, that have a metric name
			reg.MustRegister(nodeid.Gauge)
		}
	}
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg})
}
