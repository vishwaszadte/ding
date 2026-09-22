# ding Windows Installer
# Usage: irm https://raw.githubusercontent.com/vishwaszadte/ding/main/install.ps1 | iex

$ErrorActionPreference = 'Stop'
$Repo = "vishwaszadte/ding"
$InstallDir = "$env:LOCALAPPDATA\Programs\ding"

Write-Host "🔔 Installing ding for Windows..." -ForegroundColor Cyan

# Detect Architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}

# Fetch latest release tag
try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    $Tag = $Release.tag_name
} catch {
    $Tag = "v1.0.0"
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$Tag/ding_windows_$Arch.zip"
$TempZip = "$env:TEMP\ding.zip"

Write-Host "📦 Downloading $DownloadUrl..." -ForegroundColor Gray
Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Expand-Archive -Path $TempZip -DestinationPath $InstallDir -Force
Remove-Item -Force $TempZip

Write-Host "✅ ding installed to $InstallDir\ding.exe" -ForegroundColor Green

# Add to user PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    $env:Path += ";$InstallDir"
    Write-Host "✨ Added $InstallDir to user PATH." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Run 'ding test' in a new terminal to verify your desktop notifications!" -ForegroundColor Green
