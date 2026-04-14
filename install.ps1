# SYNKRO INSTALLER - One-Command Installation Script for Windows
# Copyright (c) 2024 Synkro Project

$VERSION = "1.0.0"
$INSTALL_DIR = "$env:USERPROFILE\.synkro"
$BIN_DIR = "$INSTALL_DIR\bin"
$CONFIG_DIR = "$INSTALL_DIR\config"
$DATA_DIR = "$INSTALL_DIR\data"
$MODELS_DIR = "$INSTALL_DIR\models"
$BINARY_NAME = "synkro.exe"

# Colors
function Write-Color($Message, $Color = "White") {
    Write-Host $Message -ForegroundColor $Color
}

function Banner() {
    Write-Color "╔══════════════════════════════════════╗" -Color "Cyan"
    Write-Color "║                                        ║" -Color "Cyan"
    Write-Color "║        SYNKRO INSTALLER v$VERSION       ║" -Color "Cyan"
    Write-Color "║      Motor de Contexto Inteligente      ║" -Color "Cyan"
    Write-Color "║                                        ║" -Color "Cyan"
    Write-Color "╚══════════════════════════════════════╝" -Color "Cyan"
    Write-Host ""
}

function Check-Dependencies {
    Write-Color "📦 Checking dependencies..." -Color "Cyan"
    
    # Check if Go is installed
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $GO_VERSION = go version | Select-String -Pattern '\d+\.\d+\.\d+'
        Write-Color "✓ Go found: $GO_VERSION" -Color "Green"
    }
    else {
        Write-Color "⚠ Go not found (will use pre-built binary)" -Color "Yellow"
    }
    
    # Check disk space
    $REQUIRED_SPACE = 500 # MB
    $AVAILABLE_SPACE = (Get-PSDrive $env:HOMEDRIVE.Substring(0,1)).Free / 1MB
    
    if ($AVAILABLE_SPACE -lt $REQUIRED_SPACE) {
        Write-Color "✗ Insufficient disk space" -Color "Red"
        Write-Host "  Required: ${REQUIRED_SPACE}MB, Available: ${AVAILABLE_SPACE}MB"
        exit 1
    }
    else {
        Write-Color "✓ Sufficient disk space (${AVAILABLE_SPACE}MB available)" -Color "Green"
    }
}

function Detect-Platform {
    Write-Color "🖥️ Detecting platform..." -Color "Cyan"
    
    $PROCESSOR_ARCH = $env:PROCESSOR_ARCHITECTURE
    
    if ($PROCESSOR_ARCH -eq "AMD64") {
        $PLATFORM = "windows-amd64"
        Write-Color "✓ Detected platform: $PLATFORM" -Color "Green"
    }
    elseif ($PROCESSOR_ARCH -eq "ARM64") {
        $PLATFORM = "windows-arm64"
        Write-Color "✓ Detected platform: $PLATFORM" -Color "Green"
    }
    else {
        Write-Color "✗ Unsupported architecture: $PROCESSOR_ARCH" -Color "Red"
        exit 1
    }
}

function Create-Directories {
    Write-Color "📁 Creating directories..." -Color "Cyan"
    
    New-Item -ItemType Directory -Force -Path $BIN_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $CONFIG_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $DATA_DIR | Out-Null
    New-Item -ItemType Directory -Force -Path $MODELS_DIR | Out-Null
    
    Write-Color "✓ Directories created in $INSTALL_DIR" -Color "Green"
}

function Download-Binary {
    Write-Color "⬇️  Downloading Synkro binary..." -Color "Cyan"
    
    $TEMP_DIR = Join-Path $env:TEMP "synkro"
    $BINARY_PATH = Join-Path $TEMP_DIR $BINARY_NAME
    
    New-Item -ItemType Directory -Force -Path $TEMP_DIR | Out-Null
    
    $RELEASES_URL = "https://github.com/rodascaar/synkro/releases/latest"
    
    # Try to download from releases
    try {
        $DOWNLOAD_URL = "https://github.com/rodascaar/synkro/releases/latest/download/synkro-$PLATFORM.exe"
        
        Write-Host "Downloading from: $DOWNLOAD_URL"
        
        Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $BINARY_PATH -UseBasicParsing
        
        Write-Color "✓ Binary downloaded successfully" -Color "Green"
    }
    catch {
        Write-Color "⚠ Pre-built binary not available" -Color "Yellow"
        Write-Host "  Building from source..."
        Build-From-Source
        return
    }
    
    # Move binary to installation directory
    Move-Item -Force -Path $BINARY_PATH -Destination "$BIN_DIR\$BINARY_NAME"
    
    Write-Color "✓ Binary installed to $BIN_DIR\$BINARY_NAME" -Color "Green"
}

function Build-From-Source {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Color "✗ Go is required to build from source" -Color "Red"
        Write-Host "  Install from: https://go.dev/dl/"
        exit 1
    }
    
    Write-Host "Building Synkro from source..."
    
    $TEMP_DIR = Join-Path $env:TEMP "synkro"
    New-Item -ItemType Directory -Force -Path $TEMP_DIR | Out-Null
    
    Set-Location $TEMP_DIR
    
    # Clone repository
    git clone --depth 1 https://github.com/rodascaar/synkro.git
    
    Set-Location synkro
    
    # Build binary
    $env:CGO_ENABLED = "1"
    go build -o "$BIN_DIR\$BINARY_NAME" .\cmd\synkro\
    
    Write-Color "✓ Synkro built and installed" -Color "Green"
    
    # Cleanup
    Set-Location $env:USERPROFILE
    Remove-Item -Recurse -Force $TEMP_DIR
}

