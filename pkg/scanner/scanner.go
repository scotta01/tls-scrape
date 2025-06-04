package scanner

import (
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"net"
)

// ScanDomainsInternal is an internal function that scans a list of domains for TLS certificates
// It handles chunking the domains for concurrent processing and error handling
// The chunkSize parameter controls how many domains are processed in each chunk
// The port parameter specifies which port to connect to for TLS scanning
func ScanDomainsInternal(domains []string, concurrency int, chunkSize int, port int) ([]*scraper.CertDetails, map[string]error) {
	// Chunk the domains for concurrent processing
	chunks := scraper.ChunkSlice(domains, chunkSize)

	// Collect all certificate details
	var allCertDetails []*scraper.CertDetails
	allErrors := make(map[string]error)

	for _, chunk := range chunks {
		details, err := scraper.ScrapeTLS(chunk, concurrency, port)
		if err != nil {
			if multiErr, ok := err.(*scraper.MultiError); ok {
				for domain, e := range multiErr.Errors {
					allErrors[domain] = e
				}
			}
		}

		allCertDetails = append(allCertDetails, details...)
	}

	return allCertDetails, allErrors
}

// ScanIPAddressesInternal is an internal function that scans a list of IP addresses for TLS certificates
// It handles chunking the IPs for concurrent processing and error handling
// The chunkSize parameter controls how many IPs are processed in each chunk
func ScanIPAddressesInternal(ips []net.IP, port int, concurrency int, chunkSize int) ([]*scraper.IPCertDetails, map[string]error) {
	// Chunk the IPs for concurrent processing
	chunks := scraper.ChunkIPSlice(ips, chunkSize)

	// Collect all certificate details
	var allCertDetails []*scraper.IPCertDetails
	allErrors := make(map[string]error)

	for _, chunk := range chunks {
		details, err := scraper.ScrapeIPTLS(chunk, port, concurrency)
		if err != nil {
			if multiErr, ok := err.(*scraper.MultiError); ok {
				for ip, e := range multiErr.Errors {
					allErrors[ip] = e
				}
			}
		}

		allCertDetails = append(allCertDetails, details...)
	}

	return allCertDetails, allErrors
}
