package dns

import (
	"fmt"
	"net"
	"strings"
)

type Config struct {
	Servers []string `yaml:"servers"`
	Mode    string   `yaml:"mode"` // custom, system, vpn
}

func ValidateServers(servers []string) error {
	for _, s := range servers {
		if net.ParseIP(s) == nil {
			return fmt.Errorf("invalid DNS server IP: %s", s)
		}
	}
	return nil
}

func FormatForWireGuard(servers []string) string {
	return strings.Join(servers, ", ")
}

func DefaultServers() []string {
	return []string{"1.1.1.1", "9.9.9.9"}
}

func (c *Config) GetServersForClient(clientDNSOverride []string) []string {
	if len(clientDNSOverride) > 0 {
		return clientDNSOverride
	}
	if len(c.Servers) > 0 {
		return c.Servers
	}
	return DefaultServers()
}
