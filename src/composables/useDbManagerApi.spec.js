import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'

// Behavioural tests for the db-manager client.
//
// The bug these exist for was invisible to a green suite: the DB manager was 100%
// unreachable behind its own auth gate, and nothing noticed because no test ever
// called the endpoint. A token can also stop being valid while the page is open -
// signed out in another tab, or api_token cleared by hand in the registry - and the
// old code noticed nothing then either: DbManagerView decided "logged in" from the
// presence of a localStorage key, so the user got a permanently broken shell instead
// of the login form.
//
// So these execute the real module against a stubbed fetch and localStorage, rather
// than asserting on its source text. The node environment has no window or
// localStorage, so the few globals it needs are stubbed here - the same approach
// useApi.spec.js uses, for the same reason.

const AUTH_FAILURE = {
  success: false,
  message: 'Not authenticated. Sign in to the database manager.',
}

let store
let fetchMock
let api

// Re-imported per test. useDbManagerApi.js remembers which dead token it has already
// reported (so several in-flight failures do not each re-render the login form), and
// that state lives as long as the module. A single shared import would let one test's
// report silence the next one's listener, which reads as a flake rather than as the
// dedupe it actually is.
async function loadApi() {
  vi.resetModules()
  api = await import('./useDbManagerApi.js')
  return api
}

// Answer with a JSON envelope.
const replyJson = (body, status = 200) =>
  new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })

// The auth failure, exactly as db_manager_api.php emits it.
const replyAuthFailure = () => replyJson(AUTH_FAILURE)

function lastBody() {
  return JSON.parse(fetchMock.mock.calls[fetchMock.mock.calls.length - 1][1].body)
}

beforeEach(async () => {
  await loadApi()
  store = new Map()
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
    clear: () => store.clear(),
  })
  vi.stubGlobal('window', {
    location: {
      href: 'https://world-automobile.com/cars/',
      origin: 'https://world-automobile.com',
      protocol: 'https:',
      host: 'world-automobile.com',
      hostname: 'world-automobile.com',
      port: '',
      pathname: '/cars/',
    },
  })
  fetchMock = vi.fn(async () => replyJson({ success: true, data: [] }))
  vi.stubGlobal('fetch', fetchMock)
})

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('the token', () => {
  it('is read from the db-manager session', () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    expect(api.getDbManagerToken()).toBe('abc123')
  })

  it('is absent before login', () => {
    expect(api.getDbManagerToken()).toBe(null)
  })

  it('is absent for a corrupt value rather than throwing', () => {
    // A partial write must not take navigation down with it.
    store.set('db_manager_user', '{"token":')
    expect(api.getDbManagerToken()).toBe(null)
    store.set('db_manager_user', '{"user":"admin"}')
    expect(api.getDbManagerToken()).toBe(null)
  })

  it('is absent for a session with no token field', () => {
    // This is the shape a pre-migration session has, and is why the view cannot treat
    // the key's presence as proof of a live session.
    store.set('db_manager_user', JSON.stringify({ id: 1, user: 'admin' }))
    expect(api.getDbManagerToken()).toBe(null)
  })

  it('is cleared locally on logout', () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    api.clearDbManagerToken()
    expect(api.getDbManagerToken()).toBe(null)
  })
})

