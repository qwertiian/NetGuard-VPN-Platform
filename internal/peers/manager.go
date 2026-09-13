package peers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/netguard-vpn/netguard/internal/storage"
	"github.com/netguard-vpn/netguard/internal/network"
	"github.com/netguard-vpn/netguard/internal/vpn"
	"github.com/netguard-vpn/netguard/internal/config"
)

var nameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

type Manager struct {
	db      *storage.Database
	alloc   *network.Allocator
	keyDir  string
	cfg     *config.Config
}

func NewManager(db *storage.Database, alloc *network.Allocator, keyDir string, cfg *config.Config) *Manager {
	return &Manager{
		db:     db,
		alloc:  alloc,
		keyDir: keyDir,
		cfg:    cfg,
	}
}

func (m *Manager) AddPeer(name string) (*storage.Peer, string, error) {
	if !nameRegex.MatchString(name) {
		return nil, "", fmt.Errorf("invalid peer name format")
	}

	existing, err := m.db.GetPeer(name)
	if err == nil && existing != nil {
		return nil, "", fmt.Errorf("peer %s already exists", name)
	}

	privKey, pubKey, err := generateKeys()
	if err != nil {
		return nil, "", err
	}
	
	psk, err := generatePresharedKey()
	if err != nil {
		return nil, "", err
	}

	if err := os.MkdirAll(m.keyDir, 0700); err != nil {
		return nil, "", err
	}

	privPath := filepath.Join(m.keyDir, name+".key")
	if err := os.WriteFile(privPath, []byte(privKey), 0600); err != nil {
		return nil, "", err
	}

	pskPath := filepath.Join(m.keyDir, name+".psk")
	if err := os.WriteFile(pskPath, []byte(psk), 0600); err != nil {
		os.Remove(privPath)
		return nil, "", err
	}

	ip, err := m.alloc.AllocateIP()
	if err != nil {
		os.Remove(privPath)
		os.Remove(pskPath)
		return nil, "", err
	}

	peer := &storage.Peer{
		Name:             name,
		PublicKey:        pubKey,
		PresharedKeyPath: pskPath,
		PrivateKeyPath:   privPath,
		VPNIPv4:          ip,
		AllowedIPs:       ip + "/32",
		Status:           "active",
		CreatedAt:        time.Now(),
	}

	if err := m.db.CreatePeer(peer); err != nil {
		m.alloc.ReleaseIP(ip)
		os.Remove(privPath)
		os.Remove(pskPath)
		return nil, "", err
	}

	serverPubKeyBytes, _ := os.ReadFile(filepath.Join(m.keyDir, "server.pub"))

	clientCfgParams := vpn.ClientConfigParams{
		PrivateKey:   privKey,
		Address:      ip + "/32",
		DNS:          strings.Join(m.cfg.DNS.Servers, ", "),
		ServerPubKey: strings.TrimSpace(string(serverPubKeyBytes)),
		PresharedKey: psk,
		Endpoint:     formatEndpoint(m.cfg.Server.Endpoint, m.cfg.Server.ListenPort),
		AllowedIPs:   "0.0.0.0/0",
		Keepalive:    25,
	}

	clientCfg := vpn.GenerateClientConfigFile(clientCfgParams)
	return peer, clientCfg, nil
}

func (m *Manager) ListPeers() ([]*storage.Peer, error) {
	return m.db.ListPeers()
}

func (m *Manager) GetPeer(name string) (*storage.Peer, error) {
	return m.db.GetPeer(name)
}

func (m *Manager) RevokePeer(name string) error {
	peer, err := m.db.GetPeer(name)
	if err != nil {
		return err
	}
	peer.Status = "revoked"
	return m.db.UpdatePeer(peer)
}

func (m *Manager) RemovePeer(name string) error {
	peer, err := m.db.GetPeer(name)
	if err != nil {
		return err
	}

	if err := m.db.DeletePeer(name); err != nil {
		return err
	}

	ip := strings.Split(peer.AllowedIPs, "/")[0]
	m.alloc.ReleaseIP(ip)

	os.Remove(filepath.Join(m.keyDir, name+".key"))
	os.Remove(filepath.Join(m.keyDir, name+".psk"))

	return nil
}

