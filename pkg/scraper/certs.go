// Package scraper provides functionality for scraping TLS certs, with a focus on
// simple return values, but the underlying x509 is still accessible, if required.
package scraper

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"sync"
	"time"
)

// CertDetails encapsulates various details about a certificate obtained
// from a scraped domain.
type CertDetails struct {
	Domain     string              `json:"domain"`
	Serial     string              `json:"serial"`
	NotBefore  string              `json:"not_before"`
	NotAfter   string              `json:"not_after"`
	Issuer     string              `json:"issuer"`
	CRL        []string            `json:"crl"`
	OCSPServer []string            `json:"ocsp_server"`
	CertChain  []*x509.Certificate `json:"cert_chain"`
	// Certificate validation information
	Valid          bool     `json:"valid"`
	ValidationErrs []string `json:"validation_errors,omitempty"`
}

// Dialer is an interface for types that can dial and establish network
// connections.
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

// GetLeafCert returns the leaf (or main) certificate from the scraped details.
// Returns nil if the certificate chain is empty.
func (cd *CertDetails) GetLeafCert() *x509.Certificate {
	if cd.CertChain == nil || len(cd.CertChain) == 0 {
		return nil
	}
	return cd.CertChain[0]
}

// GetIssuerCert returns the issuer's certificate from the scraped details.
// Returns nil if the certificate chain doesn't include an issuer certificate.
func (cd *CertDetails) GetIssuerCert() *x509.Certificate {
	if cd.CertChain == nil || len(cd.CertChain) < 2 {
		return nil
	}
	return cd.CertChain[1]
}

// GetCertChain returns the entire chain of certificates from the scraped details.
// Returns nil if the certificate chain is empty.
func (cd *CertDetails) GetCertChain() []*x509.Certificate {
	return cd.CertChain
}

// fetchFromDomain retrieves the certificate details from the provided domain.
func (cd *CertDetails) fetchFromDomain(domain string) error {
	// Create a TLS configuration that skips certificate verification
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	return cd.fetchFromDomainWithDialer(domain, &tls.Dialer{
		Config: tlsConfig,
	})
}

