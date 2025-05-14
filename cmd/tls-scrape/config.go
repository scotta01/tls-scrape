package main

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"strings"
)

// Config holds all the configuration parameters for the application
type Config struct {
	FQDN         string
	FilePath     string
	CSVHeader    string
	OutputDir    string
	Concurrency  int
	PrettyJSON   bool
	BundleOutput bool
	IPAddr       string
	Subnet       string
	Port         int
}

// bindEnvWithFallback binds environment variables to viper with fallback to lowercase
func bindEnvWithFallback(key string) {
	if value, exists := os.LookupEnv(strings.ToUpper(key)); exists {
		viper.Set(key, value)
	} else if value, exists := os.LookupEnv(strings.ToLower(key)); exists {
		viper.Set(key, value)
	}
}

// setupFlags initializes command line flags and binds them to viper
func setupFlags() {
	// Bind environment variables
	bindEnvWithFallback("fqdn")
	bindEnvWithFallback("filepath")
	bindEnvWithFallback("header")
	bindEnvWithFallback("outdir")
	bindEnvWithFallback("concurrency")
	bindEnvWithFallback("prettyjson")
	bindEnvWithFallback("bundle")
	bindEnvWithFallback("ip")
	bindEnvWithFallback("subnet")
	bindEnvWithFallback("port")

	// Define command line flags
	pflag.String("fqdn", "", "Fully Qualified Domain Name")
	pflag.String("filepath", "", "Path to the websites CSV file")
	pflag.String("header", "url", "Column header to look for in the CSV")
	pflag.String("outdir", "", "Output path for JSON file")
	pflag.Int("concurrency", 10, "Maximum number of concurrent TLS connections")
	pflag.Bool("prettyjson", false, "Pretty print JSON output")
	pflag.Bool("bundle", false, "Bundle all output into a single JSON file")
	pflag.String("ip", "", "IP address to scan")
	pflag.String("subnet", "", "Subnet in CIDR notation to scan (e.g., 192.168.1.0/24)")
	pflag.Int("port", 443, "Port to connect to for TLS scanning")

	pflag.Parse()
	_ = viper.BindPFlags(pflag.CommandLine)
}

// loadConfig loads the configuration from viper into a Config struct
func loadConfig() Config {
	return Config{
		FQDN:         viper.GetString("fqdn"),
		FilePath:     viper.GetString("filepath"),
		CSVHeader:    viper.GetString("header"),
		OutputDir:    viper.GetString("outdir"),
		Concurrency:  viper.GetInt("concurrency"),
		PrettyJSON:   viper.GetBool("prettyjson"),
		BundleOutput: viper.GetBool("bundle"),
		IPAddr:       viper.GetString("ip"),
		Subnet:       viper.GetString("subnet"),
		Port:         viper.GetInt("port"),
	}
}

// validateConfig checks for mutually exclusive options and returns an error message if invalid
func validateConfig(config Config) (bool, string) {
	// Check for mutually exclusive options
	optionCount := 0
	if config.FQDN != "" {
		optionCount++
	}
	if config.FilePath != "" {
		optionCount++
	}
	if config.IPAddr != "" {
		optionCount++
	}
	if config.Subnet != "" {
		optionCount++
	}

	if optionCount > 1 {
		return false, "You can only specify one of: fqdn, filepath, ip, or subnet."
	}
	if optionCount == 0 {
		return false, "You must specify one of: fqdn, filepath, ip, or subnet."
	}

	return true, ""
}
