#!/usr/bin/env bash
#
# Deploy the Cars app to a client server.
#
#   ./deploy/deploy.sh <folder> <host> <db_code> [user@server] [webroot]
#   ./deploy/deploy.sh <folder> <host> <db_code> --dry-run
#
#   folder     app folder name, any name (cars, mig_26, acme)
#   host       domain or IP the app is served from
#   db_code    the database THIS client uses, e.g. db_9a7f4e0b2fa8134e0ea0.
#              Written to <folder>/db_code.json on the server, so one dist/ can
#              serve clients on different databases.
#   user@server  default root@163.245.214.125
#   webroot    default /var/www/<host>
#
# Builds, verifies the build is host- and folder-agnostic, then rsyncs.
# Deliberately does NOT use rsync --delete: it would remove anything else in the
# folder (client uploads, per-server config) that is not in dist/. Old hashed
# assets are harmless; a wiped folder is not.
#
# db_code.json and api/config.local.php are EXCLUDED from the rsync on purpose.
# Both are per server: shipping either one from this machine would point the
# client at the wrong database, or overwrite its credentials.

set -euo pipefail

die() { echo "error: $*" >&2; exit 1; }

usage() { sed -n '3,28p' "$0" | sed 's/^# \{0,1\}//'; exit 2; }

# --dry-run must be able to appear anywhere without being mistaken for a
# positional, otherwise it is too easy to type the real command by accident.
DRY_RUN=0
POSITIONAL=()
for arg in "$@"; do
  case "$arg" in
    --dry-run|-n) DRY_RUN=1 ;;
    *) POSITIONAL+=("$arg") ;;
  esac
done
set -- ${POSITIONAL[@]+"${POSITIONAL[@]}"}

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then usage; fi

[ $# -ge 3 ] || usage

FOLDER=$1
HOST=$2
DB_CODE=$3
SSH_TARGET=${4:-root@163.245.214.125}
WEBROOT=${5:-/var/www/$HOST}
SSH_KEY=${SSH_KEY:-$HOME/.ssh/cars_deploy}
RSYNC_SSH="ssh -i $SSH_KEY"

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
REPO_DIR=$(cd -- "$SCRIPT_DIR/.." && pwd)
TARGET="$WEBROOT/$FOLDER"

# Reuse the same validation the nginx renderer applies, so a bad folder name
# cannot produce a half-configured server.
"$SCRIPT_DIR/render-nginx.sh" "$FOLDER" "$HOST" "$WEBROOT" /dev/null >/dev/null

# db_code is interpolated into JSON, so constrain it to the shape the app emits.
case "$DB_CODE" in
  db_[a-f0-9]*) ;;
  *) die "db_code must look like db_<hex>, got: '$DB_CODE'" ;;
esac

cd "$REPO_DIR"

# --- build --------------------------------------------------------------------
echo "==> building"
npm run build

# --- verify the build is portable ---------------------------------------------
# A single dist/ is the whole point: no client hostname, no /folder, no :8000.
#
# Comments are stripped first. The source explains the rules in prose that
# necessarily mentions example hosts and folders ("a server on 192.168.x.x",
# "https://host/cars/api"), and index.html ships those comments verbatim. Left
# in, they would fail every legitimate build and train people to ignore the gate.
# What must not exist is a hardcoded host in *executable* code.
echo "==> verifying build is host/folder agnostic"
fail=0
scan() { # $1 = extra grep -E pattern
  # `[^:]` before `//` is load-bearing: a naive s://.*:: treats the "//" in
  # "https://" as a comment and silently deletes every real URL from the scan.
  # grep -q exits on first match, so sed sees SIGPIPE; that is expected.
  find dist -type f \( -name '*.js' -o -name '*.css' -o -name '*.html' \) -print0 \
    | xargs -0 sed -E -e 's#(^|[^:])//.*#\1#' -e 's#/\*.*\*/##g' 2>/dev/null \
    | grep -qE "$1" 2>/dev/null
}
for pat in 'world-automobile' 'localhost:8000' '192\.168\.'; do
  if scan "$pat"; then
    echo "    FAIL: dist code contains '$pat'" >&2
    fail=1
  fi
