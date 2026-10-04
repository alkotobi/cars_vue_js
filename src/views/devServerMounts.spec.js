import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync, statSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { fileURLToPath } from 'node:url'

// vite.config.js decides which client folders the dev server serves as their own
// mount. It used to be a hardcoded list of one, and the failure it caused was
// invisible from the browser.
//
// A client provisioned after that list was written had no mount and no API proxy
// entry, so Vite's static middleware answered /<client>/api/api.php itself. Two
// rounds of the same bug looked like two different ones:
//
//   1. before server.fs.deny, the response was a 200 whose body was the PHP file's
//      own source. The SPA booted normally - the hashed assets are in the folder -
//      and every request then died in JSON.parse() with PHP comments as the
//      "response".
//   2. after server.fs.deny, the response was
//      403 The request url "/Users/.../nn/api/api.php" is outside of Vite serving
//      allow list
//      which reads like a filesystem permission problem and is not one.
//
// Both were fixed by restarting the dev server, which is the tell: anything whose fix
// is "restart it" was a snapshot taken at startup rather than a fact about the disk.

const viteConfig = readFileSync(new URL('../../vite.config.js', import.meta.url), 'utf8')
const repoRoot = fileURLToPath(new URL('../../', import.meta.url))

const resolveMount = viteConfig.slice(
  viteConfig.indexOf('const resolveFolderMount'),
  viteConfig.indexOf('const serveFolderMounts'),
)
const middleware = viteConfig.slice(
  viteConfig.indexOf('const serveFolderMounts'),
  viteConfig.indexOf('// Get environment'),
)
const rootHandler = viteConfig.slice(
  viteConfig.indexOf('const serveNothingAtRoot'),
  viteConfig.indexOf('export default defineConfig'),
)

