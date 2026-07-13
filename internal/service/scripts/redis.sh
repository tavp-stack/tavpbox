#!/bin/bash
# Redis installer
set -e

echo "Installing Redis..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y redis-server
    systemctl start redis-server 2>/dev/null || service redis-server start 2>/dev/null || true
elif command -v apk &>/dev/null; then
    apk add redis
    rc-service redis start 2>/dev/null || true
elif command -v dnf &>/dev/null; then
    dnf install -y redis
    systemctl start redis 2>/dev/null || true
fi

echo "✓ Redis installed!"
