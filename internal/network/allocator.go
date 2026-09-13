package network

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

type Allocator struct {
	subnet    *net.IPNet
	serverIP  net.IP
	usedIPs   map[string]bool
}

// NewAllocator creates a new IP allocator for the given subnet.
func NewAllocator(subnetStr string, serverAddrStr string) (*Allocator, error) {
	_, subnet, err := net.ParseCIDR(subnetStr)
	if err != nil {
		return nil, fmt.Errorf("invalid subnet: %w", err)
	}

	serverIP := net.ParseIP(serverAddrStr)
	if serverIP == nil {
		return nil, errors.New("invalid server address")
	}

	if !subnet.Contains(serverIP) {
		return nil, errors.New("server address is not in the subnet")
	}

	return &Allocator{
		subnet:   subnet,
		serverIP: serverIP,
		usedIPs:  make(map[string]bool),
	}, nil
}

// AllocateIP returns the next available IP in the subnet.
func (a *Allocator) AllocateIP() (string, error) {
	ip := a.subnet.IP.To4()
	if ip == nil {
		return "", errors.New("only IPv4 is supported")
	}

	startIP := binary.BigEndian.Uint32(ip)
	mask := binary.BigEndian.Uint32(a.subnet.Mask)
	
	// Start from 1 to avoid network address, end at ^mask - 1 to avoid broadcast
	for i := uint32(1); i < ^mask; i++ {
		nextIPNum := startIP + i
		nextIP := make(net.IP, 4)
		binary.BigEndian.PutUint32(nextIP, nextIPNum)

		ipStr := nextIP.String()
		if ipStr == a.serverIP.String() {
			continue
		}

		if !a.usedIPs[ipStr] {
			a.usedIPs[ipStr] = true
			return ipStr, nil
		}
	}

	return "", errors.New("no available IPs in subnet")
}

// ReserveIP marks an IP as used.
func (a *Allocator) ReserveIP(ip string) error {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return errors.New("invalid IP address")
	}
	if !a.subnet.Contains(parsedIP) {
		return errors.New("IP is not in subnet")
	}
	if ip == a.serverIP.String() {
		return errors.New("cannot reserve server IP")
	}
	if a.usedIPs[ip] {
		return errors.New("IP is already in use")
	}
	a.usedIPs[ip] = true
	return nil
}

// ReleaseIP frees an IP.
func (a *Allocator) ReleaseIP(ip string) error {
	if !a.usedIPs[ip] {
		return errors.New("IP is not in use")
	}
	delete(a.usedIPs, ip)
	return nil
}

// IsAvailable checks if an IP is available.
func (a *Allocator) IsAvailable(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil || !a.subnet.Contains(parsedIP) || ip == a.serverIP.String() {
		return false
	}
	return !a.usedIPs[ip]
}

// UsedCount returns the number of used IPs.
func (a *Allocator) UsedCount() int {
	return len(a.usedIPs)
}

// AvailableCount returns the number of available IPs.
func (a *Allocator) AvailableCount() int {
	mask := binary.BigEndian.Uint32(a.subnet.Mask)
	totalIPs := int(^mask) - 1 // Exclude network and broadcast
	
	// Exclude server IP if it's within the valid range
	return totalIPs - 1 - len(a.usedIPs)
}
