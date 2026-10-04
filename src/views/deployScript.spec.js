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

describe('deploy.sh: the root-executed library protection step', () => {
  it('delivers the target path to the remote script instead of concatenating it', () => {
    // `bash -s --` is the form that works: the script arrives on stdin and the argument
    // after -- becomes $1. The concatenated form is what silently did nothing.
    expect(protectCode).toMatch(/ssh -i "\$SSH_KEY" "\$SSH_TARGET" bash -s -- "\$TARGET"/)
    expect(protectCode).not.toMatch(/"\$SSH_TARGET" '\n/)
    // The old shape put a bare `_` after the quoted script to mean "this is $0".
    expect(protectCode).not.toMatch(/'\s+_\s+"\$TARGET"/)
  })

  it('reads the path from the positional argument it was actually given', () => {
    // $1 is the target. Reading anything else - a literal /api/... path, or $2 because
    // of an off-by-one in the argument list - is the original bug wearing a new hat.
    expect(protectCode).toMatch(/f="\$1\/api\/lib\/tenant-provision\.php"/)
    expect(protectCode).not.toMatch(/f="\/api\/lib/)
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
