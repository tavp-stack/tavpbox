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
redis-server --daemonize yes >/dev/null 2>&1

# Start PHP-FPM
php-fpm --daemonize >/dev/null 2>&1

# Wait for PHP-FPM socket
sleep 1

# Start Nginx
nginx >/dev/null 2>&1

# Start Mailpit
if [ -f /usr/local/bin/mailpit ]; then
    nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &
fi

# Health check - restart dead services
exec sh -c '
while true; do
    sleep 10
    if [ -f /usr/local/bin/mailpit ] && ! pgrep -x mailpit >/dev/null 2>&1; then
        nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &
    fi
    if ! pgrep nginx >/dev/null 2>&1; then
        nginx >/dev/null 2>&1
    fi
done
'
