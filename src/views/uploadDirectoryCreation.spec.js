import { describe, it, expect } from 'vitest'
import { readFileSync, statSync } from 'node:fs'
import { execFileSync } from 'node:child_process'

// POST must create the destination folder; GET must never create anything.
//
// The security rewrite of upload.php replaced the old mkdir-then-check with a bare
// is_dir(), which is the safer half of the trade but broke a folder the app does not
// ship. A new car's files/cars/<id>/<type> does not exist until its first upload, so
// the very upload meant to create the folder was refused with "Invalid upload
// directory" - the symptom that started this. Every tenant in the repo has an empty
// files/cars, which is the same fact seen from the other end.
//
// The fix has to add creation back without giving up containment, and that is the
// part worth pinning: mkdir on a caller-supplied path is exactly how a webshell
// lands outside the tenant. So the checks below run the real helper against a real
// temporary root rather than reading the source and hoping.

const UPLOAD_PHP = readFileSync(new URL('../../api/upload.php', import.meta.url), 'utf8')

const hasPhp = (() => {
  try {
    execFileSync('php', ['-v'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
})()

// Both helpers are pure path arithmetic, so they can be lifted straight out of
// upload.php and exercised without a database, a tenant or an HTTP request. The
// regexes pull the bodies verbatim - a renamed or dropped function fails here
// rather than silently testing a stale copy.
const extractFunction = (name) => {
  const match = UPLOAD_PHP.match(
    new RegExp(`function ${name}\\([\\s\\S]*?\\n}\\n`),
  )
  if (!match) {
    throw new Error(`${name} not found in api/upload.php`)
  }
  return match[0]
}

const runHelper = (root, relative) => {
  const script = `
    ${extractFunction('upload_resolve_within_root')}
    ${extractFunction('upload_prepare_within_root')}
    echo upload_prepare_within_root($argv[1], $argv[2]);
  `
  return execFileSync('php', ['-r', script, '--', root, relative], {
    encoding: 'utf8',
  })
}

// mktemp hands back /var/folders/... on macOS, which is a symlink to
// /private/var/folders/.... The helper returns realpath() output by design - it
// has to, or the containment it just proved could be escaped by a later path
// that is not resolved - so the expectation has to be the resolved root too.
const tempRoot = () => {
  const raw = execFileSync('mktemp', ['-d', '-t', 'uploadtest'], { encoding: 'utf8' }).trim()
  return execFileSync('php', ['-r', 'echo realpath($argv[1]);', '--', raw], {
    encoding: 'utf8',
  }).trim()
}

describe.runIf(hasPhp)('upload.php destination folders', () => {
  it('creates a nested destination folder that does not exist yet', () => {
    const root = tempRoot()
    const result = runHelper(root, 'files/cars/4242/bl')

    expect(result).toBe(`${root}/files/cars/4242/bl`)
    expect(isDirectory(`${root}/files/cars/4242/bl`)).toBe(true)
  })

  it('resolves a folder that already exists without touching it', () => {
    const root = tempRoot()
    execFileSync('mkdir', ['-p', `${root}/files/ids`])

    expect(runHelper(root, 'files/ids')).toBe(`${root}/files/ids`)
  })

  it('refuses a parent-directory segment instead of creating it somewhere unexpected', () => {
    const root = tempRoot()
    execFileSync('mkdir', ['-p', `${root}/files/ids`])

    // Contained is not good enough. Collapsing this would create <root>/etc,
    // which is a directory no caller asked for.
    expect(runHelper(root, 'files/../../etc')).toBe('')
    expect(isDirectory(`${root}/etc`)).toBe(false)
  })

  it('never resolves outside the root, whatever it is handed', () => {
    const root = tempRoot()

    for (const attempt of [
      '/etc',
      '../escape',
      'files/../../escape',
      'files/./../escape',
      '../../../../tmp/uploadtest-escape',
    ]) {
      const result = runHelper(root, attempt)
      expect(result === '' || result.startsWith(`${root}/`)).toBe(true)
    }
  })

  it('keeps creation out of the GET branch', () => {
    // Serving a file must not have a side effect: a missing path is a 404, and it
    // must not quietly grow the tree because somebody asked for a URL.
    const getBranch = UPLOAD_PHP.slice(UPLOAD_PHP.indexOf("REQUEST_METHOD'] === 'GET'"))
    const getBranchBody = getBranch.slice(0, getBranch.indexOf('// POST'))

    expect(getBranchBody).toContain('upload_resolve_within_root')
    expect(getBranchBody).not.toContain('upload_prepare_within_root')
    expect(getBranchBody).not.toContain('mkdir')
  })

  it('still refuses a non-tenant request rather than writing into the shared root', () => {
    // upload_base_root returns the app root when there is no tenant, and
    // app_db_files_dir() is null. The mkdir that was added back must not have
    // turned that into a write the shared api/ folder will accept.
    expect(UPLOAD_PHP).toContain('function upload_base_root')
    const fn = UPLOAD_PHP.match(/function upload_base_root[\s\S]*?\n}\n/)[0]
    expect(fn).toContain('app_db_files_dir() === null')
  })
})

describe('upload.php directory rules, statically', () => {
  it('resolves the POST destination through the creating helper', () => {
    // Counting calls is not enough, and is what the first version of this test did:
    // reverting POST to upload_resolve_within_root left every behavioural test green
    // because those run the helper directly. What has to hold is the call POST makes.
    const post = UPLOAD_PHP.slice(UPLOAD_PHP.indexOf('// POST'))
    expect(post).toContain('upload_prepare_within_root(')
    expect(post).not.toMatch(/upload_resolve_within_root\(\s*\n\s*upload_base_root/)
  })
})

// readFileSync cannot see a directory - it throws EISDIR - so "was the folder
// created?" needs statSync, which is the whole question in the first test.
function isDirectory(path) {
  try {
    return statSync(path).isDirectory()
  } catch {
    return false
  }
}
