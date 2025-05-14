package scraper

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

// IPCertDetails extends CertDetails with IP-specific information
type IPCertDetails struct {
	*CertDetails
	IP             string   `json:"ip"`
	Hostname       string   `json:"hostname,omitempty"`
	HostnameInCert bool     `json:"hostname_in_cert"`
	SANs           []string `json:"sans,omitempty"`
}

// fetchFromIP retrieves the certificate details from the provided IP address
func (cd *IPCertDetails) fetchFromIP(ip net.IP, port int) error {
	// Create a TLS configuration that skips certificate verification
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	return cd.fetchFromIPWithDialer(ip, port, &tls.Dialer{
		Config: tlsConfig,
	})
}

// fetchFromIPWithDialer retrieves the certificate details from the provided IP address using a custom dialer
func (cd *IPCertDetails) fetchFromIPWithDialer(ip net.IP, port int, dialer Dialer) error {
	ipStr := ip.String()
	address := ipStr + ":" + strconv.Itoa(port)

	// Use the provided dialer to establish a connection
	conn, err := dialer.Dial("tcp", address)
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
	if len(certs) == 0 {
		return fmt.Errorf("no certificates found for IP %s", ipStr)
	}

	cert := certs[0]

	// Set the base CertDetails
	cd.CertDetails = &CertDetails{
		Domain:         ipStr,
		Serial:         cert.SerialNumber.String(),
		NotBefore:      cert.NotBefore.String(),
		NotAfter:       cert.NotAfter.String(),
		Issuer:         cert.Issuer.String(),
		CRL:            cert.CRLDistributionPoints,
		OCSPServer:     cert.OCSPServer,
		CertChain:      certs,
		Valid:          true, // Assume valid until proven otherwise
		ValidationErrs: []string{},
	}

	// Set IP-specific details
	cd.IP = ipStr

	// Perform reverse DNS lookup
	hostname, err := reverseDNSLookup(ip)
	if err == nil {
		cd.Hostname = hostname

		// Check if hostname is in the certificate
		cd.HostnameInCert = isHostnameInCert(cert, hostname)

		// If we have a hostname, validate the certificate against it
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
			DNSName:       hostname,
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
				cd.ValidationErrs = append(cd.ValidationErrs, "Certificate is not valid for hostname: "+hostname)
			case x509.UnknownAuthorityError:
				cd.ValidationErrs = append(cd.ValidationErrs, "Certificate signed by unknown authority (possibly self-signed)")
			default:
				cd.ValidationErrs = append(cd.ValidationErrs, "Certificate validation error: "+err.Error())
			}
		}
	} else {
		// If we couldn't get a hostname, validate the certificate without a hostname
		// This will at least check for expiration and other basic issues
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

		// Verify the certificate chain without a hostname
		opts := x509.VerifyOptions{
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
			case x509.UnknownAuthorityError:
				cd.ValidationErrs = append(cd.ValidationErrs, "Certificate signed by unknown authority (possibly self-signed)")
			default:
				cd.ValidationErrs = append(cd.ValidationErrs, "Certificate validation error: "+err.Error())
			}
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

	// Extract SANs from the certificate
	cd.SANs = extractSANs(cert)

	return nil
}

// isHostnameInCert checks if the hostname is in the certificate's Common Name or SANs
func isHostnameInCert(cert *x509.Certificate, hostname string) bool {
	// Check Common Name
	if cert.Subject.CommonName == hostname {
		return true
	}

	// Check SANs
	for _, san := range cert.DNSNames {
		if san == hostname {
			return true
		}
	}

	return false
}

// extractSANs extracts all Subject Alternative Names from the certificate
func extractSANs(cert *x509.Certificate) []string {
	return cert.DNSNames
}

// reverseDNSLookup performs a reverse DNS lookup for the given IP address with a timeout
func reverseDNSLookup(ip net.IP) (string, error) {
	// Create a channel to receive the lookup result
	resultChan := make(chan struct {
		names []string
		err   error
	}, 1)

	// Perform the lookup in a goroutine
	go func() {
		names, err := net.LookupAddr(ip.String())
		resultChan <- struct {
			names []string
			err   error
		}{names, err}
	}()

	// Wait for the result or timeout
	select {
	case result := <-resultChan:
		if result.err != nil {
			return "", result.err
		}
		if len(result.names) == 0 {
			return "", fmt.Errorf("no hostname found for IP %s", ip.String())
		}
		// Remove trailing dot from hostname
		return strings.TrimSuffix(result.names[0], "."), nil
	case <-time.After(5 * time.Second):
		return "", fmt.Errorf("reverse DNS lookup for IP %s timed out after 5 seconds", ip.String())
	}
}

// ScrapeIPTLS scrapes the given IP addresses for TLS certificate details
// concurrently and returns the collected information.
func ScrapeIPTLS(ips []net.IP, port int, concurrency int) ([]*IPCertDetails, error) {
	results := make(chan *IPCertDetails, len(ips))
	errorChan := make(chan map[string]error, len(ips))

	sem := make(chan struct{}, concurrency)

	var wg sync.WaitGroup

	// For each IP, fetch certificate details in a goroutine.
	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr net.IP) {
			defer wg.Done()

			sem <- struct{}{} // Acquire a concurrency token

			ipStr := ipAddr.String()
			timer := prometheus.NewTimer(scrapeDuration.WithLabelValues(ipStr))
			defer timer.ObserveDuration()

			certInfo := &IPCertDetails{}
			err := certInfo.fetchFromIP(ipAddr, port)

			<-sem // Release a concurrency token

			if err != nil {
				errorChan <- map[string]error{ipStr: err}
				totalScrapes.WithLabelValues("failed").Inc()
				return
			}
			totalScrapes.WithLabelValues("success").Inc()
			results <- certInfo
		}(ip)
	}

	// Close result channels when all scraping goroutines are done.
	go func() {
		wg.Wait()
		close(results)
		close(errorChan)
	}()

	var details []*IPCertDetails

	multiError := &MultiError{Errors: make(map[string]error)}

	for res := range results {
		details = append(details, res)
	}

	for err := range errorChan {
		for ip, e := range err {
			multiError.Errors[ip] = e
		}
	}

	if len(multiError.Errors) > 0 {
		return details, multiError
	}

	return details, nil
}

// String provides a string representation of the IP certificate details.
func (c *IPCertDetails) String() string {
	hostnameInfo := ""
	if c.Hostname != "" {
		hostnameInfo = fmt.Sprintf("Hostname:%s HostnameInCert:%t ", c.Hostname, c.HostnameInCert)
	}

	validationInfo := ""
	if !c.Valid && len(c.ValidationErrs) > 0 {
		validationInfo = fmt.Sprintf("Valid:%t ValidationErrors:%v ", c.Valid, c.ValidationErrs)
	} else {
		validationInfo = fmt.Sprintf("Valid:%t ", c.Valid)
	}

	return fmt.Sprintf(
		"IP:%s "+
			"%s"+
			"%s"+
			"Serial:%s "+
			"NotBefore:%s "+
			"NotAfter:%s "+
			"Issuer:%s "+
			"CRL:%s "+
			"OCSPServer:%s "+
			"SANs:%s",
		c.IP,
		hostnameInfo,
		validationInfo,
		c.Serial,
		c.NotBefore,
		c.NotAfter,
		c.Issuer,
		c.CRL,
		c.OCSPServer,
		strings.Join(c.SANs, ","),
	)
}
