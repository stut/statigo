package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
)

var (
	VERSION = 6

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

func checkDirExists(rootDir string) error {
	// Check the root dir exists and is accessible.
	f, err := os.Open(rootDir)
	if err != nil {
		return fmt.Errorf("does not exist: %s", err)
	}
	defer func() { _ = f.Close() }()
	var s fs.FileInfo
	s, err = f.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %s", err)
	}
	if !s.IsDir() {
		return fmt.Errorf("is not a directory")
	}
	return nil
}

func main() {
	var err error

	listenPort := os.Getenv("NOMAD_PORT_http")
	if len(listenPort) == 0 {
		listenPort = "3000"
	}
	site := os.Getenv("STATIGO_SITE")
	if len(site) == 0 {
		site = "notset"
	}
	listenAddr := flag.String("listen-addr", fmt.Sprintf(":%s", listenPort), "Address on which to listen for HTTP requests")
	rootDir := flag.String("root-dir", "./", "Root directory to serve files from")
	indexFilename := flag.String("index-filename", "index.html", "Directory index filename")
	noMetrics := flag.Bool("no-metrics", false, "Disable prometheus metrics")
	healthUrl := flag.String("health-url", "/health", "Healthcheck URL")
	metricsUrl := flag.String("metrics-url", "/metrics", "Prometheus metrics URL")
	notFoundFilename := flag.String("not-found-filename", "404.html", "Page not found content filename")
	disableApacheLogging := flag.Bool("no-request-logging", false, "Disable Apache request logging to stdout")

	flag.Parse()

	err = checkDirExists(*rootDir)
	if err != nil {
		log.Fatalf("Root directory %s", err)
	}

	// Read the 404 content. If reading fails the default content is used.
	var content []byte
	content, err = os.ReadFile(path.Join(*rootDir, *notFoundFilename))
	if err == nil {
		notFoundContent = content
	}

	// Handle healthcheck requests. No metrics, no content.
	http.HandleFunc(*healthUrl, func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Set up the metrics endpoint.
	if !*noMetrics {
		http.Handle(*metricsUrl, promhttp.Handler())
	}

	// Static file server.
	var handler http.Handler
	handler = CreateCustomHandler(site, !(*noMetrics),
		http.FileServer(CreateFileSystemNoDirList(http.Dir(*rootDir), *indexFilename)))
	if !(*disableApacheLogging) {
		handler = NewApacheLoggingHandler(handler, os.Stdout)
	}
	http.Handle("/", handler)

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
