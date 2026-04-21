package web

import (
	"log"
	"net/http"
	"strunck/uv4-middleware/globals"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func MetricsEndpoint(gs *globals.GlobalState) http.Handler {

	log.Println("Call to MetricsEndpoint")

	gs.Metrics.Registry = prometheus.NewRegistry()

	return promhttp.HandlerFor(
		gs.Metrics.Registry,
		promhttp.HandlerOpts{
			Registry: gs.Metrics.Registry})
}
