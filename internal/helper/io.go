package helper

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func ReadCSV(filename string, csvheader string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Return an error if the CSV is empty
	if len(lines) == 0 {
		return nil, errors.New("empty CSV file")
	}

	// Identify the column index based on the header
	colIndex := -1
	for index, header := range lines[0] {
		if header == csvheader {
			colIndex = index
			break
		}
	}

	// Return an error if the column header isn't found
	if colIndex == -1 {
		return nil, fmt.Errorf("column header '%s' not found", csvheader)
	}

	var websites []string
	// Start from index 1 to skip the header row
	for _, line := range lines[1:] {
		if len(line) > colIndex {
			websites = append(websites, line[colIndex])
		}
	}
	return websites, nil
}

func WriteJSON(directory string, details *scraper.CertDetails, prettyPrint bool) error {
	var data []byte
	var err error

	if prettyPrint {
		data, err = json.MarshalIndent(details, "", "  ")
	} else {
		data, err = json.Marshal(details)
	}

	if err != nil {
		return err
	}
	// Add a newline to the end of the file so that commands like tail can read it.
	data = append(data, '\n')
	filename := filepath.Join(directory, details.Domain+".json")
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteLog(details []*scraper.CertDetails) error {
	var logString []string
	for _, detail := range details {
		// Format CRL and OCSPServer slices to be more readable
		crlStr := formatStringSlice(detail.CRL)
		ocspStr := formatStringSlice(detail.OCSPServer)

		logString = append(logString, fmt.Sprintf(
			"tls-scrape "+
				"Domain:%s "+
				"Serial:%s "+
				"NotBefore:%s "+
				"NotAfter:%s "+
				"Issuer:%s "+
				"CRL:%s "+
				"OCSPServer:%s ",
			detail.Domain,
			detail.Serial,
			detail.NotBefore,
			detail.NotAfter,
			detail.Issuer,
			crlStr,
			ocspStr,
		))
	}

	for _, i := range logString {
		log.Println(i)
	}

	return nil
}

// formatStringSlice converts a slice of strings to a readable format
// Returns a comma-separated string of the slice elements, or "null" if the slice is nil or empty
func formatStringSlice(slice []string) string {
	if slice == nil || len(slice) == 0 {
		return "null"
	}
	return strings.Join(slice, ", ")
}

// WriteBundledJSON writes multiple certificate details to a single JSON file.
// The filename will be in the format "tls-scrape-bundle-YYYYMMDD-HHMMSS.json"
func WriteBundledJSON(directory string, details []*scraper.CertDetails, prettyPrint bool) error {
	if len(details) == 0 {
		return nil // Nothing to write
	}

	var data []byte
	var err error

	if prettyPrint {
		data, err = json.MarshalIndent(details, "", "  ")
	} else {
		data, err = json.Marshal(details)
	}

	if err != nil {
		return err
	}

	// Add a newline to the end of the file
	data = append(data, '\n')

	// Create a filename with timestamp
	timestamp := time.Now().Format("20060102-150405") // YYYYMMDD-HHMMSS format
	filename := filepath.Join(directory, fmt.Sprintf("tls-scrape-bundle-%s.json", timestamp))

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}

	log.Printf("Bundled %d certificate details into %s", len(details), filename)
	return nil
}
