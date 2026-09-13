package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/netguard-vpn/netguard/internal/vpn"
	"github.com/spf13/cobra"
)

var peerCmd = &cobra.Command{
	Use:   "peer",
	Short: "View live WireGuard peer status",
}

var peerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List WireGuard peers from running interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := mgr.GetStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get peer status: %w\n\nIs the VPN server running?\n  ngvpn server status", err)
		}

		if isJSON() {
			printJSON(status.Peers)
			return nil
		}

		if len(status.Peers) == 0 {
			fmt.Println("No peers connected.")
			return nil
		}

		rows := make([][]string, len(status.Peers))
		for i, p := range status.Peers {
			endpoint := p.Endpoint
			if endpoint == "" {
				endpoint = "(none)"
			}
			handshake := "never"
			if !p.LatestHandshake.IsZero() {
				handshake = formatDuration(time.Since(p.LatestHandshake))
			}
			rows[i] = []string{
				p.PublicKey[:16] + "...",
				p.AllowedIPs,
				endpoint,
				handshake,
			}
		}
		formatTable([]string{"Public Key", "Allowed IPs", "Endpoint", "Handshake"}, rows)
		return nil
	},
}

var peerStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show detailed WireGuard peer status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		status, err := mgr.GetStatus(ctx)
		if err != nil {
			return fmt.Errorf("failed to get peer status: %w", err)
		}

		if isJSON() {
			printJSON(status)
			return nil
		}

		fmt.Printf("Interface: %s\n", status.Name)
		fmt.Printf("Public Key: %s\n", status.PublicKey)
		fmt.Printf("Listen Port: %d\n\n", status.ListenPort)

		for i, p := range status.Peers {
			if i > 0 {
				fmt.Println()
			}
			fmt.Printf("Peer: %s\n", p.PublicKey)
			fmt.Printf("  Endpoint:       %s\n", p.Endpoint)
			fmt.Printf("  Allowed IPs:    %s\n", p.AllowedIPs)
			if !p.LatestHandshake.IsZero() {
				fmt.Printf("  Last Handshake: %s\n", formatDuration(time.Since(p.LatestHandshake)))
			} else {
				fmt.Printf("  Last Handshake: never\n")
			}
			fmt.Printf("  Transfer:       ↓ %s  ↑ %s\n", formatBytes(p.TransferRx), formatBytes(p.TransferTx))
		}
		return nil
	},
}

func init() {
	peerCmd.AddCommand(peerListCmd)
	peerCmd.AddCommand(peerStatusCmd)
	rootCmd.AddCommand(peerCmd)
}
