# ============================================================
# NetGuard VPN - Docker Image
# Multi-stage build for minimal production image
# ============================================================

# --- Build Stage ---
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X main.version=$(git describe --tags --always 2>/dev/null || echo docker)" \
    -o /ngvpn \
    ./cmd/ngvpn

# --- Runtime Stage ---
FROM alpine:3.19

LABEL maintainer="NetGuard Contributors"
LABEL description="Self-hosted WireGuard VPN management platform"
LABEL org.opencontainers.image.source="https://github.com/netguard-vpn/netguard"

# Install runtime dependencies
RUN apk add --no-cache \
    wireguard-tools \
    iptables \
    ip6tables \
    iproute2 \
    bash

# Copy binary
COPY --from=builder /ngvpn /usr/local/bin/ngvpn
RUN chmod 755 /usr/local/bin/ngvpn

# Create directories with secure permissions
RUN mkdir -p /etc/netguard/keys /etc/netguard/backups && \
    chmod 700 /etc/netguard /etc/netguard/keys /etc/netguard/backups

# WireGuard UDP port
EXPOSE 51820/udp

# Persistent data volume
VOLUME ["/etc/netguard", "/etc/wireguard"]

ENTRYPOINT ["ngvpn"]
CMD ["--help"]
