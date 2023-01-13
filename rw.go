package main

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	statusCode      int
	notFoundContent []byte
	notFoundSent    bool
}

func NewResponseWriter(w http.ResponseWriter, notFound []byte) *responseWriter {
	return &responseWriter{w, http.StatusOK, notFound, false}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	if code == 404 {
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