done
# The folder name may legitimately appear as a *route* path (/cars/stock), so
# only an absolute URL carrying it is a leak.
if scan "https?://[^\"']*$FOLDER/"; then
  echo "    FAIL: dist code has an absolute URL containing /$FOLDER/" >&2
  fail=1
fi
# db_code.json must not ship: it is per server and written below.
if [ -f dist/db_code.json ]; then
  echo "    FAIL: dist/db_code.json exists and would be deployed per client" >&2
  fail=1
fi
[ "$fail" -eq 0 ] || die "build is not portable - fix before deploying"
echo "    ok"

# --- upload --------------------------------------------------------------------
# The rsync excludes are asserted before anything is sent. A *.local.php holding
# this machine's DB credentials reaching a client server is unrecoverable from
# the client side, and shipping db_manager_config.local.php once is exactly what
# happened: the db-manager then failed with "Access denied for user 'root'".
# Listing the credential filenames individually is how that exclusion regressed.
RSYNC_EXCLUDES=(--exclude 'db_code.json')
API_EXCLUDES=(--exclude '*.local.php')

assert_no_credentials() { # $1 = rsync dry-run output
  if printf '%s' "$1" | grep -q '\.local\.php'; then
    die "a *.local.php credential file is in the transfer list - fix the --exclude"
  fi
}

# rsync needs -v to list anything in dry-run mode. Without it the transfer list
# is empty, so a dry run reports "(no changes)" for a 130-file deploy and looks
# like a clean bill of health. That is the opposite of what a dry run is for.
RSYNC_DRY_FLAGS=(-aznv --out-format='%n%L')

