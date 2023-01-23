package main

import (
	"net/http"
	"strconv"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	notFoundContent []byte
	notFoundSent    bool
	metricsEnabled  bool
	site            string
}

func NewResponseWriter(w http.ResponseWriter, notFound []byte, metricsEnabled bool, site string) *responseWriter {
	return &responseWriter{
		w,
		http.StatusOK,
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
	if rw.notFoundSent {
		return rw.ResponseWriter.Write(rw.notFoundContent)
	}
	return rw.ResponseWriter.Write(data)
}
