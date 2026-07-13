#!/bin/bash
# Blank Stack provisioner
set -e

echo "Installing Blank Stack..."

apt-get update -y
apt-get install -y curl wget git nginx

service nginx start 2>/dev/null || true

echo "Blank Stack installed!"
