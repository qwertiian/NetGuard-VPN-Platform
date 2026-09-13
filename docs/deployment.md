# Production Deployment Guide

Deploying NetGuard in a production environment requires attention to security, stability, and maintenance. This guide outlines best practices for a robust deployment.

## VPS Requirements

WireGuard is extremely lightweight. NetGuard can run comfortably on very small instances.

- **Minimum:** 1 vCPU, 512MB RAM, 5GB Disk
- **Recommended OS:** Ubuntu 22.04 LTS or Debian 12
- **Network:** A static public IPv4 address. High bandwidth allowance is recommended depending on your usage.

## Cloud Provider Options

NetGuard works on almost any VPS provider:
- DigitalOcean (Droplets)
- Linode / Akamai
- Hetzner Cloud (Very cost-effective)
- AWS EC2 (Ensure Security Group allows UDP 51820)
- Google Cloud Compute

## Security Hardening Checklist

Before exposing your server to the internet:

1. **SSH Security:**
   - Disable root login (`PermitRootLogin no`).
   - Disable password authentication; use SSH keys only (`PasswordAuthentication no`).
   - Change the default SSH port (optional but reduces log spam).
2. **Firewall (UFW/iptables):**
   - Deny all incoming traffic by default.
   - Allow your SSH port (e.g., 22/TCP).
   - Allow the WireGuard port (e.g., 51820/UDP).
   ```bash
   sudo ufw default deny incoming
   sudo ufw default allow outgoing
   sudo ufw allow 22/tcp
   sudo ufw allow 51820/udp
   sudo ufw enable
   ```
3. **Automatic Updates:**
   - Enable `unattended-upgrades` on Debian/Ubuntu to automatically apply security patches.

## Systemd Service Management

Always manage the NetGuard daemon via systemd (as covered in the [Installation Guide](installation.md)). This ensures it restarts on crash and starts on boot.

```bash
sudo systemctl enable netguard
sudo systemctl start netguard
```

## Log Management

By default, NetGuard logs to stdout (captured by systemd) or a file specified in `config.yaml`.
If logging to a file (`/var/log/netguard.log`), set up `logrotate` to prevent the disk from filling up.

Create `/etc/logrotate.d/netguard`:
```text
/var/log/netguard.log {
    daily
    rotate 7
    compress
    missingok
    notifempty
}
```

## Backup Schedule

Automate backups of your configuration and database. You can use cron to run the backup command daily.

```bash
# Add to root's crontab (sudo crontab -e)
0 2 * * * /usr/local/bin/ngvpn backup create /var/backups/netguard/backup-$(date +\%F).tar.gz
```
*Note: Ensure you securely transfer these backups off the server.*

## Update Process

To update NetGuard in production:

1. Backup your configuration (`ngvpn backup create ...`).
2. Stop the service: `sudo systemctl stop netguard`.
3. Download or compile the new `ngvpn` binary.
4. Replace the old binary in `/usr/local/bin/`.
5. Start the service: `sudo systemctl start netguard`.
6. Verify clients can still connect.