describe('dev server tenant mounts', () => {
  it('keeps no list of clients', () => {
    // The bug was a client existing on disk and not existing in this file. A name in
    // the config is a name that has to be edited when someone provisions.
    expect(viteConfig).not.toMatch(/FOLDER_MOUNTS/)
    expect(viteConfig).not.toMatch(/const mounts = \[/)
  })

  it('resolves a mount per request, not when the config loads', () => {
    // This is the whole fix. A list built at startup is a snapshot, and provisioning
    // happens while the dev server is running.
    expect(middleware).toMatch(/resolveFolderMount\(decodeURIComponent\(firstSegment\)\)/)
    // Inside the middleware, not beside it.
    expect(middleware.indexOf('resolveFolderMount(')).toBeGreaterThan(
      middleware.indexOf('configureServer'),
    )
    expect(viteConfig).not.toMatch(/\.map\(serveFolderMount/)
  })

  it('registers one middleware, not one plugin per folder', () => {
    // Per-folder plugins are the same snapshot in a different shape: the plugin array
    // is built once, so a folder that appears later has no middleware.
    expect(middleware).toMatch(/name: 'serve-folder-mounts'/)
    expect(viteConfig).toMatch(/^\s+serveFolderMounts,$/m)
  })

  it('recognises a tenant the way the app does', () => {
    // db_code.json is api/lib/appdb.php's test for "which database does this
    // deployment serve" - present means provisioned. api/ is what the proxy needs to
    // reach, because the dev server executes the tenant's PHP at
    // /<mount>/api/... inside that folder.
    expect(resolveMount).toMatch(/existsSync\(join\(root, name, 'db_code\.json'\)\)/)
    expect(resolveMount).toMatch(/isDirectory\(join\(root, name, 'api'\)\)/)
  })

  it('survives a folder with a db_code.json and no api/', () => {
    // Not hypothetical: mig/ and mig1/ are in the repository root right now. An
    // earlier version of this filter called statSync() straight after existsSync()
    // returned true for db_code.json; statSync() on the missing api/ threw, and the
    // config - and the dev server with it - did not load.
    // The guard itself lives at module scope, next to folderMountFor().
    expect(viteConfig).toMatch(
      /const isDirectory = \(path\) => existsSync\(path\) && statSync\(path\)\.isDirectory\(\)/,
    )
    expect(resolveMount).not.toMatch(/existsSync\([^)]*'db_code\.json'\)\) && statSync\(/)
  })

  it('validates the name before it reaches a path', () => {
    // The first path segment of a request becomes a folder name here, so it is
    // checked against the same shape api/lib/tenant-provision.php accepts rather
    // than trusted.
    expect(resolveMount).toMatch(/if \(!\/\^\[a-z\]\[a-z0-9_\]\*\$\/\.test\(name\)\) return null/)
  })

  it("proxies any mount's API by pattern, because the proxy table is also a snapshot", () => {
    // One entry per client has the same defect: the table is built when the config
    // loads. The pattern also matches paths that are not mounts, and those are
    // forwarded to a PHP server with no such file, which answers 404.
    expect(viteConfig).toMatch(/'\^\/\[\^\/\]\+\/api': \{\n\s+target: apiUrl/)
    expect(viteConfig).not.toMatch(/\.\.\.Object\.fromEntries\(\s*FOLDER_MOUNTS/)
  })

  it('never serves PHP as a document', () => {
    // A tenant's api/ copy carries config.local.php and db_manager_config.local.php.
    // Served as text, the dev server hands over the database password; reached
    // through the proxy, it is executed and prints nothing. Same rule the rendered
    // nginx config keeps: no generic PHP location.
    expect(viteConfig).toMatch(/fs: \{\s*\n\s*deny: \['\*\*\/\*\.php'\]/)
  })

  it('still refuses to serve a tenant file from outside its folder', () => {
    // Per-request resolution did not weaken the containment check: the resolved
    // folder is still the boundary, and '..' is still collapsed before comparing.
    expect(middleware).toMatch(/normalize\(relative\)/)
    expect(middleware).toMatch(/if \(!isInside\(target\)\)/)
  })
})

describe('the dev server root', () => {
  it('is registered as a plugin', () => {
    expect(rootHandler).toMatch(/name: 'serve-nothing-at-root'/)
    expect(viteConfig).toMatch(/^\s+serveNothingAtRoot,$/m)
  })

  it('answers exactly / and nothing else, and defers the rest', () => {
    // Exact equality, not startsWith('/'). A prefix test here swallows every request
    // on the server - the tenant mounts and /cars/ included - and the failure is a
    // blank page at a URL that used to work, which points at the browser.
    expect(rootHandler).toMatch(
      /if \(pathname !== '\/' && pathname !== '\/index\.html'\) \{\s+next\(\)\s+return\s+\}/,
    )
  })

  it('refuses rather than redirecting', () => {
    // The bug was the redirect, not its destination: 302 -> /cars/ is what handed the
    // browser a login form that could never authenticate. Redirecting somewhere else
    // would reproduce it somewhere else.
    expect(rootHandler).toMatch(/res\.statusCode = 404/)
    expect(rootHandler).not.toMatch(/writeHead|Location|302/)
  })

  it('runs before Vite\'s internal base redirect', () => {
    // Registered inside configureServer, not returned from it. A returned hook is a
    // post hook: Vite installs the base redirect first, so / becomes a 302 to /cars/
    // before this middleware is ever reached, and it reads as though it did nothing.
    expect(rootHandler).toMatch(/configureServer\(server\) \{\s+server\.middlewares\.use\(/)
    expect(rootHandler).not.toMatch(/configureServer\(server\) \{[\s\S]*return \(\) =>/)
  })

  it('serves no file in its place', () => {
    // The webroot root is where a welcome page belongs, and index.html is the SPA
    // shell the tenant bundles are built from. Reusing it would put the SPA back at /
    // under another name - and it would still be the app that cannot authenticate.
    expect(rootHandler).not.toMatch(/createReadStream|sendIndex|existsSync|readFileSync/)
  })

  it('answers the way production already does', () => {
    // The premise behind the 404, asserted here because it is the part that decays:
    // if a welcome page ever ships at the webroot root, dev should follow it instead.
    // The rendered config has no `location = /`, so / falls through to
    // `root __WEBROOT__; index index.html;` -> __WEBROOT__/index.html, and deploy.sh
    // only writes into $WEBROOT/$FOLDER/.
    const template = readFileSync(
      new URL('../../deploy/nginx-app.conf.template', import.meta.url),
      'utf8',
    )
    expect(template).toMatch(/root __WEBROOT__;/)
    expect(template).not.toMatch(/^\s*location = \/\s*\{/m)

    const deploy = readFileSync(new URL('../../deploy/deploy.sh', import.meta.url), 'utf8')
    expect(deploy).toMatch(/^TARGET="\$WEBROOT\/\$FOLDER"$/m)

    // Every rsync destination is a path inside the tenant folder. None of them is
    // $WEBROOT itself, which is what makes __WEBROOT__/index.html a file that does
    // not exist and the dev server's 404 the right answer to copy.
    const destinations = [...deploy.matchAll(/\$SSH_TARGET:(\$TARGET[^"']*)/g)].map((m) => m[1])
    expect(destinations.length).toBeGreaterThan(0)
    for (const dest of destinations) {
      expect(dest).toMatch(/^\$TARGET\/(\w+\/)?$/)
    }
    expect(deploy).not.toMatch(/\$SSH_TARGET:\$WEBROOT/)
  })
})

describe('what that resolution finds in this checkout', () => {
  const isDirectory = (path) => existsSync(path) && statSync(path).isDirectory()
  // Discovered the way vite.config.js's resolveFolderMount() does, rather than from a
  // hardcoded list of client names. The list was ['m', 'mig_27', 'nono', 'nn'], so when
  // those tenants were deleted and `merhab` was provisioned in their place it still
  // matched nothing and the assertion below failed on an empty list - the spec had
  // stopped describing this checkout and started describing a deleted one. Scanning
  // for db_code.json + api/ is the condition tenant-provision.php actually satisfies,
  // so a newly provisioned tenant is covered without editing this file.
  const mounts = readdirSync(repoRoot).filter(
    (name) =>
      /^[a-z][a-z0-9_]*$/.test(name) &&
      existsSync(join(repoRoot, name, 'db_code.json')) &&
      isDirectory(join(repoRoot, name, 'api')),
  )

  it('covers every client that is actually provisioned here', () => {
    // If a client is provisioned and this list does not grow, resolution has stopped
    // matching what tenant-provision.php writes - which is the original bug wearing a
    // different hat.
    expect(mounts.length).toBeGreaterThan(0)
    for (const name of mounts) {
      expect(existsSync(join(repoRoot, name, 'index.html'))).toBe(true)
      expect(existsSync(join(repoRoot, `${name}_files`))).toBe(true)
    }
  })

  it("does not mistake Vite's own directories for clients", () => {
    // public/ holds the root app's db_code.json and is in git, and dist/ has one
    // stripped by removeDbCode(). Neither is a client, and mounting them would put a
    // build folder and a source folder on the same footing as a tenant.
    expect(mounts).not.toContain('public')
    expect(mounts).not.toContain('dist')
    expect(mounts).not.toContain('api')
  })
})
