# Split Tunneling

NetGuard supports both Full Tunnel and Split Tunnel configurations. This document explains the difference and how to configure them.

## Full Tunnel (Default)

In a **Full Tunnel** setup, *all* network traffic from the client device is routed through the VPN.

- **Config Value:** `AllowedIPs = 0.0.0.0/0, ::/0`
- **Use Case:** You are on an untrusted network (like public Wi-Fi) and want to secure all your browsing, masking your IP address from the websites you visit.
- **Traffic Flow:** Client -> VPN Server -> Internet.

## Split Tunnel

In a **Split Tunnel** setup, only traffic destined for specific IP addresses or subnets is routed through the VPN. All other traffic goes through your normal, local internet connection.

- **Config Value:** e.g., `AllowedIPs = 10.77.0.0/24, 192.168.1.0/24`
- **Use Case:** You only need to access internal resources hosted behind the VPN server (like a private web app or a file server), but you want your regular web browsing (Netflix, YouTube) to go directly to the internet for better speeds and less load on the VPN server.
- **Traffic Flow (VPN subnets):** Client -> VPN Server -> Internal Network.
- **Traffic Flow (Everything else):** Client -> Local ISP -> Internet.

## How to Configure in NetGuard

NetGuard controls this behavior via the `client_defaults.allowed_ips` setting in `/etc/netguard/config.yaml`.

### Example 1: Full Tunnel (Default)
```yaml
client_defaults:
  allowed_ips:
    - "0.0.0.0/0"
```
Any client created with `ngvpn client add` while this config is active will route everything through the VPN.

### Example 2: Split Tunnel (VPN Subnet Only)
If you only want clients to access other peers on the VPN network, restrict the `AllowedIPs` to the VPN subnet:
```yaml
client_defaults:
  allowed_ips:
    - "10.77.0.0/24"
```

### Example 3: Split Tunnel (VPN + Remote LAN)
If your VPN server sits on a private network (e.g., `192.168.100.0/24`) and you want clients to access it:
```yaml
client_defaults:
  allowed_ips:
    - "10.77.0.0/24"
    - "192.168.100.0/24"
```

*Note: Changes to `client_defaults` only affect newly created clients. Existing clients must have their `.conf` files manually updated on their devices.*

## DNS Considerations

DNS behaves differently depending on the tunnel mode:

- **Full Tunnel:** It is highly recommended to configure DNS (e.g., `1.1.1.1`) in the client config to prevent DNS leaks. The client will send all DNS queries through the tunnel.
- **Split Tunnel:** If you define a DNS server in a split tunnel config, most operating systems will route *all* DNS queries to that server through the VPN, even for traffic that isn't routed through the VPN. If you don't define a DNS server, the client will use its local network's DNS. If you are trying to resolve internal hostnames on the remote network, you must configure a DNS server and accept that all DNS traffic goes over the VPN.
