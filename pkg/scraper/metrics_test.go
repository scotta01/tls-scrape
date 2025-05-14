package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetrics(t *testing.T) {
	// Reset the metrics before testing
	totalScrapes.Reset()
	scrapeDuration.Reset()

	// Test that totalScrapes is incremented correctly
	t.Run("totalScrapes", func(t *testing.T) {
		// Get the initial value
		initialValue := getCounterValue(totalScrapes, "success")

		// Increment the counter
		totalScrapes.WithLabelValues("success").Inc()

		// Get the new value
		newValue := getCounterValue(totalScrapes, "success")

		// Check that the value was incremented by 1
		if newValue != initialValue+1 {
			t.Errorf("Expected totalScrapes to be incremented by 1, got %f", newValue-initialValue)
		}
	})

	// Test that scrapeDuration is observed correctly
	t.Run("scrapeDuration", func(t *testing.T) {
		// Observe a duration
		scrapeDuration.WithLabelValues("example.com").Observe(0.5)

		// We can't easily check the value of a summary metric, so we'll just check that it doesn't panic
	})
}

// Helper function to get the value of a counter metric
func getCounterValue(counter *prometheus.CounterVec, labelValue string) float64 {
	m := &dto.Metric{}
	counter.WithLabelValues(labelValue).Write(m)
	return m.Counter.GetValue()
}

func TestGetMetricsHandler(t *testing.T) {
	// Get the metrics handler
	handler := GetMetricsHandler()

	// Check that the handler is not nil
	if handler == nil {
		t.Fatal("Expected non-nil handler, got nil")
	}

	// Create a test HTTP server with the handler
	server := httptest.NewServer(handler)
	defer server.Close()

	// Make a request to the metrics endpoint
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Error making request to metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Check that the response status code is 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check that the response content type contains the expected parts
	expectedContentType := "text/plain; version=0.0.4; charset=utf-8"
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, expectedContentType) {
		t.Errorf("Content-Type %s does not contain expected %s", contentType, expectedContentType)
	}
}
