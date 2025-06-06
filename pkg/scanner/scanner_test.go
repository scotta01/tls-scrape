package scanner

import (
	"net"
	"testing"
)

// TestScanDomainsInternal_Exists verifies that the ScanDomainsInternal function exists and can be called
func TestScanDomainsInternal_Exists(t *testing.T) {
	// This test just verifies that the function exists and can be called
	// It doesn't test the actual functionality because that would require mocking the scraper package
	domains := []string{"example.com"}
	_, _ = ScanDomainsInternal(domains, 1, 1, 443)
}

// TestScanIPAddressesInternal_Exists verifies that the ScanIPAddressesInternal function exists and can be called
func TestScanIPAddressesInternal_Exists(t *testing.T) {
	// This test just verifies that the function exists and can be called
	// It doesn't test the actual functionality because that would require mocking the scraper package
	ips := []net.IP{net.ParseIP("192.168.1.1")}
	_, _ = ScanIPAddressesInternal(ips, 443, 1, 1)
}
