package storage

import "time"

type Peer struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	PublicKey        string    `json:"public_key"`
	PrivateKeyPath   string    `json:"-"` // Never serialize
	PresharedKeyPath string    `json:"-"` // Never serialize
	VPNIPv4          string    `json:"vpn_ipv4"`
	DNSServers       string    `json:"dns_servers"`
	AllowedIPs       string    `json:"allowed_ips"`
	Keepalive        int       `json:"keepalive"`
	Status           string    `json:"status"` // active, revoked, removed
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type ServerState struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
