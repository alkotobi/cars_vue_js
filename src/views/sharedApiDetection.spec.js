import { describe, it, expect } from 'vitest'
import { execFileSync } from 'node:child_process'
import { mkdtempSync, mkdirSync, copyFileSync, writeFileSync, existsSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

// tenant_has_shared_api() answers one question: is there ONE api/ serving every
// tenant, or does each tenant folder carry its own copy of it?
//
// It decides whether provisioning writes 68 files into every client folder. Getting
// it wrong is expensive in a way that produces no error at all: a per-tenant copy
// that drifts from the shared one is not a broken site, it is a client quietly
// running an older authentication bug while the app beside it is current.
//
// The two bugs this had, both found by testing rather than reading:
//
//   1. db_manager_api.php worked the answer out by comparing api_dir against this
//      file's parent directory. For the shared layout this server actually runs -
//      /var/www/api serving /var/www/<tenant>/ - that compared /var/www/api with
//      /var/www, concluded "not shared", and handed every client its own API copy.
//      Wrong on the one machine where it mattered, and silent.
//   2. The replacement probe looked for index.html one level too high, inside api/
//      itself, where a built application never is - so it reported "shared" for every
//      installation, including per-tenant ones.
//
// Both were invisible in the rendered output, which is why these tests build the
// layouts on disk and ask.

const LIB = new URL('../../api/lib/tenant-provision.php', import.meta.url)

const hasPhp = (() => {
  try {
    execFileSync('php', ['-v'], { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
})()

/**
 * Builds an installation and asks the real function inside it.
 *
 * @param sharedApi true for /api beside the tenant folders, false for a per-tenant copy
 * @param override  value for "shared_api" in the config, to test the override
 */
function probe({ sharedApi, override = undefined }) {
  const root = mkdtempSync(join(tmpdir(), 'cars-shared-api-'))
  mkdirSync(join(root, 'api', 'lib'), { recursive: true })
  copyFileSync(LIB.pathname, join(root, 'api', 'lib', 'tenant-provision.php'))

  if (!sharedApi) {
    // The per-tenant layout: the application is the sibling of api/, which is what
    // the probe looks for. In a shared layout there is no application at the root at
    // all - only other tenants' folders.
    writeFileSync(join(root, 'index.html'), '<!doctype html>')
  }

  let configPath = '/etc/cars-deploy.json'
  let libPath = join(root, 'api', 'lib', 'tenant-provision.php')

  if (override !== undefined) {
    configPath = join(root, 'cars-deploy.json')
    writeFileSync(configPath, JSON.stringify({ shared_api: override }))

    // The library reads a hardcoded path, so the copy is repointed at the fixture.
    // The replacement value keeps its quotes: the source text being matched is the
    // quoted literal, so substituting a bare path leaves `$path = /tmp/...` and a
    // parse error - which reads as a broken test rather than a broken patch.
    const quoted = `'${configPath}'`
    const patched = execFileSync(
      'php',
      [
        '-r',
        `echo str_replace("'/etc/cars-deploy.json'", ${JSON.stringify(quoted)}, file_get_contents($argv[1]));`,
        LIB.pathname,
      ],
      { encoding: 'utf8', maxBuffer: 4 * 1024 * 1024 },
    )

    libPath = join(root, 'api', 'lib', 'patched.php')
    writeFileSync(libPath, patched)
  }

  const entry = join(root, 'entry.php')

  // A separate entry file, because the library's own CLI guard fires when it is the
  // script being run - it would print its usage text instead of the answer.
  writeFileSync(entry, `<?php require ${JSON.stringify(libPath)}; echo json_encode(tenant_has_shared_api());`)

  const answer = execFileSync('php', [entry], { encoding: 'utf8' }).trim()
  execFileSync('rm', ['-rf', root])
  return answer
}

describe.skipIf(!hasPhp)('detecting whether the API is shared', () => {
  it('sees one api/ beside the tenant folders as shared', () => {
    // /var/www/api, with /var/www/acme/ and /var/www/acme_files/ beside it.
    expect(probe({ sharedApi: true })).toBe('true')
  })

  it('sees an api/ inside a client folder as that client\'s own copy', () => {
    // /var/www/acme/api, with /var/www/acme/index.html beside it.
    expect(probe({ sharedApi: false })).toBe('false')
  })

  it('can be told explicitly, which is what the config override is for', () => {
    // A stray index.html at the deployment root would make the probe answer wrongly
    // for a shared install. The override exists for that, and has to win over the
    // probe in both directions - including false on a layout the probe reads as
    // shared.
    expect(probe({ sharedApi: true, override: false })).toBe('false')
    expect(probe({ sharedApi: false, override: true })).toBe('true')
  })

  it('is what db_manager_api.php asks, rather than working it out again', () => {
    // The duplicate derivation is the bug: it produced the wrong answer on the
    // production layout while the library's answer was right. One source of truth.
    const api = execFileSync('php', [
      '-r',
      `echo file_get_contents(${JSON.stringify(new URL('../../api/db_manager_api.php', import.meta.url).pathname)});`,
    ]).toString()
    const fn = api.slice(api.indexOf('function dbm_has_shared_api()'))
    expect(fn.slice(0, 400)).toContain('tenant_has_shared_api()')
  })
})

describe('the shared copy is what the nginx config assumes', () => {
  it('serves the API out of the shared api/ for every tenant', () => {
    // If provisioning wrote per-tenant copies, this config would still point every
    // SCRIPT_FILENAME at the shared api/ - so the copies would be dead weight, and
    // tenant_status would be reporting on a folder nothing uses.
    const template = execFileSync('php', [
      '-r',
      `$c = json_decode(file_get_contents(${JSON.stringify(new URL('../../deploy/cars-deploy.example.json', import.meta.url).pathname)}), true);
       echo $c["api_dir"];`,
    ]).toString().trim()

    expect(template).toBe('/var/www/api')
    expect(existsSync(new URL('../../deploy/cars-deploy.example.json', import.meta.url))).toBe(true)
  })
})
