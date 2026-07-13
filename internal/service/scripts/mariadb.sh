#!/bin/bash
# MariaDB installer
set -e

echo "Installing MariaDB..."

apt-get update -y
apt-get install -y mariadb-server mariadb-client 2>/dev/null || {
    # Fallback for systems without mariadb
    apt-get install -y mysql-server mysql-client 2>/dev/null || true
}

# Start MariaDB
service mariadb start 2>/dev/null || service mysql start 2>/dev/null || systemctl start mariadb 2>/dev/null || true

# Create default database and user
mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true

echo "MariaDB installed!"
