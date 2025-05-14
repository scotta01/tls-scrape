package helper

import (
	"net"
	"reflect"
	"testing"
)

func TestParseIPOrSubnet(t *testing.T) {
	tests := []struct {
		name       string
		ipOrSubnet string
		want       *IPRange
		wantErr    bool
	}{
		{
			name:       "valid IP address",
			ipOrSubnet: "192.168.1.1",
			want: &IPRange{
				Start: net.ParseIP("192.168.1.1").To4(),
				End:   net.ParseIP("192.168.1.1").To4(),
			},
			wantErr: false,
		},
		{
			name:       "valid IPv6 address",
			ipOrSubnet: "2001:db8::1",
			want: &IPRange{
				Start: net.ParseIP("2001:db8::1"),
				End:   net.ParseIP("2001:db8::1"),
			},
			wantErr: false,
		},
		{
			name:       "valid subnet",
			ipOrSubnet: "192.168.1.0/24",
			want: &IPRange{
				Start: net.ParseIP("192.168.1.0").To4(),
				End:   net.ParseIP("192.168.1.255").To4(),
			},
			wantErr: false,
		},
		{
			name:       "invalid IP address",
			ipOrSubnet: "invalid",
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid subnet",
			ipOrSubnet: "192.168.1.0/invalid",
			want:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseIPOrSubnet(tt.ipOrSubnet)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseIPOrSubnet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("ParseIPOrSubnet() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestNextIP(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
		want net.IP
	}{
		{
			name: "IPv4 address",
			ip:   net.ParseIP("192.168.1.1").To4(),
			want: net.ParseIP("192.168.1.2").To4(),
		},
		{
			name: "IPv4 address with overflow",
			ip:   net.ParseIP("192.168.1.255").To4(),
			want: net.ParseIP("192.168.2.0").To4(),
		},
		{
			name: "IPv6 address",
			ip:   net.ParseIP("2001:db8::1"),
			want: net.ParseIP("2001:db8::2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NextIP(tt.ip)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NextIP() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompareIPs(t *testing.T) {
	tests := []struct {
		name string
		ip1  net.IP
		ip2  net.IP
		want int
	}{
		{
			name: "ip1 < ip2",
			ip1:  net.ParseIP("192.168.1.1").To4(),
			ip2:  net.ParseIP("192.168.1.2").To4(),
			want: -1,
		},
		{
			name: "ip1 == ip2",
			ip1:  net.ParseIP("192.168.1.1").To4(),
			ip2:  net.ParseIP("192.168.1.1").To4(),
			want: 0,
		},
		{
			name: "ip1 > ip2",
			ip1:  net.ParseIP("192.168.1.2").To4(),
			ip2:  net.ParseIP("192.168.1.1").To4(),
			want: 1,
		},
		{
			name: "IPv4 and IPv6",
			ip1:  net.ParseIP("192.168.1.1").To4(),
			ip2:  net.ParseIP("2001:db8::1"),
			want: 0,
		},
		{
			name: "IPv4 and IPv4-mapped IPv6",
			ip1:  net.ParseIP("192.168.1.1").To4(),
			ip2:  net.ParseIP("::ffff:192.168.1.1"),
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CompareIPs(tt.ip1, tt.ip2)
			if got != tt.want {
				t.Errorf("CompareIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIPsInRange(t *testing.T) {
	tests := []struct {
		name    string
		ipRange *IPRange
		want    []net.IP
	}{
		{
			name: "single IP",
			ipRange: &IPRange{
				Start: net.ParseIP("192.168.1.1").To4(),
				End:   net.ParseIP("192.168.1.1").To4(),
			},
			want: []net.IP{net.ParseIP("192.168.1.1").To4()},
		},
		{
			name: "multiple IPs",
			ipRange: &IPRange{
				Start: net.ParseIP("192.168.1.1").To4(),
				End:   net.ParseIP("192.168.1.3").To4(),
			},
			want: []net.IP{
				net.ParseIP("192.168.1.1").To4(),
				net.ParseIP("192.168.1.2").To4(),
				net.ParseIP("192.168.1.3").To4(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetIPsInRange(tt.ipRange)
			if len(got) != len(tt.want) {
				t.Errorf("GetIPsInRange() got %d IPs, want %d IPs", len(got), len(tt.want))
				return
			}
			for i := range got {
				if !reflect.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("GetIPsInRange()[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		want    int
		wantErr bool
	}{
		{
			name:    "empty string (default port)",
			port:    "",
			want:    443,
			wantErr: false,
		},
		{
			name:    "valid port",
			port:    "8080",
			want:    8080,
			wantErr: false,
		},
		{
			name:    "port 0",
			port:    "0",
			want:    0,
			wantErr: true,
		},
		{
			name:    "port too large",
			port:    "65536",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid port",
			port:    "invalid",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParsePort() = %v, want %v", got, tt.want)
			}
		})
	}
}
