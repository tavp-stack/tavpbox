#!/bin/bash
# PostgreSQL installer
set -e

echo "Installing PostgreSQL..."

apt-get update -y
apt-get install -y postgresql postgresql-client

service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true

# Create default database and user
su - postgres -c "psql -c \"CREATE USER app WITH PASSWORD 'app' CREATEDB;\"" 2>/dev/null || true
su - postgres -c "psql -c \"CREATE DATABASE app OWNER app;\"" 2>/dev/null || true

echo "PostgreSQL installed!"
