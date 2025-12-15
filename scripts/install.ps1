# xbuilder Windows Installation Script
# Usage:
#   iwr -useb https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/install.ps1 | iex
#   & ([scriptblock]::Create((iwr -useb https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/install.ps1).Content)) -Version v1.0.0

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

# Color output functions
function Write-Color {
    param(
        [string]$Text,
        [string]$Color = "White"
    )
    Write-Host $Text -ForegroundColor $Color
}

function Write-Info { Write-Color "[INFO] $args" "Cyan" }
function Write-Success { Write-Color "[SUCCESS] $args" "Green" }
function Write-Warn { Write-Color "[WARN] $args" "Yellow" }
function Write-Err { Write-Color "[ERROR] $args" "Red"; exit 1 }

# Detect architecture
function Get-Arch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    switch ($arch) {
        "X64"   { return "amd64" }
        "Arm64" { return "arm64" }
        default { Write-Err "Unsupported architecture: $arch" }
    }
}

# Get latest version from GitHub API
function Get-LatestVersion {
    try {
        $release = Invoke-RestMethod -Uri "$GITHUB_API/latest" -Headers @{ "User-Agent" = "PowerShell" }
        return $release.tag_name
    }
    catch {
        Write-Err "Failed to get latest version: $_"
    }
}

# Calculate SHA256 hash
function Get-FileSHA256 {
    param([string]$FilePath)
    $hash = Get-FileHash -Path $FilePath -Algorithm SHA256
    return $hash.Hash.ToLower()
}

# Main installation function
function Install-Xbuilder {
    Write-Host ""
    Write-Color "================================================" "Magenta"
    Write-Color "       xbuilder Installation Script (Windows)   " "Magenta"
    Write-Color "       Build & Deploy Pipeline CLI Tool         " "Magenta"
    Write-Color "================================================" "Magenta"
    Write-Host ""

    # Detect environment
    $os = "windows"
    $arch = Get-Arch
    Write-Info "Detected system: $os/$arch"

    # Get version
    if ([string]::IsNullOrEmpty($Version)) {
        $Version = Get-LatestVersion
    }
    Write-Info "Installing version: $Version"

    # Build download URL
    $binaryFile = "$BINARY_NAME-$os-$arch.exe"
    $downloadUrl = "$GITHUB_DOWNLOAD/$Version/$binaryFile"
    $checksumUrl = "$GITHUB_DOWNLOAD/$Version/checksums.txt"

    # Create temp directory
    $tmpDir = Join-Path $env:TEMP "xbuilder-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    try {
        # Download binary
        Write-Info "Downloading $binaryFile..."
        $tmpBinary = Join-Path $tmpDir $binaryFile
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpBinary -UseBasicParsing

        # Download checksum
        Write-Info "Downloading checksum file..."
        $tmpChecksum = Join-Path $tmpDir "checksums.txt"
        Invoke-WebRequest -Uri $checksumUrl -OutFile $tmpChecksum -UseBasicParsing

        # Verify checksum
        Write-Info "Verifying file integrity..."
        $checksumContent = Get-Content $tmpChecksum
        $expectedHash = ($checksumContent | Where-Object { $_ -match $binaryFile } | ForEach-Object { ($_ -split '\s+')[0] })

        if ($expectedHash) {
            $actualHash = Get-FileSHA256 $tmpBinary
            if ($actualHash -ne $expectedHash) {
                Write-Err "Checksum mismatch!`nExpected: $expectedHash`nActual: $actualHash"
            }
            Write-Success "Checksum verification passed"
        }
        else {
            Write-Warn "Checksum not found, skipping verification"
        }

        # Create install directory
        Write-Info "Installing to $INSTALL_DIR..."
        if (-not (Test-Path $INSTALL_DIR)) {
            New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
        }

        # Install binary
        $targetPath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
        Move-Item -Path $tmpBinary -Destination $targetPath -Force

        Write-Success "$BINARY_NAME $Version installed successfully!"

        # Check PATH
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if ($userPath -notlike "*$INSTALL_DIR*") {
            Write-Host ""
            Write-Warn "Install directory ($INSTALL_DIR) is not in PATH"
            Write-Host ""
            Write-Host "Choose one of the following methods to add to PATH:" -ForegroundColor White
            Write-Host ""
            Write-Host "Method 1 (Recommended): Auto add to user PATH" -ForegroundColor Cyan
            Write-Host '  [Environment]::SetEnvironmentVariable("Path", "$env:Path;' + $INSTALL_DIR + '", "User")' -ForegroundColor Gray
            Write-Host ""
            Write-Host "Method 2: Manual add" -ForegroundColor Cyan
            Write-Host "  System Properties -> Advanced -> Environment Variables -> User Variables -> Path -> Add: $INSTALL_DIR" -ForegroundColor Gray
            Write-Host ""
            Write-Host "Method 3: Temporary (current session only)" -ForegroundColor Cyan
            Write-Host '  $env:Path += ";' + $INSTALL_DIR + '"' -ForegroundColor Gray
        }

        Write-Host ""
        Write-Host "Usage: " -NoNewline
        Write-Color "$BINARY_NAME --help" "Green"
        Write-Host ""
    }
    finally {
        # Cleanup temp directory
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Run
Install-Xbuilder
