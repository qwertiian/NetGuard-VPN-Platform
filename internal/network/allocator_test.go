package network

import (
	"testing"
)

func TestAllocator(t *testing.T) {
	alloc, err := NewAllocator("10.0.0.0/29", "10.0.0.1")
	if err != nil {
		t.Fatalf("Failed to create allocator: %v", err)
	}

	// Subnet /29 has 8 IPs: 10.0.0.0 to 10.0.0.7
	// Network: .0, Broadcast: .7, Server: .1
	// Available: .2, .3, .4, .5, .6 (5 IPs)

	if alloc.AvailableCount() != 5 {
		t.Errorf("Expected 5 available IPs, got %d", alloc.AvailableCount())
	}

	ip, err := alloc.AllocateIP()
	if err != nil {
		t.Fatalf("Failed to allocate IP: %v", err)
	}
	if ip != "10.0.0.2" {
		t.Errorf("Expected 10.0.0.2, got %s", ip)
	}

	if alloc.UsedCount() != 1 {
		t.Errorf("Expected 1 used IP, got %d", alloc.UsedCount())
	}

	err = alloc.ReserveIP("10.0.0.4")
	if err != nil {
		t.Fatalf("Failed to reserve IP: %v", err)
	}

	if !alloc.IsAvailable("10.0.0.3") {
		t.Error("10.0.0.3 should be available")
	}

	// exhaust the rest
	alloc.AllocateIP() // .3
	alloc.AllocateIP() // .5
	alloc.AllocateIP() // .6

	_, err = alloc.AllocateIP()
	if err == nil {
		t.Error("Expected error when allocating from exhausted subnet")
	}

	err = alloc.ReleaseIP("10.0.0.4")
	if err != nil {
		t.Fatalf("Failed to release IP: %v", err)
	}

	ip, err = alloc.AllocateIP()
	if err != nil {
		t.Fatalf("Failed to allocate after release: %v", err)
	}
	if ip != "10.0.0.4" {
		t.Errorf("Expected 10.0.0.4, got %s", ip)
	}
}

func TestAllocatorInvalid(t *testing.T) {
	_, err := NewAllocator("invalid", "10.0.0.1")
	if err == nil {
		t.Error("Expected error for invalid subnet")
	}

	_, err = NewAllocator("10.0.0.0/24", "192.168.1.1")
	if err == nil {
		t.Error("Expected error when server IP is outside subnet")
	}
}
