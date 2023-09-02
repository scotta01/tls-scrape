package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
)

// totalScrapes is a counter metric to track the number of domains scraped.
// The metric includes labels to differentiate between successful and failed scrapes.
var (
	totalScrapes = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tls_scrapes_total",
			Help: "Total number of domains scraped.",
		},
		[]string{"status"}, // "status" can be "success" or "failed"
	)

	// scrapeDuration is a summary metric to capture the duration taken to scrape TLS information from domains.
	// It provides latency quantiles for each domain.
	scrapeDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "tls_scrape_duration_seconds",
			Help: "Duration of the TLS scraping process in seconds.",
		},
		[]string{"domain"}, // The domain for which the scrape duration is being measured
	)
)

// init function registers the Prometheus metrics during package initialization.
func init() {
	prometheus.MustRegister(totalScrapes)
	prometheus.MustRegister(scrapeDuration)
}

// GetMetricsHandler returns a HTTP handler for the Prometheus metrics.
// This can be attached to an HTTP server to expose the metrics endpoint.
func GetMetricsHandler() http.Handler {
	return promhttp.Handler()
}
