# Quick Start Guide

Get your NetGuard VPN up and running in just 5 minutes. This guide covers the fastest path from installation to connecting your first client.

## 1. Install NetGuard

Ensure you have installed NetGuard and the underlying WireGuard tools as detailed in the [Installation Guide](installation.md). 

```bash
# Verify installation
ngvpn --version
```

## 2. Initialize the Server

Initialize the server configuration. This sets up your keys, default subnet (10.77.0.0/24), and database.

```bash
sudo ngvpn server init
```

## 3. Start the Server

Bring up the WireGuard interface and start the NetGuard daemon.

```bash
sudo ngvpn server start
```
*(Note: If you configured the systemd service, you can use `sudo systemctl start netguard` instead).*

## 4. Add a Client

Create a new configuration for your first device (e.g., your laptop).

```bash
sudo ngvpn client add laptop
```

## 5. Export the Client Configuration

Export the generated client configuration to a `.conf` file so you can transfer it to your device.

```bash
sudo ngvpn client export laptop > laptop.conf
```

## 6. (Alternative) Scan QR Code

If you are setting up a mobile device (iOS/Android), it's much easier to scan a QR code directly from your terminal.

```bash
sudo ngvpn client qr laptop
```
This will display a QR code in your terminal window that you can scan with the WireGuard mobile app.

## 7. Import Configuration on Client Device

- **Desktop (Windows/macOS/Linux):** Import the `laptop.conf` file into your WireGuard client application.
- **Mobile (iOS/Android):** Open the WireGuard app, select "Add from QR code," and scan the code generated in the previous step.

## 8. Connect and Verify

1. Activate the tunnel in your WireGuard client.
2. Verify the connection by pinging the NetGuard server's internal VPN IP address (typically `10.77.0.1`):

```bash
ping 10.77.0.1
```

Congratulations! You have successfully set up NetGuard and connected your first device.
