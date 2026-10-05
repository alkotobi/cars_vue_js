import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync, statSync, mkdtempSync, mkdirSync, writeFileSync } from 'node:fs'
import { execFileSync } from 'node:child_process'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

// The repository root is not a tenant, and must not be able to behave like one.
//
// It used to be one. api.php resolved which database to use from db_code.json, and
// when there was no db_code.json it fell back to config.php's db_name - a single
// fixed name. That is only harmless while the root serves nothing, because the
// moment one database is named there, every request that resolved nothing opened
// that one. The name it held was merhab_cars: a live tenant with real users,
// password hashes and api_tokens. So an unresolved request was not an error, it was
// a working session on somebody's data, and the only symptom was a response nobody
// had asked a question about.
//
// The two ways to fix it - delete the root db_code.json, or stop naming a database
// in configuration - each fix half of it. With the file gone but the fallback left,
// the root still opens the tenant, now with nothing on disk to say that is a
// mistake. With the fallback gone but the file left, the root is still a tenant.
// Both have to hold, and neither is visible from reading any one file, so both are
// pinned here.
//
// The failure mode this guards against is also why the checks are static as well as
// behavioural: reintroducing the fallback is a one-line change that looks like
// removing a special case.

const root = new URL('../../', import.meta.url)
const API_PHP = readFileSync(new URL('../../api/api.php', import.meta.url), 'utf8')
const CONFIG_PHP = readFileSync(new URL('../../api/config.php', import.meta.url), 'utf8')
const CONFIG_EXAMPLE = readFileSync(new URL('../../api/config.example.php', import.meta.url), 'utf8')
const BACKUP_PHP = readFileSync(new URL('../../api/backup.php', import.meta.url), 'utf8')
const BACKUP_SIMPLE = readFileSync(new URL('../../api/backup_simple_web.php', import.meta.url), 'utf8')
const BACKUP_DUMP = readFileSync(new URL('../../api/lib/backup_dump.php', import.meta.url), 'utf8')
const APPDB = readFileSync(new URL('../../api/lib/appdb.php', import.meta.url), 'utf8')

