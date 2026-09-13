package config

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			ListenPort: 51820,
			Interface:  "wg0",
			Endpoint:   "127.0.0.1",
		},
		Network: NetworkConfig{
			IPv4Subnet:    "10.77.0.0/24",
			ServerAddress: "10.77.0.1",
		},
		DNS: DNSConfig{
			Servers: []string{"1.1.1.1", "8.8.8.8"},
		},
		Routing: RoutingConfig{
			Mode: "full",
		},
		Security: SecurityConfig{
			KeyDirectory: "/etc/netguard/keys",
		},
		Storage: StorageConfig{
			DataDirectory: "/etc/netguard",
			Database:      "netguard.db",
		},
	}
}
