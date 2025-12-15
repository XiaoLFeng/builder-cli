# xbuilder Windows Update Script
# Usage:
#   iwr -useb https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/update.ps1 | iex
#   & ([scriptblock]::Create((iwr -useb https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/update.ps1).Content)) -Version v1.0.0

param(
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"

# Configuration
$REPO = "XiaoLFeng/builder-cli"
$BINARY_NAME = "xbuilder"
$INSTALL_DIR = "$env:USERPROFILE\.local\bin"
$GITHUB_API = "https://api.github.com/repos/$REPO/releases"
$GITHUB_DOWNLOAD = "https://github.com/$REPO/releases/download"

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }
function Write-Err { Write-Host "[ERROR] $args" -ForegroundColor Red; exit 1 }

function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { Write-Err "Unsupported architecture: $arch" }
    }
}

function Get-CurrentVersion {
    $binary = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
    if (Test-Path $binary) {
        try {
            $output = & $binary --version 2>&1
            if ($output -match 'v?(\d+\.\d+\.\d+)') {
                return $matches[0]
            }
        }
        catch {}
    }
    return ""
}

function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "$GITHUB_API/latest" -Headers @{ "User-Agent" = "PowerShell" }
        return $release.tag_name
    }
    catch {
        Write-Err "Failed to get latest version: $_"
    }
}

function Compare-SemVer {
    param([string]$v1, [string]$v2)

    $v1 = $v1 -replace '^v', ''
    $v2 = $v2 -replace '^v', ''

    $v1Parts = $v1.Split('.') | ForEach-Object { [int]$_ }
    $v2Parts = $v2.Split('.') | ForEach-Object { [int]$_ }

    for ($i = 0; $i -lt [Math]::Max($v1Parts.Length, $v2Parts.Length); $i++) {
        $n1 = if ($i -lt $v1Parts.Length) { $v1Parts[$i] } else { 0 }
        $n2 = if ($i -lt $v2Parts.Length) { $v2Parts[$i] } else { 0 }

        if ($n1 -gt $n2) { return 1 }
        if ($n1 -lt $n2) { return -1 }
    }
    return 0
}

function Get-FileSHA256 {
    param([string]$FilePath)
    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    return $hash.Hash.ToLower()
}

function Update-Xbuilder {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Magenta
    Write-Host "       xbuilder Update Script (Windows)         " -ForegroundColor Magenta
    Write-Host "================================================" -ForegroundColor Magenta
    Write-Host ""

    $os = "windows"
    $arch = Get-Arch
    Write-Info "Detected system: $os/$arch"

    # Check current version
    $currentVersion = Get-CurrentVersion
    if ([string]::IsNullOrEmpty($currentVersion)) {
        Write-Warn "$BINARY_NAME is not installed, will perform fresh install"
        $currentVersion = "v0.0.0"
    }
    else {
        Write-Info "Current version: $currentVersion"
    }

    # Get target version
    if ([string]::IsNullOrEmpty($Version)) {
        $Version = Get-LatestVersion
    }
    Write-Info "Target version: $Version"

    # Compare versions
    $cmp = Compare-SemVer $Version $currentVersion
    switch ($cmp) {
        0 {
            Write-Success "Already at latest version ($currentVersion)"
            return
        }
        1 {
            Write-Info "Upgrading from $currentVersion to $Version"
        }
        -1 {
            Write-Warn "Target version ($Version) is older than current version ($currentVersion)"
            $confirm = Read-Host "Confirm downgrade? [y/N]"
            if ($confirm -ne 'y' -and $confirm -ne 'Y') {
                Write-Info "Operation cancelled"
                return
            }
        }
    }

    # Create temp directory
    $tmpDir = Join-Path $env:TEMP "xbuilder-update-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        $binaryFile = "$BINARY_NAME-$os-$arch.exe"
        $downloadUrl = "$GITHUB_DOWNLOAD/$Version/$binaryFile"
        $checksumUrl = "$GITHUB_DOWNLOAD/$Version/checksums.txt"

        Write-Info "Downloading $binaryFile..."
        $tmpBinary = Join-Path $tmpDir $binaryFile
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpBinary -UseBasicParsing

        Write-Info "Downloading checksum..."
        $tmpChecksum = Join-Path $tmpDir "checksums.txt"
        Invoke-WebRequest -Uri $checksumUrl -OutFile $tmpChecksum -UseBasicParsing

        $checksumContent = Get-Content $tmpChecksum
        $expectedHash = ($checksumContent | Where-Object { $_ -match $binaryFile } | ForEach-Object { ($_ -split '\s+')[0] })

        if ($expectedHash) {
            $actualHash = Get-FileSHA256 $tmpBinary
            if ($actualHash -ne $expectedHash) {
                Write-Err "Checksum mismatch!"
            }
            Write-Success "Checksum verification passed"
        }

        # Backup old version
        $targetPath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
        if (Test-Path $targetPath) {
            Write-Info "Backing up old version..."
            Copy-Item $targetPath "$targetPath.backup" -Force
        }

        # Install new version
        Write-Info "Installing new version..."
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }
        Move-Item -Path $tmpBinary -Destination $targetPath -Force

        # Cleanup backup
        if (Test-Path "$targetPath.backup") {
            Remove-Item "$targetPath.backup" -Force
        }

        Write-Success "$BINARY_NAME updated to $Version!"
    }
    finally {
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

Update-Xbuilder
