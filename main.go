package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	listenPort := os.Getenv("NOMAD_PORT_http")
	if len(listenPort) == 0 {
		listenPort = "3000"
	}

	listenAddr := flag.String("listen-addr", fmt.Sprintf(":%s", listenPort),
		"Address on which to listen for HTTP requests")
	rootDir := flag.String("root-dir", "./", "Root directory to serve files from")
	flag.Parse()

	log.Printf("Serving from %s on %s", *rootDir, *listenAddr)
	http.Handle("/", http.FileServer(http.Dir(*rootDir)))
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
