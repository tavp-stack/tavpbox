#!/bin/bash
# TAVPBox start script for PHP containers

# Start MariaDB
mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
mysqld --user=mysql --datadir=/var/lib/mysql &

# Start Redis
redis-server --daemonize yes 2>/dev/null || true

# Start PHP-FPM
/usr/sbin/php-fpm8.2 --daemonize 2>/dev/null || true

# Start Nginx
nginx 2>/dev/null || true

# Start Mailpit
nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &

# Keep container alive
while true; do sleep 3600; done
