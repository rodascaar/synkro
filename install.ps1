# Synkro PowerShell Installer v1.0
# Supports: Windows 10+
param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

$REPO = "rodascaar/synkro"
$VERSION = "latest"
$INSTALL_DIR = "$env:USERPROFILE\.local\bin"
$BINARY_NAME = "synkro.exe"

Write-Host "🚀 Installing Synkro v1..." -ForegroundColor Cyan

# Detectar arquitectura
$ARCH = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "windows-arm64" } else { "windows-amd64" }
Write-Host "📦 Detected platform: $ARCH" -ForegroundColor Green

$DOWNLOAD_URL = "https://github.com/$REPO/releases/latest/download/synkro-$ARCH"
$TEMP_DIR = $env:TEMP

# Crear directorio temporal
$TEMP_PATH = Join-Path $TEMP_DIR "synkro-install"
New-Item -ItemType Directory -Path $TEMP_PATH -Force | Out-Null

# Intentar descargar binario
Write-Host "📥 Downloading Synkro..." -ForegroundColor Yellow
try {
    Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile "$TEMP_PATH\synkro.exe" -UseBasicParsing
    Write-Host "✅ Binary downloaded successfully" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Pre-built binary not available" -ForegroundColor Yellow
    Write-Host "🔨 Compiling from source instead..." -ForegroundColor Yellow

    # Fallback a compilación
    if (!(Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Host "❌ Go not installed" -ForegroundColor Red
        Write-Host "Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
        exit 1
    }

    $SOURCE_DIR = Join-Path $TEMP_PATH "synkro-source"
    git clone --depth 1 "https://github.com/$REPO.git" $SOURCE_DIR

    Push-Location $SOURCE_DIR
    $env:CGO_ENABLED = "1"
    go build -tags sqlite_fts5 -ldflags="-s -w" -o "$TEMP_PATH\synkro.exe" .\cmd\synkro\
    Pop-Location

    Remove-Item -Recurse -Force $SOURCE_DIR
}

# Crear directorio de instalación
if (!(Test-Path $INSTALL_DIR)) {
    New-Item -ItemType Directory -Path $INSTALL_DIR -Force | Out-Null
}

# Instalar binario
Write-Host "📦 Installing to $INSTALL_DIR..." -ForegroundColor Yellow
Copy-Item -Force "$TEMP_PATH\synkro.exe" "$INSTALL_DIR\synkro.exe"

# Agregar al PATH
$env:PATH = "$INSTALL_DIR;$env:PATH"
[System.Environment]::SetEnvironmentVariable('Path', "$env:PATH", 'User')

# Limpiar
Remove-Item -Recurse -Force $TEMP_PATH

# Verificar
if (Test-Path "$INSTALL_DIR\synkro.exe") {
    Write-Host ""
    Write-Host "✅ Synkro installed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🎯 Quick Start:" -ForegroundColor Cyan
    Write-Host "   synkro.exe init              # Initialize database" -ForegroundColor White
    Write-Host "   synkro.exe add --help        # Add your first memory" -ForegroundColor White
    Write-Host "   synkro.exe tui               # Launch TUI" -ForegroundColor White
    Write-Host "   synkro.exe mcp               # Start MCP server" -ForegroundColor White
    Write-Host ""
    Write-Host "📚 Documentation:" -ForegroundColor Cyan
    Write-Host "   https://github.com/$REPO" -ForegroundColor White
} else {
    Write-Host "❌ Installation failed" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🎉 Installation complete!" -ForegroundColor Green
