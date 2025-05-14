package scraper

import (
	"errors"
	"net"
	"reflect"
	"testing"
)

func TestChunkSlice(t *testing.T) {
	tests := []struct {
		name      string
		slice     []string
		chunkSize int
		want      [][]string
	}{
		{
			name:      "empty slice",
			slice:     []string{},
			chunkSize: 2,
			want:      [][]string{},
		},
		{
			name:      "slice smaller than chunk size",
			slice:     []string{"a", "b"},
			chunkSize: 3,
			want:      [][]string{{"a", "b"}},
		},
		{
			name:      "slice equal to chunk size",
			slice:     []string{"a", "b", "c"},
			chunkSize: 3,
			want:      [][]string{{"a", "b", "c"}},
		},
		{
			name:      "slice larger than chunk size",
			slice:     []string{"a", "b", "c", "d", "e"},
			chunkSize: 2,
			want:      [][]string{{"a", "b"}, {"c", "d"}, {"e"}},
		},
		{
			name:      "chunk size of 1",
			slice:     []string{"a", "b", "c"},
			chunkSize: 1,
			want:      [][]string{{"a"}, {"b"}, {"c"}},
		},
		{
			name:      "chunk size of 0",
			slice:     []string{"a", "b", "c"},
			chunkSize: 0,
			want:      [][]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunkSlice(tt.slice, tt.chunkSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChunkSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChunkIPSlice(t *testing.T) {
	ip1 := net.ParseIP("192.168.1.1")
	ip2 := net.ParseIP("192.168.1.2")
	ip3 := net.ParseIP("192.168.1.3")
	ip4 := net.ParseIP("192.168.1.4")
	ip5 := net.ParseIP("192.168.1.5")

	tests := []struct {
		name      string
		slice     []net.IP
		chunkSize int
		want      [][]net.IP
	}{
		{
			name:      "empty slice",
			slice:     []net.IP{},
			chunkSize: 2,
			want:      [][]net.IP{},
		},
		{
			name:      "slice smaller than chunk size",
			slice:     []net.IP{ip1, ip2},
			chunkSize: 3,
			want:      [][]net.IP{{ip1, ip2}},
		},
		{
			name:      "slice equal to chunk size",
			slice:     []net.IP{ip1, ip2, ip3},
			chunkSize: 3,
			want:      [][]net.IP{{ip1, ip2, ip3}},
		},
		{
			name:      "slice larger than chunk size",
			slice:     []net.IP{ip1, ip2, ip3, ip4, ip5},
			chunkSize: 2,
			want:      [][]net.IP{{ip1, ip2}, {ip3, ip4}, {ip5}},
		},
		{
			name:      "chunk size of 1",
			slice:     []net.IP{ip1, ip2, ip3},
			chunkSize: 1,
			want:      [][]net.IP{{ip1}, {ip2}, {ip3}},
		},
		{
			name:      "chunk size of 0",
			slice:     []net.IP{ip1, ip2, ip3},
			chunkSize: 0,
			want:      [][]net.IP{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ChunkIPSlice(tt.slice, tt.chunkSize)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ChunkIPSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsConnectionError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "non-connection error",
			err:  errors.New("some other error"),
			want: false,
		},
		{
			name: "dial tcp error",
			err:  errors.New("dial tcp: connection refused"),
			want: true,
		},
		{
			name: "connect error",
			err:  errors.New("connect: network is unreachable"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsConnectionError(tt.err)
			if got != tt.want {
				t.Errorf("IsConnectionError() = %v, want %v", got, tt.want)
			}
		})
	}
}
