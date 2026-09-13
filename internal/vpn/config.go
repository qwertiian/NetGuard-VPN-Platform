package vpn

import (
	"fmt"
	"strings"
)

type ServerConfigParams struct {
	PrivateKey   string
	Address      string
	ListenPort   int
	ExtInterface string
	Subnet       string
	Peers        []PeerConfigParams
}

type PeerConfigParams struct {
	PublicKey    string
	PresharedKey string
	AllowedIPs   string
	Comment      string
}

type ClientConfigParams struct {
	PrivateKey   string
	Address      string
	DNS          string
	ServerPubKey string
	PresharedKey string
	Endpoint     string
	AllowedIPs   string
	Keepalive    int
}

func GenerateServerConfigFile(params ServerConfigParams) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", params.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", params.Address))
	sb.WriteString(fmt.Sprintf("ListenPort = %d\n", params.ListenPort))
	
	if params.ExtInterface != "" && params.Subnet != "" {
		postUp := fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s -o %s -j MASQUERADE -m comment --comment \"netguard: nat\"; iptables -A FORWARD -i wg0 -o %s -j ACCEPT -m comment --comment \"netguard: forward-out\"; iptables -A FORWARD -i %s -o wg0 -m state --state RELATED,ESTABLISHED -j ACCEPT -m comment --comment \"netguard: forward-in\"", params.Subnet, params.ExtInterface, params.ExtInterface, params.ExtInterface)
		postDown := fmt.Sprintf("iptables -t nat -D POSTROUTING -s %s -o %s -j MASQUERADE -m comment --comment \"netguard: nat\"; iptables -D FORWARD -i wg0 -o %s -j ACCEPT -m comment --comment \"netguard: forward-out\"; iptables -D FORWARD -i %s -o wg0 -m state --state RELATED,ESTABLISHED -j ACCEPT -m comment --comment \"netguard: forward-in\"", params.Subnet, params.ExtInterface, params.ExtInterface, params.ExtInterface)
		sb.WriteString(fmt.Sprintf("PostUp = %s\n", postUp))
		sb.WriteString(fmt.Sprintf("PostDown = %s\n", postDown))
	}

	for _, peer := range params.Peers {
		sb.WriteString("\n[Peer]\n")
		if peer.Comment != "" {
			if !strings.HasPrefix(peer.Comment, "#") {
				sb.WriteString("# ")
			}
			sb.WriteString(peer.Comment + "\n")
		}
		sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
		sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", peer.AllowedIPs))
	}

	return strings.TrimSpace(sb.String()) + "\n"
}

func GenerateClientConfigFile(params ClientConfigParams) string {
	var sb strings.Builder

	sb.WriteString("[Interface]\n")
	sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", params.PrivateKey))
	sb.WriteString(fmt.Sprintf("Address = %s\n", params.Address))
	if params.DNS != "" {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", params.DNS))
	}
	
	sb.WriteString("\n[Peer]\n")
	sb.WriteString(fmt.Sprintf("PublicKey = %s\n", params.ServerPubKey))
	if params.PresharedKey != "" {
		sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", params.PresharedKey))
	}
	sb.WriteString(fmt.Sprintf("Endpoint = %s\n", params.Endpoint))
	sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", params.AllowedIPs))
	if params.Keepalive > 0 {
		sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", params.Keepalive))
	}

	return sb.String()
}
