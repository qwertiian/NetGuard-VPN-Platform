package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Server.ListenPort != 51820 {
		t.Errorf("Expected DefaultConfig to have ListenPort 51820, got %d", cfg.Server.ListenPort)
	}
	if cfg.Network.IPv4Subnet != "10.77.0.0/24" {
		t.Errorf("Expected DefaultConfig to have IPv4Subnet 10.77.0.0/24, got %s", cfg.Network.IPv4Subnet)
	}
}

func TestConfigLoadSave(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.yaml")

	cfg := DefaultConfig()
	cfg.Server.Endpoint = "example.com"

	err := cfg.Save(configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	loadedCfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedCfg.Server.Endpoint != cfg.Server.Endpoint {
		t.Errorf("Expected endpoint %s, got %s", cfg.Server.Endpoint, loadedCfg.Server.Endpoint)
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := DefaultConfig()
	
	// Default config should be valid
	if err := cfg.Validate(); err != nil {
		t.Errorf("Default config should be valid, got: %v", err)
	}

	// Invalid port
	cfg.Server.ListenPort = 70000
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for invalid port")
	}
	cfg.Server.ListenPort = 51820

	// Empty endpoint
	cfg.Server.Endpoint = ""
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for empty endpoint")
	}
	cfg.Server.Endpoint = "example.com"

	// Invalid CIDR
	cfg.Network.IPv4Subnet = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for invalid CIDR")
	}
	cfg.Network.IPv4Subnet = "10.77.0.0/24"

	// Invalid Routing Mode
	cfg.Routing.Mode = "invalid"
	if err := cfg.Validate(); err == nil {
		t.Error("Expected error for invalid routing mode")
	}
}
