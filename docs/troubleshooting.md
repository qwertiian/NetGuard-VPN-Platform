# Troubleshooting

If you encounter issues while using NetGuard, use this guide to diagnose and resolve the most common problems.

## Diagnostic Commands

The best place to start troubleshooting is the built-in diagnostic tool.

### `ngvpn doctor`
Run this command to check for common misconfigurations and system issues.
```bash
sudo ngvpn doctor
```
It will verify:
- System requirements and dependencies
- File permissions
- Required kernel modules (e.g., `wireguard`)
- Network configuration (IP forwarding, NAT rules)
- DNS resolution

## Common Problems & Solutions

### VPN connects, but there is no internet access
This is the most common issue. The client establishes a connection (handshake), but traffic isn't routed properly.

**Solutions:**
1. **IP Forwarding:** Ensure IPv4 (and IPv6 if used) forwarding is enabled on the server.
   ```bash
   sysctl net.ipv4.ip_forward
   ```
   It should return `1`. If not, enable it in `/etc/sysctl.conf`.
2. **NAT/Masquerade:** Ensure your server's firewall is NATting the VPN traffic out to the internet. NetGuard attempts to configure this automatically, but custom iptables/nftables rules might interfere.
3. **Firewall:** Ensure the firewall allows traffic from the VPN subnet (`10.77.0.0/24`) to the external interface.

### Handshake fails (cannot connect)
The client attempts to connect, but no data is transferred, and the "Latest Handshake" timer does not update.

**Solutions:**
1. **Endpoint/Port:** Ensure the `endpoint` in your client config points to the correct public IP/domain of the server and the correct UDP port (default `51820`).
2. **Firewall:** Ensure UDP port `51820` is open on the server's external firewall (e.g., AWS Security Group, UFW, iptables).
3. **Keys:** Verify that the public key on the server matches the private key on the client, and vice versa. 
4. **Endpoint Network:** Ensure the network the client is on does not block outbound UDP traffic on non-standard ports.

### DNS is not working
You can access internet resources via IP address (e.g., `ping 8.8.8.8`) but not by domain name.

**Solutions:**
1. **Client DNS Config:** Check the `DNS` setting in the client's `.conf` file. Ensure it points to a valid, reachable DNS server (e.g., `1.1.1.1` or the server's internal VPN IP if running a local resolver).
2. **DNS Leaks:** If using full tunnel, ensure the OS is respecting the WireGuard DNS settings. See [DNS Configuration](dns.md) for more details.

### Slow speeds
Connection is established, but throughput is lower than expected.

**Solutions:**
1. **MTU Mismatch:** A common cause of poor performance is MTU issues. Try lowering the MTU in the NetGuard config (e.g., from `1420` to `1360` or `1280`) and recreating the client configs.
2. **Server Load:** Check the CPU and network load on your VPS. WireGuard is very efficient, but cheap VPS instances might have restrictive network throttling.

### Permission Denied
You see "permission denied" errors when running `ngvpn` commands.

**Solutions:**
1. Ensure you are running commands with `sudo` or as the `root` user, as NetGuard manages network interfaces and reads sensitive key files.
2. Check the permissions of `/etc/netguard/config.yaml`. It should be owned by `root` and readable only by `root` (`chmod 600`).

## Reading WireGuard Logs

To see raw WireGuard events, you can enable dynamic debugging in the kernel (if supported by your OS):

```bash
echo module wireguard +p > /sys/kernel/debug/dynamic_debug/control
dmesg -wT | grep wireguard
```

## Getting Help / Bug Reports

If you've exhausted this guide and need to report a bug, please include the output of the following command in your issue report (redact any sensitive public IPs or keys if they appear):

```bash
sudo ngvpn doctor --json
```
