package ocsp

import (
	"crypto/x509"
	"testing"
)

// TestGetOCSPResp_NoServer tests the GetOCSPResp method when no OCSP server is specified
func TestGetOCSPResp_NoServer(t *testing.T) {
	// Create a certificate with no OCSP server
	cert := &x509.Certificate{
		OCSPServer: []string{},
	}
	issuer := &x509.Certificate{}

	// Create an OCSPChecker
	checker := &OCSPChecker{
		Certificate: cert,
		Issuer:      issuer,
	}

	// Call the method
	_, err := checker.GetOCSPResp()
	if err == nil {
		t.Errorf("Expected error for no OCSP server, got nil")
		return
	}

	errMsg := err.Error()
	if errMsg != "no OCSP server specified in cert" {
		t.Errorf("Expected error message 'no OCSP server specified in cert', got '%s'", errMsg)
	}
}

// TestCheckOCSPStatus_Error tests the CheckOCSPStatus method when GetOCSPResp returns an error
func TestCheckOCSPStatus_Error(t *testing.T) {
	// Create a certificate with no OCSP server to force an error
	cert := &x509.Certificate{
		OCSPServer: []string{},
	}
	issuer := &x509.Certificate{}

	// Create an OCSPChecker
	checker := &OCSPChecker{
		Certificate: cert,
		Issuer:      issuer,
	}

	// Call the method
	err := checker.CheckOCSPStatus()
	if err == nil {
		t.Errorf("Expected error from CheckOCSPStatus, got nil")
		return
	}

	errMsg := err.Error()
	if errMsg != "no OCSP server specified in cert" {
		t.Errorf("Expected error message 'no OCSP server specified in cert', got '%s'", errMsg)
	}
}
