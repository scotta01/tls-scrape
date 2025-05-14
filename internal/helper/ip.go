package helper

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

// IPRange represents a range of IP addresses
type IPRange struct {
	Start net.IP
	End   net.IP
}

// ParseIPOrSubnet parses an IP address or subnet in CIDR notation
// and returns a range of IP addresses
func ParseIPOrSubnet(ipOrSubnet string) (*IPRange, error) {
	// Check if it's a CIDR notation
	if strings.Contains(ipOrSubnet, "/") {
		return parseSubnet(ipOrSubnet)
	}

	// It's a single IP address
	ip := net.ParseIP(ipOrSubnet)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipOrSubnet)
	}

	// Convert to IPv4 if it's an IPv4-mapped IPv6 address
	if ip.To4() != nil {
		ip = ip.To4()
	}

	return &IPRange{
		Start: ip,
		End:   ip,
	}, nil
}

// parseSubnet parses a subnet in CIDR notation and returns a range of IP addresses
func parseSubnet(cidr string) (*IPRange, error) {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}

	// Convert to IPv4 if it's an IPv4-mapped IPv6 address
	if ip.To4() != nil {
		ip = ip.To4()
	}

	// Calculate the start and end IP addresses
	start := ip.Mask(ipNet.Mask)
	end := make(net.IP, len(start))
	copy(end, start)

	// Calculate the end IP by flipping all the masked bits to 1
	for i := 0; i < len(ipNet.Mask); i++ {
		end[i] |= ^ipNet.Mask[i]
	}

	return &IPRange{
		Start: start,
		End:   end,
	}, nil
}

// NextIP returns the next IP address after the given IP
func NextIP(ip net.IP) net.IP {
	next := make(net.IP, len(ip))
	copy(next, ip)

	for i := len(next) - 1; i >= 0; i-- {
		next[i]++
		if next[i] > 0 {
			break
		}
	}

	return next
}

// CompareIPs compares two IP addresses
// Returns -1 if ip1 < ip2, 0 if ip1 == ip2, 1 if ip1 > ip2
func CompareIPs(ip1, ip2 net.IP) int {
	// Ensure both IPs are of the same type (IPv4 or IPv6)
	if len(ip1) != len(ip2) {
		if ip1.To4() != nil && ip2.To4() != nil {
			ip1 = ip1.To4()
			ip2 = ip2.To4()
		} else {
			// Cannot compare different IP versions
			return 0
		}
	}

	for i := 0; i < len(ip1); i++ {
		if ip1[i] < ip2[i] {
			return -1
		}
		if ip1[i] > ip2[i] {
			return 1
		}
	}
	return 0
}

// GetIPsInRange returns all IP addresses in the given range
func GetIPsInRange(ipRange *IPRange) []net.IP {
	var ips []net.IP
	for ip := ipRange.Start; CompareIPs(ip, ipRange.End) <= 0; ip = NextIP(ip) {
		newIP := make(net.IP, len(ip))
		copy(newIP, ip)
		ips = append(ips, newIP)
	}
	return ips
}

// ParsePort parses a port string and returns the port number
func ParsePort(port string) (int, error) {
	if port == "" {
		return 443, nil // Default to HTTPS port
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return 0, fmt.Errorf("invalid port number: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return 0, fmt.Errorf("port number out of range: %d", portNum)
	}

	return portNum, nil
}
