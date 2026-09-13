package commands

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/netguard-vpn/netguard/internal/backup"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Manage configuration backups",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup of VPN configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		output, _ := cmd.Flags().GetString("output")

		if output == "" {
			timestamp := time.Now().Format("20060102-150405")
			output = filepath.Join(cfg.Storage.DataDirectory, "backups", fmt.Sprintf("netguard-backup-%s.tar.gz", timestamp))
		}

		printWarning("Backups contain sensitive key material. Store securely.")

		mgr := backup.NewManager(cfg.Storage.DataDirectory, cfg.Storage.Database)
		if err := mgr.CreateBackup(output); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}

		printSuccess(fmt.Sprintf("Backup created: %s", output))
		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore [file]",
	Short: "Restore VPN configuration from backup",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		archivePath := args[0]
		cfg := getConfig()

		if !confirmAction("Restore from backup? This will overwrite current configuration.") {
			fmt.Println("Cancelled.")
			return nil
		}

		mgr := backup.NewManager(cfg.Storage.DataDirectory, cfg.Storage.Database)
		if err := mgr.RestoreBackup(archivePath); err != nil {
			return fmt.Errorf("failed to restore backup: %w", err)
		}

		printSuccess("Backup restored successfully")
		fmt.Println("Restart the server to apply: ngvpn server restart")
		return nil
	},
}

func init() {
	backupCreateCmd.Flags().StringP("output", "o", "", "output file path")

	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	rootCmd.AddCommand(backupCmd)
}
