#!/bin/bash
# Laravel Stack provisioner
set -e

PHP_VER="${PHP_VERSION:-8.3}"

echo "═══════════════════════════════════════"
echo "  Installing Laravel Stack"
echo "  PHP ${PHP_VER}"
echo "═══════════════════════════════════════"

if command -v apt-get &>/dev/null; then
    PKG="apt"
elif command -v apk &>/dev/null; then
    PKG="apk"
elif command -v dnf &>/dev/null; then
    PKG="dnf"
fi

echo "[1/4] Installing PHP ${PHP_VER}..."
if [ "$PKG" = "apt" ]; then
    apt-get update
    apt-get install -y software-properties-common
    add-apt-repository -y ppa:ondrej/php
    apt-get update
    apt-get install -y \
        php${PHP_VER}-fpm php${PHP_VER}-cli php${PHP_VER}-common \
        php${PHP_VER}-curl php${PHP_VER}-mbstring php${PHP_VER}-xml \
        php${PHP_VER}-zip php${PHP_VER}-bcmath php${PHP_VER}-intl \
        php${PHP_VER}-gd php${PHP_VER}-readline \
        php${PHP_VER}-sqlite3 php${PHP_VER}-mysql php${PHP_VER}-pgsql \
        php${PHP_VER}-opcache php${PHP_VER}-redis
fi

echo "[2/4] Installing Nginx..."
apt-get install -y nginx

echo "[3/4] Installing Composer..."
curl -sS https://getcomposer.org/installer | php -- \
    --install-dir=/usr/local/bin --filename=composer

echo "[4/4] Configuring Nginx..."
cat > /etc/nginx/nginx.conf <<'NGINX'
worker_processes auto;
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 100M;
    server {
        listen 80 default_server;
        root /var/www/html/public;
        index index.php;
        location / {
            try_files $uri $uri/ /index.php?$query_string;
        }
        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include fastcgi_params;
        }
    }
}
NGINX

systemctl enable php${PHP_VER}-fpm nginx
systemctl start php${PHP_VER}-fpm nginx

echo "✓ Laravel stack installed!"
