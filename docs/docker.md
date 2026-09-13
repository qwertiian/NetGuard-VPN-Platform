# Docker Deployment Guide

NetGuard can be deployed using Docker, which is useful for keeping your host system clean or integrating into existing containerized infrastructure.

## Requirements

- Docker installed on the host.
- Docker Compose (v2 recommended).
- The host must have the WireGuard kernel module loaded (`modprobe wireguard`).

## `docker-compose.yml`

Create a `docker-compose.yml` file with the following configuration:

```yaml
version: '3.8'
services:
  netguard:
    image: ghcr.io/netguard-vpn/netguard:latest
    container_name: netguard
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    environment:
      # Optional: Override config values via ENV vars
      - NG_SERVER_ENDPOINT=vpn.example.com
      - NG_SERVER_PORT=51820
    volumes:
      # Mount configuration and database directory
      - ./netguard-data:/etc/netguard
      # Mount lib/modules to allow loading kernel modules if needed
      - /lib/modules:/lib/modules:ro
    ports:
      # Map the WireGuard UDP port
      - "51820:51820/udp"
    sysctls:
      - net.ipv4.ip_forward=1
      - net.ipv4.conf.all.src_valid_mark=1
    restart: unless-stopped
```

## Required Capabilities

WireGuard creates network interfaces and manipulates routing tables. Therefore, the container requires the `NET_ADMIN` capability. 
`SYS_MODULE` is also recommended to ensure the container can interact with the host's WireGuard kernel module.

## Volume Mounts

You **must** mount a volume to `/etc/netguard` to persist your configuration, keys, and SQLite database. If you do not do this, your VPN will reset every time the container restarts, and clients will lose connectivity.

In the example above, data is stored in the `./netguard-data` directory on the host.

## Networking Considerations

When running in Docker, NetGuard operates inside its own network namespace. 

- Port `51820/UDP` must be published to the host (`ports` section) so external clients can reach it.
- The `sysctls` section is crucial. It enables IP forwarding inside the container, allowing traffic to flow from the WireGuard interface out to the internet.

## Initializing via Docker

To initialize the server for the first time using Docker:

```bash
docker compose run --rm netguard ngvpn server init
```

Then, start the service:

```bash
docker compose up -d
```

## Managing Clients via Docker

To run CLI commands, execute them inside the running container:

```bash
# Add a client
docker compose exec netguard ngvpn client add my-laptop

# Export a client config
docker compose exec netguard ngvpn client export my-laptop > my-laptop.conf

# Show QR code
docker compose exec netguard ngvpn client qr my-phone
```

## Limitations and Caveats

- **Host Kernel Module:** The Docker container relies on the host OS having WireGuard support in the kernel. This is standard in modern Linux kernels (5.6+), but older systems may need it installed manually.
- **Firewall Integration:** NetGuard attempts to manage iptables rules automatically. In a Docker environment, these rules apply inside the container. Ensure your host's firewall allows the published Docker port.