describe('requests carry the credential', () => {
  it('attaches the token', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest('get_databases')
    expect(lastBody().token).toBe('abc123')
  })

  it('posts, so the token never reaches an access log', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest('get_databases')
    expect(fetchMock.mock.calls[0][1].method).toBe('POST')
    expect(String(fetchMock.mock.calls[0][0])).not.toMatch(/\?/)
  })

  it('puts the action and the fields in the body', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest('delete_db_update', { id: 7 })
    expect(lastBody()).toEqual({ action: 'delete_db_update', id: 7, token: 'abc123' })
  })

  it('resolves the URL the same way the rest of the app does', async () => {
    // Not `https://host/api`: the API ships inside the app folder, so that 404s on
    // every /folder deploy. Compared against the shared resolver rather than a
    // hard-coded string, because it legitimately differs between environments - in
    // dev, import.meta.env.DEV is true and PHP is served from the root without the
    // mount point. What matters is that this client cannot drift from api.php's.
    const { resolveApiBaseUrl } = await import('../utils/basePath.js')
    const base = resolveApiBaseUrl({
      override: import.meta.env?.VITE_API_BASE_URL,
      protocol: 'https:',
      hostname: 'world-automobile.com',
      port: '',
      basePath: '/cars/',
      isDev: import.meta.env.DEV,
    })

    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest('get_databases')

    const url = String(fetchMock.mock.calls[0][0])
    expect(url).toBe(`${base}/db_manager_api.php`)
    // The bug this guards against was a doubled segment.
    expect(url).not.toMatch(/api\/api\//)
  })

  it('honours an explicit base URL', async () => {
    // EditDbCodeJson.vue passes its api-base-url prop through instead of re-resolving.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest(
      'read_db_code_json',
      { database_id: 1 },
      { baseUrl: 'https://other.example/custom' },
    )
    expect(String(fetchMock.mock.calls[0][0])).toBe(
      'https://other.example/custom/db_manager_api.php',
    )
  })

  it('sends no token for an anonymous action', async () => {
    // login and signup run before there is a token; one leaked here would imply the
    // gate is optional.
    await api.dbManagerRequest('login', { user: 'admin', pass: 'x' }, { anonymous: true })
    expect(lastBody()).toEqual({ action: 'login', user: 'admin', pass: 'x' })
  })

  it('still sends one for an anonymous action if a session happens to exist', async () => {
    // A stale session in another tab must not leak into a login attempt.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'abc123' })
    await api.dbManagerRequest('login', { user: 'admin', pass: 'x' }, { anonymous: true })
    expect(lastBody().token).toBeUndefined()
  })
})

describe('a revoked session is recognised', () => {
  it('spots the gate rejection', () => {
    expect(api.isDbManagerSessionLost(AUTH_FAILURE)).toBe(true)
  })

  it('does not mistake an ordinary failure for a dead session', () => {
    // Logging users out over a failed backup or a bad version number would be worse
    // than the problem it was meant to solve.
    expect(api.isDbManagerSessionLost({ success: false, message: 'File already exists' })).toBe(
      false,
    )
    expect(api.isDbManagerSessionLost({ success: true, data: [] })).toBe(false)
    expect(api.isDbManagerSessionLost(null)).toBe(false)
    expect(api.isDbManagerSessionLost(undefined)).toBe(false)
  })

  it('tolerates the message being reworded after the prefix', () => {
    expect(
      api.isDbManagerSessionLost({ success: false, message: 'Not authenticated. Try again.' }),
    ).toBe(true)
  })

  it('still needs success to be false', () => {
    expect(api.isDbManagerSessionLost({ success: true, message: 'Not authenticated.' })).toBe(false)
  })
})

