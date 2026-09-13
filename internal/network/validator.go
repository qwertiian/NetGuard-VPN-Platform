package network

import (
	"errors"
	"fmt"
	"net"
	"regexp"
)

var interfaceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_=+.-]{1,15}$`)

// ValidateCIDR checks if a string is a valid CIDR notation.
func ValidateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return fmt.Errorf("invalid CIDR notation: %w", err)
	}
	return nil
}

// ValidateIP checks if a string is a valid IP address.
func ValidateIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("invalid IP address")
	}
	return nil
}

// ValidateIPInSubnet checks if an IP belongs to a given subnet.
func ValidateIPInSubnet(ip string, subnet string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("invalid IP address")
	}

	_, parsedSubnet, err := net.ParseCIDR(subnet)
	if err != nil {
		return fmt.Errorf("invalid subnet: %w", err)
	}

	if !parsedSubnet.Contains(parsedIP) {
		return errors.New("IP is not within the subnet")
	}

	return nil
}

// ValidateSubnetsNoOverlap checks that a list of subnets do not overlap.
func ValidateSubnetsNoOverlap(subnets []string) error {
	var parsedSubnets []*net.IPNet
	for _, s := range subnets {
		_, parsed, err := net.ParseCIDR(s)
		if err != nil {
			return fmt.Errorf("invalid subnet %s: %w", s, err)
		}
		parsedSubnets = append(parsedSubnets, parsed)
	}

	for i := 0; i < len(parsedSubnets); i++ {
		for j := i + 1; j < len(parsedSubnets); j++ {
			if parsedSubnets[i].Contains(parsedSubnets[j].IP) || parsedSubnets[j].Contains(parsedSubnets[i].IP) {
				return fmt.Errorf("subnets %s and %s overlap", subnets[i], subnets[j])
			}
		}
	}

	return nil
}

// ValidatePort checks if a port is within the valid range (1-65535).
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d: must be between 1 and 65535", port)
	}
	return nil
}

// ValidateInterfaceName checks if a network interface name is valid.
func ValidateInterfaceName(name string) error {
	if !interfaceNameRegex.MatchString(name) {
		return fmt.Errorf("invalid interface name %s: must match ^[a-zA-Z0-9_=+.-]{1,15}$", name)
	}
	return nil
}
