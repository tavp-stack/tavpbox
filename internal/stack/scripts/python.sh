#!/bin/bash
# Python Stack provisioner
set -e

echo "═══════════════════════════════════════"
echo "  Installing Python Stack"
echo "═══════════════════════════════════════"

echo "[1/3] Installing Python..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y python3 python3-pip python3-venv python3-dev
elif command -v apk &>/dev/null; then
    apk add python3 py3-pip py3-virtualenv
elif command -v dnf &>/dev/null; then
    dnf install -y python3 python3-pip python3-devel
fi

echo "[2/3] Installing Nginx..."
apt-get install -y nginx 2>/dev/null || apk add nginx 2>/dev/null || dnf install -y nginx 2>/dev/null

echo "[3/3] Configuring Nginx..."
cat > /etc/nginx/nginx.conf <<'NGINX'
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    server {
        listen 80;
        root /var/www/html;
        index index.html;
        location / {
            proxy_pass http://127.0.0.1:8000;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
NGINX

systemctl enable nginx 2>/dev/null || true
systemctl start nginx 2>/dev/null || true

echo "✓ Python stack installed!"
