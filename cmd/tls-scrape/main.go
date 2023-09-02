package main

import (
	"github.com/scotta01/tls-scrape/internal/helper"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
)

func bindEnvWithFallback(key string) {
	if value, exists := os.LookupEnv(strings.ToUpper(key)); exists {
		viper.Set(key, value)
	} else if value, exists := os.LookupEnv(strings.ToLower(key)); exists {
		viper.Set(key, value)
	}
}

func init() {
	bindEnvWithFallback("fqdn")
	bindEnvWithFallback("filepath")
	bindEnvWithFallback("header")
	bindEnvWithFallback("outfile")
	bindEnvWithFallback("concurrency")

	pflag.String("fqdn", "", "Fully Qualified Domain Name")
	pflag.String("filepath", "", "Path to the websites CSV file")
	pflag.String("header", "url", "Column header to look for in the CSV")
	pflag.String("outfile", "", "Output path for JSON file")
	pflag.Int("concurrency", 10, "Maximum number of concurrent TLS connections")
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	if err != nil {
		return
	}

}

func main() {
	fqdn := viper.GetString("fqdn")
	filepath := viper.GetString("filepath")
	csvHeader := viper.GetString("header")
	output := viper.GetString("outfile")
	concurrency := viper.GetInt("concurrency")

	if fqdn != "" && filepath != "" {
		log.Fatal("You can only pass either fqdn or filepath and header, but not both.")
	}
	if fqdn == "" && filepath == "" {
		log.Fatal("You must pass either fqdn or filepath.")
	}

	var websites []string
	var err error

	if fqdn != "" {
		websites = []string{fqdn}
	} else {
		websites, err = helper.ReadCSV(filepath, csvHeader)
		if err != nil {
			log.Fatalf("error reading CSV: %v", err)
		}
	}

	details, err := scraper.ScrapeTLS(websites, concurrency)
	if err != nil {
		if multiErr, ok := err.(*scraper.MultiError); ok {
			for domain, e := range multiErr.Errors {
				log.Printf("Failed to scrape domain %s with error: %s", domain, e.Error())
			}
		} else {
			log.Fatal("Error scraping TLS:", err)
		}
	}

	if output != "" {
		err = helper.WriteJSON(output, details)
		if err != nil {
			log.Fatal("Error writing JSON:", err)
		}
	}
	err = helper.WriteLog(details)
	if err != nil {
		log.Fatal("Error writing log:", err)
	}

}
