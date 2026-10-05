import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'

// deploy/deploy.sh is the thing that puts code on the server, and one step in it runs
// as root over ssh rather than through rsync. These tests exist because that step was
// broken in a way that looked exactly like working.
//
// What went wrong
// ---------------
// The step protected api/lib/tenant-provision.php, the library the nginx renderer
// executes as root. It was written as
//
//   ssh -i "$SSH_KEY" "$SSH_TARGET" '... $1 ...' _ "$TARGET"
//
// which reads like it passes $TARGET as the first positional argument. It does not.
// ssh concatenates the remote command and its arguments into a single string and hands
// that to the login shell, so the remote shell saw the script with $1 UNSET, and then
// tried to run `_ /var/www/world-automobile.com/cars` as a second command. Inside the
// script that made $f "/api/lib/tenant-provision.php", so `[ -f "$f" ]` was false and it
// exited 0 immediately.
//
// The consequences were all silent, which is why it survived every earlier deploy:
//
//   - no output, because the fix is only reported when before != after;
//   - no non-zero status, so deploy.sh carried on and reported success;
//   - the file stayed owned by whatever rsync wrote, and `rsync -a` preserves the local
//     checkout's uid - so on the production server it was left owned by uid 501, an
//     account that does not exist there.
//
// Nothing complained until the renderer was actually installed and run, and its
// ownership check refused: "is owned by 'UNKNOWN', not root". That check is the reason
// this was caught rather than exploited - it treats the file as root's code, which it is.
const DEPLOY = readFileSync('deploy/deploy.sh', 'utf8')

// The step, sliced out on its own so these assertions cannot be satisfied by a
// similarly-shaped string somewhere unrelated in a 271-line script.
const protect = DEPLOY.slice(
  DEPLOY.indexOf('protecting the root-executed library'),
  DEPLOY.indexOf('# --- post-deploy verification'),
)

