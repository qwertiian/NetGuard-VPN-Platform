package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/netguard-vpn/netguard/internal/network"
	"github.com/netguard-vpn/netguard/internal/peers"
	"github.com/netguard-vpn/netguard/internal/storage"
	qrterminal "github.com/mdp/qrterminal/v3"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Manage VPN clients",
}

func getPeerManager() (*peers.Manager, *storage.Database, error) {
	cfg := getConfig()
	db, err := storage.Open(cfg.Storage.Database)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open database: %w\n\nHave you initialized the server?\n  ngvpn server init --endpoint <your-server-ip>", err)
	}
	if err := db.Initialize(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	alloc, err := network.NewAllocator(cfg.Network.IPv4Subnet, cfg.Network.ServerAddress)
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to create IP allocator: %w", err)
	}

	mgr := peers.NewManager(db, alloc, cfg.Security.KeyDirectory, cfg)
	if err := mgr.LoadAllocatedIPs(); err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("failed to load allocated IPs: %w", err)
	}

	return mgr, db, nil
}

var clientAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new VPN client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		peer, clientConfig, err := mgr.AddPeer(name)
		if err != nil {
			return fmt.Errorf("failed to add client '%s': %w", name, err)
		}

		if isJSON() {
			printJSON(peer)
			return nil
		}

		printSuccess(fmt.Sprintf("Client '%s' added successfully", name))
		fmt.Fprintf(os.Stderr, "  VPN IP: %s\n", peer.VPNIPv4)
		fmt.Fprintf(os.Stderr, "\nClient configuration:\n")
		fmt.Fprintf(os.Stderr, "────────────────────\n")
		fmt.Fprintln(os.Stdout, clientConfig)
		fmt.Fprintf(os.Stderr, "\nSave this config or use:\n")
		fmt.Fprintf(os.Stderr, "  ngvpn client export %s > %s.conf\n", name, name)
		fmt.Fprintf(os.Stderr, "  ngvpn client qr %s\n", name)
		return nil
	},
}

var clientListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all VPN clients",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		peerList, err := mgr.ListPeers()
		if err != nil {
			return fmt.Errorf("failed to list clients: %w", err)
		}

		if isJSON() {
			printJSON(peerList)
			return nil
		}

		if len(peerList) == 0 {
			fmt.Println("No clients configured.")
			return nil
		}

		rows := make([][]string, len(peerList))
		for i, p := range peerList {
			statusIcon := "●"
			if p.Status == "active" {
				statusIcon = "\033[32m●\033[0m"
			} else if p.Status == "revoked" {
				statusIcon = "\033[31m●\033[0m"
			}
			rows[i] = []string{p.Name, p.VPNIPv4, statusIcon + " " + p.Status, p.CreatedAt.Format("2006-01-02")}
		}
		formatTable([]string{"Name", "VPN IP", "Status", "Created"}, rows)
		return nil
	},
}

var clientShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show details of a VPN client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		peer, err := mgr.GetPeer(name)
		if err != nil {
			return fmt.Errorf("client '%s' not found: %w", name, err)
		}

		if isJSON() {
			printJSON(peer)
			return nil
		}

		fmt.Printf("Client: %s\n", peer.Name)
		fmt.Printf("═══════%s\n", strings.Repeat("═", len(peer.Name)))
		fmt.Printf("ID:         %s\n", peer.ID)
		fmt.Printf("VPN IP:     %s\n", peer.VPNIPv4)
		fmt.Printf("Public Key: %s\n", peer.PublicKey)
		fmt.Printf("Status:     %s\n", peer.Status)
		fmt.Printf("AllowedIPs: %s\n", peer.AllowedIPs)
		fmt.Printf("Keepalive:  %ds\n", peer.Keepalive)
		fmt.Printf("Created:    %s\n", peer.CreatedAt.Format("2006-01-02 15:04:05"))
		return nil
	},
}

var clientRevokeCmd = &cobra.Command{
	Use:   "revoke [name]",
	Short: "Revoke a VPN client's access",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !confirmAction(fmt.Sprintf("Revoke access for client '%s'?", name)) {
			fmt.Println("Cancelled.")
			return nil
		}
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := mgr.RevokePeer(name); err != nil {
			return fmt.Errorf("failed to revoke client '%s': %w", name, err)
		}
		printSuccess(fmt.Sprintf("Client '%s' revoked", name))
		fmt.Println("Restart the server to apply: ngvpn server restart")
		return nil
	},
}

var clientRemoveCmd = &cobra.Command{
	Use:   "remove [name]",
	Short: "Remove a VPN client completely",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !confirmAction(fmt.Sprintf("Remove client '%s' completely? This cannot be undone.", name)) {
			fmt.Println("Cancelled.")
			return nil
		}
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		if err := mgr.RemovePeer(name); err != nil {
			return fmt.Errorf("failed to remove client '%s': %w", name, err)
		}
		printSuccess(fmt.Sprintf("Client '%s' removed", name))
		fmt.Println("Restart the server to apply: ngvpn server restart")
		return nil
	},
}

var clientExportCmd = &cobra.Command{
	Use:   "export [name]",
	Short: "Export WireGuard configuration for a client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		clientConfig, err := mgr.ExportConfig(name)
		if err != nil {
			return fmt.Errorf("failed to export config for '%s': %w", name, err)
		}
		fmt.Print(clientConfig)
		return nil
	},
}

var clientQRCmd = &cobra.Command{
	Use:   "qr [name]",
	Short: "Display QR code for a client's configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		clientConfig, err := mgr.ExportConfig(name)
		if err != nil {
			return fmt.Errorf("failed to export config for '%s': %w", name, err)
		}

		fmt.Fprintf(os.Stderr, "Scan this QR code with the WireGuard mobile app:\n\n")
		qrterminal.GenerateHalfBlock(clientConfig, qrterminal.L, os.Stdout)
		return nil
	},
}

var clientRotateCmd = &cobra.Command{
	Use:   "rotate [name]",
	Short: "Rotate keys for a VPN client",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		if !confirmAction(fmt.Sprintf("Rotate keys for client '%s'? The old configuration will stop working.", name)) {
			fmt.Println("Cancelled.")
			return nil
		}
		mgr, db, err := getPeerManager()
		if err != nil {
			return err
		}
		defer db.Close()

		_, newConfig, err := mgr.RotateKeys(name)
		if err != nil {
			return fmt.Errorf("failed to rotate keys for '%s': %w", name, err)
		}

		printSuccess(fmt.Sprintf("Keys rotated for client '%s'", name))
		fmt.Fprintf(os.Stderr, "\nNew configuration:\n")
		fmt.Fprintf(os.Stderr, "──────────────────\n")
		fmt.Fprintln(os.Stdout, newConfig)
		fmt.Fprintf(os.Stderr, "\nRestart the server to apply: ngvpn server restart\n")
		return nil
	},
}

func init() {
	clientAddCmd.Flags().StringSlice("dns", nil, "DNS servers")
	clientAddCmd.Flags().Int("keepalive", 25, "persistent keepalive interval")
	clientAddCmd.Flags().Bool("full-tunnel", true, "route all traffic through VPN")

	clientCmd.AddCommand(clientAddCmd)
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientShowCmd)
	clientCmd.AddCommand(clientRevokeCmd)
	clientCmd.AddCommand(clientRemoveCmd)
	clientCmd.AddCommand(clientExportCmd)
	clientCmd.AddCommand(clientQRCmd)
	clientCmd.AddCommand(clientRotateCmd)
	rootCmd.AddCommand(clientCmd)
}
