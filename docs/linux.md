# Linux Client Setup

This guide covers connecting a Linux machine (desktop or server) to your NetGuard VPN using the standard `wireguard-tools`.

## 1. Install WireGuard Tools

Install the necessary packages using your distribution's package manager.

**Ubuntu / Debian:**
```bash
sudo apt update
sudo apt install wireguard-tools
```

**Fedora:**
```bash
sudo dnf install wireguard-tools
```

**Arch Linux:**
```bash
sudo pacman -S wireguard-tools
```

## 2. Generate and Transfer Configuration

On your NetGuard server, generate a configuration for your Linux client:

```bash
sudo ngvpn client add my-linux-client
sudo ngvpn client export my-linux-client > wg0.conf
```

Securely transfer `wg0.conf` to the Linux client machine.

## 3. Setup the Configuration

On the Linux client machine, move the configuration file to the standard WireGuard directory and set the correct permissions:

```bash
sudo mv wg0.conf /etc/wireguard/wg0.conf
sudo chown root:root /etc/wireguard/wg0.conf
sudo chmod 600 /etc/wireguard/wg0.conf
```

## 4. Connect and Disconnect

You will use the `wg-quick` command to manage the tunnel.

**Start the connection:**
```bash
sudo wg-quick up wg0
```

**Stop the connection:**
```bash
sudo wg-quick down wg0
```

## 5. Enable Auto-Start on Boot (Optional)

If you want the VPN to connect automatically whenever the Linux machine boots, enable the systemd service:

```bash
sudo systemctl enable wg-quick@wg0
sudo systemctl start wg-quick@wg0
```

## 6. Verify the Connection

Check the status of the WireGuard interface:

```bash
sudo wg show
```
This should display the interface details, the peer (your NetGuard server), and the "latest handshake" time.

Verify your routing by checking your public IP:
```bash
curl ifconfig.me
```
It should return the IP address of your NetGuard server.

## Troubleshooting Linux

- **`wg-quick: command not found`:** Ensure `wireguard-tools` is installed and in your PATH.
- **Resolvconf errors:** If `wg-quick up` fails with errors related to DNS or `resolvconf`, ensure you have a tool like `systemd-resolved`, `resolvconf`, or `openresolv` installed to handle DNS changes.
