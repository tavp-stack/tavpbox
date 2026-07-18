#!/bin/bash
# fix-nginx.sh - Fix nginx config for TAVP projects (webroot: .)

set -e

CONTAINER="$1"
if [ -z "$CONTAINER" ]; then
    echo "Usage: fix-nginx.sh <container_name>"
    exit 1
fi

echo "=== Fixing nginx config in $CONTAINER ==="

# Create proper nginx config
podman exec "$CONTAINER" bash -c '
cat > /etc/nginx/sites-available/default << '\''NGINX'\''
server {
    listen 80 default_server;
    root /var/www/html;
    index index.php index.html;
    
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    
    location ~ /\.ht {
        deny all;
    }
}
'\''NGINX'\''
'

# Test and reload
podman exec "$CONTAINER" nginx -t
podman exec "$CONTAINER" nginx -s reload

echo "=== Nginx config fixed and reloaded ==="
