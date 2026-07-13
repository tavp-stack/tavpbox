#!/bin/bash
# TAVP Stack provisioner
set -e

PHP_VER="${PHP_VERSION:-8.3}"
NODE_VER="${NODE_VERSION:-20}"

echo "═══════════════════════════════════════"
echo "  Installing TAVP Stack"
echo "  PHP ${PHP_VER} · Node ${NODE_VER}"
echo "═══════════════════════════════════════"

# Detect package manager
if command -v apt-get &>/dev/null; then
    PKG="apt"
elif command -v apk &>/dev/null; then
    PKG="apk"
elif command -v dnf &>/dev/null; then
    PKG="dnf"
fi

# Install PHP
echo "[1/5] Installing PHP ${PHP_VER}..."
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
elif [ "$PKG" = "apk" ]; then
    PHP_SHORT=$(echo $PHP_VER | tr -d '.')
    apk update
    apk add \
        php${PHP_SHORT}-fpm php${PHP_SHORT}-cli php${PHP_SHORT}-common \
        php${PHP_SHORT}-curl php${PHP_SHORT}-mbstring php${PHP_SHORT}-xml \
        php${PHP_SHORT}-zip php${PHP_SHORT}-bcmath php${PHP_SHORT}-intl \
        php${PHP_SHORT}-gd php${PHP_SHORT}-pdo php${PHP_SHORT}-pdo_mysql \
        php${PHP_SHORT}-pgsql php${PHP_SHORT}-sqlite3 \
        php${PHP_SHORT}-opcache php${PHP_SHORT}-pecl-redis
elif [ "$PKG" = "dnf" ]; then
    dnf install -y epel-release
    dnf module enable php:${PHP_VER} -y
    dnf install -y \
        php-fpm php-cli php-common php-curl php-mbstring \
        php-xml php-zip php-bcmath php-intl php-gd \
        php-pdo php-mysqlnd php-pgsql php-sqlite3 \
        php-opcache php-redis
fi

# Install Nginx
echo "[2/5] Installing Nginx..."
if [ "$PKG" = "apt" ]; then
    apt-get install -y nginx
elif [ "$PKG" = "apk" ]; then
    apk add nginx
elif [ "$PKG" = "dnf" ]; then
    dnf install -y nginx
fi

# Install Composer
echo "[3/5] Installing Composer..."
curl -sS https://getcomposer.org/installer | php -- \
    --install-dir=/usr/local/bin --filename=composer

# Install Node.js
echo "[4/5] Installing Node.js ${NODE_VER}..."
if [ "$PKG" = "apt" ]; then
    curl -fsSL https://deb.nodesource.com/setup_${NODE_VER}.x | bash -
    apt-get install -y nodejs
elif [ "$PKG" = "apk" ]; then
    apk add nodejs npm
elif [ "$PKG" = "dnf" ]; then
    curl -fsSL https://rpm.nodesource.com/setup_${NODE_VER}.x | bash -
    dnf install -y nodejs
fi

# Configure Nginx
echo "[5/5] Configuring Nginx..."
cat > /etc/nginx/nginx.conf <<'NGINX'
worker_processes auto;
pid /run/nginx.pid;

events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;
    
    sendfile on;
    tcp_nopush on;
    keepalive_timeout 65;
    client_max_body_size 100M;
    
    gzip on;
    gzip_vary on;
    gzip_proxied any;
    gzip_comp_level 6;
    gzip_types text/plain text/css application/json application/javascript
               text/xml application/xml application/xml+rss text/javascript
               image/svg+xml;

    server {
        listen 80 default_server;
        listen [::]:80 default_server;
        
        root /var/www/html/public;
        index index.php index.html;

        location / {
            try_files $uri $uri/ /index.php?_url=$uri&$args;
        }

        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include fastcgi_params;
            fastcgi_read_timeout 300;
        }

        location ~ /\.ht {
            deny all;
        }

        location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
            expires 1y;
            add_header Cache-Control "public, immutable";
        }
    }
}
NGINX

# Start services
if command -v systemctl &>/dev/null; then
    systemctl enable php${PHP_VER}-fpm nginx
    systemctl start php${PHP_VER}-fpm nginx
elif command -v rc-service &>/dev/null; then
    rc-service php-fpm start 2>/dev/null || true
    rc-service nginx start
    rc-update add php-fpm default 2>/dev/null || true
    rc-update add nginx default
fi

echo ""
echo "═══════════════════════════════════════"
echo "  ✓ TAVP Stack installed!"
echo "  PHP ${PHP_VER} · Nginx · Node ${NODE_VER}"
echo "═══════════════════════════════════════"
