package firewall

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

func DetectFirewallTool() string {
	if _, err := exec.LookPath("nftables"); err == nil {
		return "nftables"
	}
	if _, err := exec.LookPath("iptables"); err == nil {
		return "iptables"
	}
	return "none"
}

func DetectDefaultInterface() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "ip", "route", "show", "default").CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ip route command failed: %w", err)
	}

	parts := strings.Fields(string(out))
	for i, p := range parts {
		if p == "dev" && i+1 < len(parts) {
			return parts[i+1], nil
		}
	}
	return "", fmt.Errorf("could not determine default interface from output")
}

func IsSSHListening() bool {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:22", 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func NewManager() (Manager, error) {
	tool := DetectFirewallTool()
	switch tool {
	case "iptables":
		return NewIPTablesManager(), nil
	default:
		return nil, fmt.Errorf("unsupported or missing firewall tool: %s", tool)
	}
}
