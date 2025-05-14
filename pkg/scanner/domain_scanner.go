package scanner

import (
	"github.com/scotta01/tls-scrape/internal/helper"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"log"
)

// DomainScannerConfig holds the configuration for domain scanning
type DomainScannerConfig struct {
	FQDN         string
	FilePath     string
	CSVHeader    string
	OutputDir    string
	Concurrency  int
	PrettyJSON   bool
	BundleOutput bool
}

// ScanDomains is a higher-level function that scans domains for TLS certificates using the provided configuration.
// It uses the ScanDomainsInternal function to perform the actual scanning.
func ScanDomains(config DomainScannerConfig) ([]*scraper.CertDetails, map[string]error) {
	var websites []string
	var err error

	if config.FQDN != "" {
		websites = []string{config.FQDN}
	} else {
		websites, err = helper.ReadCSV(config.FilePath, config.CSVHeader)
		if err != nil {
			log.Fatalf("error reading CSV: %v", err)
		}
	}

	// Use the ScanDomainsInternal function to handle chunking and processing
	details, errors := ScanDomainsInternal(websites, config.Concurrency, config.Concurrency)

	// Handle errors
	for domain, e := range errors {
		// Check if it's a connection error
		if scraper.IsConnectionError(e) {
			log.Printf("Skipping domain %s: Connection failed (domain may be unreachable or not running a TLS service)", domain)
		} else {
			log.Printf("Skipping domain %s: %s", domain, e.Error())
		}
	}

	// If bundling output, collect all certificate details
	var allCertDetails []*scraper.CertDetails

	if config.OutputDir != "" {
		if config.BundleOutput {
			// Collect certificate details for bundled output
			allCertDetails = append(allCertDetails, details...)
		} else {
			// Write individual JSON files
			for _, detail := range details {
				err = helper.WriteJSON(config.OutputDir, detail, config.PrettyJSON)
				if err != nil {
					log.Printf("Error writing JSON for domain %s: %v", detail.Domain, err)
				}
			}
		}
	}

	err = helper.WriteLog(details)
	if err != nil {
		log.Printf("Error writing log: %v", err)
	}

	// Write bundled output if requested
	if config.OutputDir != "" && config.BundleOutput && len(allCertDetails) > 0 {
		err := helper.WriteBundledJSON(config.OutputDir, allCertDetails, config.PrettyJSON)
		if err != nil {
			log.Printf("Error writing bundled JSON: %v", err)
		}
	}

	return details, errors
}