func (m *Manager) RotateKeys(name string) (*storage.Peer, string, error) {
	peer, err := m.db.GetPeer(name)
	if err != nil {
		return nil, "", err
	}

	privKey, pubKey, err := generateKeys()
	if err != nil {
		return nil, "", err
	}
	
	psk, err := generatePresharedKey()
	if err != nil {
		return nil, "", err
	}

	privPath := filepath.Join(m.keyDir, name+".key")
	if err := os.WriteFile(privPath, []byte(privKey), 0600); err != nil {
		return nil, "", err
	}

	pskPath := filepath.Join(m.keyDir, name+".psk")
	if err := os.WriteFile(pskPath, []byte(psk), 0600); err != nil {
		return nil, "", err
	}

	peer.PublicKey = pubKey
	peer.PresharedKeyPath = pskPath
	peer.PrivateKeyPath = privPath
	
	if err := m.db.UpdatePeer(peer); err != nil {
		return nil, "", err
	}

	serverPubKeyBytes, _ := os.ReadFile(filepath.Join(m.keyDir, "server.pub"))

	clientCfgParams := vpn.ClientConfigParams{
		PrivateKey:   privKey,
		Address:      peer.AllowedIPs,
		DNS:          strings.Join(m.cfg.DNS.Servers, ", "),
		ServerPubKey: strings.TrimSpace(string(serverPubKeyBytes)),
		PresharedKey: psk,
		Endpoint:     formatEndpoint(m.cfg.Server.Endpoint, m.cfg.Server.ListenPort),
		AllowedIPs:   "0.0.0.0/0",
		Keepalive:    25,
	}

	clientCfg := vpn.GenerateClientConfigFile(clientCfgParams)
	return peer, clientCfg, nil
}

func (m *Manager) ExportConfig(name string) (string, error) {
	peer, err := m.db.GetPeer(name)
	if err != nil {
		return "", err
	}

	privKeyBytes, err := os.ReadFile(filepath.Join(m.keyDir, name+".key"))
	if err != nil {
		return "", fmt.Errorf("could not read private key: %v", err)
	}

	pskBytes, _ := os.ReadFile(filepath.Join(m.keyDir, name+".psk"))
	serverPubKeyBytes, _ := os.ReadFile(filepath.Join(m.keyDir, "server.pub"))

	clientCfgParams := vpn.ClientConfigParams{
		PrivateKey:   strings.TrimSpace(string(privKeyBytes)),
		Address:      peer.AllowedIPs,
		DNS:          strings.Join(m.cfg.DNS.Servers, ", "),
		ServerPubKey: strings.TrimSpace(string(serverPubKeyBytes)),
		PresharedKey: strings.TrimSpace(string(pskBytes)),
		Endpoint:     formatEndpoint(m.cfg.Server.Endpoint, m.cfg.Server.ListenPort),
		AllowedIPs:   "0.0.0.0/0",
		Keepalive:    25,
	}

	return vpn.GenerateClientConfigFile(clientCfgParams), nil
}

func (m *Manager) GetServerPeersConfig() ([]vpn.PeerConfigParams, error) {
	peers, err := m.db.ListPeers()
	if err != nil {
		return nil, err
	}

	var params []vpn.PeerConfigParams
	for _, p := range peers {
		if p.Status == "active" {
			pskBytes, _ := os.ReadFile(filepath.Join(m.keyDir, p.Name+".psk"))
			params = append(params, vpn.PeerConfigParams{
				PublicKey:    p.PublicKey,
				PresharedKey: strings.TrimSpace(string(pskBytes)),
				AllowedIPs:   p.AllowedIPs,
				Comment:      "# " + p.Name,
			})
		}
	}
	return params, nil
}

func (m *Manager) LoadAllocatedIPs() error {
	peers, err := m.db.ListPeers()
	if err != nil {
		return err
	}
	for _, p := range peers {
		ip := strings.Split(p.AllowedIPs, "/")[0]
		m.alloc.ReserveIP(ip)
	}
	return nil
}

func generateKeys() (string, string, error) {
	privCmd := exec.Command("wg", "genkey")
	privOut, err := privCmd.Output()
	if err != nil {
		return "", "", err
	}
	privKey := strings.TrimSpace(string(privOut))

	pubCmd := exec.Command("wg", "pubkey")
	pubCmd.Stdin = strings.NewReader(privKey)
	pubOut, err := pubCmd.Output()
	if err != nil {
		return "", "", err
	}
	pubKey := strings.TrimSpace(string(pubOut))

	return privKey, pubKey, nil
}

func generatePresharedKey() (string, error) {
	cmd := exec.Command("wg", "genpsk")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func formatEndpoint(endpoint string, port int) string {
	// If the endpoint contains a colon (like IPv6) and doesn't already have brackets, wrap it
	if strings.Contains(endpoint, ":") && !strings.HasPrefix(endpoint, "[") {
		return fmt.Sprintf("[%s]:%d", endpoint, port)
	}
	return fmt.Sprintf("%s:%d", endpoint, port)
}
