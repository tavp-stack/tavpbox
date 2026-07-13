#!/bin/bash
# Redis installer
set -e

echo "Installing Redis..."

apt-get update -y
apt-get install -y redis-server

service redis-server start 2>/dev/null || systemctl start redis-server 2>/dev/null || true

echo "Redis installed!"
