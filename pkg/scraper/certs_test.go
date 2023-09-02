package scraper

import (
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"io"
	"math/big"
	"net"
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
	}
	expected := "Domain:www.jetbrains.com Serial:12070828292658740519284007523384970881 NotBefore:2023-02-28 00:00:00 +0000 UTC NotAfter:2024-02-09 23:59:59 +0000 UTC Issuer:CN=Amazon RSA 2048 M02,O=Amazon,C=US CRL:[http://crl.r2m02.amazontrust.com/r2m02.crl] OCSPServer:[http://ocsp.r2m02.amazontrust.com]"
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
			err := cd.fetchFromDomainWithDialer("example.com", tt.dialer)
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
