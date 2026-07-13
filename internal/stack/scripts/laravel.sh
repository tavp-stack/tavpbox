#!/bin/bash
# Laravel Stack provisioner
set -e

echo "Installing Laravel Stack..."

apt-get update -y
apt-get install -y php-fpm php-cli php-common php-curl php-mbstring php-xml php-zip php-bcmath php-intl php-gd php-mysql php-pgsql php-sqlite3 php-redis nginx curl git

curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer 2>/dev/null || true

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html/public;
    index index.php;
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
}
NGINX

service php*-fpm start 2>/dev/null || true
service nginx start 2>/dev/null || true

echo "Laravel Stack installed!"
