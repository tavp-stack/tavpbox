#!/bin/bash
# TAVPBox start script for PHP containers

# Start PHP-FPM
service php8.2-fpm start 2>/dev/null || true

# Start Nginx
nginx 2>/dev/null || true

# Keep container alive
while true; do sleep 3600; done
