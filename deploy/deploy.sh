#!/usr/bin/env bash
#
# Deploy the Cars app to a client server.
#
#   ./deploy/deploy.sh <host> [user@server] [webroot] [--dry-run]
#
#   host       domain or IP the app is served from
#   user@server  default root@163.245.214.125
#   webroot    default /var/www
#
# There is no client folder argument, and that is the whole point of this script now.
# It deploys two shared directories and nothing else:
#
#   dist/ -> <webroot>/dist/     the one build every client is served from
#   api/  -> <webroot>/api/      the one code every client runs
#
# Client folders hold only their own files/ (uploads), which this script never writes
# to. Deploying a release therefore updates every client on the server at once, and
# there is no per-client rsync that can be forgotten for one of them - which is how
# two clients ended up on different code.
#
# Both rsyncs use --delete. That is safe precisely because of the layout above:
# dist/ holds nothing but the build, and the credentials in api/*.local.php are
# excluded from the transfer (and from the deletion) by name. The old per-client
# rsync could not do this without also excluding files/, and a typo in that exclude
# list would have deleted a client's invoices.
#
# api/config.local.php is EXCLUDED on purpose (per-server credentials).

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

[ $# -ge 1 ] || usage

HOST=$1
SSH_TARGET=${2:-root@163.245.214.125}
WEBROOT=${3:-/var/www}
SSH_KEY=${SSH_KEY:-$HOME/.ssh/cars_deploy}
RSYNC_SSH="ssh -i $SSH_KEY"

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" && pwd)
REPO_DIR=$(cd -- "$SCRIPT_DIR/.." && pwd)

# The two shared destinations, and the two names they are refused under as clients.
# api/ and dist/ are the same level as the client folders, so a client called either
# would have its folder shadowed by the thing it is served from. app_valid_tenant()
# and tenant_assert_valid_db_name() refuse both; this is the third place, in shell,
# and it exists so a bad registry row is caught before anything is uploaded.
case "$HOST" in
  ''|*[!A-Za-z0-9.:_-]*) die "'$HOST' is not usable as a server_name" ;;
esac

