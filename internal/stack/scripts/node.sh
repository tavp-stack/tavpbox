#!/bin/bash
# Node.js Stack provisioner
set -e

NODE_VER="${NODE_VERSION:-20}"

echo "═══════════════════════════════════════"
echo "  Installing Node.js Stack"
echo "  Node ${NODE_VER}"
echo "═══════════════════════════════════════"

echo "[1/3] Installing Node.js ${NODE_VER}..."
if command -v apt-get &>/dev/null; then
    curl -fsSL https://deb.nodesource.com/setup_${NODE_VER}.x | bash -
    apt-get install -y nodejs
elif command -v apk &>/dev/null; then
    apk add nodejs npm
elif command -v dnf &>/dev/null; then
    curl -fsSL https://rpm.nodesource.com/setup_${NODE_VER}.x | bash -
    dnf install -y nodejs
fi

echo "[2/3] Installing global packages..."
npm install -g yarn pnpm pm2

echo "[3/3] Configuring Nginx..."
apt-get install -y nginx 2>/dev/null || apk add nginx 2>/dev/null || dnf install -y nginx 2>/dev/null

cat > /etc/nginx/nginx.conf <<'NGINX'
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    server {
        listen 80;
        root /var/www/html;
        index index.html;
        location / {
            proxy_pass http://127.0.0.1:3000;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
        }
    }
}
NGINX

systemctl enable nginx 2>/dev/null || true
systemctl start nginx 2>/dev/null || true

echo "✓ Node.js stack installed!"
