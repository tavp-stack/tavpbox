#!/bin/bash
# Mailpit installer
set -e

echo "Installing Mailpit..."

curl -sL https://github.com/axllent/mailpit/releases/latest/download/mailpit_linux_amd64.tar.gz | tar xz -C /usr/local/bin/ 2>/dev/null || {
    echo "Mailpit download failed, skipping..."
    exit 0
}

# Create systemd service
cat > /etc/systemd/system/mailpit.service <<'EOF'
[Unit]
Description=Mailpit
After=network.target

[Service]
ExecStart=/usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025
Restart=always

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload 2>/dev/null || true
systemctl enable mailpit 2>/dev/null || true
systemctl start mailpit 2>/dev/null || true

echo "Mailpit installed!"
