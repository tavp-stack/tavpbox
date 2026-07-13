#!/bin/bash
# phpMyAdmin installer
set -e

echo "Installing phpMyAdmin..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y phpmyadmin
    ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true
elif command -v apk &>/dev/null; then
    apk add phpmyadmin
    ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true
elif command -v dnf &>/dev/null; then
    dnf install -y phpMyAdmin
    ln -sf /usr/share/phpMyAdmin /var/www/html/pma 2>/dev/null || true
fi

echo "✓ phpMyAdmin installed!"
