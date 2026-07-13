# TAVPBox - Windows Global Installer
# Usage: powershell -ExecutionPolicy Bypass -File install-windows.ps1
# Or: iex (irm 'https://raw.githubusercontent.com/tavp-stack/tavpbox/main/install/install-windows.ps1')

param(
    [switch]$SkipWSL,
    [switch]$SkipLXD
)

$ErrorActionPreference = "Stop"

# ── Colors ────────────────────────────────────────────────────
$Green = "`e[32m"
$Yellow = "`e[33m"
$Red = "`e[31m"
$Cyan = "`e[36m"
$Bold = "`e[1m"
$Reset = "`e[0m"

# ── Banner ────────────────────────────────────────────────────
Write-Host ""
Write-Host "${Cyan}${Bold}╔══════════════════════════════════════════════════════════════╗${Reset}"
Write-Host "${Cyan}${Bold}║                                                              ║${Reset}"
Write-Host "${Cyan}${Bold}║   ⚡ TAVPBox - LXC-based Dev Environment                     ║${Reset}"
Write-Host "${Cyan}${Bold}║   Like Lando, but lighter RAM                                ║${Reset}"
Write-Host "${Cyan}${Bold}║                                                              ║${Reset}"
Write-Host "${Cyan}${Bold}╚══════════════════════════════════════════════════════════════╝${Reset}"
Write-Host ""

# ── Step 1: Check/Install WSL2 ────────────────────────────────
Write-Host "${Bold}[1/5] Checking WSL2...${Reset}"
$wslStatus = wsl --status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ${Yellow}⚠ WSL2 not found${Reset}"
    Write-Host "  ${Cyan}→ Installing WSL2...${Reset}"
    
    # Enable WSL feature
    Write-Host "  ${Cyan}→ Enabling Windows Subsystem for Linux...${Reset}"
    dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart | Out-Null
    
    # Enable Virtual Machine Platform
    Write-Host "  ${Cyan}→ Enabling Virtual Machine Platform...${Reset}"
    dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart | Out-Null
    
    # Set WSL 2 as default
    Write-Host "  ${Cyan}→ Setting WSL 2 as default...${Reset}"
    wsl --set-default-version 2 | Out-Null
    
    Write-Host "  ${Green}✓ WSL2 installed${Reset}"
    Write-Host "  ${Yellow}⚠ Please REBOOT your computer and run this script again${Reset}"
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 0
} else {
    Write-Host "  ${Green}✓ WSL2 is available${Reset}"
}

# ── Step 2: Check/Install Ubuntu ──────────────────────────────
Write-Host "${Bold}[2/5] Checking Ubuntu WSL...${Reset}"
$distros = wsl --list --quiet 2>&1
if ($distros -notmatch "Ubuntu") {
    Write-Host "  ${Yellow}⚠ Ubuntu not found${Reset}"
    Write-Host "  ${Cyan}→ Installing Ubuntu...${Reset}"
    
    wsl --install Ubuntu --no-launch | Out-Null
    
    # Wait for Ubuntu to register
    $maxWait = 60
    $waited = 0
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 2
        $waited += 2
        $distros = wsl --list --quiet 2>&1
        if ($distros -match "Ubuntu") {
            break
        }
        Write-Host "  ${Cyan}→ Waiting for Ubuntu... ($waited seconds)${Reset}"
    }
    
    if ($distros -notmatch "Ubuntu") {
        Write-Host "  ${Red}✗ Ubuntu installation timed out${Reset}"
        Write-Host "  ${Yellow}Please install manually: wsl --install -d Ubuntu${Reset}"
        Read-Host "Press Enter to exit"
        exit 1
    }
    
    Write-Host "  ${Green}✓ Ubuntu installed${Reset}"
} else {
    Write-Host "  ${Green}✓ Ubuntu is available${Reset}"
}

# Set Ubuntu as default
wsl --set-default Ubuntu 2>&1 | Out-Null

# ── Step 3: Check/Install LXD ────────────────────────────────
Write-Host "${Bold}[3/5] Checking LXD...${Reset}"
$lxdCheck = wsl -d Ubuntu -- bash -c "which lxc 2>/dev/null || echo 'not found'"
if ($lxdCheck -match "not found") {
    Write-Host "  ${Yellow}⚠ LXD not found${Reset}"
    Write-Host "  ${Cyan}→ Installing LXD...${Reset}"
    
    # Install LXD via snap
    Write-Host "  ${Cyan}→ Installing LXD via snap...${Reset}"
    wsl -d Ubuntu -- sudo snap install lxd 2>&1 | Out-Null
    
    # Initialize LXD
    Write-Host "  ${Cyan}→ Initializing LXD...${Reset}"
    wsl -d Ubuntu -- sudo lxd init --auto 2>&1 | Out-Null
    
    # Add user to lxd group
    Write-Host "  ${Cyan}→ Configuring permissions...${Reset}"
    wsl -d Ubuntu -- sudo usermod -aG lxd root 2>&1 | Out-Null
    
    Write-Host "  ${Green}✓ LXD installed${Reset}"
} else {
    Write-Host "  ${Green}✓ LXD is available${Reset}"
}

