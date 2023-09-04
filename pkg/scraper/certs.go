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
}

// Dialer is an interface for types that can dial and establish network
// connections.
type Dialer interface {
	Dial(network, address string) (net.Conn, error)
}

// GetLeafCert returns the leaf (or main) certificate from the scraped details.
func (cd *CertDetails) GetLeafCert() *x509.Certificate {
	return cd.CertChain[0]
}

// GetIssuerCert returns the issuer's certificate from the scraped details.
func (cd *CertDetails) GetIssuerCert() *x509.Certificate {
	return cd.CertChain[1]
}

// GetCertChain returns the entire chain of certificates from the scraped details.
func (cd *CertDetails) GetCertChain() []*x509.Certificate {
	return cd.CertChain
}

// fetchFromDomain retrieves the certificate details from the provided domain.
func (cd *CertDetails) fetchFromDomain(domain string) error {
	return cd.fetchFromDomainWithDialer(domain, &tls.Dialer{})
}

// fetchFromDomainWithDialer retrieves the certificate details from
// the provided domain using a custom dialer.
func (cd *CertDetails) fetchFromDomainWithDialer(domain string, dialer Dialer) error {
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
	return fmt.Sprintf(
		"Domain:%s "+
			"Serial:%s "+
			"NotBefore:%s "+
			"NotAfter:%s "+
			"Issuer:%s "+
			"CRL:%s "+
			"OCSPServer:%s",
		c.Domain,
		c.Serial,
		c.NotBefore,
		c.NotAfter,
		c.Issuer,
		c.CRL,
		c.OCSPServer,
	)
}
