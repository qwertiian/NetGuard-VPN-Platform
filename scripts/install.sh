#!/usr/bin/env bash
# NetGuard VPN - Installation Script
# https://github.com/netguard-vpn/netguard
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_banner() {
    echo -e "${BLUE}"
    echo "  _   _      _    ____                     _ "
    echo " | \ | | ___| |_ / ___|_   _  __ _ _ __ __| |"
    echo " |  \| |/ _ \ __| |  _| | | |/ _\` | '__/ _\` |"
    echo " | |\  |  __/ |_| |_| | |_| | (_| | | | (_| |"
    echo " |_| \_|\___|\__|\____|\__,_|\__,_|_|  \__,_|"
    echo ""
    echo "  Self-Hosted WireGuard VPN Management"
    echo -e "${NC}"
}

log_info()    { echo -e "${GREEN}[✓]${NC} $1"; }
log_warn()    { echo -e "${YELLOW}[⚠]${NC} $1"; }
log_error()   { echo -e "${RED}[✗]${NC} $1"; }
log_step()    { echo -e "${BLUE}[→]${NC} $1"; }

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        echo "  Run: sudo bash $0"
        exit 1
    fi
}

detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS_ID="$ID"
        OS_VERSION="$VERSION_ID"
    else
        log_error "Cannot detect operating system"
        exit 1
    fi

    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64)  ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        arm64)   ARCH="arm64" ;;
        *)
            log_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    log_info "Detected OS: $OS_ID $OS_VERSION ($ARCH)"
}

install_wireguard() {
    if command -v wg &>/dev/null; then
        log_info "WireGuard already installed: $(wg --version 2>/dev/null || echo 'installed')"
        return
    fi

    log_step "Installing WireGuard..."
    case "$OS_ID" in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y -qq wireguard wireguard-tools
            ;;
        fedora)
            dnf install -y -q wireguard-tools
            ;;
        arch|manjaro)
            pacman -S --noconfirm --needed wireguard-tools
            ;;
        centos|rhel|rocky|alma)
            dnf install -y -q epel-release
            dnf install -y -q wireguard-tools
            ;;
        *)
            log_error "Unsupported distribution: $OS_ID"
            echo "  Please install WireGuard manually and re-run this script."
            exit 1
            ;;
    esac
    log_info "WireGuard installed"
}

build_ngvpn() {
    INSTALL_DIR="/usr/local/bin"
    SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
    PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

    if [[ -f "$PROJECT_DIR/cmd/ngvpn/main.go" ]] && command -v go &>/dev/null; then
        GO_VERSION=$(go version | grep -oP '\d+\.\d+' | head -1)
        log_info "Go $GO_VERSION found, building from source..."
        cd "$PROJECT_DIR"
        CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo dev)" -o "$INSTALL_DIR/ngvpn" ./cmd/ngvpn
        log_info "Built and installed ngvpn to $INSTALL_DIR/ngvpn"
    else
        log_warn "Go not found or not in project directory"
        log_step "Downloading pre-built binary..."
        LATEST_URL="https://github.com/netguard-vpn/netguard/releases/latest/download/netguard_linux_${ARCH}.tar.gz"
        if curl -fsSL "$LATEST_URL" -o /tmp/ngvpn.tar.gz 2>/dev/null; then
            tar -xzf /tmp/ngvpn.tar.gz -C "$INSTALL_DIR" ngvpn
            chmod 755 "$INSTALL_DIR/ngvpn"
            rm -f /tmp/ngvpn.tar.gz
            log_info "Downloaded and installed ngvpn to $INSTALL_DIR/ngvpn"
        else
            log_error "Failed to download binary. Please build from source:"
            echo "  cd $PROJECT_DIR && make build && sudo make install"
            exit 1
        fi
    fi
}

setup_directories() {
    log_step "Creating directories..."
    mkdir -p /etc/netguard/keys
    mkdir -p /etc/netguard/backups
    chmod 700 /etc/netguard
    chmod 700 /etc/netguard/keys
    chmod 700 /etc/netguard/backups
    log_info "Created /etc/netguard/ with secure permissions"
}

setup_systemd() {
    log_step "Setting up systemd service..."
    cat > /etc/systemd/system/netguard-wg.service << 'EOF'
[Unit]
Description=NetGuard WireGuard VPN
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
RemainAfterExit=yes
ExecStart=/usr/bin/wg-quick up wg0
ExecStop=/usr/bin/wg-quick down wg0
ExecReload=/bin/bash -c 'wg syncconf wg0 <(wg-quick strip wg0)'

[Install]
WantedBy=multi-user.target
EOF
    systemctl daemon-reload
    log_info "Systemd service created: netguard-wg.service"
    log_info "Enable with: systemctl enable netguard-wg.service"
}

verify_installation() {
    log_step "Verifying installation..."
    if command -v ngvpn &>/dev/null; then
        VERSION=$(ngvpn --version 2>&1 | head -1)
        log_info "ngvpn installed: $VERSION"
    else
        log_error "ngvpn not found in PATH"
        exit 1
    fi

    if command -v wg &>/dev/null; then
        log_info "WireGuard available"
    else
        log_warn "WireGuard not found"
    fi
}

print_success() {
    echo ""
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    echo -e "${GREEN}  NetGuard installed successfully!${NC}"
    echo -e "${GREEN}════════════════════════════════════════${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Initialize the server:"
    echo "     ngvpn server init --endpoint <your-server-ip>"
    echo ""
    echo "  2. Start the VPN:"
    echo "     ngvpn server start"
    echo ""
    echo "  3. Add a client:"
    echo "     ngvpn client add <name>"
    echo ""
    echo "  4. Run health check:"
    echo "     ngvpn doctor"
    echo ""
}

# Main
print_banner
check_root
detect_os

echo ""
echo "This script will:"
echo "  • Install WireGuard (if not present)"
echo "  • Install ngvpn to /usr/local/bin/"
echo "  • Create /etc/netguard/ directory"
echo "  • Create systemd service"
echo ""
read -p "Continue? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Cancelled."
    exit 0
fi
echo ""

install_wireguard
build_ngvpn
setup_directories
setup_systemd
verify_installation
print_success
