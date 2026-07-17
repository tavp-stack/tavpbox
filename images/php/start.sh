#!/bin/bash
# TAVPBox start script for PHP containers

# Start MariaDB (init if needed)
mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
if [ ! -f /var/lib/mysql/ibdata1 ]; then
    mariadb-install-db --user=mysql --datadir=/var/lib/mysql 2>/dev/null || true
    chown -R mysql:mysql /var/lib/mysql
fi
nohup mariadbd --user=mysql --datadir=/var/lib/mysql --socket=/run/mysqld/mysqld.sock --pid-file=/run/mysqld/mysqld.pid > /var/log/mariadb.log 2>&1 &

# Wait for MariaDB
sleep 2

# Start Redis
redis-server --daemonize yes 2>/dev/null || true

# Start PHP-FPM
php-fpm 2>/dev/null || true

# Wait for PHP-FPM socket
sleep 1

# Start Nginx (retry if fails)
for i in 1 2 3; do
    nginx 2>/dev/null && break
    sleep 1
done

# Start Mailpit
if [ -f /usr/local/bin/mailpit ]; then
    nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &
fi

# Keep container alive - restart services if they die
while true; do
    sleep 10
    if [ -f /usr/local/bin/mailpit ] && ! pgrep -x mailpit > /dev/null 2>&1; then
        nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &
    fi
done
