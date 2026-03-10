#!/usr/bin/env bash
# V1Claw installer
# Downloads a pre-built binary from GitHub Releases — no Go required.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/amit-vikramaditya/V1Claw/main/install.sh | bash
#
# Or with a specific version:
#   curl -fsSL .../install.sh | bash -s -- --version v1.2.0

set -euo pipefail

REPO="amit-vikramaditya/V1Claw"
BINARY="v1claw"
INSTALL_DIR="${INSTALL_DIR:-}"   # can be overridden by environment

# ── helpers ──────────────────────────────────────────────────────────────────
info()  { printf "  \033[34m→\033[0m  %s\n" "$*"; }
ok()    { printf "  \033[32m✓\033[0m  %s\n" "$*"; }
warn()  { printf "  \033[33m⚠\033[0m  %s\n" "$*"; }
fail()  { printf "  \033[31m✗\033[0m  %s\n" "$*"; exit 1; }

# ── detect OS + arch ─────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
  Darwin) GOOS="Darwin"  ;;
  Linux)  GOOS="Linux"   ;;
  *)      fail "Unsupported OS: $OS. Download a binary manually from https://github.com/$REPO/releases" ;;
esac

case "$ARCH" in
  x86_64)          GOARCH="x86_64" ;;
  arm64|aarch64)   GOARCH="arm64"  ;;
  armv7l|armv6l)   GOARCH="armv7"  ;;
  *)               fail "Unsupported architecture: $ARCH. Download a binary manually from https://github.com/$REPO/releases" ;;
esac

# ── resolve version ───────────────────────────────────────────────────────────
VERSION=""
for arg in "$@"; do
  case "$arg" in
    --version=*) VERSION="${arg#--version=}" ;;
    --version)   shift; VERSION="$1" ;;
  esac
done

if [ -z "$VERSION" ]; then
  info "Fetching latest release version…"
  if command -v curl &>/dev/null; then
    VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
      | grep '"tag_name"' | cut -d'"' -f4)
  elif command -v wget &>/dev/null; then
    VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" \
      | grep '"tag_name"' | cut -d'"' -f4)
  else
    fail "Neither curl nor wget found. Install one of them and retry."
  fi
fi

if [ -z "$VERSION" ]; then
  fail "Could not determine the latest version. Pass --version vX.Y.Z explicitly."
fi

ok "Version: $VERSION"

# ── build download URL ────────────────────────────────────────────────────────
ARCHIVE="${BINARY}_${GOOS}_${GOARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"

# ── install location ──────────────────────────────────────────────────────────
if [ -z "$INSTALL_DIR" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
    mkdir -p "$INSTALL_DIR"
  fi
fi

# ── download + extract ────────────────────────────────────────────────────────
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

info "Downloading $ARCHIVE…"
if command -v curl &>/dev/null; then
  curl -fsSL "$URL" -o "$TMP/$ARCHIVE" || fail "Download failed. Check https://github.com/$REPO/releases for available files."
else
  wget -qO "$TMP/$ARCHIVE" "$URL"       || fail "Download failed. Check https://github.com/$REPO/releases for available files."
fi

info "Extracting…"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

if [ ! -f "$TMP/$BINARY" ]; then
  fail "Binary '$BINARY' not found in the archive. This is unexpected — please report it."
fi

chmod +x "$TMP/$BINARY"
mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
ok "Installed to $INSTALL_DIR/$BINARY"

# ── PATH check ────────────────────────────────────────────────────────────────
if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
  warn "$INSTALL_DIR is not in your PATH."

  SHELL_RC=""
  case "${SHELL:-}" in
    */zsh)  SHELL_RC="$HOME/.zshrc" ;;
    */bash)
      if   [ -f "$HOME/.bashrc" ];       then SHELL_RC="$HOME/.bashrc"
      elif [ -f "$HOME/.bash_profile" ]; then SHELL_RC="$HOME/.bash_profile"
      fi ;;
    */fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
  esac

  if [ -n "$SHELL_RC" ]; then
    printf '\n# V1Claw\nexport PATH="$PATH:%s"\n' "$INSTALL_DIR" >> "$SHELL_RC"
    ok "Added $INSTALL_DIR to PATH in $SHELL_RC"
    warn "Run: source $SHELL_RC   (or open a new terminal)"
  else
    warn "Add this to your shell profile manually:"
    printf '    export PATH="$PATH:%s"\n' "$INSTALL_DIR"
  fi
fi

# ── done ─────────────────────────────────────────────────────────────────────
echo ""
echo "  ✅  V1Claw $VERSION is installed!"
echo ""
echo "  Next step — run the 2-minute setup wizard:"
echo ""
echo "    v1claw onboard"
echo ""
echo "  Or silent setup (replace YOUR_KEY):"
echo ""
echo "    v1claw onboard --auto --provider gemini --api-key YOUR_KEY"
echo ""
