package firewall

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var validInterface = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

type IPTablesManager struct{}

func NewIPTablesManager() *IPTablesManager {
	return &IPTablesManager{}
}

func (m *IPTablesManager) validateInterface(name string) error {
	if !validInterface.MatchString(name) {
		return fmt.Errorf("invalid interface name: %s", name)
	}
	return nil
}

func (m *IPTablesManager) runCMD(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "iptables", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("iptables %v failed: %w (%s)", args, err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (m *IPTablesManager) Setup(wgInterface, extInterface, subnet string) error {
	if err := m.validateInterface(wgInterface); err != nil {
		return err
	}
	if err := m.validateInterface(extInterface); err != nil {
		return err
	}

	rules := [][]string{
		{"-A", "FORWARD", "-i", wgInterface, "-o", extInterface, "-m", "comment", "--comment", "netguard: forward_wg_to_ext", "-j", "ACCEPT"},
		{"-A", "FORWARD", "-i", extInterface, "-o", wgInterface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-m", "comment", "--comment", "netguard: forward_ext_to_wg", "-j", "ACCEPT"},
		{"-t", "nat", "-A", "POSTROUTING", "-s", subnet, "-o", extInterface, "-m", "comment", "--comment", "netguard: masquerade", "-j", "MASQUERADE"},
	}

	for _, rule := range rules {
		_ = m.runCMD(rule...)
	}
	return nil
}

func (m *IPTablesManager) Teardown(wgInterface, extInterface, subnet string) error {
	if err := m.validateInterface(wgInterface); err != nil {
		return err
	}
	if err := m.validateInterface(extInterface); err != nil {
		return err
	}

	rules := [][]string{
		{"-D", "FORWARD", "-i", wgInterface, "-o", extInterface, "-m", "comment", "--comment", "netguard: forward_wg_to_ext", "-j", "ACCEPT"},
		{"-D", "FORWARD", "-i", extInterface, "-o", wgInterface, "-m", "state", "--state", "RELATED,ESTABLISHED", "-m", "comment", "--comment", "netguard: forward_ext_to_wg", "-j", "ACCEPT"},
		{"-t", "nat", "-D", "POSTROUTING", "-s", subnet, "-o", extInterface, "-m", "comment", "--comment", "netguard: masquerade", "-j", "MASQUERADE"},
	}

	for _, rule := range rules {
		_ = m.runCMD(rule...)
	}
	return nil
}

func (m *IPTablesManager) IsConfigured(wgInterface string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "iptables-save")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to check iptables: %w", err)
	}

	return strings.Contains(string(out), "netguard: forward_wg_to_ext") && strings.Contains(string(out), wgInterface), nil
}

func (m *IPTablesManager) EnsureSSHAccess() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "iptables", "-C", "INPUT", "-p", "tcp", "--dport", "22", "-j", "ACCEPT", "-m", "comment", "--comment", "netguard: ssh_access")
	if err := cmd.Run(); err == nil {
		return nil
	}

	return m.runCMD("-I", "INPUT", "1", "-p", "tcp", "--dport", "22", "-j", "ACCEPT", "-m", "comment", "--comment", "netguard: ssh_access")
}

func (m *IPTablesManager) BackupRules(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "iptables-save").CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to backup rules: %w", err)
	}
	return os.WriteFile(path, out, 0600)
}

func (m *IPTablesManager) RestoreRules(path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
    
    file, err := os.Open(path)
    if err != nil {
        return fmt.Errorf("failed to open rules file: %w", err)
    }
    defer file.Close()
    
	cmd := exec.CommandContext(ctx, "iptables-restore")
    cmd.Stdin = file
	if err := cmd.Run(); err != nil {
        return fmt.Errorf("failed to restore rules: %w", err)
    }
    return nil
}
