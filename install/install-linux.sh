#!/bin/bash
set -e

echo "╔══════════════════════════════════════════╗"
echo "║   TAVPBox — Linux Installation           ║"
echo "╚══════════════════════════════════════════╝"

if [ "$EUID" -ne 0 ]; then
    echo "Please run as root: sudo bash $0"
    exit 1
fi

echo "[1/3] Installing LXD..."
if command -v lxd &>/dev/null; then
    echo "  LXD already installed"
else
    snap install lxd
    lxd init --auto
fi

echo "[2/3] Installing dependencies..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y caddy dnsmasq jq curl
elif command -v dnf &>/dev/null; then
    dnf install -y caddy dnsmasq jq curl
elif command -v pacman &>/dev/null; then
    pacman -Sy --noconfirm caddy dnsmasq jq curl
fi

echo "[3/3] Installing TAVPBox..."
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
esac

if [ -f "./tavpbox-linux-${ARCH}" ]; then
    install -m 755 "./tavpbox-linux-${ARCH}" /usr/local/bin/tavpbox
else
    echo "  Downloading latest release..."
    curl -sSL "https://github.com/tavp-stack/tavpbox/releases/latest/download/tavpbox-linux-${ARCH}" \
        -o /usr/local/bin/tavpbox
    chmod +x /usr/local/bin/tavpbox
fi

mkdir -p ~/.tavpbox/{boxes,plugins,snapshots}

echo ""
echo "✓ Installation complete!"
echo "  Run: tavpbox init"
