#!/usr/bin/env bash
# ============================================================
# NetGuard VPN - Network Namespace Integration Test
# ============================================================
# Tests WireGuard tunnel establishment using Linux network namespaces.
# This creates an isolated test environment on a single machine.
#
# Requirements:
#   - Root privileges
#   - wireguard-tools installed
#   - iproute2 installed
#   - Linux kernel with WireGuard support (5.6+)
#
# This test does NOT require:
#   - Internet access
#   - A VPS or cloud server
#   - Any external infrastructure
set -euo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m'

PASSED=0
FAILED=0
SKIPPED=0

log_pass() { echo -e "  ${GREEN}✓ PASS${NC}: $1"; ((PASSED++)); }
log_fail() { echo -e "  ${RED}✗ FAIL${NC}: $1"; ((FAILED++)); }
log_skip() { echo -e "  ${YELLOW}○ SKIP${NC}: $1"; ((SKIPPED++)); }
log_info() { echo -e "  [→] $1"; }

# Cleanup function - always runs on exit
cleanup() {
    echo ""
    echo "Cleaning up test namespaces..."
    ip netns del ns-server 2>/dev/null || true
    ip netns del ns-client 2>/dev/null || true
    ip link del veth-srv 2>/dev/null || true
    rm -f /tmp/ngtest-server.key /tmp/ngtest-server.pub
    rm -f /tmp/ngtest-client.key /tmp/ngtest-client.pub
    echo "Cleanup complete."
}
trap cleanup EXIT

# Prerequisite checks
check_prerequisites() {
    echo "Checking prerequisites..."

    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: Must be run as root${NC}"
        exit 1
    fi

    if ! command -v wg &>/dev/null; then
        echo -e "${RED}Error: wireguard-tools not installed${NC}"
        exit 1
    fi

    if ! command -v ip &>/dev/null; then
        echo -e "${RED}Error: iproute2 not installed${NC}"
        exit 1
    fi

    # Check WireGuard kernel support
    if ! ip link add wgtest type wireguard 2>/dev/null; then
        echo -e "${RED}Error: WireGuard kernel module not available${NC}"
        exit 1
    fi
    ip link del wgtest 2>/dev/null || true

    log_pass "All prerequisites met"
}

# Setup test topology
setup_topology() {
    echo ""
    echo "Setting up test topology..."
    echo "  ns-client (192.168.100.2) <--veth--> ns-server (192.168.100.1)"
    echo "  ns-client (10.77.0.2/wg0) <--WireGuard--> ns-server (10.77.0.1/wg0)"

    # Create namespaces
    ip netns add ns-server
    ip netns add ns-client

    # Create veth pair (simulates physical network)
    ip link add veth-srv type veth peer name veth-cli
    ip link set veth-srv netns ns-server
    ip link set veth-cli netns ns-client

    # Configure server namespace
    ip netns exec ns-server ip addr add 192.168.100.1/24 dev veth-srv
    ip netns exec ns-server ip link set veth-srv up
    ip netns exec ns-server ip link set lo up

    # Configure client namespace
    ip netns exec ns-client ip addr add 192.168.100.2/24 dev veth-cli
    ip netns exec ns-client ip link set veth-cli up
    ip netns exec ns-client ip link set lo up

    # Generate WireGuard keys
    wg genkey | tee /tmp/ngtest-server.key | wg pubkey > /tmp/ngtest-server.pub
    wg genkey | tee /tmp/ngtest-client.key | wg pubkey > /tmp/ngtest-client.pub
    chmod 600 /tmp/ngtest-server.key /tmp/ngtest-client.key

    SRV_PUB=$(cat /tmp/ngtest-server.pub)
    CLI_PUB=$(cat /tmp/ngtest-client.pub)

    # Setup WireGuard on server
    ip netns exec ns-server ip link add wg0 type wireguard
    ip netns exec ns-server wg set wg0 \
        listen-port 51820 \
        private-key /tmp/ngtest-server.key \
        peer "$CLI_PUB" allowed-ips 10.77.0.2/32
    ip netns exec ns-server ip addr add 10.77.0.1/24 dev wg0
    ip netns exec ns-server ip link set wg0 up

    # Setup WireGuard on client
    ip netns exec ns-client ip link add wg0 type wireguard
    ip netns exec ns-client wg set wg0 \
        private-key /tmp/ngtest-client.key \
        peer "$SRV_PUB" endpoint 192.168.100.1:51820 allowed-ips 10.77.0.0/24
    ip netns exec ns-client ip addr add 10.77.0.2/24 dev wg0
    ip netns exec ns-client ip link set wg0 up

    log_pass "Test topology created"
}

