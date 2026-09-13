# Installation Guide

Welcome to the NetGuard VPN installation guide. NetGuard is an open-source, self-hosted WireGuard VPN management platform. This guide will walk you through setting up NetGuard on a Linux server.

## Prerequisites

Before you begin, ensure you have the following:

- **OS:** A Linux server (Ubuntu 22.04+, Debian 11+, Fedora 38+, or similar).
- **Access:** Root or `sudo` access to the server.
- **Dependencies:** 
  - Git
  - Go 1.22 or higher (for compiling from source)
  - `make` (optional, for convenience)
- **Network:** A public IP address and an open UDP port (default is 51820).

## Quick Install Method

The quickest way to install NetGuard is by compiling it from source.

```bash
# Clone the repository
git clone https://github.com/netguard-vpn/netguard.git
cd netguard

# Build the binary
make build

# Install to /usr/local/bin
sudo make install
```

## Package Manager Installation (Planned)

In the future, NetGuard will be available via native package managers (APT, RPM, etc.). This feature is currently **Planned**.

## Verify Installation

Once installed, verify that the `ngvpn` binary is available in your PATH:

```bash
ngvpn --version
```
You should see output similar to `NetGuard VPN v1.0.0`.

## Install WireGuard Tools

NetGuard manages WireGuard under the hood. You must install the core WireGuard tools on your system.

### Ubuntu / Debian

```bash
sudo apt update
sudo apt install wireguard wireguard-tools
```

### Fedora

```bash
sudo dnf install wireguard-tools
```

## Post-Installation Setup

After installing the binary and WireGuard tools, initialize your NetGuard server configuration.

```bash
sudo ngvpn server init
```

This command will:
1. Generate the necessary server private and public keys.
2. Create the default configuration file at `/etc/netguard/config.yaml`.
3. Set up the local database for managing clients.

## Systemd Service Setup

To ensure NetGuard starts automatically on boot and runs in the background, you can set it up as a systemd service.

Create a service file at `/etc/systemd/system/netguard.service`:

```ini
[Unit]
Description=NetGuard VPN Server
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/ngvpn server start
Restart=on-failure
User=root

[Install]
WantedBy=multi-user.target
```

Enable and start the service:

```bash
sudo systemctl enable netguard
sudo systemctl start netguard
```

Check the status to ensure it's running smoothly:

```bash
sudo systemctl status netguard
```
