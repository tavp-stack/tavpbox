#!/bin/bash
# Blank Stack provisioner
set -e

echo "═══════════════════════════════════════"
echo "  Installing Blank Stack"
echo "═══════════════════════════════════════"

echo "[1/2] Updating packages..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y curl wget git
elif command -v apk &>/dev/null; then
    apk update
    apk add curl wget git
elif command -v dnf &>/dev/null; then
    dnf install -y curl wget git
fi

echo "[2/2] Installing Nginx..."
apt-get install -y nginx 2>/dev/null || apk add nginx 2>/dev/null || dnf install -y nginx 2>/dev/null

systemctl enable nginx 2>/dev/null || true
systemctl start nginx 2>/dev/null || true

echo "✓ Blank stack installed!"
