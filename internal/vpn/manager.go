package vpn

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var ifaceRegex = regexp.MustCompile(`^[a-zA-Z0-9_=+.-]{1,15}$`)

type Manager struct {
	configPath  string
	wgInterface string
}

func NewManager(wgInterface, configDir string) *Manager {
	return &Manager{
		wgInterface: wgInterface,
		configPath:  filepath.Join(configDir, wgInterface+".conf"),
	}
}

func (m *Manager) GenerateServerConfig(cfg ServerConfigParams) (string, error) {
	return GenerateServerConfigFile(cfg), nil
}

func (m *Manager) WriteServerConfig(content, path string) error {
	return os.WriteFile(path, []byte(content), 0600)
}

func (m *Manager) Start(ctx context.Context) error {
	if !ifaceRegex.MatchString(m.wgInterface) {
		return fmt.Errorf("invalid interface name")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wg-quick", "up", m.configPath)
	return cmd.Run()
}

func (m *Manager) Stop(ctx context.Context) error {
	if !ifaceRegex.MatchString(m.wgInterface) {
		return fmt.Errorf("invalid interface name")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wg-quick", "down", m.configPath)
	return cmd.Run()
}

func (m *Manager) Restart(ctx context.Context) error {
	if err := m.Stop(ctx); err != nil {
		// Ignore stop error in case it's not running
	}
	return m.Start(ctx)
}

func (m *Manager) IsRunning() (bool, error) {
	if !ifaceRegex.MatchString(m.wgInterface) {
		return false, fmt.Errorf("invalid interface name")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "ip", "link", "show", m.wgInterface)
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (m *Manager) SyncConfig(ctx context.Context) error {
	if !ifaceRegex.MatchString(m.wgInterface) {
		return fmt.Errorf("invalid interface name")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wg", "syncconf", m.wgInterface, m.configPath)
	return cmd.Run()
}

func (m *Manager) EnableIPForwarding() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sysctl", "-w", "net.ipv4.ip_forward=1")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to enable IP forwarding runtime: %v", err)
	}
	return os.WriteFile("/etc/sysctl.d/99-netguard.conf", []byte("net.ipv4.ip_forward=1\n"), 0644)
}

func (m *Manager) IsIPForwardingEnabled() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "sysctl", "-n", "net.ipv4.ip_forward")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "1", nil
}
