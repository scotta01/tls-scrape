package scraper

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net"
	"os"
	"runtime/debug"
	"testing"
	"time"
)

func TestCertDetailsString(t *testing.T) {
	cd := &CertDetails{
		Domain:     "www.jetbrains.com",
		Serial:     "12070828292658740519284007523384970881",
		NotBefore:  "2023-02-28 00:00:00 +0000 UTC",
		NotAfter:   "2024-02-09 23:59:59 +0000 UTC",
		Issuer:     "CN=Amazon RSA 2048 M02,O=Amazon,C=US",
		CRL:        []string{"http://crl.r2m02.amazontrust.com/r2m02.crl"},
		OCSPServer: []string{"http://ocsp.r2m02.amazontrust.com"},
		Valid:      false,
	}
	expected := "Domain:www.jetbrains.com Valid:false Serial:12070828292658740519284007523384970881 NotBefore:2023-02-28 00:00:00 +0000 UTC NotAfter:2024-02-09 23:59:59 +0000 UTC Issuer:CN=Amazon RSA 2048 M02,O=Amazon,C=US CRL:[http://crl.r2m02.amazontrust.com/r2m02.crl] OCSPServer:[http://ocsp.r2m02.amazontrust.com]"
	if cd.String() != expected {
		t.Errorf("expected %s \n got %s", expected, cd.String())
	}
}

type mockDialer struct {
	conn net.Conn
	err  error
}

func (m *mockDialer) Dial(network, address string) (net.Conn, error) {
	return &mockTLSConn{
		Conn:  m.conn,
		state: generateMockConnectionState(),
	}, m.err
}

type mockConn struct {
	net.Conn
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	return 0, io.EOF
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	return len(b), nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockTLSConn) Close() error {
	return nil
}

type mockTLSConn struct {
	net.Conn
	state tls.ConnectionState
}

func (m *mockTLSConn) ConnectionState() tls.ConnectionState {
	return m.state
}

func generateMockConnectionState() tls.ConnectionState {
	notBefore, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "2023-02-28 00:00:00 +0000 UTC")
	if err != nil {
		panic(err)
	}
	notAfter, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "2024-02-09 23:59:59 +0000 UTC")
	if err != nil {
		panic(err)
	}

	return tls.ConnectionState{
		PeerCertificates: []*x509.Certificate{
			{
				SerialNumber: big.NewInt(1234567890),
				NotBefore:    notBefore,
				NotAfter:     notAfter,
				Issuer: pkix.Name{
					CommonName:   "Amazon RSA 2048 M02",
					Organization: []string{"Amazon"},
					Country:      []string{"US"},
				},
				CRLDistributionPoints: []string{"http://crl.r2m02.amazontrust.com/r2m02.crl"},
				OCSPServer:            []string{"http://ocsp.r2m02.amazontrust.com"},
			},
		},
	}
}

func TestGetCertMethods(t *testing.T) {
	// Create a test certificate chain
	cert1 := &x509.Certificate{
		SerialNumber: big.NewInt(1234567890),
	}
	cert2 := &x509.Certificate{
		SerialNumber: big.NewInt(9876543210),
	}

	certChain := []*x509.Certificate{cert1, cert2}

	cd := &CertDetails{}
	cd.CertChain = certChain

	// Test GetLeafCert
	t.Run("GetLeafCert", func(t *testing.T) {
		got := cd.GetLeafCert()
		if got != cert1 {
			t.Errorf("GetLeafCert() = %v, want %v", got, cert1)
		}
	})

	// Test GetIssuerCert
	t.Run("GetIssuerCert", func(t *testing.T) {
		got := cd.GetIssuerCert()
		if got != cert2 {
			t.Errorf("GetIssuerCert() = %v, want %v", got, cert2)
		}
	})

	// Test GetCertChain
	t.Run("GetCertChain", func(t *testing.T) {
		got := cd.GetCertChain()
		if len(got) != len(certChain) {
			t.Errorf("GetCertChain() returned %d certificates, want %d", len(got), len(certChain))
		}
		for i, cert := range got {
			if cert != certChain[i] {
				t.Errorf("GetCertChain()[%d] = %v, want %v", i, cert, certChain[i])
			}
		}
	})

	// Test with empty cert chain
	cd = &CertDetails{}

	t.Run("GetLeafCert with empty chain", func(t *testing.T) {
		got := cd.GetLeafCert()
		if got != nil {
			t.Errorf("GetLeafCert() = %v, want nil", got)
		}
	})

	t.Run("GetIssuerCert with empty chain", func(t *testing.T) {
		got := cd.GetIssuerCert()
		if got != nil {
			t.Errorf("GetIssuerCert() = %v, want nil", got)
		}
	})

	t.Run("GetCertChain with empty chain", func(t *testing.T) {
		got := cd.GetCertChain()
		if got != nil {
			t.Errorf("GetCertChain() = %v, want nil", got)
		}
	})
}

