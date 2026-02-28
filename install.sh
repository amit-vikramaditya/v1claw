#!/usr/bin/env bash

# V1Claw Quick Installer Shell Script
# This script compiles the V1Claw binary and automatically adds it to the system PATH.

set -e

echo "🚀 Starting V1Claw Installation..."

# 1. Check for Go
if ! command -v go &> /dev/null; then
    echo "❌ Error: Go is not installed. V1Claw requires Go to compile."
    echo "Please visit https://go.dev/doc/install to install Go."
    exit 1
fi

# 2. Build the binary using the Makefile
echo "🔨 Compiling binary..."
make build

# 3. Determine Installation Directory
INSTALL_DIR="$HOME/.local/bin"
mkdir -p "$INSTALL_DIR"

# Check if the build succeeded and binary exists
if [ ! -f "build/v1claw" ]; then
    echo "❌ Error: Compilation failed. 'build/v1claw' not found."
    exit 1
fi

# 4. Move the binary to the local bin
echo "📦 Installing V1Claw to $INSTALL_DIR..."
cp build/v1claw "$INSTALL_DIR/v1claw"
chmod +x "$INSTALL_DIR/v1claw"

# 5. Check if INSTALL_DIR is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo "⚠️  $INSTALL_DIR is not in your PATH."
    
    # Try to detect the user's shell config
    SHELL_RC=""
    if [[ "$SHELL" == *"zsh"* ]]; then
        SHELL_RC="$HOME/.zshrc"
    elif [[ "$SHELL" == *"bash"* ]]; then
        if [ -f "$HOME/.bashrc" ]; then
            SHELL_RC="$HOME/.bashrc"
        elif [ -f "$HOME/.bash_profile" ]; then
            SHELL_RC="$HOME/.bash_profile"
        fi
    fi

    if [ -n "$SHELL_RC" ]; then
        echo "Adding $INSTALL_DIR to your PATH in $SHELL_RC..."
        echo -e "\n# V1Claw Path" >> "$SHELL_RC"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\"" >> "$SHELL_RC"
        echo "✅ PATH updated! Please run 'source $SHELL_RC' or restart your terminal."
    else
        echo "❌ Could not automatically detect your shell configuration file."
        echo "Please manually add the following line to your shell profile (e.g., ~/.bashrc or ~/.zshrc):"
        echo "export PATH=\"\$PATH:$INSTALL_DIR\""
    fi
else
    echo "✅ $INSTALL_DIR is already in your PATH."
fi

echo ""
echo "🎉 V1Claw is successfully installed!"
echo "Run 'v1claw onboard' to begin your setup."
