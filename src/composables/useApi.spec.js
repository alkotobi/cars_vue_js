import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'

// Regression tests for the production incident where `isLocalhost` was deleted
// during an API-URL refactor while four call sites still referenced it. Nothing
// failed at build or import time: the module parsed fine and only threw
// ReferenceError when a request actually ran, so every API call, the cookie
// verification, the version check and the alerts broke on a live site.
//
// These tests execute the real code paths, so the same mistake fails CI here
// instead of in production. useApi.js reads window.location and localStorage at
// call time and the vitest environment is `node`, so the few globals it needs
// are stubbed here rather than pulling in jsdom for two identifiers.

const SOURCE = readFileSync(new URL('./useApi.js', import.meta.url), 'utf8')

function stubBrowser({ hostname = 'world-automobile.com', protocol = 'https:' } = {}) {
  const store = new Map()
  vi.stubGlobal('window', {
    location: {
      href: `${protocol}//${hostname}/cars/`,
      origin: `${protocol}//${hostname}`,
      protocol,
      host: hostname,
      hostname,
      pathname: '/cars/',
    },
  })
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
    clear: () => store.clear(),
  })
}

const json = (body) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  })

// useApi.js loads per-server config from db_code.json before any request, so a
// bare mock leaves callApi rejecting on config and the real request is never
// reached. Answering that one file lets the genuine API URL be asserted.
function stubFetch() {
  const calls = []
  const fetchMock = vi.fn(async (url) => {
    const href = String(url)
    calls.push(href)
    if (href.includes('db_code.json')) {
      return json({ db_name: 'merhab', uplod_path: 'mig_files' })
    }
    return json({ success: true, data: {} })
  })
  vi.stubGlobal('fetch', fetchMock)
  return calls
}

describe('useApi isLocalhost guards', () => {
  beforeEach(() => {
    stubBrowser()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    vi.resetModules()
  })

  it('declares isLocalhost at module scope', () => {
    // Every call site references this name. Deleting the declaration is what
    // shipped the incident, so pin it here.
    expect(SOURCE).toMatch(/(?:const|let|var)\s+isLocalhost\s*=/)
  })

  it('derives isLocalhost from the build, not the hostname', () => {
    // A `hostname.startsWith('192.168.')` test silently disabled cookie
    // verification on a server genuinely deployed on a LAN.
    expect(SOURCE).toMatch(/isLocalhost\s*=\s*Boolean\(import\.meta\.env\?*\.?DEV\)/)
    expect(SOURCE).not.toMatch(/startsWith\(\s*['"]192\.168\./)
  })

  it('runs a request without ReferenceError (the shipped failure)', async () => {
    // The exact production failure: module scope read window.location, and the
    // missing binding threw as soon as a request ran.
    const calls = stubFetch()
    const { useApi } = await import('./useApi')
    const api = useApi()

    await expect(api.callApi({ action: 'noop' })).resolves.toBeDefined()
    expect(calls.length).toBeGreaterThan(0)
  })

  // NOTE: the folder/domain/IP matrix for the API base lives in
  // src/utils/basePath.spec.js, where import.meta.env.DEV can be controlled.
  // Vitest sets DEV=true here, so useApi takes the dev branch (php -S on :8000)
  // and the production URL shape cannot be asserted from this file.

  it('builds the API base from the shared resolver, not inline host logic', () => {
    expect(SOURCE).toContain('resolveApiBaseUrl')
    // Inline `protocol}//${hostname}` building is what drifted between the five
    // copies of this helper before it was centralised.
    expect(SOURCE).not.toMatch(/`\$\{protocol\}\/\/\$\{hostname\}/)
  })

  it('has no module constant that is referenced but never declared', () => {
    // Scoped to the names this file defines itself. A general scan is not
    // viable: SQL lives in string literals here, so SELECT/JOIN/FROM read as
    // undefined identifiers and the check becomes noise nobody keeps.
    const required = [
      'hostname',
      'protocol',
      'getBasePath',
      'BASE_PATH',
      'isLocalhost',
      'API_BASE_URL',
    ]
    const undeclared = required.filter(
      (n) => !new RegExp(`\\b(?:const|let|var|function)\\s+${n}\\b`).test(SOURCE),
    )
    expect(undeclared).toEqual([])
  })
})
