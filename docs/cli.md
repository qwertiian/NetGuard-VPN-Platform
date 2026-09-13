# Command Line Interface (CLI) Reference

The `ngvpn` CLI is the primary way to interact with and manage NetGuard.

## Global Flags

These flags can be used with any command:
- `--config, -c <path>`: Specify a custom configuration file path (default: `/etc/netguard/config.yaml`).
- `--json, -j`: Output results in JSON format (useful for scripting).
- `--verbose, -v`: Enable verbose/debug logging.
- `--yes, -y`: Automatically answer "yes" to all confirmation prompts.
- `--version`: Print the current version and exit.
- `--help, -h`: Show help for the given command.

---

## Server Commands

Manage the WireGuard server instance.

### `ngvpn server init`
Initializes a new NetGuard installation. Generates keys, creates the default configuration file, and sets up the database.
- **Example:** `sudo ngvpn server init`

### `ngvpn server start`
Starts the WireGuard interface (`wg0`) and applies routing/NAT rules.
- **Example:** `sudo ngvpn server start`

### `ngvpn server stop`
Stops the WireGuard interface and removes associated network rules.
- **Example:** `sudo ngvpn server stop`

### `ngvpn server restart`
Restarts the server (equivalent to stop then start).
- **Example:** `sudo ngvpn server restart`

### `ngvpn server status`
Displays the current status of the WireGuard interface and daemon.
- **Example:** `ngvpn server status`

---

## Client Commands

Manage VPN clients (peers).

### `ngvpn client add <name>`
Creates a new client with the given name.
- **Example:** `sudo ngvpn client add my-laptop`

### `ngvpn client list`
Lists all configured clients and their IP addresses.
- **Example:** `sudo ngvpn client list`

### `ngvpn client show <name>`
Shows detailed information about a specific client.
- **Example:** `sudo ngvpn client show my-laptop`

### `ngvpn client revoke <name>`
Revokes a client's access without deleting their configuration. They will no longer be able to connect.
- **Example:** `sudo ngvpn client revoke my-laptop`

### `ngvpn client remove <name>`
Permanently deletes a client and their configuration from the database.
- **Example:** `sudo ngvpn client remove my-laptop`

### `ngvpn client export <name>`
Prints the WireGuard `.conf` file content for the client to stdout.
- **Example:** `sudo ngvpn client export my-laptop > laptop.conf`

### `ngvpn client qr <name>`
Generates an ASCII QR code in the terminal for the client's configuration.
- **Example:** `sudo ngvpn client qr my-phone`

### `ngvpn client rotate <name>`
Generates a new keypair for the client. **Note:** The client must be updated with the new configuration.
- **Example:** `sudo ngvpn client rotate my-laptop`

---

## Peer Commands

Interact directly with the active WireGuard session.

### `ngvpn peer list`
Lists actively connected peers and their data usage.
- **Example:** `sudo ngvpn peer list`

### `ngvpn peer status`
Shows detailed runtime status of all peers.
- **Example:** `sudo ngvpn peer status`

---

## Diagnostic Commands

### `ngvpn doctor`
Runs a suite of diagnostic tests to check for configuration errors, missing dependencies, firewall issues, and network connectivity.
- **Example:** `sudo ngvpn doctor`

---

## Backup Commands

### `ngvpn backup create`
Creates a backup archive containing the configuration, database, and generated keys.
- **Example:** `sudo ngvpn backup create ./backup.tar.gz`

### `ngvpn backup restore <file>`
Restores NetGuard from a backup archive.
- **Example:** `sudo ngvpn backup restore ./backup.tar.gz`

---

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Success |
| `1` | General error (e.g., config invalid, command failed) |
| `2` | CLI usage error (e.g., invalid flags) |
| `3` | Permission denied (must run as root) |
| `4` | Dependency missing (e.g., wireguard-tools not found) |
