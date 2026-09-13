package config

import (
	"errors"
	"fmt"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Network  NetworkConfig  `yaml:"network"`
	DNS      DNSConfig      `yaml:"dns"`
	Routing  RoutingConfig  `yaml:"routing"`
	Security SecurityConfig `yaml:"security"`
	Storage  StorageConfig  `yaml:"storage"`
}

type ServerConfig struct {
	Endpoint   string `yaml:"endpoint"`
	ListenPort int    `yaml:"listen_port"`
	Interface  string `yaml:"interface"`
}

type NetworkConfig struct {
	IPv4Subnet    string `yaml:"ipv4_subnet"`
	ServerAddress string `yaml:"server_address"`
}

type DNSConfig struct {
	Servers []string `yaml:"servers"`
}

type RoutingConfig struct {
	Mode        string   `yaml:"mode"` // "full" or "split"
	SplitRoutes []string `yaml:"split_routes"`
}

type SecurityConfig struct {
	KeyDirectory string `yaml:"key_directory"`
}

type StorageConfig struct {
	DataDirectory string `yaml:"data_directory"`
	Database      string `yaml:"database"`
}

// Load loads the configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}

// Save saves the configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to serialize config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate validates the configuration fields.
func (c *Config) Validate() error {
	if c.Server.ListenPort < 1 || c.Server.ListenPort > 65535 {
		return fmt.Errorf("invalid listen port: %d, must be between 1 and 65535", c.Server.ListenPort)
	}

	if c.Server.Endpoint == "" {
		return errors.New("server endpoint cannot be empty")
	}

	if _, _, err := net.ParseCIDR(c.Network.IPv4Subnet); err != nil {
		return fmt.Errorf("invalid ipv4_subnet: %w", err)
	}

	if c.Routing.Mode != "full" && c.Routing.Mode != "split" {
		return fmt.Errorf("invalid routing mode: %s, must be 'full' or 'split'", c.Routing.Mode)
	}

	return nil
}

// FindConfigFile searches for a configuration file in common paths.
func FindConfigFile() (string, error) {
	if path := os.Getenv("NGVPN_CONFIG_PATH"); path != "" {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	commonPaths := []string{
		"./config.yaml",
		"/etc/netguard/config.yaml",
	}

	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", errors.New("no configuration file found")
}
