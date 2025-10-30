# Installation Guide

This guide covers building auxbox from source and installing it on various platforms.

## Prerequisites

- **Go 1.25.1 or later** - [Download from golang.org](https://golang.org/dl/)
- **Git** - For cloning the repository

## Build from Source

### 1. Clone the Repository

```bash
git clone https://github.com/cerberussg/auxbox
cd auxbox
```

### 2. Build the Binary

```bash
go build -o auxbox cmd/auxbox/*.go
```

This creates an `auxbox` executable in the current directory.

## Platform-Specific Installation

After building, you need to move the binary to a location in your system PATH.

### Linux (including Arch Linux)

**Option 1: User Local Installation (Recommended)**

```bash
# Create local bin directory if it doesn't exist
mkdir -p ~/.local/bin

# Move the binary
mv auxbox ~/.local/bin/

# Add to PATH (if not already configured)
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Option 2: System-Wide Installation**

```bash
# Requires sudo privileges
sudo mv auxbox /usr/local/bin/

# Verify installation
which auxbox
```

### macOS

**Option 1: User Local Installation (Recommended)**

```bash
# Create user bin directory
mkdir -p ~/bin

# Move the binary
mv auxbox ~/bin/

# Add to PATH for zsh (default on modern macOS)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Or for bash users:
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Option 2: System-Wide Installation**

```bash
# Requires sudo privileges
sudo mv auxbox /usr/local/bin/

# Verify installation
which auxbox
```

### Windows

**Option 1: User Local Installation**

```cmd
REM Move to Windows Apps directory
move auxbox.exe %USERPROFILE%\AppData\Local\Microsoft\WindowsApps\

REM Verify installation
where auxbox
```

**Option 2: Custom Directory**

```cmd
REM Create dedicated directory
mkdir C:\auxbox
move auxbox.exe C:\auxbox\

REM Add C:\auxbox to PATH:
REM 1. Open System Properties (Win + Pause/Break)
REM 2. Click "Advanced system settings"
REM 3. Click "Environment Variables"
REM 4. Under "User variables" or "System variables", find "Path"
REM 5. Click "Edit" and add "C:\auxbox"
REM 6. Click OK to save

REM Restart your terminal and verify
where auxbox
```

## Cross-Platform Builds

You can build auxbox for different platforms from any system:

### Build for Linux (amd64)
```bash
GOOS=linux GOARCH=amd64 go build -o auxbox-linux cmd/auxbox/*.go
```

### Build for macOS (amd64)
```bash
GOOS=darwin GOARCH=amd64 go build -o auxbox-macos cmd/auxbox/*.go
```

### Build for macOS (ARM64/Apple Silicon)
```bash
GOOS=darwin GOARCH=arm64 go build -o auxbox-macos-arm64 cmd/auxbox/*.go
```

### Build for Windows (amd64)
```bash
GOOS=windows GOARCH=amd64 go build -o auxbox.exe cmd/auxbox/*.go
```

## Verification

After installation, verify auxbox is working:

```bash
# Check version
auxbox --version

# Display help
auxbox --help
```

## Updating

To update auxbox to the latest version:

```bash
# Navigate to the repository
cd auxbox

# Pull latest changes
git pull origin master

# Rebuild
go build -o auxbox cmd/auxbox/*.go

# Move to your installation location (example for Linux)
mv auxbox ~/.local/bin/
```

## Uninstallation

To remove auxbox from your system:

```bash
# Linux/macOS (user local installation)
rm ~/.local/bin/auxbox

# Linux/macOS (system-wide installation)
sudo rm /usr/local/bin/auxbox

# Windows (user local)
del %USERPROFILE%\AppData\Local\Microsoft\WindowsApps\auxbox.exe

# Windows (custom directory)
del C:\auxbox\auxbox.exe
rmdir C:\auxbox
```

## Troubleshooting

### "auxbox: command not found"

Your PATH is not configured correctly. Verify the binary location is in your PATH:

```bash
# Linux/macOS
echo $PATH

# Windows
echo %PATH%
```

Ensure the directory containing auxbox appears in the output.

### Permission Denied (Linux/macOS)

If you encounter permission errors, ensure the binary is executable:

```bash
chmod +x ~/.local/bin/auxbox
```

### Build Errors

Ensure you have Go 1.25.1 or later installed:

```bash
go version
```

If your Go version is outdated, update from [golang.org](https://golang.org/dl/).

## Next Steps

Once installed, see the [User Guide](USER_GUIDE.md) to start using auxbox.
