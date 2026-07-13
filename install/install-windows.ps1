# TAVPBox - Windows Global Installer
# Usage: powershell -ExecutionPolicy Bypass -File install-windows.ps1

$ErrorActionPreference = "Stop"

# Helper: Clean WSL output (UTF-16 encoding)
function Get-WslList {
    $raw = wsl --list --quiet 2>&1
    $clean = $raw -replace '\s+', ' ' -replace '\x00', ''
    return $clean.Trim()
}

# Banner
Write-Host ""
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host "  TAVPBox - LXC-based Dev Environment" -ForegroundColor Cyan
Write-Host "  Like Lando, but lighter RAM" -ForegroundColor Cyan
Write-Host "========================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check WSL2
Write-Host "[1/5] Checking WSL2..." -ForegroundColor Yellow
$wslStatus = wsl --status 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "  ! WSL2 not found" -ForegroundColor Red
    Write-Host "  > Installing WSL2..." -ForegroundColor Cyan
    Start-Process powershell -Verb RunAs -Wait -ArgumentList "-NoProfile -Command `"dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart; dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart; wsl --set-default-version 2`""
    Write-Host "  + WSL2 features enabled" -ForegroundColor Green
    Write-Host "  ! Please REBOOT your computer and run this script again" -ForegroundColor Yellow
    Read-Host "Press Enter to exit"
    exit 0
} else {
    Write-Host "  + WSL2 is available" -ForegroundColor Green
}

# Step 2: Check/Install Ubuntu
Write-Host "[2/5] Checking Ubuntu WSL..." -ForegroundColor Yellow
$distros = Get-WslList
$hasUbuntu = $distros -match "Ubuntu"

if (-not $hasUbuntu) {
    Write-Host "  ! Ubuntu not found" -ForegroundColor Red
    Write-Host "  > Installing Ubuntu..." -ForegroundColor Cyan
    wsl --install Ubuntu --no-launch 2>&1 | Out-Null
    $maxWait = 120
    $waited = 0
    $found = $false
    while ($waited -lt $maxWait) {
        Start-Sleep -Seconds 3
        $waited += 3
        $distros = Get-WslList
        if ($distros -match "Ubuntu") {
            $found = $true
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
    Write-Host ""
    Write-Host "  ! Ubuntu needs first-time setup" -ForegroundColor Yellow
    Write-Host "  > A terminal will open for you to create a username and password" -ForegroundColor Cyan
    Write-Host "  > After setup is complete, close the terminal and press Enter here" -ForegroundColor Cyan
    Write-Host ""
    Read-Host "Press Enter to open Ubuntu setup"
    Start-Process wsl -ArgumentList "-d Ubuntu"
    Read-Host "Press Enter after Ubuntu setup is complete"
} else {
    Write-Host "  + Ubuntu is available" -ForegroundColor Green
}

wsl --set-default Ubuntu 2>&1 | Out-Null

# Step 3: Check/Install LXD
Write-Host "[3/5] Checking LXD..." -ForegroundColor Yellow
$lxdCheck = wsl -d Ubuntu -- bash -c "which lxc 2>/dev/null"
if ($lxdCheck -notmatch "lxc") {
    Write-Host "  ! LXD not found" -ForegroundColor Red
    Write-Host "  > Installing LXD..." -ForegroundColor Cyan
    Write-Host "  > Installing LXD via snap..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo snap install lxd 2>&1 | Out-Null
    Write-Host "  > Initializing LXD..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo lxd init --auto 2>&1 | Out-Null
    Write-Host "  > Configuring permissions..." -ForegroundColor Cyan
    wsl -d Ubuntu -- sudo usermod -aG lxd root 2>&1 | Out-Null
    Write-Host "  + LXD installed" -ForegroundColor Green
} else {
    Write-Host "  + LXD is available" -ForegroundColor Green
}

# Step 4: Install TAVPBox
Write-Host "[4/5] Installing TAVPBox..." -ForegroundColor Yellow
$installDir = "$env:LOCALAPPDATA\tavpbox"
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir | Out-Null
}
$binaryPath = "$installDir\tavpbox.exe"
$sourcePath = "C:\Users\JT\Desktop\tavpbox-windows-amd64.exe"
if (Test-Path $sourcePath) {
    Copy-Item $sourcePath $binaryPath -Force
    Write-Host "  + TAVPBox installed" -ForegroundColor Green
} else {
    Write-Host "  X Binary not found at $sourcePath" -ForegroundColor Red
    Read-Host "Press Enter to exit"
    exit 1
}
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$installDir", "User")
    $env:PATH += ";$installDir"
    Write-Host "  + Added to PATH" -ForegroundColor Green
}
$globalWrapper = "$env:LOCALAPPDATA\Microsoft\WindowsApps\tavpbox.bat"
$globalContent = "@echo off`r`n`"$binaryPath`" %*"
Set-Content -Path $globalWrapper -Value $globalContent -Encoding ASCII
Write-Host "  + TAVPBox installed globally" -ForegroundColor Green

# Step 5: Verify
Write-Host "[5/5] Verifying installation..." -ForegroundColor Yellow
$tavpboxCheck = & $binaryPath version 2>&1
if ($tavpboxCheck -match "tavpbox") {
    Write-Host "  + TAVPBox is working" -ForegroundColor Green
} else {
    Write-Host "  ! TAVPBox verification failed" -ForegroundColor Yellow
}

# Success
Write-Host ""
Write-Host "========================================================" -ForegroundColor Green
Write-Host "  + TAVPBox installed successfully!" -ForegroundColor Green
Write-Host "========================================================" -ForegroundColor Green
Write-Host ""
Write-Host "Quick Start:" -ForegroundColor White
Write-Host "  1. Open a NEW terminal (PowerShell/CMD)" -ForegroundColor Cyan
Write-Host "  2. tavpbox init          - Setup your environment" -ForegroundColor Cyan
Write-Host "  3. tavpbox create        - Create a new container" -ForegroundColor Cyan
Write-Host "  4. tavpbox list          - List all containers" -ForegroundColor Cyan
Write-Host "  5. tavpbox ssh <name>    - Enter a container" -ForegroundColor Cyan
Write-Host ""
Write-Host "Commands:" -ForegroundColor White
Write-Host "  tavpbox --help        - Show all commands" -ForegroundColor Cyan
Write-Host "  tavpbox version       - Show version" -ForegroundColor Cyan
Write-Host ""
Write-Host "Documentation:" -ForegroundColor White
Write-Host "  https://docs.tavp.web.id/guide/tavpbox.html" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: You MUST open a NEW terminal for PATH changes to take effect" -ForegroundColor Yellow
Write-Host ""
Read-Host "Press Enter to exit"
