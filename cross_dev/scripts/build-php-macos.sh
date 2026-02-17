#!/usr/bin/env bash
# Build a portable PHP CLI for macOS with MySQL support (pdo_mysql, mysqli).
# Uses static-php-cli (https://static-php.dev/) to produce a single binary.
#
# Usage:
#   ./build-php-macos.sh [OUTPUT_DIR]
#
# Default OUTPUT_DIR: cross_dev/bundle/php-macos/
# Result: OUTPUT_DIR/php (executable), OUTPUT_DIR/php.ini (optional)
#
# Requires: curl, tar (or xz), Xcode Command Line Tools (clang, make).
# Optional: composer (if building spc from source instead of using the binary).

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CROSS_DEV_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
OUTPUT_DIR="${1:-$CROSS_DEV_DIR/bundle/php-macos}"
BUILD_DIR="$CROSS_DEV_DIR/build-php-static"
SPC_DIR="$BUILD_DIR/static-php-cli"

# Extensions needed for Cars API: MySQL (PDO + mysqli), JSON, cURL, OpenSSL, etc.
# mysqlnd is required for pdo_mysql and mysqli.
EXTENSIONS="bcmath,curl,fileinfo,json,mbstring,mysqlnd,mysqli,openssl,pdo,pdo_mysql,phar,posix,tokenizer,zlib"

echo "=== Portable PHP for macOS (with MySQL) ==="
echo "  Output directory: $OUTPUT_DIR"
echo "  Build directory:  $BUILD_DIR"
echo "  Extensions:       $EXTENSIONS"
echo ""

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  arm64)  SPC_ARCH="aarch64" ;;
  x86_64) SPC_ARCH="x86_64"  ;;
  *)      echo "Unsupported arch: $ARCH"; exit 1 ;;
esac
echo "  Detected arch:    $ARCH (spc: $SPC_ARCH)"
echo ""

mkdir -p "$BUILD_DIR"
cd "$BUILD_DIR"

# --- 1. Get static-php-cli binary (spc) ---
# Manual download (if curl fails): open URL in browser, save as build-php-static/spc, chmod +x spc
#   Apple Silicon: https://dl.static-php.dev/static-php-cli/spc-bin/nightly/spc-macos-aarch64
#   Intel:         https://dl.static-php.dev/static-php-cli/spc-bin/nightly/spc-macos-x86_64
SPC_URL="https://dl.static-php.dev/static-php-cli/spc-bin/nightly/spc-macos-${SPC_ARCH}"
SPC_BIN="$BUILD_DIR/spc"

download_spc() {
  echo ">>> Downloading static-php-cli (spc) for macOS $SPC_ARCH ..."
  curl -fsSL --retry 3 --retry-delay 5 -o "$SPC_BIN" "$SPC_URL"
  chmod +x "$SPC_BIN"
  # Remove quarantine so Gatekeeper doesn't block "unidentified developer"
  xattr -d com.apple.quarantine "$SPC_BIN" 2>/dev/null || true
}

if [[ ! -f "$SPC_BIN" ]] || [[ ! -x "$SPC_BIN" ]]; then
  download_spc
elif ! "$SPC_BIN" --version &>/dev/null; then
  echo ">>> Existing spc binary invalid or incomplete; re-downloading ..."
  rm -f "$SPC_BIN"
  download_spc
else
  echo ">>> Using existing spc binary."
fi
# Clear quarantine so Gatekeeper doesn't block (e.g. after manual download)
xattr -d com.apple.quarantine "$SPC_BIN" 2>/dev/null || true
./spc --version || true
echo ""

# --- 2. Download PHP and dependency sources ---
if [[ ! -d "$BUILD_DIR/source/php-src" ]]; then
  echo ">>> Downloading PHP and dependencies (this may take a few minutes) ..."
  ./spc download --for-extensions="${EXTENSIONS}" --with-php=8.3
  echo "    Done."
else
  echo ">>> Source already present; skipping download. (Delete $BUILD_DIR/source to re-download.)"
fi
echo ""

# --- 3. Build PHP CLI with extensions ---
if [[ ! -f "$BUILD_DIR/buildroot/bin/php" ]]; then
  echo ">>> Building PHP CLI (this will take a while) ..."
  ./spc build "${EXTENSIONS}" --build-cli --debug 2>&1 | tee "$BUILD_DIR/build.log" || true
  if [[ ! -f "$BUILD_DIR/buildroot/bin/php" ]]; then
    echo "    Build failed. Check $BUILD_DIR/build.log"
    exit 1
  fi
  echo "    Done."
else
  echo ">>> buildroot/bin/php already exists; skipping build."
fi
echo ""

# --- 4. Verify and copy to output ---
PHP_BIN="$BUILD_DIR/buildroot/bin/php"
"$PHP_BIN" -v
echo ""
echo ">>> Checking extensions ..."
"$PHP_BIN" -m | grep -E "^(curl|json|mbstring|mysqli|pdo_mysql|openssl|phar|zlib)" || true
echo ""

mkdir -p "$OUTPUT_DIR"
cp -f "$PHP_BIN" "$OUTPUT_DIR/php"
chmod +x "$OUTPUT_DIR/php"

# Optional: minimal php.ini for built-in server (e.g. memory limit, date zone)
if [[ ! -f "$OUTPUT_DIR/php.ini" ]]; then
  cat > "$OUTPUT_DIR/php.ini" << 'INI'
; Portable PHP for CrossDev (macOS)
memory_limit = 256M
date.timezone = UTC
INI
  echo ">>> Created $OUTPUT_DIR/php.ini"
fi

echo "=== Done ==="
echo "  PHP binary:  $OUTPUT_DIR/php"
echo "  Config:      $OUTPUT_DIR/php.ini"
echo ""
echo "  Test run:    $OUTPUT_DIR/php -S 127.0.0.1:8765 -t /path/to/api"
echo ""
