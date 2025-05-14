// Package main provides the entry point for the tls-scrape CLI tool.
// The code has been refactored into multiple files for better organization:
// - main.go: Entry point and high-level flow control
// - config.go: Configuration handling and flag setup
package main

import (
	"github.com/scotta01/tls-scrape/pkg/scanner"
	"log"
)

func init() {
	setupFlags()
}

func main() {
	config := loadConfig()

	valid, errMsg := validateConfig(config)
	if !valid {
		log.Fatal(errMsg)
	}

	// Handle IP or subnet scanning
	if config.IPAddr != "" || config.Subnet != "" {
		// Create IP scanner configuration
		ipConfig := scanner.IPScannerConfig{
			IPAddr:       config.IPAddr,
			Subnet:       config.Subnet,
			Port:         config.Port,
			OutputDir:    config.OutputDir,
			Concurrency:  config.Concurrency,
			PrettyJSON:   config.PrettyJSON,
			BundleOutput: config.BundleOutput,
		}

		// Use the scanner package to scan IP addresses
		scanner.ScanIPAddresses(ipConfig)
		return
	}

	// Handle domain scanning
	// Create domain scanner configuration
	domainConfig := scanner.DomainScannerConfig{
		FQDN:         config.FQDN,
		FilePath:     config.FilePath,
		CSVHeader:    config.CSVHeader,
		OutputDir:    config.OutputDir,
		Concurrency:  config.Concurrency,
		PrettyJSON:   config.PrettyJSON,
		BundleOutput: config.BundleOutput,
	}

	// Use the scanner package to scan domains
	scanner.ScanDomains(domainConfig)
}
