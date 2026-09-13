# Security Model

NetGuard is built on top of WireGuard, a modern, highly secure VPN protocol. This document outlines the security model of NetGuard, what it protects, and what it does *not* protect.

## What WireGuard Protects

WireGuard utilizes state-of-the-art cryptography (Curve25519, ChaCha20, Poly1305, BLAKE2) to ensure:

- **In-Transit Encryption:** All data sent between the client and the server is fully encrypted. Anyone intercepting the traffic (like an ISP or malicious actor on public Wi-Fi) will only see encrypted UDP packets.
- **Authentication:** Only clients with valid cryptographic keys recognized by the server can connect.
- **Perfect Forward Secrecy:** WireGuard rotates session keys frequently, meaning a compromised session key cannot be used to decrypt past traffic.
- **Identity Hiding:** WireGuard does not respond to unauthenticated packets, making the server "invisible" to port scanners looking for open services.

## What NetGuard/WireGuard Does NOT Protect

It is crucial to understand the limitations of a VPN:

- **Anonymity:** NetGuard does *not* make you anonymous on the internet. Your server provider knows who you are, and websites you visit can still track you via cookies, browser fingerprinting, and account logins.
- **Endpoint Security:** NetGuard does not protect your device from malware, viruses, or phishing attacks.
- **Traffic Analysis:** While the content is encrypted, adversaries can still observe the timing and size of the packets you send and receive.

## Key Management Best Practices

NetGuard handles key generation, but it relies on you to protect them:

1. **Server Private Key:** The `private_key` in `/etc/netguard/config.yaml` is the core secret of your VPN. Never share it.
2. **Client Private Keys:** When you export a client configuration (via `ngvpn client export`), it contains the client's private key. Treat these `.conf` files like passwords. Once loaded onto a device, delete the `.conf` file from the server.
3. **Key Rotation:** If a device is lost or compromised, immediately use `ngvpn client revoke <name>` or `ngvpn client remove <name>` to block access.

## File Permission Requirements

NetGuard strictly enforces file permissions. If permissions are too loose, NetGuard may refuse to start.

- `/etc/netguard/config.yaml`: Must be owned by `root` with `0600` permissions (`-rw-------`).
- `/var/lib/netguard/db.sqlite`: Must be owned by `root` with `0600` permissions.

## Backup Security

When using `ngvpn backup create`, the resulting archive contains your server's private key and all client configurations.

- **Store Backups Securely:** Do not store backups in publicly accessible web directories.
- **Encrypt Backups:** Consider encrypting the backup archive (e.g., using GPG) before transferring it off-site.

## Update Policy

NetGuard relies on the underlying OS for the core WireGuard implementation.
- Keep your server's OS up to date to receive kernel security patches.
- Update the `ngvpn` binary regularly as new releases become available.

## Responsible Usage

As a self-hosted VPN, you are responsible for the traffic exiting your server. Ensure you comply with the terms of service of your hosting provider, as abuse (spam, malicious activity) may lead to your server being terminated.
