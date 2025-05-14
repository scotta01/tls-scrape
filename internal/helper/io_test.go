package helper

import (
	"encoding/csv"
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"os"
	"path/filepath"
	"testing"
)

func TestReadCSV(t *testing.T) {
	// Create a temporary CSV file for testing
	tmpFile, err := os.CreateTemp("", "test-*.csv")
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data to the CSV file
	writer := csv.NewWriter(tmpFile)
	testData := [][]string{
		{"url", "name", "description"},
		{"example.com", "Example", "An example website"},
		{"test.com", "Test", "A test website"},
	}
	err = writer.WriteAll(testData)
	if err != nil {
		t.Fatalf("Failed to write to CSV file: %v", err)
	}
	writer.Flush()
	tmpFile.Close()

	// Test cases
	tests := []struct {
		name      string
		filename  string
		csvheader string
		want      []string
		wantErr   bool
	}{
		{
			name:      "valid CSV with url header",
			filename:  tmpFile.Name(),
			csvheader: "url",
			want:      []string{"example.com", "test.com"},
			wantErr:   false,
		},
		{
			name:      "valid CSV with name header",
			filename:  tmpFile.Name(),
			csvheader: "name",
			want:      []string{"Example", "Test"},
			wantErr:   false,
		},
		{
			name:      "valid CSV with non-existent header",
			filename:  tmpFile.Name(),
			csvheader: "nonexistent",
			want:      nil,
			wantErr:   true,
		},
		{
			name:      "non-existent file",
			filename:  "nonexistent.csv",
			csvheader: "url",
			want:      nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadCSV(tt.filename, tt.csvheader)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("ReadCSV() got = %v, want %v", got, tt.want)
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("ReadCSV() got[%d] = %v, want[%d] = %v", i, v, i, tt.want[i])
					}
				}
			}
		})
	}

	// Test empty CSV file
	emptyFile, err := os.CreateTemp("", "empty-*.csv")
	if err != nil {
		t.Fatalf("Failed to create empty temporary file: %v", err)
	}
	defer os.Remove(emptyFile.Name())
	emptyFile.Close()

	_, err = ReadCSV(emptyFile.Name(), "url")
	if err == nil {
		t.Errorf("ReadCSV() with empty file should return error")
	}
}

func TestWriteJSON(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-json")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test data
	testDetails := &scraper.CertDetails{
		Domain:     "example.com",
		Serial:     "123456",
		NotBefore:  "2020-01-01",
		NotAfter:   "2021-01-01",
		Issuer:     "Test CA",
		CRL:        []string{"http://crl.example.com"},
		OCSPServer: []string{"http://ocsp.example.com"},
	}

	// Test cases
	tests := []struct {
		name        string
		directory   string
		details     *scraper.CertDetails
		prettyPrint bool
		wantErr     bool
	}{
		{
			name:        "write JSON with pretty print",
			directory:   tmpDir,
			details:     testDetails,
			prettyPrint: true,
			wantErr:     false,
		},
		{
			name:        "write JSON without pretty print",
			directory:   tmpDir,
			details:     testDetails,
			prettyPrint: false,
			wantErr:     false,
		},
		{
			name:        "write to non-existent directory",
			directory:   filepath.Join(tmpDir, "nonexistent"),
			details:     testDetails,
			prettyPrint: false,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteJSON(tt.directory, tt.details, tt.prettyPrint)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check if the file was created
				filename := filepath.Join(tt.directory, tt.details.Domain+".json")
				if _, err := os.Stat(filename); os.IsNotExist(err) {
					t.Errorf("WriteJSON() did not create file %s", filename)
				}
			}
		})
	}
}

func TestWriteLog(t *testing.T) {
	// Create test data
	testDetails := []*scraper.CertDetails{
		{
			Domain:     "example.com",
			Serial:     "123456",
			NotBefore:  "2020-01-01",
			NotAfter:   "2021-01-01",
			Issuer:     "Test CA",
			CRL:        []string{"http://crl.example.com"},
			OCSPServer: []string{"http://ocsp.example.com"},
		},
		{
			Domain:     "test.com",
			Serial:     "654321",
			NotBefore:  "2020-02-01",
			NotAfter:   "2021-02-01",
			Issuer:     "Test CA 2",
			CRL:        []string{"http://crl.test.com"},
			OCSPServer: []string{"http://ocsp.test.com"},
		},
	}

	// Test WriteLog
	err := WriteLog(testDetails)
	if err != nil {
		t.Errorf("WriteLog() error = %v", err)
	}
}

func TestWriteBundledJSON(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "test-bundled-json")
	if err != nil {
		t.Fatalf("Failed to create temporary directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test data
	testDetails := []*scraper.CertDetails{
		{
			Domain:     "example.com",
			Serial:     "123456",
			NotBefore:  "2020-01-01",
			NotAfter:   "2021-01-01",
			Issuer:     "Test CA",
			CRL:        []string{"http://crl.example.com"},
			OCSPServer: []string{"http://ocsp.example.com"},
		},
		{
			Domain:     "test.com",
			Serial:     "654321",
			NotBefore:  "2020-02-01",
			NotAfter:   "2021-02-01",
			Issuer:     "Test CA 2",
			CRL:        []string{"http://crl.test.com"},
			OCSPServer: []string{"http://ocsp.test.com"},
		},
	}

	// Test cases
	tests := []struct {
		name        string
		directory   string
		details     []*scraper.CertDetails
		prettyPrint bool
		wantErr     bool
	}{
		{
			name:        "write bundled JSON with pretty print",
			directory:   tmpDir,
			details:     testDetails,
			prettyPrint: true,
			wantErr:     false,
		},
		{
			name:        "write bundled JSON without pretty print",
			directory:   tmpDir,
			details:     testDetails,
			prettyPrint: false,
			wantErr:     false,
		},
		{
			name:        "write to non-existent directory",
			directory:   filepath.Join(tmpDir, "nonexistent"),
			details:     testDetails,
			prettyPrint: false,
			wantErr:     true,
		},
		{
			name:        "write empty details",
			directory:   tmpDir,
			details:     []*scraper.CertDetails{},
			prettyPrint: false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := WriteBundledJSON(tt.directory, tt.details, tt.prettyPrint)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteBundledJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(tt.details) > 0 {
				// Check if a file with the expected prefix was created
				files, err := filepath.Glob(filepath.Join(tt.directory, "tls-scrape-bundle-*.json"))
				if err != nil {
					t.Errorf("Failed to list files in directory: %v", err)
					return
				}
				if len(files) == 0 {
					t.Errorf("WriteBundledJSON() did not create any files")
				}
			}
		})
	}
}
