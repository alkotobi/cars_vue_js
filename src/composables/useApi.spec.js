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

describe('loginErrorText tells a refused login apart by its cause', () => {
  beforeEach(() => {
    // useApi reads window while it initialises, so the dynamic import in the first
    // case below needs one in place before it runs.
    stubBrowser()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.resetModules()
  })

  // LoginView used to render auth.invalidCredentials for every unsuccessful login, so
  // a wrong password and an unreachable database produced the same sentence. The
  // cost was not cosmetic: the message named the wrong cause, so the reader reset a
  // password that did not need resetting and never learned the server was down.
  const t = (key) => key

  it('only calls bad credentials bad credentials', async () => {
    const { loginErrorText } = await import('./useApi')

    // api/actions/auth.php answers a wrong username AND a wrong password with this
    // one code - it cannot tell them apart, and neither can the reader.
    expect(loginErrorText(t, { success: false, code: 'invalid_credentials' })).toBe(
      'auth.invalidCredentials',
    )
  })

  it('names an unreachable database as such', async () => {
    const { loginErrorText } = await import('./useApi')

    expect(loginErrorText(t, { success: false, code: 'db_unavailable' })).toBe(
      'auth.serverUnavailable',
    )
  })

  it('never guesses credentials from a code it does not know', async () => {
    const { loginErrorText } = await import('./useApi')

    // The deliberate difference from apiErrorText()/colorErrorText(), which return ''
    // for unknown codes so the caller can fall back to a message about its own
    // action. A login form has one button and no action to name, so there is nothing
    // to fall back to - but claiming bad credentials it cannot prove would be worse
    // than a generic failure.
    expect(loginErrorText(t, { success: false, code: 'brand_new_code' })).toBe('auth.loginError')
    expect(loginErrorText(t, { success: false })).toBe('auth.loginError')
    expect(loginErrorText(t, {})).toBe('auth.loginError')
    expect(loginErrorText(t, null)).toBe('auth.loginError')
    expect(loginErrorText(t, undefined)).toBe('auth.loginError')
  })

  it('is a named export, not a member of what useApi() returns', async () => {
    const { loginErrorText, useApi } = await import('./useApi')

    // This is the whole bug this case exists for. The three error translators are
    // top-level exports of the module, sibling to useApi() rather than part of its
    // return value, so `const { loginErrorText } = useApi()` destructures to undefined
    // and every login died on "loginErrorText is not a function" - only once a
    // password was actually wrong, because the success branch never calls it.
    expect(typeof loginErrorText).toBe('function')

    // Same trap, asserted against the real object rather than the source: the
    // translators are absent from it, which is why they must be imported by name.
    const api = useApi()
    expect(api.loginErrorText).toBeUndefined()
    expect(api.apiErrorText).toBeUndefined()
    expect(api.colorErrorText).toBeUndefined()
    // And the members that DO come off it, so the contrast is pinned rather than
    // assumed.
    expect(typeof api.callApi).toBe('function')
  })

  it('is imported by the login form as a named import', () => {
    const loginView = readFileSync(new URL('../views/LoginView.vue', import.meta.url), 'utf8')

    expect(loginView).toMatch(
      /import\s*\{[^}]*\bloginErrorText\b[^}]*\}\s*from\s*'\.\.\/composables\/useApi'/,
    )
    // Belt and braces, and the assertion whose absence let the bug through: a
    // `/useApi\(\)[\s\S]*loginErrorText/` style check is satisfied by the broken
    // destructuring itself, because the symbol appears after the call it was wrongly
    // pulled from. Matching the destructure directly is what actually constrains it.
    expect(loginView).not.toMatch(/const\s*\{[^}]*\bloginErrorText\b[^}]*\}\s*=\s*useApi\(\)/)

    expect(loginView).toContain('loginErrorText(t, result)')
    // The old line must not survive anywhere on the failure branch.
    expect(loginView).not.toMatch(/error\.value = t\('auth\.invalidCredentials'\)/)
    // And the envelope is logged, because the copy above is only a summary of it.
    expect(loginView).toContain("console.error('Login rejected:', result)")
  })

  it('has copy for the new code in every locale', async () => {
    const { loginErrorText } = await import('./useApi')

    // A key that exists only in en.json renders as the raw key string to every other
    // reader, which is the exact failure this file exists to stop.
    for (const locale of ['en', 'ar', 'fr', 'zh']) {
      const messages = JSON.parse(
        readFileSync(new URL(`../locales/${locale}.json`, import.meta.url), 'utf8'),
      )
      const translate = (key) => messages.auth?.[key.replace('auth.', '')]

      expect(messages.auth.serverUnavailable).toBeTruthy()
      expect(loginErrorText(translate, { code: 'db_unavailable' })).toBe(
        messages.auth.serverUnavailable,
      )
      expect(loginErrorText(translate, { code: 'invalid_credentials' })).toBe(
        messages.auth.invalidCredentials,
      )
      expect(loginErrorText(translate, { code: 'nope' })).toBe(messages.auth.loginError)
    }
  })
})
