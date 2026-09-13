# macOS Client Setup

This guide walks you through connecting a macOS device (MacBook, iMac, Mac mini) to your NetGuard VPN.

## 1. Install WireGuard

There are two primary ways to install WireGuard on macOS:

**Option A: Mac App Store (Recommended)**
This is the easiest method and provides a graphical user interface.
1. Open the **App Store**.
2. Search for "WireGuard".
3. Install the official WireGuard app by WireGuard LLC.

**Option B: Homebrew (CLI)**
If you prefer the command line, you can install the CLI tools using Homebrew.
```bash
brew install wireguard-tools
```

## 2. Generate Client Configuration

On your NetGuard server, generate a configuration for your Mac:

```bash
sudo ngvpn client add my-macbook
sudo ngvpn client export my-macbook > my-macbook.conf
```

Transfer `my-macbook.conf` securely to your Mac.

## 3. Import and Connect

### Using the GUI App (App Store)

1. Open the WireGuard application. You will see a WireGuard icon in your macOS menu bar (top right).
2. Click the icon and select **"Manage Tunnels"** or click the **"+"** button at the bottom left of the WireGuard window.
3. Select **"Import tunnel(s) from file"**.
4. Choose the `my-macbook.conf` file. You may be prompted to allow WireGuard to add VPN configurations; authenticate with your Touch ID or password.
5. To connect, simply click **"Activate"** in the app, or use the menu bar icon to toggle the connection.

### Using the CLI (Homebrew)

If you installed via Homebrew, the process is similar to Linux:

1. Move the config to a secure location (e.g., `/etc/wireguard/`, though you will need to create this directory).
   ```bash
   sudo mkdir -p /etc/wireguard
   sudo mv my-macbook.conf /etc/wireguard/wg0.conf
   sudo chmod 600 /etc/wireguard/wg0.conf
   ```
2. Start the tunnel:
   ```bash
   sudo wg-quick up wg0
   ```
3. Stop the tunnel:
   ```bash
   sudo wg-quick down wg0
   ```

## 4. Verify the Connection

Open Terminal and verify your traffic is routing correctly:

```bash
# Check routing
curl ifconfig.me
```
The output should match your NetGuard server's public IP.

## Troubleshooting macOS

- **App Store version doesn't connect:** Ensure you have granted WireGuard permission to create VPN connections in **System Settings > Network**.
- **DNS Issues:** If you can ping IPs but not domains, check the `DNS` setting in your tunnel configuration within the WireGuard app. macOS can sometimes be strict about DNS routing. Try using `1.1.1.1` in the DNS field if you are having issues.
