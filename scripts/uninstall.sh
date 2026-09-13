#!/usr/bin/env bash
# NetGuard VPN - Uninstallation Script
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

log_info()  { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn()  { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error() { echo -e "${RED}[✗]${NC} $1"; }

if [[ $EUID -ne 0 ]]; then
    log_error "This script must be run as root"
    exit 1
fi

echo "NetGuard VPN Uninstaller"
echo "═══════════════════════"
echo ""
echo "This will remove:"
echo "  • ngvpn binary from /usr/local/bin/"
echo "  • netguard-wg systemd service"
echo "  • netguard sysctl configuration"
echo "  • netguard-tagged iptables rules"
echo ""
echo -e "${YELLOW}Note: WireGuard itself will NOT be removed.${NC}"
echo ""
read -p "Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi
echo ""

# Stop WireGuard
if ip link show wg0 &>/dev/null; then
    log_warn "Stopping WireGuard interface wg0..."
    wg-quick down wg0 2>/dev/null || true
    log_info "WireGuard interface stopped"
fi

# Stop and disable systemd service
if systemctl is-active netguard-wg.service &>/dev/null; then
    systemctl stop netguard-wg.service 2>/dev/null || true
fi
if systemctl is-enabled netguard-wg.service &>/dev/null; then
    systemctl disable netguard-wg.service 2>/dev/null || true
fi
if [[ -f /etc/systemd/system/netguard-wg.service ]]; then
    rm -f /etc/systemd/system/netguard-wg.service
    systemctl daemon-reload
    log_info "Removed systemd service"
fi

# Remove binary
if [[ -f /usr/local/bin/ngvpn ]]; then
    rm -f /usr/local/bin/ngvpn
    log_info "Removed /usr/local/bin/ngvpn"
fi

# Remove sysctl config
if [[ -f /etc/sysctl.d/99-netguard.conf ]]; then
    rm -f /etc/sysctl.d/99-netguard.conf
    sysctl --system &>/dev/null || true
    log_info "Removed sysctl configuration"
fi

# Remove netguard-tagged iptables rules
if command -v iptables-save &>/dev/null; then
    if iptables-save 2>/dev/null | grep -q "netguard:"; then
        log_warn "Removing netguard-tagged iptables rules..."
        iptables-save | grep -v "netguard:" | iptables-restore 2>/dev/null || true
        log_info "Removed netguard iptables rules"
    fi
fi

# Ask about configuration data
echo ""
if [[ -d /etc/netguard ]]; then
    echo -e "${YELLOW}Configuration directory /etc/netguard/ still exists.${NC}"
    echo -e "${YELLOW}It may contain private keys and VPN configuration.${NC}"
    echo ""
    read -p "Remove /etc/netguard/ and all its contents? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf /etc/netguard
        log_info "Removed /etc/netguard/"
    else
        log_warn "Kept /etc/netguard/ — remove manually when ready"
    fi
fi

# Remove WireGuard config
if [[ -f /etc/wireguard/wg0.conf ]]; then
    read -p "Remove /etc/wireguard/wg0.conf? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f /etc/wireguard/wg0.conf
        log_info "Removed WireGuard configuration"
    else
        log_warn "Kept /etc/wireguard/wg0.conf"
    fi
fi

echo ""
log_info "NetGuard uninstalled successfully"
echo ""
echo "Note: WireGuard packages were not removed."
echo "To remove WireGuard: apt remove wireguard (or equivalent)"
