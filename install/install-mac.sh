#!/bin/bash
set -e

echo "╔══════════════════════════════════════════╗"
echo "║   TAVPBox — macOS Installation           ║"
echo "║   (uses Lima VM)                         ║"
echo "╚══════════════════════════════════════════╝"

echo "[1/3] Installing dependencies..."
brew install lima jq

echo "[2/3] Creating Lima VM..."
cat > /tmp/lima-tavpbox.yaml <<'EOF'
arch: "default"
images:
  - location: "https://cloud-images.ubuntu.com/releases/24.04/release/ubuntu-24.04-server-cloudimg-amd64.img"
cpus: 2
memory: "1GB"
disk: "20GB"
mounts:
  - location: "~"
    writable: true
provision:
  - mode: system
    script: |
      #!/bin/bash
      snap install lxd
      lxd init --auto
      apt-get update
      apt-get install -y caddy dnsmasq jq curl
portForwards:
  - guestPortRange: [80, 80]
    hostPortRange: [8080, 8080]
  - guestPort: 8025
    hostPort: 8025
EOF

limactl start --name tavpbox /tmp/lima-tavpbox.yaml

echo "[3/3] Installing TAVPBox in VM..."
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    arm64)   ARCH="arm64" ;;
esac

limactl copy "./tavpbox-linux-${ARCH}" tavpbox:/usr/local/bin/tavpbox
limactl shell tavpbox chmod +x /usr/local/bin/tavpbox

cat > /usr/local/bin/tavpbox <<'WRAPPER'
#!/bin/bash
limactl shell tavpbox tavpbox "$@"
WRAPPER
chmod +x /usr/local/bin/tavpbox

echo ""
echo "✓ Installation complete!"
echo "  Run: tavpbox init"
