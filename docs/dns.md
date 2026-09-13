# DNS Configuration

Proper DNS configuration is essential for privacy and connectivity when using a VPN. NetGuard allows you to configure which DNS servers your clients use.

## How DNS Works with WireGuard

When a WireGuard client connects, it can optionally override the operating system's default DNS settings. 
This is controlled by the `DNS = ...` line in the client's `.conf` file.

## DNS Modes

### 1. Custom Public DNS (Default behavior)
By default, NetGuard configures clients to use public DNS resolvers (like Cloudflare or Google). 

```yaml
# /etc/netguard/config.yaml
client_defaults:
  dns:
    - "1.1.1.1"
    - "8.8.8.8"
```
- **Pros:** Fast, reliable, prevents DNS leaks (your local ISP won't see your queries).
- **Cons:** Cannot resolve internal hostnames on your private network.

### 2. System DNS (No DNS specified)
If you remove the `dns` section from `client_defaults`, the generated `.conf` files will not include a `DNS` line.
- **Pros:** Best for Split Tunnel setups where you only want to access IPs, not hostnames.
- **Cons:** In a Full Tunnel, this causes a **DNS Leak**. Your web traffic goes over the VPN, but your DNS queries go to your local ISP, revealing which sites you visit.

### 3. Self-Hosted / VPN DNS (Planned)
*Future Feature:* Pointing the DNS to an internal resolver (e.g., Pi-hole or AdGuard Home) running on the VPN server or within the VPN subnet (e.g., `10.77.0.1`).
- **Pros:** Ad-blocking, privacy, ability to resolve internal `.local` hostnames.
- **Cons:** Requires additional setup on the server.

## Preventing DNS Leaks

If you are using a Full Tunnel (`0.0.0.0/0`), you **must** specify a DNS server in your configuration to ensure your queries are encrypted and routed through the VPN. Use public resolvers like `1.1.1.1` (Cloudflare) or `9.9.9.9` (Quad9).

## Testing DNS

Once connected to the VPN, you can verify your DNS is working and not leaking:

**Linux/macOS:**
```bash
# Check which server is resolving
dig example.com
# Alternatively
nslookup example.com
```

**DNS Leak Test Websites:**
Visit a site like [dnsleaktest.com](https://www.dnsleaktest.com/) while connected to the VPN. The results should show the ISP of your chosen DNS provider (e.g., Cloudflare), not your local home or coffee shop ISP.
