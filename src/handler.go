package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

type CustomHandler struct {
	site              string
	metricsEnabled    bool
	requestLogEnabled bool
	nextHandler       http.Handler
}

func CreateCustomHandler(site string, metricsEnabled bool, requestLogEnabled bool, nextHandler http.Handler) CustomHandler {
	return CustomHandler{
		site:              site,
		metricsEnabled:    metricsEnabled,
		requestLogEnabled: requestLogEnabled,
		nextHandler:       nextHandler,
	}
}

func (handler CustomHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var timer *prometheus.Timer = nil
	url := ""
	if handler.metricsEnabled {
		url = req.URL.Path
		timer = prometheus.NewTimer(httpDuration.WithLabelValues(handler.site, url))
	}

	// This allows us to mark a metric for the response code.
	rw := NewResponseWriter(w, notFoundContent, handler.metricsEnabled, handler.site)

	handler.nextHandler.ServeHTTP(rw, req)

	handler.logRequest(req, rw)

	if handler.metricsEnabled {
		if rw.statusCode >= 200 && rw.statusCode <= 399 {
			httpRequests.WithLabelValues(handler.site, url).Inc()
			timer.ObserveDuration()
		}
	}
}