# Test: Ping through WireGuard tunnel
test_tunnel_ping() {
    echo ""
    echo "Test 1: Tunnel connectivity (ping)"
    if ip netns exec ns-client ping -c 3 -W 5 10.77.0.1 &>/dev/null; then
        log_pass "Client can ping server through WireGuard tunnel"
    else
        log_fail "Client cannot ping server through tunnel"
    fi
}

# Test: Verify WireGuard handshake
test_handshake() {
    echo ""
    echo "Test 2: WireGuard handshake"
    HANDSHAKE=$(ip netns exec ns-server wg show wg0 latest-handshakes | awk '{print $2}')
    if [[ -n "$HANDSHAKE" ]] && [[ "$HANDSHAKE" != "0" ]]; then
        log_pass "Handshake completed (timestamp: $HANDSHAKE)"
    else
        log_fail "No handshake detected"
    fi
}

# Test: Verify traffic counters
test_traffic_counters() {
    echo ""
    echo "Test 3: Traffic counters"
    TRANSFER=$(ip netns exec ns-server wg show wg0 transfer)
    RX=$(echo "$TRANSFER" | awk '{print $2}')
    TX=$(echo "$TRANSFER" | awk '{print $3}')
    if [[ "$RX" -gt 0 ]] && [[ "$TX" -gt 0 ]]; then
        log_pass "Traffic detected (RX: $RX bytes, TX: $TX bytes)"
    else
        log_fail "No traffic on counters (RX: $RX, TX: $TX)"
    fi
}

# Test: Verify peer isolation (reverse direction)
test_reverse_ping() {
    echo ""
    echo "Test 4: Reverse connectivity (server → client)"
    if ip netns exec ns-server ping -c 3 -W 5 10.77.0.2 &>/dev/null; then
        log_pass "Server can ping client through tunnel"
    else
        log_fail "Server cannot ping client"
    fi
}

# Test: Peer revocation
test_peer_revocation() {
    echo ""
    echo "Test 5: Peer revocation"
    CLI_PUB=$(cat /tmp/ngtest-client.pub)

    # Remove peer from server
    ip netns exec ns-server wg set wg0 peer "$CLI_PUB" remove
    log_info "Removed client peer from server"

    # Wait for existing connections to time out
    sleep 2

    # Try to ping - should fail
    if ip netns exec ns-client ping -c 2 -W 3 10.77.0.1 &>/dev/null; then
        log_fail "Client can still reach server after revocation"
    else
        log_pass "Client correctly blocked after peer revocation"
    fi

    # Re-add peer for cleanup
    ip netns exec ns-server wg set wg0 peer "$CLI_PUB" allowed-ips 10.77.0.2/32
}

# Print results
print_results() {
    echo ""
    echo "═══════════════════════════"
    echo "Test Results"
    echo "═══════════════════════════"
    echo -e "  ${GREEN}Passed:  $PASSED${NC}"
    echo -e "  ${RED}Failed:  $FAILED${NC}"
    echo -e "  ${YELLOW}Skipped: $SKIPPED${NC}"
    echo "═══════════════════════════"

    if [[ $FAILED -eq 0 ]]; then
        echo -e "  ${GREEN}ALL TESTS PASSED${NC}"
        exit 0
    else
        echo -e "  ${RED}SOME TESTS FAILED${NC}"
        exit 1
    fi
}

# Main
echo "NetGuard VPN - Network Namespace Test"
echo "═════════════════════════════════════"
check_prerequisites
setup_topology
test_tunnel_ping
test_handshake
test_traffic_counters
test_reverse_ping
test_peer_revocation
print_results
