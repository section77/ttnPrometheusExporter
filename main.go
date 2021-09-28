package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"ttnPrometheusExporter/exporter"
)

var (
	listenAddress = flag.String("web.listen-address", ":9101", "Address to listen on for web interface.")
	metricPath    = flag.String("web.metrics-path", "/metrics", "Path under which to expose metrics.")
)

func main() {
	log.Print("Starting ttn_exporter")
	log.Fatal(serverMetrics(*listenAddress, *metricPath))
}
func serverMetrics(listenAddress, metricsPath string) error {
	http.Handle(metricsPath, promhttp.Handler())
	apiToken := os.Getenv("TTN_TOKEN")
	gwName := os.Getenv("TTN_GATEWAY_NAME")

	exporter.Register(apiToken, gwName)
	return http.ListenAndServe(listenAddress, nil)
}
