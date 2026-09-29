#!/usr/bin/env bash
#
# Render deploy/nginx-app.conf.template for a given folder + domain.
#
#   ./deploy/render-nginx.sh <folder> <server_name_or_ip> [webroot] [php_sock]
#
#   folder         app folder name, any name, no slashes  (cars, mig_26, acme)
#   server_name    domain (TLS via Let's Encrypt) or a raw IP (see note below)
#   webroot        default /var/www/<server_name>
#   php_sock       default: newest /run/php/php*-fpm.sock found on THIS machine
#
# Writes the rendered file to stdout and, with --write, to
# deploy/out/nginx-<folder>.conf for review before installing.
#
# NOTE on a raw IP: the template hardcodes a Let's Encrypt certificate path.
# Let's Encrypt will not issue for a bare IP, and https://IP/ requires a cert
# with that IP as a SAN. For an IP deploy, install the rendered file, then change
# the second server block to `listen 80;` only and drop the ssl_* lines — or
# obtain a cert that covers the IP.

set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

usage() {
  sed -n '3,20p' "$0" | sed 's/^# \{0,1\}//'
  exit 2
}

[ $# -ge 2 ] || usage

FOLDER=$1
SERVER_NAME=$2
WEBROOT=${3:-}
PHP_SOCK=${4:-}
WRITE=0

for arg in "$@"; do
  [ "$arg" = "--write" ] && WRITE=1
done

# --- validate the folder name -------------------------------------------------
# It is pasted into nginx config verbatim, so anything that could terminate the
# token and start a new directive must be rejected. This is the check that stops
# a bad client-supplied name from silently disabling the deny-all rules.
# Single allowlist, so the accepted set is defined in exactly one place. A leading
# dot is excluded on purpose: a dot-prefixed folder would be swallowed by the
# `location ~ /\.` deny rule and 403 the whole app.
if ! printf '%s' "$FOLDER" | grep -Eq '^[A-Za-z0-9_][A-Za-z0-9._-]*$'; then
  die "invalid folder name: '$FOLDER' (start with a letter, digit or _;" \
      "then letters, digits, . - and _; no slashes or spaces)"
fi

case "$SERVER_NAME" in
  *[[:space:]]*|*';'*|*'{'*|*'}'*|*'#'*) die "invalid server_name: '$SERVER_NAME'" ;;
esac

# --- defaults that depend on this machine -------------------------------------
[ -n "$WEBROOT" ] || WEBROOT="/var/www/$SERVER_NAME"

if [ -z "$PHP_SOCK" ]; then
  PHP_SOCK=$(ls -1 /run/php/php*-fpm.sock 2>/dev/null | sort -V | tail -1 || true)
  [ -n "$PHP_SOCK" ] || die "no /run/php/php*-fpm.sock found; pass the socket as arg 4"
fi

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
TEMPLATE="$SCRIPT_DIR/nginx-app.conf.template"
[ -f "$TEMPLATE" ] || die "template not found: $TEMPLATE"

# Substitute with | as the delimiter: folder names may contain _ and . but never
# | (rejected above), so no escaping is needed.
sed -e "s|__FOLDER__|$FOLDER|g" \
    -e "s|__WEBROOT__|$WEBROOT|g" \
    -e "s|__SERVER_NAME__|$SERVER_NAME|g" \
    -e "s|__PHP_SOCK__|$PHP_SOCK|g" \
    "$TEMPLATE"

# Belt and braces: a leftover placeholder means the template and this script have
# drifted apart, and the output would contain literal __FOLDER__.
if [ "$WRITE" -eq 1 ]; then
  mkdir -p "$SCRIPT_DIR/out"
  OUT="$SCRIPT_DIR/out/nginx-$FOLDER.conf"
  sed -e "s|__FOLDER__|$FOLDER|g" \
      -e "s|__WEBROOT__|$WEBROOT|g" \
      -e "s|__SERVER_NAME__|$SERVER_NAME|g" \
      -e "s|__PHP_SOCK__|$PHP_SOCK|g" \
      "$TEMPLATE" > "$OUT"
  if grep -q '__[A-Z_]*__' "$OUT"; then
    die "unsubstituted placeholder left in $OUT"
  fi
  echo "wrote $OUT" >&2
  if command -v nginx >/dev/null 2>&1; then
    nginx -t 2>&1 | sed 's/^/  /' >&2 || true
  fi
fi
