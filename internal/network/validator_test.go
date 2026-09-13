package network

import (
	"testing"
)

func TestValidateCIDR(t *testing.T) {
	if err := ValidateCIDR("10.0.0.0/24"); err != nil {
		t.Errorf("Valid CIDR failed: %v", err)
	}
	if err := ValidateCIDR("invalid"); err == nil {
		t.Error("Invalid CIDR passed")
	}
}

func TestValidateIP(t *testing.T) {
	if err := ValidateIP("192.168.1.1"); err != nil {
		t.Errorf("Valid IP failed: %v", err)
	}
	if err := ValidateIP("999.999.999.999"); err == nil {
		t.Error("Invalid IP passed")
	}
}

func TestValidateIPInSubnet(t *testing.T) {
	if err := ValidateIPInSubnet("10.0.0.5", "10.0.0.0/24"); err != nil {
		t.Errorf("IP in subnet failed: %v", err)
	}
	if err := ValidateIPInSubnet("192.168.1.5", "10.0.0.0/24"); err == nil {
		t.Error("IP outside subnet passed")
	}
}

func TestValidateSubnetsNoOverlap(t *testing.T) {
	err := ValidateSubnetsNoOverlap([]string{"10.0.0.0/24", "192.168.1.0/24"})
	if err != nil {
		t.Errorf("Non-overlapping subnets failed: %v", err)
	}

	err = ValidateSubnetsNoOverlap([]string{"10.0.0.0/16", "10.0.0.0/24"})
	if err == nil {
		t.Error("Overlapping subnets passed")
	}
}

func TestValidatePort(t *testing.T) {
	if err := ValidatePort(80); err != nil {
		t.Errorf("Valid port failed: %v", err)
	}
	if err := ValidatePort(0); err == nil {
		t.Error("Port 0 passed")
	}
	if err := ValidatePort(70000); err == nil {
		t.Error("Port 70000 passed")
	}
}

func TestValidateInterfaceName(t *testing.T) {
	if err := ValidateInterfaceName("wg0"); err != nil {
		t.Errorf("Valid interface name failed: %v", err)
	}
	if err := ValidateInterfaceName("very_long_interface_name_that_exceeds_15_chars"); err == nil {
		t.Error("Too long interface name passed")
	}
	if err := ValidateInterfaceName("wg0@invalid"); err == nil {
		t.Error("Invalid characters passed")
	}
}
