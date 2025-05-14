package scraper

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"net"
	"os"
	"strings"
	"testing"
)

// Mock dialer for testing fetchFromIPWithDialer
type mockIPDialer struct {
	conn net.Conn
	err  error
}

func (m *mockIPDialer) Dial(network, address string) (net.Conn, error) {
	return &mockTLSConn{
		Conn:  m.conn,
		state: generateMockConnectionState(),
	}, m.err
}

func TestFetchFromIPWithDialer(t *testing.T) {
	tests := []struct {
		name               string
		dialer             Dialer
		expectedErr        string
		expectedIP         string
		expectedSerial     string
		expectedNotBefore  string
		expectedNotAfter   string
		expectedIssuer     string
		expectedCRL        string
		expectedOCSPServer string
	}{
		{
			name: "failed to dial",
			dialer: &mockIPDialer{
				err: errors.New("mock dial error"),
			},
			expectedErr: "mock dial error",
		},
		{
			name: "successful dial",
			dialer: &mockIPDialer{
				conn: &mockConn{},
			},
			expectedIP:         "192.0.2.1",
			expectedSerial:     "1234567890",
			expectedNotBefore:  "2023-02-28 00:00:00 +0000 UTC",
			expectedNotAfter:   "2024-02-09 23:59:59 +0000 UTC",
			expectedIssuer:     "CN=Amazon RSA 2048 M02,O=Amazon,C=US",
			expectedCRL:        "http://crl.r2m02.amazontrust.com/r2m02.crl",
			expectedOCSPServer: "http://ocsp.r2m02.amazontrust.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &IPCertDetails{}
			ip := net.ParseIP("192.0.2.1") // TEST-NET-1 (RFC 5737)
			err := cd.fetchFromIPWithDialer(ip, 443, tt.dialer)

			if tt.expectedErr == "" && err != nil {
				t.Errorf("expected no error, got: %v", err)
			} else if tt.expectedErr != "" && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			// Only check details if no error is expected
			if tt.expectedErr == "" {
				if cd.IP != tt.expectedIP {
					t.Errorf("expected IP %s, got %s", tt.expectedIP, cd.IP)
				}
				if cd.CertDetails.Serial != tt.expectedSerial {
					t.Errorf("expected serial %s, got %s", tt.expectedSerial, cd.CertDetails.Serial)
				}
				if cd.CertDetails.NotBefore != tt.expectedNotBefore {
					t.Errorf("expected NotBefore %s, got %s", tt.expectedNotBefore, cd.CertDetails.NotBefore)
				}
				if cd.CertDetails.NotAfter != tt.expectedNotAfter {
					t.Errorf("expected NotAfter %s, got %s", tt.expectedNotAfter, cd.CertDetails.NotAfter)
				}
				if cd.CertDetails.Issuer != tt.expectedIssuer {
					t.Errorf("expected issuer %s, got %s", tt.expectedIssuer, cd.CertDetails.Issuer)
				}
				if len(cd.CertDetails.CRL) > 0 && cd.CertDetails.CRL[0] != tt.expectedCRL {
					t.Errorf("expected CRL %s, got %s", tt.expectedCRL, cd.CertDetails.CRL[0])
				}
				if len(cd.CertDetails.OCSPServer) > 0 && cd.CertDetails.OCSPServer[0] != tt.expectedOCSPServer {
					t.Errorf("expected OCSPServer %s, got %s", tt.expectedOCSPServer, cd.CertDetails.OCSPServer[0])
				}
			}
		})
	}
}

func TestReverseDNSLookup(t *testing.T) {
	// This test uses real DNS lookups, so it might be flaky depending on network conditions
	// We'll test with well-known IP addresses that should have stable DNS entries

	// Skip this test if SKIP_NETWORK_TESTS environment variable is set
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network-dependent test")
	}

	tests := []struct {
		name    string
		ip      string
		wantErr bool
	}{
		{
			name:    "localhost",
			ip:      "127.0.0.1",
			wantErr: false,
		},
		{
			name:    "invalid IP",
			ip:      "999.999.999.999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil && !tt.wantErr {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}

			hostname, err := reverseDNSLookup(ip)

			if (err != nil) != tt.wantErr {
				t.Errorf("reverseDNSLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && hostname == "" {
				t.Errorf("reverseDNSLookup() returned empty hostname for %s", tt.ip)
			}
		})
	}
}

