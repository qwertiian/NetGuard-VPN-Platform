# Uninstalling NetGuard

If you need to completely remove NetGuard from your system, follow these steps to ensure a clean removal.

## Clean Removal Process

### 1. Stop Services

First, stop the NetGuard daemon and bring down the WireGuard interface. This will disconnect any active clients.

```bash
sudo systemctl stop netguard
sudo systemctl disable netguard
```

Alternatively, if not using systemd:
```bash
sudo ngvpn server stop
```

### 2. Remove Firewall Rules

If NetGuard was managing iptables/nftables rules, stopping the server should clean them up automatically. However, if you added custom UFW rules, you should remove them:

```bash
# Example if you used UFW
sudo ufw delete allow 51820/udp
```

### 3. Remove Configuration and Database

**WARNING: This step is irreversible and will delete all client configurations, keys, and the server identity.**

Remove the configuration directory and database:

```bash
sudo rm -rf /etc/netguard
sudo rm -rf /var/lib/netguard
```

### 4. Remove the Binary

Remove the `ngvpn` executable:

```bash
sudo rm /usr/local/bin/ngvpn
```

### 5. Remove Systemd Service

Remove the service file and reload the systemd daemon:

```bash
sudo rm /etc/systemd/system/netguard.service
sudo systemctl daemon-reload
```

## Using the Uninstall Script

If you installed NetGuard from the source repository, you can use the provided uninstall script (if available) or the makefile target:

```bash
cd netguard
sudo make uninstall
```

## What is NOT Removed

Uninstalling NetGuard from your server does **not** automatically remove the client configurations (`.conf` files) from your users' devices (laptops, phones). 

Those devices will simply fail to connect until the tunnels are manually deleted within the WireGuard client applications on each respective device.