function Setup-Environment {
    Write-Color "⚙️  Setting up environment..." -Color "Cyan"
    
    # Create environment file
    $ENV_FILE = Join-Path $CONFIG_DIR "synkro.ps1"
    
    @"
# Synkro environment configuration
`$ENV_FILE = @"
`$INSTALL_DIR = "$INSTALL_DIR"
`$BIN_DIR = "$BIN_DIR"
`$CONFIG_DIR = "$CONFIG_DIR"
`$DATA_DIR = "$DATA_DIR"
`$MODELS_DIR = "$MODELS_DIR"

# Add to PATH
`$PATH = "$BIN_DIR;`$env:PATH"
"@ | Out-File -FilePath $ENV_FILE -Encoding UTF8
    
    Write-Color "✓ Environment configured" -Color "Green"
}

function Install-PowerShell-Integration {
    Write-Color "🐚 Installing PowerShell integration..." -Color "Cyan"
    
    $PROFILE_PATH = $PROFILE.CurrentUserCurrentHostCurrentLocation
    
    if (-not (Select-String -Path $PROFILE_PATH -Pattern "Synkro" -Quiet)) {
        Add-Content -Path $PROFILE_PATH -Value "`n# Synkro integration`n. `$env:SYNKRO_INIT = `"$ENV_FILE"`n"
        
        Write-Color "✓ PowerShell integration added to $PROFILE_PATH" -Color "Green"
        Write-Color "⚠ Please restart PowerShell to apply changes" -Color "Yellow"
    }
    else {
        Write-Color "⚠ PowerShell integration already exists" -Color "Yellow"
    }
}

function Initialize-Database {
    Write-Color "💾 Initializing database..." -Color "Cyan"
    
    $DB_PATH = Join-Path $DATA_DIR "memory.db"
    
    # Check if database already exists
    if (Test-Path $DB_PATH) {
        Write-Color "⚠ Database already exists at $DB_PATH" -Color "Yellow"
        $RESPONSE = Read-Host "Do you want to reinitialize? (y/N)"
        
        if ($RESPONSE -ne "y" -and $RESPONSE -ne "Y") {
            Write-Color "✓ Keeping existing database" -Color "Green"
            return
        }
    }
    
    # Initialize database
    $SYNKRO_BIN = Join-Path $BIN_DIR $BINARY_NAME
    & $SYNKRO_BIN init
    
    if ($LASTEXITCODE -eq 0) {
        Write-Color "✓ Database initialized" -Color "Green"
    }
    else {
        Write-Color "✗ Failed to initialize database" -Color "Red"
        exit 1
    }
}

function Download-Default-Model {
    Write-Color "🤖 Setting up embedding model..." -Color "Cyan"
    
    Write-Host ""
    Write-Host "Synkro can use semantic embeddings for better search accuracy."
    Write-Host "Do you want to download the default embedding model (all-MiniLM-L6-v2)?"
    Write-Host ""
    
    $RESPONSE = Read-Host "Download default model? (Y/n)"
    
    if ($RESPONSE -eq "n" -or $RESPONSE -eq "N") {
        Write-Color "⚠ Skipping model download" -Color "Yellow"
        Write-Host "  You can download models later with: synkro model download <model-name>"
        return
    }
    
    # Download default model
    Write-Host "Downloading default model..."
    
    $SYNKRO_BIN = Join-Path $BIN_DIR $BINARY_NAME
    & $SYNKRO_BIN model download all-MiniLM-L6-v2
    
    if ($LASTEXITCODE -eq 0) {
        Write-Color "✓ Default model installed" -Color "Green"
    }
    else {
        Write-Color "⚠ Model download failed, but Synkro is ready to use" -Color "Yellow"
        Write-Host "  You can download models later with: synkro model list"
    }
}

function Print-Success {
    Write-Host ""
    Write-Color "╔══════════════════════════════════════╗" -Color "Green"
    Write-Color "║       SYNKRO INSTALLED SUCCESSFULLY!      ║" -Color "Green"
    Write-Color "╚══════════════════════════════════════╝" -Color "Green"
    Write-Host ""
    Write-Color "🚀 Getting Started:" -Color "Cyan"
    Write-Host ""
    Write-Host "1. Open a new PowerShell window or restart current one"
    Write-Host "2. Initialize database (if not already done):"
    Write-Color "   synkro init" -Color "Yellow"
    Write-Host ""
    Write-Host "3. Add your first memory:"
    Write-Color "   synkro add --title `"My first memory`" --content `"Hello Synkro!`" --type note" -Color "Yellow"
    Write-Host ""
    Write-Host "4. Start TUI:"
    Write-Color "   synkro tui" -Color "Yellow"
    Write-Host ""
    Write-Color "📚 Documentation:" -Color "Cyan"
    Write-Host "  Visit: https://github.com/rodascaar/synkro/wiki"
    Write-Host ""
    Write-Color "💡 Tips:" -Color "Cyan"
    Write-Host "  - Use 'synkro --help' for all commands"
    Write-Host "  - Check available models with 'synkro model list'"
    Write-Host "  - Run MCP server with 'synkro mcp'"
    Write-Host ""
}

# Main installation flow
Banner
Check-Dependencies
Detect-Platform
Create-Directories
Download-Binary
Setup-Environment
Install-PowerShell-Integration
Initialize-Database
Download-Default-Model
Print-Success
