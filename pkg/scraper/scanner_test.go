package scraper

import (
	"net"
	"os"
	"testing"
)

// TestScanDomains_Chunking tests that ScanDomains correctly chunks the input domains
func TestScanDomains_Chunking(t *testing.T) {
	// This test verifies that ScanDomains correctly chunks the input domains
	// We can't easily mock ScrapeTLS, so we'll just test with a small number of domains
	// that are unlikely to exist, which will result in errors but still test the chunking logic

	// Skip this test if SKIP_NETWORK_TESTS environment variable is set
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network-dependent test")
	}

	domains := []string{"nonexistent1.example", "nonexistent2.example", "nonexistent3.example"}
	concurrency := 2
	chunkSize := 2

	// Call ScanDomains
	details, errors := ScanDomains(domains, concurrency, chunkSize, 443)

	// We expect all domains to fail (since they don't exist), so details should be empty
	// and errors should contain all domains
	if len(details) != 0 {
		t.Errorf("Expected 0 details, got %d", len(details))
	}

	// Check that we got errors for all domains
	if len(errors) != len(domains) {
		t.Errorf("Expected %d errors, got %d", len(domains), len(errors))
	}

	// Check that each domain has an error
	for _, domain := range domains {
		if _, ok := errors[domain]; !ok {
			t.Errorf("Expected error for domain %s, but none was found", domain)
		}
	}
}

// TestScanIPAddresses_Chunking tests that ScanIPAddresses correctly chunks the input IPs
func TestScanIPAddresses_Chunking(t *testing.T) {
	// This test verifies that ScanIPAddresses correctly chunks the input IPs
	// We can't easily mock ScrapeIPTLS, so we'll just test with a small number of IPs
	// that are unlikely to respond to TLS, which will result in errors but still test the chunking logic

	// Skip this test if SKIP_NETWORK_TESTS environment variable is set
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network-dependent test")
	}

	ips := []net.IP{
		net.ParseIP("192.0.2.1"), // TEST-NET-1 (RFC 5737)
		net.ParseIP("192.0.2.2"),
		net.ParseIP("192.0.2.3"),
	}
	port := 12345 // Unlikely to have a TLS server
	concurrency := 2
	chunkSize := 2

	// Call ScanIPAddresses
	details, errors := ScanIPAddresses(ips, port, concurrency, chunkSize)

	// We expect all IPs to fail (since they're TEST-NET IPs), so details should be empty
	// and errors should contain all IPs
	if len(details) != 0 {
		t.Errorf("Expected 0 details, got %d", len(details))
	}

	// Check that we got errors for all IPs
	if len(errors) != len(ips) {
		t.Errorf("Expected %d errors, got %d", len(ips), len(errors))
	}

	// Check that each IP has an error
	for _, ip := range ips {
		if _, ok := errors[ip.String()]; !ok {
			t.Errorf("Expected error for IP %s, but none was found", ip.String())
		}
	}
}

// TestScanDomains_EmptyInput tests that ScanDomains handles empty input correctly
func TestScanDomains_EmptyInput(t *testing.T) {
	domains := []string{}
	concurrency := 2
	chunkSize := 2

	details, errors := ScanDomains(domains, concurrency, chunkSize, 443)

	if len(details) != 0 {
		t.Errorf("Expected 0 details for empty input, got %d", len(details))
	}

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors for empty input, got %d", len(errors))
	}
}

// TestScanIPAddresses_EmptyInput tests that ScanIPAddresses handles empty input correctly
func TestScanIPAddresses_EmptyInput(t *testing.T) {
	ips := []net.IP{}
	port := 443
	concurrency := 2
	chunkSize := 2

	details, errors := ScanIPAddresses(ips, port, concurrency, chunkSize)

	if len(details) != 0 {
		t.Errorf("Expected 0 details for empty input, got %d", len(details))
	}

	if len(errors) != 0 {
		t.Errorf("Expected 0 errors for empty input, got %d", len(errors))
	}
}