if [ "$DRY_RUN" -eq 1 ]; then
  echo "==> DRY RUN: nothing will be written to $SSH_TARGET"
  echo
  echo "--- dist/ -> $TARGET/ ---"
  DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${RSYNC_EXCLUDES[@]}" -e "$RSYNC_SSH" dist/ "$SSH_TARGET:$TARGET/" 2>&1 || true)
  printf '%s\n' "${DRY:-    (no changes)}"
  assert_no_credentials "$DRY"

  echo
  echo "--- api/ -> $TARGET/api/ ---"
  DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${API_EXCLUDES[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$TARGET/api/" 2>&1 || true)
  printf '%s\n' "${DRY:-    (no changes)}"
  assert_no_credentials "$DRY"

  echo
  echo "--- would write ---"
  echo "    $TARGET/db_code.json  ->  { \"db_code\": \"$DB_CODE\" }"
  echo
  echo "DRY RUN complete. Re-run without --dry-run to apply."
  exit 0
fi

echo "==> uploading to $SSH_TARGET:$TARGET/"
ssh -i "$SSH_KEY" "$SSH_TARGET" "mkdir -p '$TARGET/api'"

rsync -az "${RSYNC_EXCLUDES[@]}" -e "$RSYNC_SSH" dist/ "$SSH_TARGET:$TARGET/"

# Re-assert against the real target before transferring, so a typo in the
# exclude list is caught here rather than after the credentials have landed.
DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${API_EXCLUDES[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$TARGET/api/" 2>&1 || true)
assert_no_credentials "$DRY"
rsync -az "${API_EXCLUDES[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$TARGET/api/"


# --- per-server files ----------------------------------------------------------
echo "==> writing per-server db_code.json"
printf '{ "db_code": "%s" }\n' "$DB_CODE" \
  | ssh -i "$SSH_KEY" "$SSH_TARGET" "cat > '$TARGET/db_code.json' && chown www-data:www-data '$TARGET/db_code.json'"

if ! ssh -i "$SSH_KEY" "$SSH_TARGET" "test -f '$TARGET/api/config.local.php'"; then
  echo "    WARNING: api/config.local.php is missing on the server." >&2
  echo "    Create it (it holds the DB password and is never deployed):" >&2
  echo "      ssh $SSH_TARGET \"cp '$TARGET/api/config.example.php' '$TARGET/api/config.local.php'\"" >&2
  echo "      ssh $SSH_TARGET \"chown www-data:www-data '$TARGET/api/config.local.php'\"" >&2
  echo "      # then fill in the credentials for THIS client's database" >&2
fi

# --- post-deploy verification ---------------------------------------------------
# Checks the pieces that are easy to get wrong and hard to notice: the SPA
# fallback, the deny rules, and that the API is actually reachable.
echo "==> verifying"
if command -v curl >/dev/null 2>&1; then
  SCHEME=https
  # A deep link must serve the SPA, and the entry script it references must come
  # back as JavaScript. Those are separate failure modes: the SPA fallback makes
  # the HTML 200 even when a deep link is broken, so checking the page alone
  # passes on a site that renders a blank tab.
  #
  # Without a <base> in index.html the relative "./index.<hash>.js" resolves
  # against the document, so /folder/sell-bills/5 requests
  # /folder/sell-bills/index.<hash>.js, misses, and gets index.html with
  # Content-Type: text/html -> "Failed to load module script ... text/html".
  # render-nginx.sh injects <base> via sub_filter; this is the check that it is
  # actually in the served HTML, and that the asset is real JS.
  for path in "login" "sell-bills/5"; do
    PAGE=$(curl -s --max-time 10 "$SCHEME://$HOST/$FOLDER/$path" || true)
    CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/$FOLDER/$path" || echo 000)
    case "$CODE" in
      200|301|302) echo "    ok  /$FOLDER/$path -> $CODE" ;;
      *) echo "    note /$FOLDER/$path -> $CODE (check the nginx ^~ fallback)" >&2 ;;
    esac

    if printf '%s' "$PAGE" | grep -q "<base href=\"/$FOLDER/\">"; then
      echo "    ok  /$FOLDER/$path has <base href=\"/$FOLDER/\">"
    else
      echo "    FAIL /$FOLDER/$path has no <base> - deep links will 404 the entry script" >&2
      echo "         re-apply: ./deploy/render-nginx.sh $FOLDER $HOST $WEBROOT --write" >&2
    fi

    # Resolve the entry script the way a browser would, then require JS back.
    ASSET=$(printf '%s' "$PAGE" | grep -oE '(src|href)="\./index\.[^"]+\.js"' | head -1 \
      | sed -E 's/.*"\.\/([^"]+)"/\1/')
    if [ -n "$ASSET" ]; then
      ACT=$(curl -s -o /dev/null -w '%{http_code} %{content_type}' --max-time 10 \
        "$SCHEME://$HOST/$FOLDER/$ASSET" || echo "000 none")
      case "$ACT" in
        *"application/javascript"*|*"text/javascript"*)
          echo "    ok  /$FOLDER/$ASSET -> $ACT" ;;
        *)
          echo "    FAIL /$FOLDER/$ASSET -> $ACT (expected JavaScript, got $ACT)" >&2 ;;
      esac
    else
      echo "    note /$FOLDER/$path has no relative entry script reference" >&2
    fi
  done

  for f in "api/install.php" "api/setup.sql" "api/export.sql"; do
    CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/$FOLDER/$f" || echo 000)
    if [ "$CODE" = "403" ] || [ "$CODE" = "404" ]; then
      echo "    ok  /$FOLDER/$f -> $CODE (blocked)"
    else
      echo "    FAIL /$FOLDER/$f -> $CODE (must be 403/404)" >&2
    fi
  done

  CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/$FOLDER/db_code.json" || echo 000)
  echo "    ok  /$FOLDER/db_code.json -> $CODE"
else
  echo "    skipped (no curl on this machine)"
fi

cat <<EOF

Done. Next, once (per server):
  ssh $SSH_TARGET "cp '$TARGET/../nginx-app.conf' /etc/nginx/sites-available/default" 2>/dev/null || true
  ./deploy/render-nginx.sh $FOLDER $HOST $WEBROOT --write
  # review deploy/out/nginx-$FOLDER.conf, then:
  ssh $SSH_TARGET 'cp /etc/nginx/sites-available/default /etc/nginx/sites-available/default.bak.\$(date +%s)'
  ssh $SSH_TARGET 'cp - < deploy/out/nginx-$FOLDER.conf > /etc/nginx/sites-available/default && nginx -t && systemctl reload nginx'
EOF
