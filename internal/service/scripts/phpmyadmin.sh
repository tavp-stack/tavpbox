#!/bin/bash
# phpMyAdmin installer
set -e

echo "Installing phpMyAdmin..."

apt-get update -y
apt-get install -y phpmyadmin 2>/dev/null || {
    echo "phpMyAdmin not available in repos, skipping..."
    exit 0
}

ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true

echo "phpMyAdmin installed!"
