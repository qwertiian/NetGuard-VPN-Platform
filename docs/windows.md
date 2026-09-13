# Windows Client Setup

This guide walks you through connecting a Windows PC to your NetGuard VPN.

## 1. Download WireGuard

1. Go to the official WireGuard website: [https://www.wireguard.com/install/](https://www.wireguard.com/install/)
2. Download the Windows Installer.
3. Run the installer to install the WireGuard client.

## 2. Generate Client Configuration

On your NetGuard server, generate a configuration for your Windows machine if you haven't already:

```bash
sudo ngvpn client add my-windows-pc
sudo ngvpn client export my-windows-pc > my-windows-pc.conf
```

Securely transfer `my-windows-pc.conf` to your Windows machine (e.g., via SFTP, secure email, or a password manager).

## 3. Import the Configuration

1. Open the WireGuard application on Windows.
2. Click the **"Import tunnel(s) from file"** button.
3. Select the `my-windows-pc.conf` file you transferred.

*Alternatively, if your laptop has a webcam, you can generate a QR code on the server (`sudo ngvpn client qr my-windows-pc`) and use a QR scanning app to grab the text, though file transfer is usually easier for desktop.*

## 4. Connect and Disconnect

- To connect, select the newly imported tunnel in the WireGuard list and click **"Activate"**.
- The status should change to "Active," and you will see data being sent and received.
- To disconnect, click **"Deactivate"**.

## 5. Verify the Connection

Open PowerShell or Command Prompt to verify your traffic is routing correctly.

1. **Ping the VPN Server:**
   ```powershell
   ping 10.77.0.1
   ```
2. **Check your public IP:**
   ```powershell
   Invoke-RestMethod -Uri "https://ifconfig.me"
   ```
   The IP address returned should be the public IP of your NetGuard server, not your local internet connection.

## Troubleshooting Windows

- **"Unable to create Wintun interface":** This usually indicates a driver conflict or permission issue. Try running WireGuard as Administrator or reinstalling the client.
- **Connected, but no internet:** Ensure that the "Block untunneled traffic (kill switch)" option is unchecked if you are on a network that requires local access (like a hotel captive portal) before establishing the full tunnel. Check your server-side firewall rules as described in the [Troubleshooting Guide](troubleshooting.md).
