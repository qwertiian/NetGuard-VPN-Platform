package commands

import (
	"fmt"
	"os"

	"github.com/netguard-vpn/netguard/internal/doctor"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run system health checks",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := getConfig()
		docCfg := &doctor.Config{
			Interface: cfg.Server.Interface,
			ConfigDir: "/etc/wireguard",
			KeyDir:    cfg.Security.KeyDirectory,
			Subnet:    cfg.Network.IPv4Subnet,
			DBPath:    cfg.Storage.Database,
		}

		if isJSON() {
			report := doctor.RunDiagnostics(docCfg)
			printJSON(report)
			if !report.Healthy {
				os.Exit(1)
			}
			return nil
		}

		fmt.Println("NetGuard VPN Doctor")
		fmt.Println("═══════════════════")
		fmt.Println()

		report := doctor.RunDiagnostics(docCfg)

		for _, check := range report.Checks {
			var icon string
			switch check.Status {
			case "pass":
				icon = "\033[32m✓\033[0m"
			case "fail":
				icon = "\033[31m✗\033[0m"
			case "warn":
				icon = "\033[33m⚠\033[0m"
			case "skip":
				icon = "\033[90m○\033[0m"
			}
			fmt.Printf("%s %s\n", icon, check.Name)
			if check.Message != "" && check.Status != "pass" {
				fmt.Printf("  %s\n", check.Message)
			}
			if check.Remediation != "" && check.Status == "fail" {
				fmt.Printf("  Fix: %s\n", check.Remediation)
			}
		}

		fmt.Println()
		if report.Healthy {
			fmt.Println("\033[32mResult: HEALTHY\033[0m")
		} else {
			fmt.Println("\033[31mResult: UNHEALTHY\033[0m")
			fmt.Println("\nReview the issues above and run 'ngvpn doctor' again after fixing.")
			os.Exit(1)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}
