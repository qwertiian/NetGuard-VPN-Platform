package vpn

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type PeerStatus struct {
	PublicKey           string
	Endpoint            string
	AllowedIPs          string
	LatestHandshake     time.Time
	TransferRx          int64
	TransferTx          int64
	PersistentKeepalive int
}

type InterfaceStatus struct {
	Name       string
	PublicKey  string
	ListenPort int
	Peers      []PeerStatus
}

func (m *Manager) GetStatus(ctx context.Context) (*InterfaceStatus, error) {
	if !ifaceRegex.MatchString(m.wgInterface) {
		return nil, fmt.Errorf("invalid interface name")
	}
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "wg", "show", m.wgInterface, "dump")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	status, err := ParseWgDump(string(out))
	if err != nil {
		return nil, err
	}
	status.Name = m.wgInterface
	return status, nil
}

func ParseWgDump(output string) (*InterfaceStatus, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, fmt.Errorf("empty dump")
	}

	ifaceParts := strings.Split(lines[0], "\t")
	if len(ifaceParts) < 4 {
		return nil, fmt.Errorf("invalid interface line")
	}

	port, _ := strconv.Atoi(ifaceParts[2])
	status := &InterfaceStatus{
		PublicKey:  ifaceParts[1],
		ListenPort: port,
	}

	for i := 1; i < len(lines); i++ {
		parts := strings.Split(lines[i], "\t")
		if len(parts) < 8 {
			continue
		}
		
		ts, _ := strconv.ParseInt(parts[4], 10, 64)
		rx, _ := strconv.ParseInt(parts[5], 10, 64)
		tx, _ := strconv.ParseInt(parts[6], 10, 64)
		keepalive, _ := strconv.Atoi(parts[7])

		var handshake time.Time
		if ts > 0 {
			handshake = time.Unix(ts, 0)
		}

		status.Peers = append(status.Peers, PeerStatus{
			PublicKey:           parts[0],
			Endpoint:            parts[2],
			AllowedIPs:          parts[3],
			LatestHandshake:     handshake,
			TransferRx:          rx,
			TransferTx:          tx,
			PersistentKeepalive: keepalive,
		})
	}

	return status, nil
}
