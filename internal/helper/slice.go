package helper

import (
	"github.com/scotta01/tls-scrape/pkg/scraper"
	"net"
)

// ChunkSlice divides a string slice into chunks of the specified size
func ChunkSlice(slice []string, chunkSize int) [][]string {
	return scraper.ChunkSlice(slice, chunkSize)
}

// ChunkIPSlice divides a net.IP slice into chunks of the specified size
func ChunkIPSlice(slice []net.IP, chunkSize int) [][]net.IP {
	return scraper.ChunkIPSlice(slice, chunkSize)
}
