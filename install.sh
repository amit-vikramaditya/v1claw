#!/usr/bin/env bash
# V1Claw installer
# Downloads a pre-built binary from GitHub Releases — no Go required.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/amit-vikramaditya/v1claw/main/install.sh | bash
#
# Or with a specific version:
#   curl -fsSL .../install.sh | bash -s -- --version v1.2.0

set -euo pipefail

REPO="amit-vikramaditya/v1claw"
BINARY="v1claw"
INSTALL_DIR="${INSTALL_DIR:-}"   # can be overridden by environment
INSTALL_DIR_EXPLICIT=0
if [ -n "$INSTALL_DIR" ]; then
  INSTALL_DIR_EXPLICIT=1
fi

# ── helpers ──────────────────────────────────────────────────────────────────
info()  { printf "  \033[34m→\033[0m  %s\n" "$*"; }
ok()    { printf "  \033[32m✓\033[0m  %s\n" "$*"; }
warn()  { printf "  \033[33m⚠\033[0m  %s\n" "$*"; }
fail()  { printf "  \033[31m✗\033[0m  %s\n" "$*"; exit 1; }

download_to() {
  local url="$1"
  local output="$2"
  if command -v curl &>/dev/null; then
    curl -fsSL "$url" -o "$output"
  elif command -v wget &>/dev/null; then
    wget -qO "$output" "$url"
  else
    fail "Neither curl nor wget found. Install one of them and retry."
  fi
}

fetch_latest_release_version() {
  local api="https://api.github.com/repos/$REPO/releases/latest"
  if command -v curl &>/dev/null; then
    curl -fsSL "$api" | grep '"tag_name"' | cut -d'"' -f4
  elif command -v wget &>/dev/null; then
    wget -qO- "$api" | grep '"tag_name"' | cut -d'"' -f4
  else
    return 1
  fi
}

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
FORCE_SOURCE=0
for arg in "$@"; do
  case "$arg" in
    --version=*) VERSION="${arg#--version=}" ;;
    --version)   shift; VERSION="$1" ;;
    --source)    FORCE_SOURCE=1 ;;
  esac
done

if [ "$FORCE_SOURCE" -ne 1 ] && [ "${V1CLAW_INSTALL_MODE:-}" != "source" ] && [ -z "$VERSION" ]; then
  info "Fetching latest release version…"
  VERSION="$(fetch_latest_release_version || true)"
fi

INSTALL_MODE="release"
if [ "$FORCE_SOURCE" -eq 1 ] || [ "${V1CLAW_INSTALL_MODE:-}" = "source" ]; then
  INSTALL_MODE="source"
elif [ -z "$VERSION" ]; then
  INSTALL_MODE="source"
  warn "No published GitHub release was found for $REPO."
  if ! command -v go &>/dev/null; then
    fail "No release is available and Go is not installed. Build from source manually after installing Go, or publish a release first."
  fi
  warn "Falling back to a source build because Go is installed."
else
  ok "Version: $VERSION"
fi

if [ "$INSTALL_MODE" = "source" ] && ! command -v go &>/dev/null; then
  fail "Source installation requires Go, but 'go' was not found in PATH."
fi

# ── install location ──────────────────────────────────────────────────────────
if [ -z "$INSTALL_DIR" ]; then
  if [ -w "/usr/local/bin" ]; then
    INSTALL_DIR="/usr/local/bin"
  else
    INSTALL_DIR="$HOME/.local/bin"
  fi
fi
mkdir -p "$INSTALL_DIR"

# ── download + extract ────────────────────────────────────────────────────────
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

if [ "$INSTALL_MODE" = "release" ]; then
  ARCHIVE="${BINARY}_${GOOS}_${GOARCH}.tar.gz"
  URL="https://github.com/$REPO/releases/download/$VERSION/$ARCHIVE"

  info "Downloading $ARCHIVE…"
  download_to "$URL" "$TMP/$ARCHIVE" || fail "Download failed. Check https://github.com/$REPO/releases for available files."

  info "Extracting…"
  tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

  if [ ! -f "$TMP/$BINARY" ]; then
    fail "Binary '$BINARY' not found in the archive. This is unexpected — please report it."
  fi
else
  SOURCE_DESC="source build"
  if [ -f "./go.mod" ] && [ -d "./cmd/$BINARY" ]; then
    SOURCE_DIR="$(pwd)"
    SOURCE_DESC="local source build"
    info "Using local source checkout at $SOURCE_DIR…"
  else
    SOURCE_ARCHIVE="$TMP/source.tar.gz"
    SOURCE_DESC="source build from main"
    info "Downloading source archive…"
    download_to "https://github.com/$REPO/archive/refs/heads/main.tar.gz" "$SOURCE_ARCHIVE" || fail "Source download failed."

    info "Extracting source archive…"
    tar -xzf "$SOURCE_ARCHIVE" -C "$TMP"

    SOURCE_DIR="$(find "$TMP" -maxdepth 1 -type d -name 'v1claw-*' | head -n 1)"
    [ -n "$SOURCE_DIR" ] || fail "Could not locate extracted source directory."
  fi

  info "Building from source with $(go version | awk '{print $3}')…"
  (
    cd "$SOURCE_DIR"
    go build -o "$TMP/$BINARY" ./cmd/$BINARY
  ) || fail "Source build failed."
fi

chmod +x "$TMP/$BINARY"
mv "$TMP/$BINARY" "$INSTALL_DIR/$BINARY"
ok "Installed to $INSTALL_DIR/$BINARY"

# ── PATH check ────────────────────────────────────────────────────────────────
case ":$PATH:" in
  *":$INSTALL_DIR:"*) PATH_HAS_INSTALL_DIR=1 ;;
  *) PATH_HAS_INSTALL_DIR=0 ;;
esac

if [ "$PATH_HAS_INSTALL_DIR" -ne 1 ]; then
  warn "$INSTALL_DIR is not in your PATH."

  if [ "$INSTALL_DIR_EXPLICIT" -eq 1 ]; then
    warn "INSTALL_DIR was set explicitly, so your shell profile was not modified."
    warn "Add it to PATH manually if you want to run 'v1claw' without the full path."
  else

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
      if ! grep -Fq "$INSTALL_DIR" "$SHELL_RC" 2>/dev/null; then
        printf '\n# V1Claw\nexport PATH="$PATH:%s"\n' "$INSTALL_DIR" >> "$SHELL_RC"
        ok "Added $INSTALL_DIR to PATH in $SHELL_RC"
      else
        ok "$INSTALL_DIR already exists in $SHELL_RC"
      fi
      warn "Run: source $SHELL_RC   (or open a new terminal)"
    else
      warn "Add this to your shell profile manually:"
      printf '    export PATH="$PATH:%s"\n' "$INSTALL_DIR"
    fi
  fi
fi

# ── done ─────────────────────────────────────────────────────────────────────
echo ""
if [ "$INSTALL_MODE" = "release" ]; then
  echo "  ✅  V1Claw $VERSION is installed!"
else
  echo "  ✅  V1Claw ($SOURCE_DESC) is installed!"
fi
echo ""
echo "  Next step — run the 2-minute setup wizard:"
echo ""
echo "    v1claw onboard"
echo ""
echo "  Or silent setup (replace YOUR_KEY):"
echo ""
echo "    v1claw onboard --auto --provider gemini --api-key YOUR_KEY"
echo ""
echo "  For CI or offline setup, add:"
echo ""
echo "    --skip-test"
echo ""
