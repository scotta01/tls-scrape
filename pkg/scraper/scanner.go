package scraper

import (
	"net"
)

// ScanDomains is a lower-level function that scans a list of domains for TLS certificates.
// It is used by the scanner package to implement higher-level scanning functionality.
// It handles chunking the domains for concurrent processing and error handling.
// The chunkSize parameter controls how many domains are processed in each chunk.
func ScanDomains(domains []string, concurrency int, chunkSize int) ([]*CertDetails, map[string]error) {
	// Chunk the domains for concurrent processing
	chunks := ChunkSlice(domains, chunkSize)

	// Collect all certificate details
	var allCertDetails []*CertDetails
	allErrors := make(map[string]error)

	for _, chunk := range chunks {
		details, err := ScrapeTLS(chunk, concurrency)
		if err != nil {
			if multiErr, ok := err.(*MultiError); ok {
				for domain, e := range multiErr.Errors {
					allErrors[domain] = e
				}
			}
		}

		allCertDetails = append(allCertDetails, details...)
	}

	return allCertDetails, allErrors
}

// ScanIPAddresses is a lower-level function that scans a list of IP addresses for TLS certificates.
// It is used by the scanner package to implement higher-level scanning functionality.
// It handles chunking the IPs for concurrent processing and error handling.
// The chunkSize parameter controls how many IPs are processed in each chunk.
func ScanIPAddresses(ips []net.IP, port int, concurrency int, chunkSize int) ([]*IPCertDetails, map[string]error) {
	// Chunk the IPs for concurrent processing
	chunks := ChunkIPSlice(ips, chunkSize)

	// Collect all certificate details
	var allCertDetails []*IPCertDetails
	allErrors := make(map[string]error)

	for _, chunk := range chunks {
		details, err := ScrapeIPTLS(chunk, port, concurrency)
		if err != nil {
			if multiErr, ok := err.(*MultiError); ok {
				for ip, e := range multiErr.Errors {
					allErrors[ip] = e
				}
			}
		}

		allCertDetails = append(allCertDetails, details...)
	}

	return allCertDetails, allErrors
}
