import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'

// Onboarding a client used to mean SSH: create the database, copy the build, write
// the nginx block. The API for all of that exists now, but the screen still only
// offered "Create", which runs the legacy schema-only action - no reference data, no
// admin account, no folders, no URL. So the operator's options were a shell or a
// half-built client, and the second one looked identical to the first until someone
// tried to log in.
//
// These tests are about the shape of the flow rather than the rendering: what the
// screen asks for before it acts, what it refuses to do, and what it keeps on screen
// when something fails partway. All three are decisions that are invisible in a
// screenshot and easy to undo by accident.

const DATABASES = readFileSync(
  new URL('../components/db-manager/Databases.vue', import.meta.url),
  'utf8',
)
const DBM_API = readFileSync(new URL('../../api/db_manager_api.php', import.meta.url), 'utf8')
const LIB = readFileSync(new URL('../../api/lib/tenant-provision.php', import.meta.url), 'utf8')

const functionBody = (name) => {
  const start = DATABASES.indexOf(`const ${name} = `)

  expect(start, `${name} should exist in Databases.vue`).toBeGreaterThan(-1)

  // Take the declaration to the next top-level `const`, which is where these
  // handlers end. Reading to a fixed line count would silently start asserting
  // about a neighbour as soon as one of these grew.
  const rest = DATABASES.slice(start + 1)
  const next = rest.search(/^const \w+ = /m)

  return next === -1 ? rest : rest.slice(0, next)
}

describe('the Provision screen', () => {
  it('is reachable from a registry row, and is not the legacy Create button', () => {
    expect(DATABASES).toMatch(/@click="openProvisionModal\(db\)"/)
    expect(DATABASES).toMatch(/class="btn-provision"/)
  })

  it('reads the server and the client before it offers to change either', () => {
    // Both requests are read-only and independent, so they go together. Doing this
    // first is what lets the dialog say "webroot is not writable" instead of letting
    // the operator press a button that fails on it.
    const body = functionBody('openProvisionModal')

    expect(body).toMatch(/deployment_config/)
    expect(body).toMatch(/tenant_status/)
    expect(body).toMatch(/Promise\.all/)
  })

  it('will not start provisioning while the server is not ready', () => {
    // The alternative is a run that creates the database and then fails on
    // permissions, leaving a half-built client and an error that does not say which
    // of the six prerequisites was the problem.
    expect(DATABASES).toMatch(/:disabled="provisioning \|\| provisionLoading \|\| !serverConfig\?\.ready"/)
  })

  it('keeps the root step out of the provisioning step', () => {
    // reload_nginx is the only privileged operation, and the only one whose effects
    // reach beyond this client - including taking other clients offline if the
    // registry read goes wrong. It is a button the operator presses, never a
    // side effect of setting a client up.
    expect(DATABASES).toMatch(/@click="applyNginx"/)
    expect(functionBody('confirmProvision')).not.toMatch(/reload_nginx/)
    expect(functionBody('openProvisionModal')).not.toMatch(/reload_nginx/)
  })

  it('surfaces the server log instead of only the failure', () => {
    // A run that stops at "duplicate column" is diagnosable from the lines above it,
    // and those lines exist only in the response. Showing just the message throws
    // away the only evidence, and the operator has no access to error_log.
    expect(DATABASES).toMatch(/result\?\.data\?\.log/)
    expect(DATABASES).toMatch(/class="provision-log"/)
  })

  it('leaves a partial run in place instead of clearing the evidence', () => {
    expect(DATABASES).toMatch(/log stays on screen/)
  })
})

describe('what blocks which button', () => {
  // The bug this pins down: every check was folded into one boolean called `ready`,
  // and Provision was gated on it. Two of those checks - the root-only nginx ones -
  // have nothing to do with provisioning, so on any machine that is not the
  // production server the button was permanently disabled. The flow could then only
  // be exercised by calling the API by hand, which is how it ended up with a bug in
  // the first place: nothing in the UI ever ran it.
  it('does not gate provisioning on the nginx setup', () => {
    // Provisioning runs as the web user and creates a database, folders and
    // db_code.json. It never needs root.
    expect(DATABASES).toMatch(/:disabled="provisioning \|\| provisionLoading \|\| !serverConfig\?\.ready"/)
    expect(DATABASES).not.toMatch(
      /:disabled="provisioning \|\| provisionLoading \|\| ![^"]*nginx_ready[^"]*"/,
    )
  })

  it('gates Apply nginx on the nginx setup instead', () => {
    expect(DATABASES).toMatch(/applyingNginx \|\| provisioning \|\| !serverConfig\?\.nginx_ready/)
  })

  it('tells the two apart on the server, so neither question hides the other', () => {
    expect(DBM_API).toMatch(/'needs' => 'provision'/)
    expect(DBM_API).toMatch(/'needs' => 'nginx'/)
    expect(DBM_API).toMatch(/'nginx_ready' => \$unmet\('nginx'\) === \[\]/)
    expect(DATABASES).toMatch(/nginx not set up/)
    expect(DATABASES).toMatch(/ready to provision/)
  })

  it('says what is missing by name, per group', () => {
    // "Not ready" with no list is what sent this looking for a permissions problem
    // that was really a root setup step.
    expect(DBM_API).toMatch(/'unmet' => \$unmet\('provision'\)/)
    expect(DATABASES).toMatch(/serverConfig\.unmet\.join/)
    expect(DATABASES).toMatch(/serverConfig\.nginx_unmet\.join/)
  })

  it('does not block on a missing build, and leaves a way to add it later', () => {
    // tenant_deploy_app() threw when there was no build to copy, so the run stopped at
    // the last step with the client already half-written. Now the copy is skipped with
    // a message - and the button that performs it must therefore not be hidden behind
    // a fully-complete status, or there is no way back.
    expect(LIB).toMatch(/app copy: skipped, no build to copy/)
    expect(DATABASES).toMatch(/v-if="hasCanonicalBuild"/)
    expect(DATABASES).not.toMatch(/v-if="provisionStatus\?\.ready"[\s\S]{0,120}@click="deployApp"/)
  })
})

describe('what the browser is allowed to decide', () => {
  it('does not choose which database a client is built from', () => {
    // The template is server configuration, not a request parameter. A client that
    // could name its own seed source would be choosing what reference data - and
    // which admin credentials - it receives.
    expect(functionBody('confirmProvision')).not.toMatch(/seed_source/)
    expect(functionBody('confirmProvision')).not.toMatch(/seed_source|seedSource/)
  })

  it('never builds a filesystem path', () => {
    // Paths are derived server-side from a validated database name. A path assembled
    // in the browser is a path the server did not check, and this component is the
    // one place a future "custom install path" field would be wired in.
    for (const handler of ['confirmProvision', 'applyNginx', 'deployApp']) {
      expect(functionBody(handler)).not.toMatch(/\/var\/www|webroot|\.\.\//)
    }
  })

  it('sends the database name the server already knows, and nothing else', () => {
    const body = functionBody('confirmProvision')

    expect(body).toMatch(/dbManagerRequest\('provision_tenant', \{\s*db_name:/)
  })
})

describe('the legacy Create button', () => {
  it('is kept, and says what it does not do', () => {
    // It is still the right thing for a schema-only rebuild of an existing client,
    // and removing it would take that away. What was missing was the distinction:
    // both buttons created a database and only one of them made a usable client.
    expect(DATABASES).toMatch(/@click="createTables\(db\)"/)
    expect(DATABASES).toMatch(/Prefer Provision/)
  })

  it('still does not claim to provision anything', () => {
    expect(functionBody('confirmCreateTables')).not.toMatch(/provision/)
  })
})
