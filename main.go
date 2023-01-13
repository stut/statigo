package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	VERSION        = 2
	responseStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statigo_status_total",
			Help: "HTTP response status.",
		},
		[]string{"site", "status"},
	)
	httpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "statigo_requests_total",
			Help: "HTTP requests total.",
		},
		[]string{"site", "path"})
	httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "statigo_response_time_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"site", "path"})
	notFoundContent = []byte("<html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p><a href=\"/\">Go to the homepage &raquo;</a></p></body></html>")
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
	healthUrl := flag.String("health-url", "/health", "Healthcheck URL")
	metricsUrl := flag.String("metrics-url", "/metrics", "Prometheus metrics URL")
	notFoundFilename := flag.String("not-found-filename", "404.html", "Page not found content filename")
	flag.Parse()

	// Read the 404 content. If reading fails the default content is used.
	content, err := os.ReadFile(path.Join(*rootDir, *notFoundFilename))
	if err == nil {
		notFoundContent = content
	}

	// Handle healthcheck requests. No metrics, no content.
	http.HandleFunc(*healthUrl, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(204)
	})

	// Set up the metrics endpoint.
	if !*noMetrics {
		http.Handle(*metricsUrl, promhttp.Handler())
	}

	// Static file server.
	fileServer := http.FileServer(http.Dir(*rootDir))
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var timer *prometheus.Timer
		var url string
		if !*noMetrics {
			url = req.URL.Path
			timer = prometheus.NewTimer(httpDuration.WithLabelValues(site, url))
		}

		// This allows us to access to response code.
		rw := NewResponseWriter(w, notFoundContent)
		fileServer.ServeHTTP(rw, req)

		if !*noMetrics {
			responseStatus.WithLabelValues(site, strconv.Itoa(rw.statusCode)).Inc()
			if rw.statusCode >= 200 && rw.statusCode <= 399 {
				httpRequests.WithLabelValues(site, url).Inc()
				timer.ObserveDuration()
			}
		}
	})

	log.Printf("Statigo v%d", VERSION)
	log.Printf("  Site: %s", site)
	log.Printf("  Web root: %s", *rootDir)
	log.Printf("  Listen addr: %s", *listenAddr)
	log.Printf("  Healthcheck: %s", *healthUrl)
	if !*noMetrics {
		log.Printf("  Prometheus metrics: %s", *metricsUrl)
	}

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
