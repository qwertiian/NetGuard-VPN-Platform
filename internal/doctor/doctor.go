package doctor

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Config struct {
	Interface string
	ConfigDir string
	KeyDir    string
	Subnet    string
	DBPath    string
}

type CheckResult struct {
	Name        string
	Status      string // "pass", "fail", "warn", "skip"
	Message     string
	Remediation string
}

type DoctorReport struct {
	Checks  []CheckResult
	Healthy bool
}

func RunDiagnostics(cfg *Config) *DoctorReport {
	checks := []CheckResult{
		checkWireGuardInstalled(),
		checkInterfaceExists(cfg.Interface),
		checkInterfaceActive(cfg.Interface),
		checkServerKeyExists(cfg.KeyDir),
		checkConfigValid(cfg.ConfigDir, cfg.Interface),
		checkIPForwarding(),
		checkFirewall(cfg.Interface),
		checkNAT(cfg.Subnet),
		checkDNSResolution(),
		checkPeerConfiguration(cfg.DBPath),
	}

	healthy := true
	for _, c := range checks {
		if c.Status == "fail" {
			healthy = false
			break
		}
	}

	return &DoctorReport{
		Checks:  checks,
		Healthy: healthy,
	}
}

func checkWireGuardInstalled() CheckResult {
	_, err := exec.LookPath("wg")
	if err != nil {
		return CheckResult{
			Name:        "WireGuard Installed",
			Status:      "fail",
			Message:     "wg command not found in PATH",
			Remediation: "Install wireguard-tools package using your system package manager",
		}
	}
	return CheckResult{Name: "WireGuard Installed", Status: "pass", Message: "wg command found"}
}

func checkInterfaceExists(iface string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ip", "link", "show", iface)
	if err := cmd.Run(); err != nil {
		return CheckResult{
			Name:        "Interface Exists",
			Status:      "fail",
			Message:     fmt.Sprintf("Interface %s does not exist", iface),
			Remediation: fmt.Sprintf("Run 'ngvpn start' to bring up the interface"),
		}
	}
	return CheckResult{Name: "Interface Exists", Status: "pass", Message: fmt.Sprintf("Interface %s exists", iface)}
}

func checkInterfaceActive(iface string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ip", "link", "show", iface, "up")
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "state UNKNOWN") && !strings.Contains(string(out), "state UP") {
		return CheckResult{
			Name:        "Interface Active",
			Status:      "fail",
			Message:     fmt.Sprintf("Interface %s is down", iface),
			Remediation: fmt.Sprintf("Run 'ip link set %s up'", iface),
		}
	}
	return CheckResult{Name: "Interface Active", Status: "pass", Message: fmt.Sprintf("Interface %s is up", iface)}
}

func checkServerKeyExists(keyDir string) CheckResult {
	path := keyDir + "/server.key"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return CheckResult{
			Name:        "Server Key Exists",
			Status:      "fail",
			Message:     "Server private key not found",
			Remediation: "Generate server keys or run init",
		}
	}
	return CheckResult{Name: "Server Key Exists", Status: "pass", Message: "Server key found"}
}

func checkConfigValid(configDir, iface string) CheckResult {
	path := fmt.Sprintf("%s/%s.conf", configDir, iface)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return CheckResult{
			Name:        "Config Valid",
			Status:      "fail",
			Message:     "Config file not found",
			Remediation: "Generate configuration using manager",
		}
	}
	return CheckResult{Name: "Config Valid", Status: "pass", Message: "Config file exists"}
}

func checkIPForwarding() CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sysctl", "-n", "net.ipv4.ip_forward")
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) != "1" {
		return CheckResult{
			Name:        "IP Forwarding",
			Status:      "fail",
			Message:     "IP forwarding is disabled",
			Remediation: "Run 'sysctl -w net.ipv4.ip_forward=1'",
		}
	}
	return CheckResult{Name: "IP Forwarding", Status: "pass", Message: "IP forwarding enabled"}
}

func checkFirewall(iface string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "iptables", "-S")
	out, err := cmd.Output()
	if err != nil {
		return CheckResult{Name: "Firewall Rules", Status: "skip", Message: "Cannot read iptables"}
	}
	if !strings.Contains(string(out), "netguard") {
		return CheckResult{
			Name:        "Firewall Rules",
			Status:      "warn",
			Message:     "No netguard firewall rules found",
			Remediation: "Ensure interface starts with PostUp rules",
		}
	}
	return CheckResult{Name: "Firewall Rules", Status: "pass", Message: "Firewall rules present"}
}

func checkNAT(subnet string) CheckResult {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "iptables", "-t", "nat", "-S")
	out, err := cmd.Output()
	if err != nil {
		return CheckResult{Name: "NAT Rules", Status: "skip", Message: "Cannot read nat table"}
	}
	if !strings.Contains(string(out), "MASQUERADE") {
		return CheckResult{
			Name:        "NAT Rules",
			Status:      "warn",
			Message:     "No NAT masquerade rules found",
			Remediation: "Add MASQUERADE rule for the VPN subnet",
		}
	}
	return CheckResult{Name: "NAT Rules", Status: "pass", Message: "NAT rules present"}
}

func checkDNSResolution() CheckResult {
	_, err := net.LookupHost("google.com")
	if err != nil {
		return CheckResult{
			Name:        "DNS Resolution",
			Status:      "fail",
			Message:     "Cannot resolve public DNS",
			Remediation: "Check system DNS configuration",
		}
	}
	return CheckResult{Name: "DNS Resolution", Status: "pass", Message: "DNS working"}
}

func checkPeerConfiguration(dbPath string) CheckResult {
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return CheckResult{
			Name:        "Peer Configuration",
			Status:      "warn",
			Message:     "Database not found",
			Remediation: "Init the server to create database",
		}
	}
	return CheckResult{Name: "Peer Configuration", Status: "pass", Message: "Database exists"}
}