func TestScrapeIPTLS(t *testing.T) {
	// This test is more of an integration test and might be flaky depending on network conditions
	// We'll use a small number of test IPs that are unlikely to have TLS servers

	// Skip this test if SKIP_NETWORK_TESTS environment variable is set
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network-dependent test")
	}

	ips := []net.IP{
		net.ParseIP("192.0.2.1"), // TEST-NET-1 (RFC 5737)
		net.ParseIP("192.0.2.2"),
	}
	port := 12345 // Unlikely to have a TLS server
	concurrency := 2

	details, err := ScrapeIPTLS(ips, port, concurrency)

	// We expect all IPs to fail (since they're TEST-NET IPs), so details should be empty
	if len(details) != 0 {
		t.Errorf("Expected 0 details, got %d", len(details))
	}

	// We should get an error
	if err == nil {
		t.Errorf("Expected error, got nil")
	}

	// The error should be a MultiError
	multiErr, ok := err.(*MultiError)
	if !ok {
		t.Errorf("Expected MultiError, got %T", err)
	}

	// The MultiError should contain errors for all IPs
	if len(multiErr.Errors) != len(ips) {
		t.Errorf("Expected %d errors, got %d", len(ips), len(multiErr.Errors))
	}
}

func TestIPCertDetailsString(t *testing.T) {
	cd := &IPCertDetails{
		CertDetails: &CertDetails{
			Domain:     "192.168.1.1",
			Serial:     "12345",
			NotBefore:  "2023-02-28 00:00:00 +0000 UTC",
			NotAfter:   "2024-02-09 23:59:59 +0000 UTC",
			Issuer:     "CN=Test CA,O=Test,C=US",
			CRL:        []string{"http://crl.example.com/test.crl"},
			OCSPServer: []string{"http://ocsp.example.com"},
			Valid:      true,
		},
		IP:             "192.168.1.1",
		Hostname:       "test.example.com",
		HostnameInCert: true,
		SANs:           []string{"test.example.com", "www.example.com"},
	}

	// Test with hostname
	result := cd.String()
	expectedSubstrings := []string{
		"IP:192.168.1.1",
		"Hostname:test.example.com",
		"HostnameInCert:true",
		"Valid:true",
		"Serial:12345",
		"NotBefore:2023-02-28 00:00:00 +0000 UTC",
		"NotAfter:2024-02-09 23:59:59 +0000 UTC",
		"Issuer:CN=Test CA,O=Test,C=US",
		"CRL:[http://crl.example.com/test.crl]",
		"OCSPServer:[http://ocsp.example.com]",
		"SANs:test.example.com,www.example.com",
	}

	for _, substr := range expectedSubstrings {
		if !contains(result, substr) {
			t.Errorf("Expected String() to contain %q, but it didn't. Got: %s", substr, result)
		}
	}

	// Test without hostname
	cd.Hostname = ""
	cd.HostnameInCert = false
	result = cd.String()

	// Should not contain hostname info
	if contains(result, "Hostname:") {
		t.Errorf("Expected String() not to contain hostname info when Hostname is empty, but it did. Got: %s", result)
	}
}

func TestIsHostnameInCert(t *testing.T) {
	// Create a test certificate
	cert := &x509.Certificate{
		Subject: pkix.Name{
			CommonName: "example.com",
		},
		DNSNames: []string{"www.example.com", "api.example.com"},
	}

	tests := []struct {
		name     string
		hostname string
		want     bool
	}{
		{
			name:     "hostname matches common name",
			hostname: "example.com",
			want:     true,
		},
		{
			name:     "hostname matches SAN",
			hostname: "www.example.com",
			want:     true,
		},
		{
			name:     "hostname does not match",
			hostname: "other.example.com",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isHostnameInCert(cert, tt.hostname)
			if got != tt.want {
				t.Errorf("isHostnameInCert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractSANs(t *testing.T) {
	// Create a test certificate
	cert := &x509.Certificate{
		DNSNames: []string{"example.com", "www.example.com", "api.example.com"},
	}

	got := extractSANs(cert)
	want := []string{"example.com", "www.example.com", "api.example.com"}

	if !stringSlicesEqual(got, want) {
		t.Errorf("extractSANs() = %v, want %v", got, want)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return s != "" && strings.Contains(s, substr)
}

// Helper function to check if two string slices are equal
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}