// The same step with its comments removed. The comments quote the broken form in order
// to explain it, so a negative assertion run against them fails on the explanation.
const protectCode = protect.replace(/^\s*#.*$/gm, '')

describe('deploy.sh: the shared layout it writes to', () => {
  it('takes a host and not a client folder', () => {
    // The client folder argument is what this deployment stopped doing: a release is
    // two shared directories and no per-client rsync. A folder argument would invite
    // the old invocation, which deployed a build into ONE client and left the rest on
    // whatever version they happened to be on - and nothing in the output said so.
    expect(DEPLOY).toMatch(/\[ \$# -ge 1 \] \|\| usage/)
    expect(DEPLOY).toMatch(/^HOST=\$1$/m)
    expect(DEPLOY).toMatch(/^DIST_TARGET="\$WEBROOT\/dist"$/m)
    expect(DEPLOY).toMatch(/^API_TARGET="\$WEBROOT\/api"$/m)
    expect(DEPLOY).not.toMatch(/^FOLDER=/m)
    expect(DEPLOY).not.toMatch(/\$TARGET=/)
  })

  it('writes to nothing inside a client folder', () => {
    // A client folder holds uploads. One rsync destination that reaches into it -
    // even a files/ that is excluded from the transfer - is the shape of the bug that
    // deleted invoices when an exclude list was wrong.
    const destinations = [...DEPLOY.matchAll(/\$SSH_TARGET:([^"'\s]+)/g)].map((m) => m[1])
    expect(destinations.length).toBeGreaterThan(0)
    for (const dest of destinations) {
      expect(dest).toMatch(/^\$(DIST|API)_TARGET\//)
    }
    expect(DEPLOY).not.toMatch(/\$SSH_TARGET:\$WEBROOT\/[a-z]/)
  })

  it('excludes the credential files by pattern, and never deletes them', () => {
    // --delete is what makes a removed feature actually disappear from the server,
    // and api/ needs it as much as dist/ does: there is one copy of every endpoint
    // on the box, so an endpoint deleted from the source stayed reachable on every
    // client unless the destination is pruned.
    //
    // It is only safe because *.local.php is excluded from BOTH: rsync does not
    // delete an excluded file from the destination, so a --delete on api/ cannot take
    // the server's database password with it.
    expect(DEPLOY).toMatch(/API_FLAGS=\(--delete --exclude '\*\.local\.php'\)/)
    expect(DEPLOY).toMatch(/DIST_FLAGS=\(--delete/)
    // --delete-excluded is the flag that would void the exclusion's protection.
    // Asserted against the flag arrays, not the whole file: the prose explains the
    // flag by name, so a blanket match fails on the comment that documents it.
    const flagArrays = DEPLOY.split('\n').filter((l) => /^[A-Z_]+=\(/.test(l))
    expect(flagArrays.join('\n')).not.toMatch(/--delete-excluded/)
    expect(flagArrays.join('\n')).not.toMatch(/--delete[^\n]*--exclude 'api'/)

    // Both the real transfer and the dry run that precedes it use the same flags.
    // A dry run on one set and an apply on another is how the check ends up
    // reassuring about a transfer that does something else.
    const apiTransfers = [...DEPLOY.matchAll(/rsync[^\n]*\$\{API_FLAGS\[@\]\}[^\n]*/g)].map((m) => m[0])
    expect(apiTransfers.length).toBeGreaterThanOrEqual(3) // 2 dry runs + 1 apply
    for (const t of apiTransfers) {
      expect(t).toMatch(/api\/ "\$SSH_TARGET:\$API_TARGET\/"/)
    }
    // Exactly one of them is the real transfer; it must carry -a like the others.
    expect(apiTransfers.filter((t) => t.startsWith('rsync -az ')).length).toBe(1)
  })

  it('re-reads the destination for the credentials after the api/ transfer', () => {
    // Excluding them from the transfer does not prove they survived the --delete;
    // it proves they were not sent. So every *.local.php this checkout has is
    // asserted to still exist on the server, after the prune.
    const afterApi = DEPLOY.slice(DEPLOY.indexOf('rsync -az "${API_FLAGS[@]}"'))
    expect(afterApi).toMatch(/api\/\*\.local\.php/)
    expect(afterApi).toMatch(/test -f '\$API_TARGET\/\$name'/)
    // The check has to come after the apply, not only inside the dry run.
    expect(afterApi.indexOf('test -f \'$API_TARGET/$name\'')).toBeGreaterThan(
      afterApi.indexOf('rsync -az "${API_FLAGS[@]}"'),
    )
  })

  it('verifies against a client rather than a folder it was handed', () => {
    // The post-deploy checks run over HTTP, where a client is named by the URL. It
    // reads one from the registry instead of inventing one, so it cannot end up
    // asserting against a client that was deleted.
    expect(DEPLOY).toMatch(/SELECT db_name FROM merhab_databases\.dbs WHERE is_created = 1/)
    expect(DEPLOY).toMatch(/\$SCHEME:\/\/\$HOST\/\$CLIENT\//)
  })

  it('checks the webroot /api is refused, because no client is named there', () => {
    // /api/api.php has no client in it. Serving it would mean guessing which database
    // the request meant, so the rendered config answers 404 and this asserts that.
    expect(DEPLOY).toMatch(/\$SCHEME:\/\/\$HOST\/api\/api\.php/)
  })
})

describe('deploy.sh: the root-executed library protection step', () => {
  it('delivers the target path to the remote script instead of concatenating it', () => {
    // `bash -s --` is the form that works: the script arrives on stdin and the argument
    // after -- becomes $1. The concatenated form is what silently did nothing.
    //
    // The argument is now $API_TARGET rather than a per-client $TARGET, because there
    // is one shared api/ - but the mechanism is the same and the failure is just as
    // silent, so the form is still asserted rather than assumed.
    expect(protectCode).toMatch(/ssh -i "\$SSH_KEY" "\$SSH_TARGET" bash -s -- "\$API_TARGET"/)
    expect(protectCode).not.toMatch(/"\$SSH_TARGET" '\n/)
    // The old shape put a bare `_` after the quoted script to mean "this is $0".
    expect(protectCode).not.toMatch(/'\s+_\s+"\$(API_)?TARGET"/)
  })

  it('reads the path from the positional argument it was actually given', () => {
    // $1 is the shared api/. Reading anything else - a literal /lib/... path, or $2
    // because of an off-by-one in the argument list - is the original bug wearing a
    // new hat. Note there is no api/ after $1 any more either: $1 already IS api/,
    // and concatenating api/ onto it would name a folder that does not exist.
    expect(protectCode).toMatch(/f="\$1\/lib\/tenant-provision\.php"/)
    expect(protectCode).not.toMatch(/f="\/lib\//)
    expect(protectCode).not.toMatch(/f="\$1\/api\//)
  })

  it('fails the deploy when the file root executes is missing or not root-owned', () => {
    // The old step exited 0 on a missing file, which is the state where the renderer
    // has nothing to load. And a chown that fails reports nothing on a server where the
    // uid does not resolve to a name, so the end state is asserted, not the error code.
    expect(protect).toMatch(/is missing - root would have nothing to execute/)
    expect(protect).toMatch(/could not chown/)
    expect(protect).toMatch(/not root-owned/)
    expect(protect).toMatch(/chown root:root/)
  })

  it('reports the fix when the ownership actually changed', () => {
    // Kept from the original: a deploy that quietly corrects this should say so, because
    // "it changed" is the only signal that rsync had been writing someone else's uid.
    expect(protect).toMatch(/fixed \$f: \$before -> \$after/)
  })

  it('still chowns it to root, not to the web user', () => {
    // The whole point of the step. chowning to www-data would satisfy every other
    // assertion here while handing a web-writable file to root.
    expect(protect).toMatch(/chown root:root/)
    expect(protectCode).not.toMatch(/chown (www-data|\$WEB_USER)/)
  })
})