# ── Step 4: Install TAVPBox ──────────────────────────────────
Write-Host "${Bold}[4/5] Installing TAVPBox...${Reset}"

# Create install directory
$installDir = "$env:LOCALAPPDATA\tavpbox"
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}

# Download binary
$binaryUrl = "https://github.com/tavp-stack/tavpbox/releases/latest/download/tavpbox-windows-amd64.exe"
$binaryPath = "$installDir\tavpbox.exe"

Write-Host "  ${Cyan}→ Downloading TAVPBox...${Reset}"
try {
    Invoke-WebRequest -Uri $binaryUrl -OutFile $binaryPath -UseBasicParsing
    Write-Host "  ${Green}✓ TAVPBox downloaded${Reset}"
} catch {
    Write-Host "  ${Yellow}⚠ Download failed, using local binary${Reset}"
    # Try to copy from current directory
    if (Test-Path ".\tavpbox-windows-amd64.exe") {
        Copy-Item ".\tavpbox-windows-amd64.exe" $binaryPath
    } elseif (Test-Path ".\tavpbox.exe") {
        Copy-Item ".\tavpbox.exe" $binaryPath
    } else {
        Write-Host "  ${Red}✗ No binary found${Reset}"
        Read-Host "Press Enter to exit"
        exit 1
    }
}

# Create wrapper script
$wrapperPath = "$installDir\tavpbox.ps1"
$wrapperContent = @"
# TAVPBox wrapper script
`$env:PATH += ";$installDir"
& "$binaryPath" @args
"@
Set-Content -Path $wrapperPath -Value $wrapperContent

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:PATH += ";$installDir"
    Write-Host "  ${Green}✓ Added to PATH${Reset}"
}

# Create global command
$globalWrapper = "$env:LOCALAPPDATA\Microsoft\WindowsApps\tavpbox.bat"
$globalContent = @"
@echo off
"$binaryPath" %*
"@
Set-Content -Path $globalWrapper -Value $globalContent -Encoding ASCII

Write-Host "  ${Green}✓ TAVPBox installed globally${Reset}"

# ── Step 5: Run Initial Setup ────────────────────────────────
Write-Host "${Bold}[5/5] Running initial setup...${Reset}"
Write-Host "  ${Cyan}→ Starting TAVPBox setup...${Reset}"
Write-Host ""

# Run tavpbox init
& $binaryPath init

# ── Success ──────────────────────────────────────────────────
Write-Host ""
Write-Host "${Green}${Bold}╔══════════════════════════════════════════════════════════════╗${Reset}"
Write-Host "${Green}${Bold}║                                                              ║${Reset}"
Write-Host "${Green}${Bold}║   ✓ TAVPBox installed successfully!                          ║${Reset}"
Write-Host "${Green}${Bold}║                                                              ║${Reset}"
Write-Host "${Green}${Bold}╚══════════════════════════════════════════════════════════════╝${Reset}"
Write-Host ""
Write-Host "${Bold}Quick Start:${Reset}"
Write-Host "  1. ${Cyan}tavpbox init${Reset}          - Setup your environment"
Write-Host "  2. ${Cyan}tavpbox create${Reset}        - Create a new container"
Write-Host "  3. ${Cyan}tavpbox list${Reset}          - List all containers"
Write-Host "  4. ${Cyan}tavpbox ssh <name>${Reset}    - Enter a container"
Write-Host ""
Write-Host "${Bold}Commands:${Reset}"
Write-Host "  ${Cyan}tavpbox --help${Reset}        - Show all commands"
Write-Host "  ${Cyan}tavpbox version${Reset}       - Show version"
Write-Host ""
Write-Host "${Bold}Documentation:${Reset}"
Write-Host "  ${Cyan}https://docs.tavp.web.id/guide/tavpbox.html${Reset}"
Write-Host ""
Write-Host "${Yellow}Note: You may need to restart your terminal for PATH changes to take effect${Reset}"
Write-Host ""
Read-Host "Press Enter to exit"
