#!/bin/bash
# Python Stack provisioner
set -e

echo "Installing Python Stack..."

apt-get update -y
apt-get install -y python3 python3-pip python3-venv nginx curl

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html;
    index index.html;
    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
NGINX

service nginx start 2>/dev/null || true

echo "Python Stack installed!"
