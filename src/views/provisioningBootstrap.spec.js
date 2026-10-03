import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'

// Provisioning copies a template. Nothing built one.
//
// tenant_provision() takes a seed_source and copies the reference tables out of it,
// and tenant_assert_template_clean() then refuses any template holding client data -
// which is the right pair of behaviours and still leaves the bootstrap circular. The
// only reference database available on a fresh server is the local development one,
// which is a live tenant, so the check correctly rejects it and the operator is
// left with a refusal and no documented way forward.
//
// The end-to-end run that first proved the flow together worked around this by
// writing a throwaway script to build the template. That is exactly the kind of step
// that does not survive to the second server, so --build-template exists to make it
// a documented command.

const LIB = readFileSync(new URL('../../api/lib/tenant-provision.php', import.meta.url), 'utf8')
const DBM_API = readFileSync(new URL('../../api/db_manager_api.php', import.meta.url), 'utf8')
const EXAMPLE_CONFIG = JSON.parse(
  readFileSync(new URL('../../deploy/cars-deploy.example.json', import.meta.url), 'utf8'),
)
const DEPLOYMENT_DOC = readFileSync(new URL('../../DEPLOYMENT.md', import.meta.url), 'utf8')

describe('building the template', () => {
  it('is a command, not something you write a script for', () => {
    expect(LIB).toMatch(/--build-template/)
    expect(LIB).toMatch(/function tenant_build_template/)
  })

  it('loads the credentials the CLI needs', () => {
    // The credential helpers read $db_config and $db_manager_config, which only
    // api/config.php sets. Without this the CLI entry guard fires and every verb
    // fails with "config.php has not been loaded" - so both verbs were unusable from
    // a shell and the documented usage line was a description of something that
    // could not be run.
    expect(LIB).toMatch(/require_once __DIR__ \. '\/\.\.\/config\.php'/)
    expect(LIB).toMatch(/require_once __DIR__ \. '\/\.\.\/db_manager_config\.php'/)
  })

  it('refuses to rebuild over an existing database without --force', () => {
    expect(LIB).toMatch(/refusing: \{\$dbName\} already exists/)
    expect(LIB).toMatch(/\$arg === '--force'/)
  })

  it('refuses to build a template from itself', () => {
    expect(LIB).toMatch(/would be built from itself/)
  })

  it('verifies the result before reporting success', () => {
    // The same check provisioning applies to a template it is handed, applied to the
    // one it just made. A build command that trusted its own output would defeat the
    // check that makes templates usable at all.
    const build = LIB.slice(LIB.indexOf('function tenant_build_template'))

    expect(build).toMatch(/tenant_assert_template_clean/)
    expect(build).toMatch(/tenant_seed_reference/)
  })

  it('reports where the reference rows came from', () => {
    // setup.sql carries the reference INSERTs, so on this project the schema step
    // fills those tables and the copy step moves nothing. Reporting only what was
    // copied would print "0 reference tables" for a template holding 39 permissions:
    // accurate, and useless.
    expect(build_summarises_populated_tables()).toBe(true)
  })

  it('is documented where an operator will look for it', () => {
    expect(DEPLOYMENT_DOC).toMatch(/--build-template cars_template/)
    expect(EXAMPLE_CONFIG._template_database.join(' ')).toMatch(/--build-template/)
  })
})

function build_summarises_populated_tables() {
  const build = LIB.slice(LIB.indexOf('function tenant_build_template'))

  return /TENANT_SEED_TABLES/.test(build) && /setup\.sql/.test(build)
}

describe('finding the server configuration', () => {
  // The wizard reported "template database: not set in /etc/cars-deploy.json" on a
  // development machine, and there was no way to fix it: /etc needs sudo, and the
  // only alternative was an environment variable on the PHP process - which is
  // exactly the kind of setup step that gets forgotten, and its absence shows up as
  // a disabled button rather than as a missing file.
  it('has a development fallback outside the web root', () => {
    // Outside api/ because api/ is served: this file holds registry credentials, so a
    // copy under api/ would be readable over HTTP.
    expect(LIB).toMatch(/deploy\/cars-deploy\.local\.json/)
    expect(LIB).toMatch(/dirname\(__DIR__, 2\)/)
  })

  it('is git-ignored, because it is per-machine', () => {
    const gitignore = readFileSync(new URL('../../.gitignore', import.meta.url), 'utf8')

    expect(gitignore).toMatch(/deploy\/cars-deploy\.local\.json/)
  })

  it('lets /etc win over the development file', () => {
    // The production file is root-owned and authoritative. A stale copy in a checkout
    // must never quietly become the server's configuration - and deploy.sh does not
    // ship deploy/, so the file cannot reach a server in the first place.
    // From the code, not the comment above it: the comment names the dev file too,
    // so slicing from the prose compares a sentence against a statement.
    const resolve = LIB.slice(
      LIB.indexOf("$path = '/etc/cars-deploy.json';"),
      LIB.indexOf("$defaults['config_path']"),
    )

    expect(resolve).toMatch(/\$path = '\/etc\/cars-deploy\.json'/)
    // The /etc file is only bypassed when it is genuinely absent, and the dev file is
    // inside that "absent" branch rather than beside it.
    expect(resolve).toMatch(/elseif \(!is_file\(\$path\)/)
    expect(resolve.indexOf('!is_file($path)')).toBeLessThan(
      resolve.indexOf('cars-deploy.local.json'),
    )
  })

  it('says which file answered, so a missing setting names an editable one', () => {
    // "not set in /etc/cars-deploy.json" is a dead end for someone who cannot write
    // to /etc. The detail strings now come from the file that was actually read.
    expect(LIB).toMatch(/\$defaults\['config_path'\] = is_file\(\$path\) \? \$path : ''/)
    expect(DBM_API).toMatch(/\$configWhere = \(string\) \(\$serverConfig\['config_path'\]/)
  })
})
