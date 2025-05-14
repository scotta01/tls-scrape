package scraper

import (
	"net"
	"strings"
)

// ChunkSlice divides a string slice into chunks of the specified size
func ChunkSlice(slice []string, chunkSize int) [][]string {
	chunks := [][]string{}
	if chunkSize <= 0 {
		return chunks
	}
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// ChunkIPSlice divides a net.IP slice into chunks of the specified size
func ChunkIPSlice(slice []net.IP, chunkSize int) [][]net.IP {
	chunks := [][]net.IP{}
	if chunkSize <= 0 {
		return chunks
	}
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// IsConnectionError checks if an error is a connection error
func IsConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "dial tcp") || strings.Contains(errStr, "connect:")
}
