package vpn

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateServerConfigFile(t *testing.T) {
	params := ServerConfigParams{
		PrivateKey:   "priv_key",
		Address:      "10.77.0.1/24",
		ListenPort:   51820,
		ExtInterface: "eth0",
		Subnet:       "10.77.0.0/24",
		Peers: []PeerConfigParams{
			{
				PublicKey:    "pub_key",
				PresharedKey: "psk",
				AllowedIPs:   "10.77.0.2/32",
				Comment:      "# alice",
			},
		},
	}
	cfg := GenerateServerConfigFile(params)
	assert.Contains(t, cfg, "PrivateKey = priv_key")
	assert.Contains(t, cfg, "Address = 10.77.0.1/24")
	assert.Contains(t, cfg, "ListenPort = 51820")
	assert.Contains(t, cfg, "PostUp = iptables -t nat -A POSTROUTING")
	assert.Contains(t, cfg, "# alice")
	assert.Contains(t, cfg, "PublicKey = pub_key")
	assert.Contains(t, cfg, "AllowedIPs = 10.77.0.2/32")
}

func TestGenerateClientConfigFile(t *testing.T) {
	params := ClientConfigParams{
		PrivateKey:   "client_priv",
		Address:      "10.77.0.2/32",
		DNS:          "1.1.1.1",
		ServerPubKey: "server_pub",
		Endpoint:     "vpn.example.com:51820",
		AllowedIPs:   "0.0.0.0/0",
		Keepalive:    25,
	}
	cfg := GenerateClientConfigFile(params)
	assert.Contains(t, cfg, "PrivateKey = client_priv")
	assert.Contains(t, cfg, "DNS = 1.1.1.1")
	assert.Contains(t, cfg, "PublicKey = server_pub")
	assert.Contains(t, cfg, "Endpoint = vpn.example.com:51820")
	assert.Contains(t, cfg, "PersistentKeepalive = 25")
}
