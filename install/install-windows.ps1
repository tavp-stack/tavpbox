# TAVPBox — Windows WSL2 Installation
# Run as Administrator in PowerShell

Write-Host "╔══════════════════════════════════════════╗" -ForegroundColor Cyan
Write-Host "║   TAVPBox — Windows WSL2 Setup           ║" -ForegroundColor Cyan
Write-Host "╚══════════════════════════════════════════╝" -ForegroundColor Cyan

# Check WSL2
$wslStatus = wsl --status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "[1/4] Enabling WSL2..." -ForegroundColor Yellow
    wsl --install --no-distribution
    Write-Host "  WSL2 enabled. Please REBOOT and run this script again." -ForegroundColor Red
    exit 0
}

# Install Ubuntu if not present
$distros = wsl --list --quiet
if ($distros -notmatch "Ubuntu") {
    Write-Host "[2/4] Installing Ubuntu..." -ForegroundColor Yellow
    wsl --install -d Ubuntu
    Start-Sleep -Seconds 10
}

# Install TAVPBox inside WSL
Write-Host "[3/4] Installing TAVPBox in WSL..." -ForegroundColor Yellow
wsl -d Ubuntu -- bash -c "sudo snap install lxd && sudo lxd init --auto"
wsl -d Ubuntu -- bash -c "sudo apt-get update && sudo apt-get install -y caddy dnsmasq jq curl"

# Copy binary
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
wsl -d Ubuntu -- bash -c "sudo install -m 755 /dev/stdin /usr/local/bin/tavpbox" < "tavpbox-linux-$arch"

# Create Windows wrapper
Write-Host "[4/4] Creating Windows wrapper..." -ForegroundColor Yellow
$wrapperPath = "$env:LOCALAPPDATA\Microsoft\WindowsApps\tavpbox.bat"
@"
@echo off
wsl -d Ubuntu -- tavpbox %*
"@ | Out-File -FilePath $wrapperPath -Encoding ascii

# Setup DNS
Write-Host "  Configuring DNS..." -ForegroundColor Yellow
$wslIP = (wsl -d Ubuntu -- hostname -I).Trim().Split(" ")[0]
$hostsPath = "$env:SystemRoot\System32\drivers\etc\hosts"
if (-not (Get-Content $hostsPath -Match "tavp.local")) {
    Add-Content -Path $hostsPath -Value "`n# tavpbox"
    Add-Content -Path $hostsPath -Value "127.0.0.1`t*.tavp.local"
}

# Port forwarding
Write-Host "  Setting up port forwarding..." -ForegroundColor Yellow
netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=80 connectaddress=$wslIP
netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=$wslIP

Write-Host ""
Write-Host "✓ Installation complete!" -ForegroundColor Green
Write-Host "  Run: tavpbox init" -ForegroundColor Green
