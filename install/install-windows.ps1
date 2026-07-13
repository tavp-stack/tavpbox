# TAVPBox - Windows Global Installer
# Usage: powershell -ExecutionPolicy Bypass -File install-windows.ps1

$ErrorActionPreference = "Stop"

# ── Banner ────────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  TAVPBox - LXC-based Dev Environment" -ForegroundColor Cyan
Write-Host "  Like Lando, but lighter RAM" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host ""

# ── Step 1: Check/Install WSL2 ────────────────────────────────
Write-Host "[1/5] Checking WSL2..." -ForegroundColor Yellow
$wslStatus = wsl --status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ! WSL2 not found" -ForegroundColor Red
    Write-Host "  > Installing WSL2..." -ForegroundColor Cyan
    
    # Enable WSL feature
    Write-Host "  > Enabling Windows Subsystem for Linux..." -ForegroundColor Cyan
    dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart | Out-Null
    
    # Enable Virtual Machine Platform
    Write-Host "  > Enabling Virtual Machine Platform..." -ForegroundColor Cyan
    dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart | Out-Null
    
    # Set WSL 2 as default
    Write-Host "  > Setting WSL 2 as default..." -ForegroundColor Cyan
    wsl --set-default-version 2 | Out-Null
    
    Write-Host "  + WSL2 installed" -ForegroundColor Green
    Write-Host "  ! Please REBOOT your computer and run this script again" -ForegroundColor Yellow
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 0
} else {
    Write-Host "  + WSL2 is available" -ForegroundColor Green
}

# ── Step 2: Check/Install Ubuntu ──────────────────────────────
Write-Host "[2/5] Checking Ubuntu WSL..." -ForegroundColor Yellow
$distros = wsl --list --quiet 2>&1
$hasUbuntu = $false
foreach ($d in $distros) {
    if ($d -match "Ubuntu") {
        $hasUbuntu = $true
        break
    }
}

if (-not $hasUbuntu) {
    Write-Host "  ! Ubuntu not found" -ForegroundColor Red
    Write-Host "  > Installing Ubuntu..." -ForegroundColor Cyan
    
    wsl --install Ubuntu --no-launch | Out-Null
    
    # Wait for Ubuntu to register
    $maxWait = 60
    $waited = 0
    $found = $false
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 2
        $waited += 2
        $distros = wsl --list --quiet 2>&1
        foreach ($d in $distros) {
            if ($d -match "Ubuntu") {
                $found = $true
                break
            }
        }
        if ($found) {
            break
        }
        Write-Host "  > Waiting for Ubuntu... ($waited seconds)" -ForegroundColor Cyan
    }
    
    if (-not $found) {
        Write-Host "  X Ubuntu installation timed out" -ForegroundColor Red
        Write-Host "  Please install manually: wsl --install -d Ubuntu" -ForegroundColor Yellow
        Read-Host "Press Enter to exit"
        exit 1
    }
    
    Write-Host "  + Ubuntu installed" -ForegroundColor Green
} else {
    Write-Host "  + Ubuntu is available" -ForegroundColor Green
}

# Set Ubuntu as default
wsl --set-default Ubuntu 2>&1 | Out-Null

# ── Step 3: Check/Install LXD ────────────────────────────────
Write-Host "[3/5] Checking LXD..." -ForegroundColor Yellow
$lxdCheck = wsl -d Ubuntu -- bash -c "which lxc 2>/dev/null"
if ($lxdCheck -notmatch "lxc") {
    Write-Host "  ! LXD not found" -ForegroundColor Red
    Write-Host "  > Installing LXD..." -ForegroundColor Cyan
    
    # Install LXD via snap
    Write-Host "  > Installing LXD via snap..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo snap install lxd 2>&1 | Out-Null
    
    # Initialize LXD
    Write-Host "  > Initializing LXD..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo lxd init --auto 2>&1 | Out-Null
    
    # Add user to lxd group
    Write-Host "  > Configuring permissions..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo usermod -aG lxd root 2>&1 | Out-Null
    
    Write-Host "  + LXD installed" -ForegroundColor Green
} else {
    Write-Host "  + LXD is available" -ForegroundColor Green
}

# ── Step 4: Install TAVPBox ──────────────────────────────────
Write-Host "[4/5] Installing TAVPBox..." -ForegroundColor Yellow

# Create install directory
$installDir = "$env:LOCALAPPDATA\tavpbox"
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

# Copy binary
$binaryPath = "$installDir\tavpbox.exe"
$sourcePath = "C:\Users\JT\Desktop\tavpbox-windows-amd64.exe"

if (Test-Path $sourcePath) {
    Copy-Item $sourcePath $binaryPath -Force
    Write-Host "  + TAVPBox installed" -ForegroundColor Green
} else {
    Write-Host "  X Binary not found at $sourcePath" -ForegroundColor Red
    Write-Host "  Please download from: https://github.com/tavp-stack/tavpbox/releases" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 1
}

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:PATH += ";$installDir"
    Write-Host "  + Added to PATH" -ForegroundColor Green
}

# Create global command
$globalWrapper = "$env:LOCALAPPDATA\Microsoft\WindowsApps\tavpbox.bat"
$globalContent = "@echo off`r`n`"$binaryPath`" %*"
Set-Content -Path $globalWrapper -Value $globalContent -Encoding ASCII

Write-Host "  + TAVPBox installed globally" -ForegroundColor Green

# ── Step 5: Run Initial Setup ────────────────────────────────
Write-Host "[5/5] Running initial setup..." -ForegroundColor Yellow
Write-Host "  > Starting TAVPBox setup..." -ForegroundColor Cyan
Write-Host ""

# Run tavpbox init
& $binaryPath init

# ── Success ──────────────────────────────────────────────────
Write-Host ""
Write-Host "========================================================" -ForegroundColor Green
Write-Host "  + TAVPBox installed successfully!" -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor White
Write-Host "  1. tavpbox init          - Setup your environment" -ForegroundColor Cyan
Write-Host "  2. tavpbox create        - Create a new container" -ForegroundColor Cyan
Write-Host "  3. tavpbox list          - List all containers" -ForegroundColor Cyan
Write-Host "  4. tavpbox ssh <name>    - Enter a container" -ForegroundColor Cyan
Write-Host ""
Write-Host "Commands:" -ForegroundColor White
Write-Host "  tavpbox --help        - Show all commands" -ForegroundColor Cyan
Write-Host "  tavpbox version       - Show version" -ForegroundColor Cyan
Write-Host ""
Write-Host "Documentation:" -ForegroundColor White
Write-Host "  https://docs.tavp.web.id/guide/tavpbox.html" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect" -ForegroundColor Yellow
Write-Host ""
Read-Host "Press Enter to exit"
