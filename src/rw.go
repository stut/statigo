package main

import (
	"net/http"
	"strconv"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	bytesSent       uint64
	notFoundContent []byte
	notFoundSent    bool
	metricsEnabled  bool
	site            string
}

func NewResponseWriter(w http.ResponseWriter, notFound []byte, metricsEnabled bool, site string) *responseWriter {
	return &responseWriter{
		w,
		http.StatusOK,
		0,
		notFound,
		false,
		metricsEnabled,
		site,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	if rw.metricsEnabled {
		responseStatus.WithLabelValues(rw.site, strconv.Itoa(code)).Inc()
	}
	if code == http.StatusNotFound {
		rw.ResponseWriter.Header().Set("Content-Type", "text/html; charset=utf-8")
		rw.notFoundSent = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	var err error
	var written int
	if rw.notFoundSent {
		written, err = rw.ResponseWriter.Write(rw.notFoundContent)
	} else {
		written, err = rw.ResponseWriter.Write(data)
	}
	rw.bytesSent += uint64(written)
	return written, err
}