describe('a revoked token drops the session', () => {
  it('notifies listeners and clears the credential', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    const seen = []
    const stop = api.onDbManagerSessionLost(() => seen.push('lost'))

    await api.dbManagerRequest('get_databases')
    expect(seen).toEqual(['lost'])
    expect(api.getDbManagerToken()).toBe(null)

    stop()
  })

  it('has already cleared storage by the time listeners run', async () => {
    // DbManagerView derives isLoggedIn from that same key, so a listener reading
    // storage would see the dead token still present and render the panels.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    let tokenSeenByListener = 'unset'
    const stop = api.onDbManagerSessionLost(() => {
      tokenSeenByListener = api.getDbManagerToken()
    })

    await api.dbManagerRequest('get_databases')
    expect(tokenSeenByListener).toBe(null)

    stop()
  })

  it('reports once per dead token, not once per in-flight request', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    await Promise.all([
      api.dbManagerRequest('get_databases'),
      api.dbManagerRequest('run_sql'),
      api.dbManagerRequest('update_structure'),
    ])
    expect(count).toBe(1)

    stop()
  })

  it('reports again for a different token, so a new session is not silenced', async () => {
    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    api.setDbManagerSession({ id: 1, user: 'admin', token: 'token-a' })
    fetchMock.mockImplementation(async () => replyAuthFailure())
    await api.dbManagerRequest('get_databases')

    api.setDbManagerSession({ id: 1, user: 'admin', token: 'token-b' })
    await api.dbManagerRequest('get_databases')

    expect(count).toBe(2)
    stop()
  })

  it('does not fire for an ordinary action failure', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    fetchMock.mockImplementation(async () =>
      replyJson({ success: false, message: 'No such database' }),
    )

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    const result = await api.dbManagerRequest('delete_database', { id: 99 })
    expect(result.success).toBe(false)
    expect(count).toBe(0)
    expect(api.getDbManagerToken()).toBe('good-token')

    stop()
  })

  it('does not fire when the request was anonymous', async () => {
    // A login attempt that hit the same message must not sign out another tab.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    await api.dbManagerRequest('login', { user: 'a', pass: 'b' }, { anonymous: true })
    expect(count).toBe(0)
    expect(api.getDbManagerToken()).toBe('good-token')

    stop()
  })

  it('does not sign the user out because the server was unreachable', async () => {
    // An unreachable server says nothing about whether the token is still good, and
    // logging out on a flaky connection would be worse than showing the error.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    fetchMock.mockRejectedValue(new Error('Failed to fetch'))

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    await expect(api.dbManagerRequest('get_databases')).rejects.toThrow('Failed to fetch')
    expect(count).toBe(0)
    expect(api.getDbManagerToken()).toBe('good-token')

    stop()
  })

  it('keeps working when one listener throws', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())
    vi.spyOn(console, 'error').mockImplementation(() => {})

    const seen = []
    const stopBad = api.onDbManagerSessionLost(() => {
      throw new Error('subscriber bug')
    })
    const stopGood = api.onDbManagerSessionLost(() => seen.push('lost'))

    await api.dbManagerRequest('get_databases')
    expect(seen).toEqual(['lost'])

    stopBad()
    stopGood()
  })

  it('stops notifying after unsubscribe', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)
    stop()

    await api.dbManagerRequest('get_databases')
    expect(count).toBe(0)
  })
})

describe('the streaming request path', () => {
  it('hands back an unread body so a download still works', async () => {
    // backup_databases serves .sql/.zip; parsing it as JSON here would destroy it.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    fetchMock.mockImplementation(
      async () =>
        new Response('-- mysqldump', {
          status: 200,
          headers: {
            'Content-Type': 'application/sql',
            'Content-Disposition': 'attachment; filename="backup.sql"',
          },
        }),
    )

    const response = await api.dbManagerRequestRaw('backup_databases', { database_ids: [1] })
    expect(await response.text()).toBe('-- mysqldump')
  })

  it('attaches the token', async () => {
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    await api.dbManagerRequestRaw('backup_databases')
    expect(lastBody()).toEqual({ action: 'backup_databases', token: 'good-token' })
  })

  it('still notices a JSON rejection', async () => {
    // The gate answers JSON even on an endpoint that normally streams a file, so
    // without this a revoked session here would just fail a download silently.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'dead-token' })
    fetchMock.mockImplementation(async () => replyAuthFailure())

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    await api.dbManagerRequestRaw('backup_databases')
    expect(count).toBe(1)
    expect(api.getDbManagerToken()).toBe(null)

    stop()
  })

  it('leaves the caller to report unparseable JSON', async () => {
    // Truncated or HTML error pages are not the auth envelope.
    api.setDbManagerSession({ id: 1, user: 'admin', token: 'good-token' })
    fetchMock.mockImplementation(
      async () =>
        new Response('<html>500</html>', {
          status: 500,
          headers: { 'Content-Type': 'application/json' },
        }),
    )

    let count = 0
    const stop = api.onDbManagerSessionLost(() => count++)

    const response = await api.dbManagerRequestRaw('backup_databases')
    expect(response.status).toBe(500)
    expect(count).toBe(0)
    expect(api.getDbManagerToken()).toBe('good-token')

    stop()
  })
})
