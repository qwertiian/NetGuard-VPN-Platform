package main

import (
    "os"
    "github.com/netguard-vpn/netguard/cmd/ngvpn/commands"
)

var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    commands.SetVersionInfo(version, commit, date)
    if err := commands.Execute(); err != nil {
        os.Exit(1)
    }
}
