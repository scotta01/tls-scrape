package helper

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"log"
	"os"
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

	filename := fmt.Sprintf("%s/%s.json", directory, details.Domain)
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

func WriteLog(details []*scraper.CertDetails) error {
	var logString []string
	for _, detail := range details {
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
			detail.CRL,
			detail.OCSPServer,
		))
	}

	for _, i := range logString {
		log.Println(i)
	}

	return nil
}
