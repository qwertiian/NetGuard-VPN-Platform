# Configuration Reference

NetGuard is primarily configured via a YAML configuration file. This document details all available configuration options.

## Configuration File Location

The default location for the configuration file is `/etc/netguard/config.yaml`. 
You can specify a custom location using the `--config` flag with the `ngvpn` CLI.

## Complete Configuration Example

Below is a complete example of a `config.yaml` file with comments explaining each field.

```yaml
# /etc/netguard/config.yaml

server:
  # The public IP address or domain name of your server. 
  # Clients use this to connect.
  endpoint: "vpn.example.com"
  
  # The UDP port WireGuard will listen on.
  port: 51820
  
  # The WireGuard interface name.
  interface: "wg0"
  
  # The private key for the server. KEEP THIS SECRET.
  # Automatically generated during 'ngvpn server init'.
  private_key: "<SERVER_PRIVATE_KEY>"
  
  # The IP subnet to use for the VPN network.
  subnet: "10.77.0.0/24"
  
  # The MTU (Maximum Transmission Unit) for the WireGuard interface.
  mtu: 1420

client_defaults:
  # Default DNS servers assigned to clients.
  dns:
    - "1.1.1.1"
    - "8.8.8.8"
  
  # Default AllowedIPs for clients. 
  # 0.0.0.0/0 means route all traffic through the VPN (Full Tunnel).
  allowed_ips:
    - "0.0.0.0/0"
  
  # Keepalive interval in seconds to keep NAT mappings active.
  persistent_keepalive: 25

database:
  # Path to the SQLite database used to store client data.
  path: "/var/lib/netguard/db.sqlite"

logging:
  # Log level: debug, info, warn, error
  level: "info"
  
  # Path to the log file. If empty, logs to stdout.
  file: "/var/log/netguard.log"
```

## Section Details

### `server`
Configuration for the local WireGuard server instance.
- **endpoint** (string): The publicly accessible address of the server. Required.
- **port** (int): UDP listen port. Default: `51820`.
- **interface** (string): Network interface name. Default: `wg0`.
- **private_key** (string): Base64 encoded WireGuard private key. Required.
- **subnet** (string): CIDR notation for the VPN subnet. Default: `10.77.0.0/24`.
- **mtu** (int): Interface MTU. Default: `1420`.

### `client_defaults`
Settings automatically applied to newly created clients unless overridden.
- **dns** (list of strings): DNS servers provided to the client.
- **allowed_ips** (list of strings): Subnets the client will route through the VPN. Default: `["0.0.0.0/0"]`.
- **persistent_keepalive** (int): Interval in seconds. Default: `25`.

### `database`
- **path** (string): Absolute path to the SQLite DB. Default: `/var/lib/netguard/db.sqlite`.

### `logging`
- **level** (string): Logging verbosity. Default: `info`.
- **file** (string): Log file destination. Default: stdout.

## Environment Variable Overrides

Any configuration value can be overridden using environment variables prefixed with `NG_`. The structure follows the nested YAML, using underscores (`_`).

Examples:
- `NG_SERVER_ENDPOINT=vpn.mydomain.com` overrides `server.endpoint`
- `NG_SERVER_PORT=51821` overrides `server.port`
- `NG_LOGGING_LEVEL=debug` overrides `logging.level`

## Default Values

| Setting | Default Value |
|---------|---------------|
| `server.port` | `51820` |
| `server.interface` | `wg0` |
| `server.subnet` | `10.77.0.0/24` |
| `server.mtu` | `1420` |
| `client_defaults.allowed_ips` | `["0.0.0.0/0"]` |
| `client_defaults.persistent_keepalive` | `25` |
| `database.path` | `/var/lib/netguard/db.sqlite` |
| `logging.level` | `info` |
