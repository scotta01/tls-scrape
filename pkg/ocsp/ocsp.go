// Package ocsp provides functionality for checking the Online Certificate Status Protocol (OCSP) status
// of a given certificate.
package ocsp

import (
	"bytes"
	"crypto/x509"
	"errors"
	"golang.org/x/crypto/ocsp"
	"io"
	"net/http"
)

// OCSPChecker holds the details of the certificate and its issuer.
// It provides methods to retrieve and check the OCSP response for the certificate.
type OCSPChecker struct {
	// Certificate is the certificate to be checked.
	Certificate *x509.Certificate

	// Issuer is the issuer of the certificate.
	Issuer *x509.Certificate
}

// GetOCSPResp queries the OCSP server specified in the certificate and retrieves the OCSP response.
// Returns an OCSP response or an error if the OCSP server is not specified, the request fails, or the response parsing fails.
func (o *OCSPChecker) GetOCSPResp() (*ocsp.Response, error) {
	if len(o.Certificate.OCSPServer) == 0 {
		return nil, errors.New("no OCSP server specified in cert")
	}

	ocspReq, err := ocsp.CreateRequest(o.Certificate, o.Issuer, nil)
	if err != nil {
		return nil, err
	}

	httpResp, err := http.Post(o.Certificate.OCSPServer[0], "application/ocsp-request", bytes.NewReader(ocspReq))
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	ocspResp, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	resp, err := ocsp.ParseResponse(ocspResp, o.Issuer)
	if err != nil {
		return nil, err
	}

	// If the OCSP response contains a certificate, try parsing the response again with that certificate.
	if resp.Certificate != nil {
		resp, err = ocsp.ParseResponse(ocspResp, resp.Certificate)
		if err != nil {
			return nil, err
		}
	}

	return resp, nil
}

// CheckOCSPStatus retrieves the OCSP response using GetOCSPResp and checks if the certificate status is good.
// Returns an error if the OCSP response indicates an invalid status or if fetching the OCSP response fails.
func (o *OCSPChecker) CheckOCSPStatus() error {
	ocspResp, err := o.GetOCSPResp()
	if err != nil {
		return err
	}

	if ocspResp.Status != ocsp.Good {
		return errors.New("invalid OCSP status")
	}
	return nil
}
