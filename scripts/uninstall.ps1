# xbuilder Windows Uninstall Script
# Usage:
#   iwr -useb https://raw.githubusercontent.com/XiaoLFeng/builder-cli/master/scripts/uninstall.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$BINARY_NAME = "xbuilder"
$INSTALL_DIR = "$env:USERPROFILE\.local\bin"
$CONFIG_DIR = "$env:USERPROFILE\.xbuilder"

function Write-Info { Write-Host "[INFO] $args" -ForegroundColor Cyan }
function Write-Success { Write-Host "[SUCCESS] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[WARN] $args" -ForegroundColor Yellow }

function Uninstall-Xbuilder {
    Write-Host ""
    Write-Host "================================================" -ForegroundColor Magenta
    Write-Host "       xbuilder Uninstall Script (Windows)      " -ForegroundColor Magenta
    Write-Host "================================================" -ForegroundColor Magenta
    Write-Host ""

    $binaryPath = Join-Path $INSTALL_DIR "$BINARY_NAME.exe"
    $foundBinary = $false
    $foundConfig = $false

    # Check binary file
    if (Test-Path $binaryPath) {
        $foundBinary = $true
        Write-Info "Found binary: $binaryPath"
    }
    else {
        Write-Warn "Binary not found: $binaryPath"
        # Try to find in PATH
        $otherPath = Get-Command $BINARY_NAME -ErrorAction SilentlyContinue
        if ($otherPath) {
            Write-Warn "Found $BINARY_NAME at other location: $($otherPath.Source)"
            Write-Host "To remove, manually run: Remove-Item $($otherPath.Source)" -ForegroundColor Gray
        }
    }

    # Check config directory
    if (Test-Path $CONFIG_DIR) {
        $foundConfig = $true
        Write-Info "Found config directory: $CONFIG_DIR"
    }

    # If nothing found
    if (-not $foundBinary -and -not $foundConfig) {
        Write-Warn "$BINARY_NAME does not appear to be installed"
        return
    }

    # Show what will be deleted
    Write-Host ""
    Write-Host "The following will be deleted:" -ForegroundColor White
    if ($foundBinary) {
        Write-Host "  - Binary: $binaryPath" -ForegroundColor Gray
    }
    if ($foundConfig) {
        Write-Host "  - Config directory: $CONFIG_DIR (includes all config and data)" -ForegroundColor Gray
    }
    Write-Host ""

    # Confirm uninstall
    $confirm = Read-Host "Confirm uninstall? [y/N]"
    if ($confirm -ne 'y' -and $confirm -ne 'Y') {
        Write-Info "Uninstall cancelled"
        return
    }

    # Delete binary
    if ($foundBinary) {
        Write-Info "Deleting binary..."
        Remove-Item $binaryPath -Force
        Write-Success "Deleted: $binaryPath"
    }

    # Ask about config deletion
    if ($foundConfig) {
        Write-Host ""
        $confirmConfig = Read-Host "Also delete config directory ($CONFIG_DIR)? [y/N]"
        if ($confirmConfig -eq 'y' -or $confirmConfig -eq 'Y') {
            Write-Info "Deleting config directory..."
            Remove-Item $CONFIG_DIR -Recurse -Force
            Write-Success "Deleted: $CONFIG_DIR"
        }
        else {
            Write-Info "Keeping config directory: $CONFIG_DIR"
        }
    }

    Write-Host ""
    Write-Success "$BINARY_NAME uninstall complete!"
    Write-Host ""
    Write-Host "If you installed via Scoop, use:" -ForegroundColor White
    Write-Host "  scoop uninstall $BINARY_NAME" -ForegroundColor Cyan
}

Uninstall-Xbuilder
