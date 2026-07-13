#!/bin/bash
# Node.js Stack provisioner
set -e

echo "Installing Node.js Stack..."

apt-get update -y
apt-get install -y nginx curl

curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs
npm install -g yarn pnpm 2>/dev/null || true

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
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
NGINX

service nginx start 2>/dev/null || true

echo "Node.js Stack installed!"
