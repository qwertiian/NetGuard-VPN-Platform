package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/netguard-vpn/netguard/internal/config"
	"github.com/netguard-vpn/netguard/internal/logging"
	"github.com/spf13/cobra"
)

var (
	versionStr string
	commitStr  string
	dateStr    string

	cfgFile    string
	verbose    bool
	jsonOutput bool
	yesFlag    bool

	loadedConfig *config.Config
)

// SetVersionInfo sets version information from ldflags.
func SetVersionInfo(version, commit, date string) {
	versionStr = version
	commitStr = commit
	dateStr = date
}

var rootCmd = &cobra.Command{
	Use:   "ngvpn",
	Short: "NetGuard - Self-hosted WireGuard VPN management",
	Long: `NetGuard is a self-hosted, open-source WireGuard VPN management platform.

It provides simple CLI tools to set up a VPN server, manage clients,
and generate standard WireGuard configurations for any platform.`,
	Version: "dev",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		level := "info"
		if verbose {
			level = "debug"
		}
		logging.Init(level)

		if cfgFile != "" {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				return fmt.Errorf("failed to load config %s: %w", cfgFile, err)
			}
			loadedConfig = cfg
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().BoolVarP(&yesFlag, "yes", "y", false, "skip confirmation prompts")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func getConfig() *config.Config {
	if loadedConfig != nil {
		return loadedConfig
	}
	// Try to find and load config
	path, err := config.FindConfigFile()
	if err == nil {
		cfg, err := config.Load(path)
		if err == nil {
			loadedConfig = cfg
			return cfg
		}
	}
	return config.DefaultConfig()
}

func isJSON() bool {
	return jsonOutput
}

func printJSON(v interface{}) {
	data, _ := json.MarshalIndent(v, "", "  ")
	fmt.Fprintln(os.Stdout, string(data))
}

func printSuccess(msg string) {
	fmt.Fprintf(os.Stderr, "\033[32m✓\033[0m %s\n", msg)
}

func printError(msg string) {
	fmt.Fprintf(os.Stderr, "\033[31m✗\033[0m %s\n", msg)
}

func printWarning(msg string) {
	fmt.Fprintf(os.Stderr, "\033[33m⚠\033[0m %s\n", msg)
}

func confirmAction(prompt string) bool {
	if yesFlag {
		return true
	}
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(strings.ToLower(response))
	return response == "y" || response == "yes"
}

func formatTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	for i, h := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Printf("%-*s", widths[i], strings.ToUpper(h))
	}
	fmt.Println()

	for i := range headers {
		if i > 0 {
			fmt.Print("  ")
		}
		fmt.Print(strings.Repeat("─", widths[i]))
	}
	fmt.Println()

	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				fmt.Print("  ")
			}
			if i < len(widths) {
				fmt.Printf("%-*s", widths[i], cell)
			}
		}
		fmt.Println()
	}
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds ago", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}
