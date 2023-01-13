package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

var (
	VERSION      = 2
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statigo_requests_total",
			Help: "HTTP requests total.",
		},
		[]string{"site", "path"})
	responseStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statigo_status_total",
			Help: "HTTP response status.",
		},
		[]string{"site", "status"},
	)
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "statigo_response_time_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"site", "path"})
)

func main() {
	listenPort := os.Getenv("NOMAD_PORT_http")
	if len(listenPort) == 0 {
		listenPort = "3000"
	}
	site := os.Getenv("STATIGO_SITE")
	if len(site) == 0 {
		site = "notset"
	}

	listenAddr := flag.String("listen-addr", fmt.Sprintf(":%s", listenPort),
		"Address on which to listen for HTTP requests")
	rootDir := flag.String("root-dir", "./", "Root directory to serve files from")
	noMetrics := flag.Bool("no-metrics", false, "Disable prometheus metrics")
	metricsUrl := flag.String("metrics-url", "/metrics", "Prometheus metrics URL")
	flag.Parse()

	if !*noMetrics {
		http.Handle(*metricsUrl, promhttp.Handler())
	}
	fileServer := http.FileServer(http.Dir(*rootDir))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var timer *prometheus.Timer
		path := "notset"
		if !*noMetrics {
			path = req.URL.Path
			timer = prometheus.NewTimer(httpDuration.WithLabelValues(site, path))
		}
		rw := NewResponseWriter(w)

		fileServer.ServeHTTP(rw, req)

		if !*noMetrics {
			responseStatus.WithLabelValues(site, strconv.Itoa(rw.statusCode)).Inc()
			httpRequests.WithLabelValues(site, path).Inc()
			timer.ObserveDuration()
		}
	})

	log.Printf("Statigo v%d", VERSION)
	log.Printf("  Site: %s", site)
	log.Printf("  Web root: %s", *rootDir)
	log.Printf("  Listen addr: %s", *listenAddr)
	if !*noMetrics {
		log.Printf("  Prometheus metrics: %s", *metricsUrl)
	}

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
