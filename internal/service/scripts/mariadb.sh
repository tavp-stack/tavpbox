#!/bin/bash
# MariaDB installer
set -e

echo "Installing MariaDB..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y mariadb-server mariadb-client
    mysql_install_db --user=root --datadir=/var/lib/mysql 2>/dev/null || true
    service mariadb start 2>/dev/null || systemctl start mariadb 2>/dev/null || true
    mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true
elif command -v apk &>/dev/null; then
    apk add mariadb mariadb-client mariadb-server-utils
    mysql_install_db --user=root --datadir=/var/lib/mysql 2>/dev/null || true
    rc-service mariadb start 2>/dev/null || true
elif command -v dnf &>/dev/null; then
    dnf install -y mariadb-server mariadb
    mysql_install_db --user=root --datadir=/var/lib/mysql 2>/dev/null || true
    systemctl start mariadb 2>/dev/null || true
fi

echo "✓ MariaDB installed!"
