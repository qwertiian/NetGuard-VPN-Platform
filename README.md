# NetGuard VPN Platform 🛡️

**A self-hosted, open-source WireGuard VPN management platform.**

NetGuard simplifies setting up and managing your own personal VPN server. It acts as a smart management layer around [WireGuard](https://www.wireguard.com/) — the modern, fast, and secure VPN protocol. 

With NetGuard, you don't need to manually write complex `.conf` files, manage cryptographic keys, or fight with firewall rules. The `ngvpn` CLI handles server configuration, client management, IP allocation, and cross-platform configuration export for you.

---

## 🧠 How It Works (Under the Hood)

When you deploy NetGuard, you are building a **private, encrypted tunnel** between your devices (clients) and your server. Here is exactly what happens:

1. **Traffic Interception:** When you turn on the VPN on your phone/laptop, all your internet traffic is intercepted locally.
2. **Encryption:** Your device encrypts that traffic using your server's **public key**. This encryption is virtually unbreakable (Curve25519).
3. **The Tunnel:** The encrypted data is sent over the public internet to your NetGuard server on port `51820`. Anyone snooping on your Wi-Fi (like at a cafe) only sees an encrypted stream of gibberish.
4. **Decryption & NAT:** Your Linux server receives the data, decrypts it using its **private key**, and forwards the traffic out to the internet (e.g., to `google.com`) using **Network Address Translation (NAT)**. This masks your real IP address.
5. **The Return Trip:** When the website responds, your server encrypts the response and shoots it back through the tunnel to your device.

**The Magic Result:** To the rest of the internet, it looks like you are physically sitting at your server's location, and your local Wi-Fi network has no idea what you are doing online.

---

## 💻 Supported Platforms

It is important to understand the difference between the **Server** (where NetGuard runs) and the **Client** (the devices connecting to it).

| Platform | Server (Runs `ngvpn` CLI) | Client (Connects to the VPN) |
|----------|---------------------------|------------------------------|
| **Ubuntu / Debian / Linux** | ✅ Native Support | ✅ WireGuard App |
| **macOS** | 🐳 Via Docker | ✅ WireGuard App |
| **Windows** | 🐳 Via Docker | ✅ WireGuard App |
| **iOS / Android** | ❌ Cannot host | ✅ WireGuard App (QR Scan) |

*Note: While the NetGuard server relies on Linux networking tools (like `iptables` and `sysctl`), you can run the server on Windows and macOS using **Docker**.*

---

## 🚀 Quick Start Guide (Native Linux)

If you have a Linux machine (Ubuntu, Debian, Fedora, Arch, Raspberry Pi), you can install NetGuard natively.

### 1. Install NetGuard
```bash
git clone https://github.com/netguard-vpn/netguard.git
cd netguard
sudo bash scripts/install.sh
make build && sudo cp bin/ngvpn /usr/local/bin/ngvpn
```

### 2. Initialize the Server
Replace `your-server-ip` with your machine's actual public IP address or a dynamic DNS domain name:
```bash
sudo ngvpn server init --endpoint your-server-ip
```

### 3. Start the VPN Server
```bash
sudo ngvpn server start
```

### 4. Verify Health
Run the built-in diagnostic tool to ensure your IP forwarding and firewalls are configured perfectly:
```bash
sudo ngvpn doctor
```

---

## ☁️ Free Cloud Deployment (GCP / AWS / Oracle)

If you don't want to run the server on your local machine, you can deploy NetGuard permanently for free using major cloud providers. Because NetGuard is extremely lightweight, it easily fits within the "Always Free" tiers of modern cloud hosts.

### Option 1: Google Cloud Platform (Recommended)
1. Sign up for GCP and navigate to **Compute Engine**.
2. Create a new `e2-micro` VM instance. Ensure the region is `us-central1`, `us-east1`, or `us-west1` (these are permanently free).
3. Change the Boot Disk to **Ubuntu 24.04 LTS** and crucially, select **Standard persistent disk** (up to 30GB).
4. Create a GCP **Firewall Rule** to allow incoming `UDP` traffic on port `51820`.
5. SSH into the instance and follow the Native Linux Quick Start guide above.

### Option 2: Oracle Cloud
1. Sign up for Oracle Cloud Free Tier.
2. Create an **Ampere A1 Compute** instance (up to 4 OCPUs and 24GB RAM) running Ubuntu.
3. Edit the Default Security List for your VCN to add an ingress rule for `UDP` port `51820`.
4. SSH into the instance and follow the Native Linux Quick Start guide above.

### Option 3: AWS
*Note: AWS Free Tier lasts for 12 months, not permanently.*
1. Launch an EC2 `t2.micro` or `t3.micro` instance running Ubuntu.
2. Edit the Security Group attached to your instance to allow Custom UDP on port 51820.
3. SSH into the instance and install.

---

## 🐳 Quick Start Guide (Docker - Windows/Mac/Linux)

If you don't want to install dependencies directly on your host OS, or if you are running Windows/Mac, you can run the NetGuard server entirely inside Docker.

### 1. Start the Container
```bash
cd netguard/docker
docker compose up -d
```

### 2. Initialize the Server inside Docker
```bash
docker compose exec netguard ngvpn server init --endpoint your-server-ip
```

### 3. Start the VPN Server inside Docker
```bash
docker compose exec netguard ngvpn server start
```

---

## 📱 Connecting Clients (Phones & Laptops)

Once your server is running, you can add clients (devices) to your VPN.

### 1. Add a New Client
```bash
sudo ngvpn client add my-laptop
```

### 2. Connect a Phone (QR Code)
To connect a mobile device, download the **WireGuard app** from the App Store / Play Store.
Run this command to display a QR code in your terminal:
```bash
sudo ngvpn client qr my-laptop
```
Open the app, tap **+**, select **Create from QR code**, and scan your screen!

### 3. Connect a Laptop (Config File)
To connect a laptop (Windows, Mac, or another Linux machine):
```bash
sudo ngvpn client export my-laptop > my-laptop.conf
```
Transfer the `my-laptop.conf` file to the device. Download the [WireGuard Client](https://www.wireguard.com/install/), click **Import tunnel(s) from file**, select the `.conf` file, and click **Activate**.

---

## 🛠️ CLI Reference

NetGuard features a comprehensive command-line interface. Use `ngvpn --help` to see all options.

```text
ngvpn server init          Initialize VPN server and generate keys
ngvpn server status        Show server and interface status
ngvpn server start         Start WireGuard interface and routing
ngvpn server stop          Stop WireGuard interface
ngvpn server restart       Restart WireGuard interface

ngvpn client add <name>    Add a new VPN client (assigns IP & generates keys)
ngvpn client list          List all registered clients
ngvpn client show <name>   Show client details
ngvpn client revoke <name> Revoke client access without deleting
ngvpn client remove <name> Remove client completely and free up IP
ngvpn client export <name> Export client WireGuard config
ngvpn client qr <name>     Display scannable QR code for mobile
ngvpn client rotate <name> Rotate client keys for security

ngvpn doctor               Run extensive network and system health checks
```

## 🔒 Security

- WireGuard provides authenticated encryption using **Curve25519**, **ChaCha20**, and **Poly1305**.
- Private keys are stored safely on disk with strict `0600` permissions.
- The SQLite database keeps track of allocated IP addresses to prevent IP spoofing or collisions.
- No private keys are ever leaked in standard logs or command output.

## 📝 License

[MIT License](LICENSE) — Copyright (c) 2024 NetGuard Contributors