const hasPhp = (() => {
  try {
    execFileSync('php', ['-v'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
})()

const exists = (path) => {
  try {
    readFileSync(path)
    return true
  } catch {
    return false
  }
}

// readFileSync cannot see a directory, so the tenant's files/ folder needs its own
// check. A tenant is identified by that folder now, so a helper that only reports
// files would make every tenant look unprovisioned.
const isDirectory = (path) => {
  try {
    return statSync(path).isDirectory()
  } catch {
    return false
  }
}

describe('the repository root is not a tenant', () => {
  it('has no db_code.json, which is what makes it a tenant to api.php', () => {
    // The per-tenant copy below is the one that must exist. Having it here instead
    // would make the root a tenant again, and every request through /api/ would
    // resolve to merhab_cars with no error anywhere.
    expect(exists(join(root.pathname, 'db_code.json'))).toBe(false)
  })

  it('has one tenant folder, with files and no db_code.json', () => {
    // Present so the isolation below is a real separation rather than an empty one:
    // if the tenant folder disappeared the root would have nothing to leak into and
    // every test here would still pass.
    expect(isDirectory(join(root.pathname, 'merhab_cars', 'files'))).toBe(true)
    expect(exists(join(root.pathname, 'merhab_cars', 'db_code.json'))).toBe(false)
  })

  it('does not name a database in api/config.php', () => {
    // No default, and the key is not read from a fallback. This is the value the
    // root's only remaining fallback used to come from.
    expect(CONFIG_PHP).not.toMatch(/'merhab_cars'/)
    expect(CONFIG_PHP).toMatch(
      /\$resolveConfig\('db_name',\s*'DB_NAME',\s*''\s*\)/,
    )
  })

  it('does not name a database in api/config.example.php', () => {
    // The example is copied to config.local.php on a new install, so a tenant name
    // here propagates into every machine set up from it.
    expect(CONFIG_EXAMPLE).not.toMatch(/'merhab_cars'/)
    expect(CONFIG_EXAMPLE).toMatch(/'db_name'\s*=>\s*''/)
  })
})

describe('api.php refuses to guess a database', () => {
  it('refuses unconditionally, so editing config.php cannot reopen the leak', () => {
    // Not `if (empty($db_config['dbname']))`. That was the first attempt and it is
    // the wrong guard: config.php is precisely the file someone edits to point the
    // app at a local database, and a configured value is indistinguishable from the
    // tenant that used to be hardcoded there. Guarding on emptiness only means the
    // leak closes until the next person needs a local database.
    const getDbConfig = API_PHP.slice(API_PHP.indexOf('function getDbConfig'))

    expect(getDbConfig).not.toMatch(/empty\(\$db_config\['dbname'\]\)/)
    expect(getDbConfig).not.toMatch(/return \$resolved = \$db_config;/)
  })

  it('refuses with db_unavailable when nothing resolved', () => {
    // Not a generic failure: db_unavailable is what the client already handles, and
    // it is accurate - there is no database to serve the request from.
    const afterDoc = API_PHP.slice(API_PHP.indexOf('No db_code.json'))

    expect(afterDoc).toMatch(/apiErrorDie\('db_unavailable'\);/)
  })
})

describe('nothing on the request path substitutes a configured database', () => {
  // Every one of these was an independent copy of the same fallback, so fixing
  // api.php alone would have left the leak reachable through the backup endpoints -
  // which are exactly where it does the most damage, because a backup of the wrong
  // tenant hands over its users table and password hashes.

  const cases = [
    ['api/backup.php', BACKUP_PHP],
    ['api/backup_simple_web.php', BACKUP_SIMPLE],
    // The gate on both of the above, so it had the same fallback a third time - and
    // here it decides WHO may dump the database, not only which one gets dumped.
    ['api/lib/backup_dump.php', BACKUP_DUMP],
    // The gate on auth, and the last one left holding the fallback. The other three
    // decide which database gets read; this one decides WHO may read it, so a root
    // request that resolved here was authenticated against whichever tenant
    // config.php happened to name. Confirmed live rather than reasoned about: with
    // the root pointed at merhab_cars, upload.php answered "not_authenticated" (it
    // connected and went looking for a token) where a name that does not exist
    // answers "Unknown database".
    ['api/lib/appdb.php', APPDB],
  ]

  for (const [file, source] of cases) {
    it(`${file} cannot resolve a database by falling back to config`, () => {
      // The terminal `?? $db_config['dbname'];` is the shape that made these
      // reachable: whatever configuration held became the answer, with no check in
      // between. Reading the value is still fine - it is how the refusal below is
      // written - so what is pinned is that it cannot be the last word.
      expect(source).not.toMatch(
        /app_db_name\(\)\s*\?\?\s*\$db_?[cC]onfig\[['"]dbname['"]\]\s*;/,
      )
      expect(source).toMatch(/did not resolve to a tenant/)
    })
  }

  it('the backup filename no longer carries a tenant name', () => {
    // Cosmetic on its own, but it is the shape of the habit that produced the bug:
    // a specific database name hardcoded into a path where the database is an
    // argument.
    expect(BACKUP_SIMPLE).not.toMatch(/merhab_cars_backup/)
  })
})

describe.skipIf(!hasPhp)('resolution through the real resolver', () => {
  // Built on disk rather than read out of the source, because the question is what
  // the resolver answers for a given SCRIPT_NAME, and that depends on which
  // db_code.json is on disk beside it.
  const probe = (scriptName, { withTenantFolder, withFilesFolder = false }) => {
    const dir = mkdtempSync(join(tmpdir(), 'cars-root-iso-'))
    mkdirSync(join(dir, 'api', 'lib'), { recursive: true })
    copyTree(new URL('api/', root), join(dir, 'api'))

    // The tenant's own copy, including its db_code.json, is what makes this a real
    // two-tenant arrangement rather than a single one.
    if (withTenantFolder) {
      copyTree(new URL('merhab_cars/', root), join(dir, 'merhab_cars'))
    }
    if (withFilesFolder) {
      mkdirSync(join(dir, 'merhab_cars_files', 'uploads'), { recursive: true })
    }

    const entry = join(dir, 'probe.php')
    writeFileSync(
      entry,
      `<?php
$_SERVER['SCRIPT_NAME'] = ${JSON.stringify(scriptName)};
require ${JSON.stringify(join(dir, 'api', 'lib', 'appdb.php'))};
$files = app_db_files_dir();
$uploads = app_deployment_root() . '/' . $files;
echo json_encode([
    'db' => app_db_name(),
    'files' => $files,
    'root' => app_deployment_root(),
    'uploads' => $uploads,
    'uploadsExists' => is_dir($uploads),
]);
`,
    )

    const out = execFileSync('php', [entry], { encoding: 'utf8' })
    execFileSync('rm', ['-rf', dir])
    return JSON.parse(out)
  }

  function copyTree(from, to) {
    execFileSync('cp', ['-R', `${from.pathname}.`, to])
  }

  it('resolves the tenant when reached through its own folder', () => {
    // The half that has to keep working. Isolation that broke the tenant would be
    // a worse outage than the one it prevents.
    const answer = probe('/merhab_cars/api/api.php', { withTenantFolder: true })

    expect(answer.db).toBe('merhab_cars')
    expect(answer.files).toBe('files')
  })

  it('resolves the tenant uploads to the folder that exists', () => {
    // The registry records a tenant's files_dir with a leading slash, and that
    // slash is the only thing saying the uploads are a SIBLING of the app folder
    // rather than a child of it. Lose it - a row written by hand, or a migration
    // that normalises the value - and app_deployment_root() returns the app folder
    // instead of its parent, so every upload path becomes
    // <webroot>/merhab_cars/merhab_cars_files/..., a directory that is never
    // created, and every image silently 404s while the API keeps returning a URL.
    //
    // Caught by asking whether the resolved directory exists, not by reading the
    // code: the two cases differ only in one character of one database column.
    const answer = probe('/merhab_cars/api/api.php', {
      withTenantFolder: true,
      withFilesFolder: true,
    })

    expect(answer.uploads).toMatch(/\/merhab_cars\/files$/)
    expect(answer.uploadsExists).toBe(true)
  })

  it('resolves nothing at the root, so api.php refuses instead of serving it', () => {
    // The whole point: null, not 'merhab_cars'. With the fallback removed, null is
    // what makes the request fail closed rather than land in the tenant.
    expect(probe('/api/api.php', { withTenantFolder: true }).db).toBeNull()
  })
})
