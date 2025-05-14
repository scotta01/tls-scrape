package scanner

import (
	"github.com/scotta01/tls-scrape/internal/helper"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"log"
	"strings"
)

// IPScannerConfig holds the configuration for IP scanning
type IPScannerConfig struct {
	IPAddr       string
	Subnet       string
	Port         int
	OutputDir    string
	Concurrency  int
	PrettyJSON   bool
	BundleOutput bool
}

// ScanIPAddresses is a higher-level function that scans IP addresses or subnets for TLS certificates using the provided configuration.
// It uses the ScanIPAddressesInternal function to perform the actual scanning.
func ScanIPAddresses(config IPScannerConfig) ([]*scraper.IPCertDetails, map[string]error) {
	var ipRange *helper.IPRange
	var err error

	if config.IPAddr != "" {
		ipRange, err = helper.ParseIPOrSubnet(config.IPAddr)
	} else {
		ipRange, err = helper.ParseIPOrSubnet(config.Subnet)
	}

	if err != nil {
		log.Fatalf("Error parsing IP or subnet: %v", err)
	}

	ips := helper.GetIPsInRange(ipRange)
	log.Printf("Scanning %d IP addresses on port %d", len(ips), config.Port)

	// Use the ScanIPAddressesInternal function to handle chunking and processing
	details, errors := ScanIPAddressesInternal(ips, config.Port, config.Concurrency, config.Concurrency)

	// Handle errors
	for ip, e := range errors {
		// Check if it's a connection error
		if scraper.IsConnectionError(e) {
			log.Printf("Skipping IP %s: Connection failed (IP may be unreachable or not running a TLS service)", ip)
		} else {
			log.Printf("Skipping IP %s: %s", ip, e.Error())
		}
	}

	// If bundling output, collect all certificate details
	var allCertDetails []*scraper.CertDetails

	if config.OutputDir != "" {
		if config.BundleOutput {
			// Collect certificate details for bundled output
			for _, detail := range details {
				allCertDetails = append(allCertDetails, detail.CertDetails)
			}
		} else {
			// Write individual JSON files
			for _, detail := range details {
				// Convert IPCertDetails to CertDetails for WriteJSON
				err = helper.WriteJSON(config.OutputDir, detail.CertDetails, config.PrettyJSON)
				if err != nil {
					log.Printf("Error writing JSON for IP %s: %v", detail.IP, err)
				}
			}
		}
	}

	// Log the results
	for _, detail := range details {
		log.Printf("IP: %s, Hostname: %s, Hostname in cert: %t, SANs: %s",
			detail.IP, detail.Hostname, detail.HostnameInCert, strings.Join(detail.SANs, ","))
	}

	// Write bundled output if requested
	if config.OutputDir != "" && config.BundleOutput && len(allCertDetails) > 0 {
		err = helper.WriteBundledJSON(config.OutputDir, allCertDetails, config.PrettyJSON)
		if err != nil {
			log.Printf("Error writing bundled JSON: %v", err)
		}
	}

	return details, errors
}
