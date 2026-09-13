package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/netguard-vpn/netguard/internal/config"
	"github.com/netguard-vpn/netguard/internal/firewall"
	"github.com/netguard-vpn/netguard/internal/security"
	"github.com/netguard-vpn/netguard/internal/storage"
	"github.com/netguard-vpn/netguard/internal/vpn"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage the VPN server",
}

var serverInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize VPN server configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		endpoint, _ := cmd.Flags().GetString("endpoint")
		port, _ := cmd.Flags().GetInt("port")
		subnet, _ := cmd.Flags().GetString("subnet")
		iface, _ := cmd.Flags().GetString("interface")
		dnsServers, _ := cmd.Flags().GetStringSlice("dns")

		if endpoint == "" {
			return fmt.Errorf("--endpoint is required\n\nSpecify your server's public IP or domain name:\n  ngvpn server init --endpoint vpn.example.com")
		}

		cfg := config.DefaultConfig()
		cfg.Server.Endpoint = endpoint
		cfg.Server.ListenPort = port
		cfg.Server.Interface = iface
		cfg.Network.IPv4Subnet = subnet
		if len(dnsServers) > 0 {
			cfg.DNS.Servers = dnsServers
		}

		if err := cfg.Validate(); err != nil {
			return fmt.Errorf("invalid configuration: %w", err)
		}

		for _, dir := range []string{cfg.Storage.DataDirectory, cfg.Security.KeyDirectory, filepath.Join(cfg.Storage.DataDirectory, "backups")} {
			if err := security.SecureDirectory(dir); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			printSuccess(fmt.Sprintf("Created directory: %s", dir))
		}

		serverKeyPath := filepath.Join(cfg.Security.KeyDirectory, "server.key")
		serverPubPath := filepath.Join(cfg.Security.KeyDirectory, "server.pub")

		if _, err := os.Stat(serverKeyPath); err == nil {
			printWarning("Server keys already exist, skipping key generation")
		} else {
			kp, err := security.GenerateKeyPair()
			if err != nil {
				return fmt.Errorf("failed to generate server keys: %w", err)
			}
			if err := security.SaveKeyToFile(kp.PrivateKey, serverKeyPath); err != nil {
				return fmt.Errorf("failed to save server private key: %w", err)
			}
			if err := os.WriteFile(serverPubPath, []byte(kp.PublicKey+"\n"), 0644); err != nil {
				return fmt.Errorf("failed to save server public key: %w", err)
			}
			printSuccess("Generated server key pair")
		}

		configPath := filepath.Join(cfg.Storage.DataDirectory, "config.yaml")
		if err := cfg.Save(configPath); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}
		printSuccess(fmt.Sprintf("Saved configuration: %s", configPath))

		db, err := storage.Open(cfg.Storage.Database)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}
		defer db.Close()
		if err := db.Initialize(); err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}
		printSuccess("Initialized database")

		extIface, err := firewall.DetectDefaultInterface()
		if err != nil {
			printWarning(fmt.Sprintf("Could not detect default interface: %v", err))
			extIface = "eth0"
		}
		printSuccess(fmt.Sprintf("Detected default interface: %s", extIface))

		privKey, err := security.LoadKeyFromFile(serverKeyPath)
		if err != nil {
			return fmt.Errorf("failed to read server private key: %w", err)
		}

		params := vpn.ServerConfigParams{
			PrivateKey:   privKey,
			Address:      cfg.Network.ServerAddress + "/24",
			ListenPort:   cfg.Server.ListenPort,
			ExtInterface: extIface,
			Subnet:       cfg.Network.IPv4Subnet,
			Peers:        []vpn.PeerConfigParams{},
		}
		wgConfig := vpn.GenerateServerConfigFile(params)
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")
		wgConfigPath := fmt.Sprintf("/etc/wireguard/%s.conf", cfg.Server.Interface)
		if err := mgr.WriteServerConfig(wgConfig, wgConfigPath); err != nil {
			return fmt.Errorf("failed to write WireGuard config: %w", err)
		}
		printSuccess(fmt.Sprintf("Generated WireGuard config: %s", wgConfigPath))

		if err := mgr.EnableIPForwarding(); err != nil {
			printWarning(fmt.Sprintf("Could not enable IP forwarding: %v", err))
		} else {
			printSuccess("Enabled IPv4 forwarding")
		}

		db.SetState("initialized", "true")
		db.SetState("endpoint", endpoint)
		db.SetState("ext_interface", extIface)

		fmt.Fprintf(os.Stderr, "\n\033[32mServer initialized successfully!\033[0m\n\n")
		fmt.Fprintf(os.Stderr, "Next steps:\n")
		fmt.Fprintf(os.Stderr, "  1. Start the server:    ngvpn server start\n")
		fmt.Fprintf(os.Stderr, "  2. Add a client:        ngvpn client add <name>\n")
		fmt.Fprintf(os.Stderr, "  3. Export client config: ngvpn client export <name>\n")
		fmt.Fprintf(os.Stderr, "  4. Check health:        ngvpn doctor\n")
		return nil
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show VPN server status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")
		running, _ := mgr.IsRunning()

		if isJSON() {
			status := map[string]interface{}{"running": running, "interface": cfg.Server.Interface}
			if running {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if s, err := mgr.GetStatus(ctx); err == nil {
					status["public_key"] = s.PublicKey
					status["listen_port"] = s.ListenPort
					status["peers"] = len(s.Peers)
				}
			}
			printJSON(status)
			return nil
		}

		fmt.Println("NetGuard VPN Server")
		fmt.Println("═══════════════════")
		if running {
			fmt.Println("Status:    \033[32mONLINE\033[0m")
			fmt.Printf("Interface: %s\n", cfg.Server.Interface)
			fmt.Printf("Port:      %d/UDP\n", cfg.Server.ListenPort)
			fmt.Printf("Subnet:    %s\n", cfg.Network.IPv4Subnet)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if s, err := mgr.GetStatus(ctx); err == nil {
				fmt.Printf("Public Key: %s\n", s.PublicKey)
				fmt.Printf("Peers:     %d connected\n", len(s.Peers))
				var totalRx, totalTx int64
				for _, p := range s.Peers {
					totalRx += p.TransferRx
					totalTx += p.TransferTx
				}
				fmt.Printf("Traffic:   ↓ %s  ↑ %s\n", formatBytes(totalRx), formatBytes(totalTx))
			}
		} else {
			fmt.Println("Status:    \033[31mOFFLINE\033[0m")
			fmt.Println("\nStart with: ngvpn server start")
		}
		return nil
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start WireGuard VPN interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mgr.Start(ctx); err != nil {
			return fmt.Errorf("failed to start VPN: %w\n\nCheck:\n  - WireGuard installed: wg --version\n  - Config exists: /etc/wireguard/%s.conf\n  - Run with sudo", err, cfg.Server.Interface)
		}
		printSuccess(fmt.Sprintf("VPN interface %s started", cfg.Server.Interface))
		return nil
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop WireGuard VPN interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := mgr.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop VPN: %w", err)
		}
		printSuccess(fmt.Sprintf("VPN interface %s stopped", cfg.Server.Interface))
		return nil
	},
}

var serverRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart WireGuard VPN interface",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		mgr := vpn.NewManager(cfg.Server.Interface, "/etc/wireguard")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := mgr.Restart(ctx); err != nil {
			return fmt.Errorf("failed to restart VPN: %w", err)
		}
		printSuccess(fmt.Sprintf("VPN interface %s restarted", cfg.Server.Interface))
		return nil
	},
}

func init() {
	serverInitCmd.Flags().String("endpoint", "", "server public IP or domain (required)")
	serverInitCmd.Flags().Int("port", 51820, "WireGuard listen port")
	serverInitCmd.Flags().String("subnet", "10.77.0.0/24", "VPN subnet")
	serverInitCmd.Flags().String("interface", "wg0", "WireGuard interface name")
	serverInitCmd.Flags().StringSlice("dns", []string{"1.1.1.1", "9.9.9.9"}, "DNS servers")

	serverCmd.AddCommand(serverInitCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
	serverCmd.AddCommand(serverRestartCmd)
	rootCmd.AddCommand(serverCmd)
}
