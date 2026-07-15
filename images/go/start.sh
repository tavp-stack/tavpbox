#!/bin/sh
# TAVPBox start script for Go containers

# Start Nginx
nginx 2>/dev/null || true

# Keep container alive
while true; do sleep 3600; done