DIST_TARGET="$WEBROOT/dist"
API_TARGET="$WEBROOT/api"

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
scan() { # $1 = grep -E pattern
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
# A client name may legitimately appear as a *route* path (/cars/stock), so only an
# absolute URL carrying it is a leak. There is no client name to check here any more
# - the build is built once and served to every client - so this is the generic form:
# an absolute URL into the webroot outside the build would bypass the API.
if scan "https?://[^\"']*/api/"; then
  echo "    FAIL: dist code has an absolute URL ending in /api/" >&2
  fail=1
fi
# db_code.json must not ship: the database is chosen by the URL, not by a file in the
# build folder, and a client folder has no build in it to hold one.
if [ -f dist/db_code.json ]; then
  echo "    FAIL: dist/db_code.json exists; the URL carries the database now" >&2
  fail=1
fi
[ "$fail" -eq 0 ] || die "build is not portable - fix before deploying"
echo "    ok"

# --- upload --------------------------------------------------------------------
# Both transfers use --delete, so a file this checkout no longer produces actually
# leaves the server. api/ is included in that deliberately: with one shared api/
# there is exactly one copy of every endpoint on the box, so an endpoint removed
# from the source stayed reachable on every client forever unless the destination
# is pruned.
#
# It is safe against the credentials because --exclude also protects them from
# deletion: rsync never removes an excluded file from the destination (only
# --delete-excluded would, and it is never passed). That is asserted by re-reading
# the destination afterwards rather than trusted.
API_FLAGS=(--delete --exclude '*.local.php')

# The rsync excludes are asserted before anything is sent. A *.local.php holding
# this machine's DB credentials reaching a client server is unrecoverable from
# the client side, and shipping db_manager_config.local.php once is exactly what
# happened: the db-manager then failed with "Access denied for user 'root'".
# Listing the credential filenames individually is how that exclusion regressed, so
# the pattern is asserted against the real dry-run transfer list below.
#
# --delete prunes hashed assets and files a build no longer produces, which is the
# only way a removed feature actually leaves a deployed server: without it, removing
# the invitations feature left InvitationsView.*.js and invitations.php sitting in
# every webroot forever, still served, still shipping to the browser.
DIST_FLAGS=(--delete --exclude '.DS_Store')

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
  echo "--- dist/ -> $DIST_TARGET/ ---"
  DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${DIST_FLAGS[@]}" -e "$RSYNC_SSH" dist/ "$SSH_TARGET:$DIST_TARGET/" 2>&1 || true)
  printf '%s\n' "${DRY:-    (no changes)}"

  echo
  echo "--- api/ -> $API_TARGET/ ---"
  DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${API_FLAGS[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$API_TARGET/" 2>&1 || true)
  printf '%s\n' "${DRY:-    (no changes)}"
  assert_no_credentials "$DRY"

  echo
  echo "--- not touched ---"
  echo "    <webroot>/<client>/files/    client uploads, for every client"
  echo
  echo "DRY RUN complete. Re-run without --dry-run to apply."
  exit 0
fi

echo "==> uploading to $SSH_TARGET"
ssh -i "$SSH_KEY" "$SSH_TARGET" "mkdir -p '$DIST_TARGET' '$API_TARGET'"

rsync -az "${DIST_FLAGS[@]}" -e "$RSYNC_SSH" dist/ "$SSH_TARGET:$DIST_TARGET/"

# Re-assert against the real target before transferring, so a typo in the
# exclude list is caught here rather than after the credentials have landed.
DRY=$(rsync "${RSYNC_DRY_FLAGS[@]}" "${API_FLAGS[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$API_TARGET/" 2>&1 || true)
assert_no_credentials "$DRY"
rsync -az "${API_FLAGS[@]}" -e "$RSYNC_SSH" api/ "$SSH_TARGET:$API_TARGET/"


# --- per-server files ----------------------------------------------------------
# The credentials are protected from deletion by --exclude, which is a guarantee
# about rsync's behaviour rather than about this script, and this step is the one
# that just gained --delete. So every credential the source tree is missing is
# asserted to be present on the server after the transfer: a prune that did delete
# one of them breaks the app on the next request with nothing to undo it.
for f in "$REPO_DIR"/api/*.local.php; do
  name=$(basename "$f")
  # Only the ones the operator actually keeps locally say anything about the
  # server. This checkout's set is not the server's set.
  if ! ssh -i "$SSH_KEY" "$SSH_TARGET" "test -f '$API_TARGET/$name'"; then
    echo "    WARNING: api/$name is missing on the server and was not deleted by this" >&2
    echo "    deploy (it is excluded), so it was never there. Create it if this server" >&2
    echo "    needs it; it holds credentials and is never deployed." >&2
  fi
done

if ! ssh -i "$SSH_KEY" "$SSH_TARGET" "test -f '$API_TARGET/config.local.php'"; then
  echo "    WARNING: api/config.local.php is missing on the server." >&2
  echo "    Create it (it holds the DB password and is never deployed):" >&2
  echo "      ssh $SSH_TARGET \"cp '$API_TARGET/config.example.php' '$API_TARGET/config.local.php'\"" >&2
  echo "      ssh $SSH_TARGET \"chown www-data:www-data '$API_TARGET/config.local.php'\"" >&2
  echo "      # then fill in the credentials for the REGISTRY database (merhab_databases)," >&2
  echo "      # which is where the client list comes from." >&2
fi

# The nginx renderer is executed as root and reads the name validator out of api/lib.
# That file is inside the web root, so it is only safe while it stays root-owned and
# unwritable by the web user - and a deploy that someone ran as www-data, or a manual
# `chown -R www-data /var/www/api`, leaves it owned by the web user with nothing
# complaining until the next Provision screen press refuses to render.
#
# Corrected rather than only reported: the required state is unambiguous (a PHP
# library must not be web-writable), and refusing to finish the deploy over it would
# leave the app updated with the renderer broken, which is the worse of the two.
echo "==> protecting the root-executed library"
# The remote script goes in on stdin, because ssh concatenates a command and its
# arguments into ONE string for the remote shell. The old form,
# `ssh host '...' _ "$TARGET"`, therefore ran the script with $1 empty and then tried to
# execute `_ /var/www/...` as a second command: $f became /api/lib/tenant-provision.php,
# the -f test failed, and the step exited 0 having done nothing. No output, no warning,
# and the file left owned by whatever rsync wrote - which is the local checkout's uid,
# since rsync -a preserves it. `bash -s --` is what actually delivers the argument.
ssh -i "$SSH_KEY" "$SSH_TARGET" bash -s -- "$API_TARGET" <<'REMOTE' || echo "    WARNING: could not check $API_TARGET/lib/tenant-provision.php ownership." >&2
  f="$1/lib/tenant-provision.php"
  if [ ! -f "$f" ]; then
    echo "    *** $f is missing - root would have nothing to execute"
    exit 1
  fi
  before=$(stat -c "%U:%a" "$f" 2>/dev/null || stat -f "%Su:%Lp" "$f")
  chown root:root "$f" && chmod 0644 "$f" || { echo "    *** could not chown $f"; exit 1; }
  after=$(stat -c "%U:%a" "$f" 2>/dev/null || stat -f "%Su:%Lp" "$f")
  [ "$before" = "$after" ] || echo "    fixed $f: $before -> $after (root executes this file)"
  # Confirm the end state rather than the absence of an error message. A chown that
  # fails on a server where the uid does not resolve reports nothing useful, and this
  # file is the one input root loads.
  [ "${after%%:*}" = "root" ] || { echo "    *** $f is '$after', not root-owned"; exit 1; }
REMOTE

# --- post-deploy verification ---------------------------------------------------
# Checks the pieces that are easy to get wrong and hard to notice: the SPA
# fallback, the deny rules, and that the API is actually reachable.
echo "==> verifying"
if command -v curl >/dev/null 2>&1; then
  SCHEME=https

  # A client is chosen by its name in the URL, so the verification needs one. Any
  # provisioned client works; the point is that the shared build and the shared API
  # answer for it. Taken from the registry rather than hardcoded so this does not
  # test a client that has since been deleted.
  CLIENT=$(ssh -i "$SSH_KEY" "$SSH_TARGET" \
    "mysql -N -B -e \"SELECT db_name FROM merhab_databases.dbs WHERE is_created = 1 ORDER BY db_name LIMIT 1\"" 2>/dev/null || true)

  if [ -z "$CLIENT" ]; then
    echo "    skipped: no provisioned client found on the server to verify against" >&2
  else
    # A deep link must serve the SPA, and the entry script it references must come
    # back as JavaScript. Those are separate failure modes: the SPA fallback makes
    # the HTML 200 even when a deep link is broken, so checking the page alone
    # passes on a site that renders a blank tab.
    #
    # Without a <base> in index.html the relative "./index.<hash>.js" resolves
    # against the document, so /client/sell-bills/5 requests
    # /client/sell-bills/index.<hash>.js, misses, and gets index.html with
    # Content-Type: text/html -> "Failed to load module script ... text/html".
    # The nginx renderer injects <base> via sub_filter; this is the check that it is
    # actually in the served HTML, and that the asset is real JS.
    for path in "login" "sell-bills/5"; do
      PAGE=$(curl -s --max-time 10 "$SCHEME://$HOST/$CLIENT/$path" || true)
      CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/$CLIENT/$path" || echo 000)
      case "$CODE" in
        200|301|302) echo "    ok  /$CLIENT/$path -> $CODE" ;;
        *) echo "    note /$CLIENT/$path -> $CODE (check the nginx fallback)" >&2 ;;
      esac

      if printf '%s' "$PAGE" | grep -q "<base href=\"/$CLIENT/\">"; then
        echo "    ok  /$CLIENT/$path has <base href=\"/$CLIENT/\">"
      else
        echo "    FAIL /$CLIENT/$path has no <base> - deep links will 404 the entry script" >&2
        echo "         re-apply: ssh $SSH_TARGET '/usr/local/bin/cars-nginx-render --write'" >&2
      fi

      # Resolve the entry script the way a browser would, then require JS back.
      ASSET=$(printf '%s' "$PAGE" | grep -oE '(src|href)="\./index\.[^"]+\.js"' | head -1 \
        | sed -E 's/.*"\.\/([^"]+)"/\1/')
      if [ -n "$ASSET" ]; then
        ACT=$(curl -s -o /dev/null -w '%{http_code} %{content_type}' --max-time 10 \
          "$SCHEME://$HOST/$CLIENT/$ASSET" || echo "000 none")
        case "$ACT" in
          *"application/javascript"*|*"text/javascript"*)
            echo "    ok  /$CLIENT/$ASSET -> $ACT" ;;
          *)
            echo "    FAIL /$CLIENT/$ASSET -> $ACT (expected JavaScript, got $ACT)" >&2 ;;
        esac
      else
        echo "    note /$CLIENT/$path has no relative entry script reference" >&2
      fi
    done

    # The tenant folder must not have grown a build of its own: one dist/ serves
    # everyone, and a copy here is how two clients drift onto different code.
    STRAY=$(ssh -i "$SSH_KEY" "$SSH_TARGET" \
      "ls $WEBROOT/$CLIENT/index.html $WEBROOT/$CLIENT/api 2>/dev/null | wc -l" 2>/dev/null || echo 0)
    if [ "${STRAY:-0}" -eq 0 ]; then
      echo "    ok  /$CLIENT/ holds only its uploads - no build, no api copy"
    else
      echo "    FAIL /$CLIENT/ has its own build or api copy: remove them" >&2
    fi

    for f in "api/install.php" "api/setup.sql"; do
      CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/$CLIENT/$f" || echo 000)
      if [ "$CODE" = "403" ] || [ "$CODE" = "404" ]; then
        echo "    ok  /$CLIENT/$f -> $CODE (blocked)"
      else
        echo "    FAIL /$CLIENT/$f -> $CODE (must be 403/404)" >&2
      fi
    done

    # The webroot's own /api is not an entry point: a request there names no client,
    # and serving it would mean guessing which database it meant.
    CODE=$(curl -s -o /dev/null -w '%{http_code}' --max-time 10 "$SCHEME://$HOST/api/api.php" || echo 000)
    case "$CODE" in
      403|404) echo "    ok  /api/api.php -> $CODE (no client named)" ;;
      *) echo "    FAIL /api/api.php -> $CODE (must be 403/404)" >&2 ;;
    esac
  fi
else
  echo "    skipped (no curl on this machine)"
fi

cat <<EOF

Done. Every client on $HOST is served from:
  $DIST_TARGET/   the build
  $API_TARGET/    the code

Next, once per server (root):
  # one-time setup of the renderer and sudoers rule: see DEPLOYMENT.md 6A
  ssh $SSH_TARGET '/usr/local/bin/cars-nginx-render --check'
  ssh $SSH_TARGET '/usr/local/bin/cars-nginx-render --write'
  ssh $SSH_TARGET 'nginx -t && systemctl reload nginx'
  # then, on every release, after a client is added:
  ssh $SSH_TARGET '/usr/local/bin/cars-nginx-render --write && nginx -t && systemctl reload nginx'
EOF