// fetchFromDomainWithDialer retrieves the certificate details from
// the provided domain using a custom dialer.
func (cd *CertDetails) fetchFromDomainWithDialer(domain string, dialer Dialer) error {
	// Use the provided dialer to establish a connection
	conn, err := dialer.Dial("tcp", domain+":443")
	if err != nil {
		return err
	}
	defer conn.Close()

	// ConnectionStateGetter is an interface for types that can provide
	// information about a TLS connection's state.
	type ConnectionStateGetter interface {
		ConnectionState() tls.ConnectionState
	}
	tlsGetter, ok := conn.(ConnectionStateGetter)
	if !ok {
		return fmt.Errorf("expected a ConnectionStateGetter, got %T", conn)
	}

	certs := tlsGetter.ConnectionState().PeerCertificates
	cd.CertChain = certs
	if len(certs) == 0 {
		return fmt.Errorf("no certificates found for domain %s", domain)
	}

	cert := certs[0]

	cd.Domain = domain
	cd.Serial = cert.SerialNumber.String()
	cd.NotBefore = cert.NotBefore.String()
	cd.NotAfter = cert.NotAfter.String()
	cd.Issuer = cert.Issuer.String()
	cd.CRL = cert.CRLDistributionPoints
	cd.OCSPServer = cert.OCSPServer

	// Manually validate the certificate
	cd.Valid = true // Assume valid until proven otherwise
	cd.ValidationErrs = []string{}

	// Create a certificate pool with the system root certificates
	roots, err := x509.SystemCertPool()
	if err != nil {
		// If we can't get system roots, create an empty pool
		roots = x509.NewCertPool()
	}

	// Add intermediate certificates to the pool
	intermediates := x509.NewCertPool()
	for i, cert := range certs {
		if i > 0 { // Skip the leaf certificate
			intermediates.AddCert(cert)
		}
	}

	// Verify the certificate chain
	opts := x509.VerifyOptions{
		DNSName:       domain,
		Intermediates: intermediates,
		Roots:         roots,
	}

	_, err = cert.Verify(opts)
	if err != nil {
		cd.Valid = false

		// Parse the error to get detailed validation information
		switch e := err.(type) {
		case x509.CertificateInvalidError:
			reason := "Certificate is invalid"
			switch e.Reason {
			case x509.Expired:
				reason = "Certificate has expired or is not yet valid"
			case x509.NotAuthorizedToSign:
				reason = "Certificate is not authorized to sign other certificates"
			case x509.IncompatibleUsage:
				reason = "Certificate usage is incompatible with the intended usage"
			case x509.CANotAuthorizedForThisName:
				reason = "CA is not authorized for this name"
			case x509.TooManyIntermediates:
				reason = "Too many intermediate certificates"
			default:
				reason = fmt.Sprintf("Certificate is invalid (reason code: %d)", e.Reason)
			}
			cd.ValidationErrs = append(cd.ValidationErrs, reason)
		case x509.HostnameError:
			cd.ValidationErrs = append(cd.ValidationErrs, "Certificate is not valid for domain: "+domain)
		case x509.UnknownAuthorityError:
			cd.ValidationErrs = append(cd.ValidationErrs, "Certificate signed by unknown authority (possibly self-signed)")
		default:
			cd.ValidationErrs = append(cd.ValidationErrs, "Certificate validation error: "+err.Error())
		}
	}

	// Check if the certificate is expired or not yet valid
	now := time.Now()
	if now.Before(cert.NotBefore) {
		cd.Valid = false
		cd.ValidationErrs = append(cd.ValidationErrs, "Certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		cd.Valid = false
		cd.ValidationErrs = append(cd.ValidationErrs, "Certificate has expired")
	}

	return nil
}

// ScrapeTLS scrapes the given websites for TLS certificate details
// concurrently and returns the collected information.
func ScrapeTLS(websites []string, concurrency int) ([]*CertDetails, error) {
	results := make(chan *CertDetails, len(websites))
	errorChan := make(chan map[string]error, len(websites))

	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup

	// For each website, fetch certificate details in a goroutine.
	for _, website := range websites {
		wg.Add(1)
		go func(site string) {
			defer wg.Done()

			sem <- struct{}{} // Acquire a concurrency token

			timer := prometheus.NewTimer(scrapeDuration.WithLabelValues(site))
			defer timer.ObserveDuration()

			certInfo := &CertDetails{}
			err := certInfo.fetchFromDomain(site)

			<-sem // Release a concurrency token

			if err != nil {
				errorChan <- map[string]error{site: err}
				totalScrapes.WithLabelValues("failed").Inc()
				return
			}
			totalScrapes.WithLabelValues("success").Inc()
			results <- certInfo
		}(website)
	}

	// Close result channels when all scraping goroutines are done.
	go func() {
		wg.Wait()
		close(results)
		close(errorChan)
	}()

	var details []*CertDetails

	multiError := &MultiError{Errors: make(map[string]error)}

	for res := range results {
		details = append(details, res)
	}

	for err := range errorChan {
		for domain, e := range err {
			multiError.Errors[domain] = e
		}
	}

	if len(multiError.Errors) > 0 {
		return details, multiError
	}

	return details, nil
}

// String provides a string representation of the certificate details.
func (c *CertDetails) String() string {
	validationInfo := ""
	if !c.Valid && len(c.ValidationErrs) > 0 {
		validationInfo = fmt.Sprintf("Valid:%t ValidationErrors:%v ", c.Valid, c.ValidationErrs)
	} else {
		validationInfo = fmt.Sprintf("Valid:%t ", c.Valid)
	}

	return fmt.Sprintf(
		"Domain:%s "+
			"%s"+
			"Serial:%s "+
			"NotBefore:%s "+
			"NotAfter:%s "+
			"Issuer:%s "+
			"CRL:%s "+
			"OCSPServer:%s",
		c.Domain,
		validationInfo,
		c.Serial,
		c.NotBefore,
		c.NotAfter,
		c.Issuer,
		c.CRL,
		c.OCSPServer,
	)
}
