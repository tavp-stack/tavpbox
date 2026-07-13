#!/bin/bash
# PostgreSQL installer
set -e

echo "Installing PostgreSQL..."
if command -v apt-get &>/dev/null; then
    apt-get update
    apt-get install -y postgresql postgresql-client
    service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true
    su - postgres -c "psql -c \"CREATE USER app WITH PASSWORD 'app' CREATEDB;\"" 2>/dev/null || true
    su - postgres -c "psql -c \"CREATE DATABASE app OWNER app;\"" 2>/dev/null || true
elif command -v apk &>/dev/null; then
    apk add postgresql postgresql-client
    su - postgres -c "initdb -D /var/lib/postgresql/data" 2>/dev/null || true
    rc-service postgresql start 2>/dev/null || true
elif command -v dnf &>/dev/null; then
    dnf install -y postgresql-server postgresql
    postgresql-setup --initdb 2>/dev/null || true
    systemctl start postgresql 2>/dev/null || true
fi

echo "✓ PostgreSQL installed!"