func TestScrapeTLS(t *testing.T) {
	// This test is more of an integration test and might be flaky depending on network conditions
	// We'll use a small number of test domains that are unlikely to exist

	// Skip this test if SKIP_NETWORK_TESTS environment variable is set
	if os.Getenv("SKIP_NETWORK_TESTS") != "" {
		t.Skip("Skipping network-dependent test")
	}

	domains := []string{"nonexistent1.example", "nonexistent2.example"}
	concurrency := 2

	details, err := ScrapeTLS(domains, concurrency, 443)

	// We expect all domains to fail (since they don't exist), so details should be empty
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

	// The MultiError should contain errors for all domains
	if len(multiErr.Errors) != len(domains) {
		t.Errorf("Expected %d errors, got %d", len(domains), len(multiErr.Errors))
	}
}

func TestFetchFromDomainWithDialer(t *testing.T) {
	tests := []struct {
		name               string
		dialer             Dialer
		expectedErr        string
		expectedDomain     string
		expectedSerial     string
		expectedNotBefore  string
		expectedNotAfter   string
		expectedIssuer     string
		expectedCRL        string
		expectedOCSPServer string
	}{
		{
			name: "failed to dial",
			dialer: &mockDialer{
				err: errors.New("mock dial error"),
			},
			expectedErr: "mock dial error",
		},
		{
			name: "successful dial",
			dialer: &mockDialer{
				conn: &mockTLSConn{},
			},
			expectedDomain:     "example.com",
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

			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Test panicked: %v\n%s", r, debug.Stack())
				}
			}()

			cd := &CertDetails{}
			err := cd.fetchFromDomainWithDialer("example.com", tt.dialer, 443)
			if tt.expectedErr == "" && err != nil {
				t.Errorf("expected no error, got: %v", err)
			} else if tt.expectedErr != "" && (err == nil || err.Error() != tt.expectedErr) {
				t.Errorf("expected error: %v, got: %v", tt.expectedErr, err)
			}

			// Check the certificate details
			if cd.Domain != tt.expectedDomain {
				t.Errorf("expected domain %s, got %s", tt.expectedDomain, cd.Domain)
			}
			if cd.Serial != tt.expectedSerial {
				t.Errorf("expected serial %s, got %s", tt.expectedSerial, cd.Serial)
			}
			if cd.NotBefore != tt.expectedNotBefore {
				t.Errorf("expected NotBefore %s, got %s", tt.expectedNotBefore, cd.NotBefore)
			}
			if cd.NotAfter != tt.expectedNotAfter {
				t.Errorf("expected NotAfter %s, got %s", tt.expectedNotAfter, cd.NotAfter)
			}
			if cd.Issuer != tt.expectedIssuer {
				t.Errorf("expected issuer %s, got %s", tt.expectedIssuer, cd.Issuer)
			}
			if len(cd.CRL) > 0 && cd.CRL[0] != tt.expectedCRL {
				t.Errorf("expected CRL %s, got %s", tt.expectedCRL, cd.CRL[0])
			}
			if len(cd.OCSPServer) > 0 && cd.OCSPServer[0] != tt.expectedOCSPServer {
				t.Errorf("expected OCSPServer %s, got %s", tt.expectedOCSPServer, cd.OCSPServer[0])
			}
		})
	}
}
