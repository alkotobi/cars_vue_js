import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'

// A dead api token used to reach the user as a raw string.
//
// api/lib/auth.php rejects an unverifiable token with
// { success: false, code: 'not_authenticated' } and **HTTP 200**, deliberately,
// because useApi retries any 403/429 three times with backoff. So there was no
// status, no thrown error and no central place for the app to notice: ~200 call
// sites each decided for themselves, and the choices ranged from a native
// alert() to a line of red text to nothing at all. Whichever one a given screen
// happened to use, the user was told a session had gone and given no way to act on
// it.
//
// These tests pin the three things that actually change the outcome: the code is
// recognised in one place, callApi reports it (so no screen has to), and the modal
// that leads to /login is wired into the shell. The parity assertions against the
// PHP source are the point of half of them - a rename on either side of this
// boundary fails here instead of silently disabling the whole feature.

const AUTH_PHP = readFileSync(new URL('../../api/lib/auth.php', import.meta.url), 'utf8')
const API_PHP = readFileSync(new URL('../../api/api.php', import.meta.url), 'utf8')
const SESSION_LOST_SOURCE = readFileSync(new URL('./useSessionLost.js', import.meta.url), 'utf8')
const USE_API_SOURCE = readFileSync(new URL('./useApi.js', import.meta.url), 'utf8')
const APP_VUE = readFileSync(new URL('../App.vue', import.meta.url), 'utf8')
const MODAL = readFileSync(
  new URL('../components/SessionExpiredModal.vue', import.meta.url),
  'utf8',
)

const read = (rel) => readFileSync(new URL(rel, import.meta.url), 'utf8')
const readLocale = (locale) => JSON.parse(read(`../locales/${locale}.json`))

const TOKEN = 'a'.repeat(64)
const OTHER_TOKEN = 'b'.repeat(64)

function stubBrowser() {
  const store = new Map()
  vi.stubGlobal('window', {
    location: {
      href: 'https://world-automobile.com/cars/',
      origin: 'https://world-automobile.com',
      protocol: 'https:',
      host: 'world-automobile.com',
      hostname: 'world-automobile.com',
      pathname: '/cars/',
    },
    dispatchEvent: vi.fn(),
  })
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
    clear: () => store.clear(),
  })
  return store
}

const json = (body) =>
  new Response(JSON.stringify(body), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  })

// The API envelope apiErrorDie() writes for a rejected token.
const rejected = () => ({ success: false, code: 'not_authenticated', error: 'not_authenticated' })

function storeUser(store, token = TOKEN, extra = {}) {
  store.set('user', JSON.stringify({ id: 1, username: 'merhab', token, ...extra }))
}

describe('the auth gate and the client agree on what "not authenticated" means', () => {
  it('uses the code api_auth_fail() actually emits', async () => {
    const { AUTH_FAILURE_CODE } = await import('./useSessionLost')

    // Pin the literal against the server. Renaming the code in auth.php without
    // renaming it here would make every dead session silent again.
    expect(AUTH_FAILURE_CODE).toBe('not_authenticated')
    expect(AUTH_PHP).toContain(`api_auth_fail('${AUTH_FAILURE_CODE}')`)
  })

  it('is still answered with HTTP 200, so detection has to key on the code', () => {
    // Not an assertion that 200 is good - it is the reason callApi cannot branch on
    // response.ok, and the reason it must not reuse the 403/429 retry above. The
    // shared writer defaults the status to 200 on purpose: a real 4xx became a
    // 7-second retry storm and then an opaque "HTTP 401" string.
    expect(API_PHP).toMatch(/function apiErrorDie\(\$code, \$status = 200/)
    // And the standalone guard auth.php uses outside api.php sets no status at all.
    expect(AUTH_PHP).not.toMatch(/http_response_code/)
  })

  it('exempts exactly the actions the server never gates', async () => {
    const { SESSION_LOSS_EXEMPT_ACTIONS } = await import('./useSessionLost')

    // PUBLIC_ACTIONS is the server's own list. Anything gated can legitimately
    // answer with this code, so anything public must never be read as a lost
    // session - least of all `login`, where telling someone their session just
    // ended while they are signing in would be nonsense.
    const block = API_PHP.match(/const PUBLIC_ACTIONS = \[([\s\S]*?)\];/)
    expect(block).not.toBeNull()
    const serverActions = [...block[1].matchAll(/'([a-z_]+)'/g)].map((m) => m[1])

    expect(SESSION_LOSS_EXEMPT_ACTIONS).toEqual(serverActions)
  })
})

describe('isSessionLost', () => {
  it('recognises the rejection envelope', async () => {
    const { isSessionLost } = await import('./useSessionLost')

    expect(isSessionLost(rejected())).toBe(true)
    expect(isSessionLost(rejected(), 'get_cars')).toBe(true)
  })

  it('leaves every other failure alone', async () => {
    const { isSessionLost } = await import('./useSessionLost')

    // An action that fails says something about the action. Interrupting a user
    // with a sign-in modal because a colour already exists is worse than the bug
    // this module fixes.
    expect(isSessionLost({ success: false, code: 'color_exists' })).toBe(false)
    expect(isSessionLost({ success: false, code: 'not_admin' })).toBe(false)
    expect(isSessionLost({ success: false, code: 'db_unavailable' })).toBe(false)
    expect(isSessionLost({ success: true, code: 'not_authenticated' })).toBe(false)
    expect(isSessionLost(null)).toBe(false)
    expect(isSessionLost(undefined)).toBe(false)
    expect(isSessionLost({ success: false })).toBe(false)
  })

  it('never reads a public action as a lost session', async () => {
    const { isSessionLost, SESSION_LOSS_EXEMPT_ACTIONS } = await import('./useSessionLost')

    for (const action of SESSION_LOSS_EXEMPT_ACTIONS) {
      expect(isSessionLost(rejected(), action)).toBe(false)
    }
  })
})

describe('reportSessionLost', () => {
  let store
  let mod

  beforeEach(async () => {
    vi.resetModules()
    store = stubBrowser()
    mod = await import('./useSessionLost')
    mod.armSessionLost()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    vi.resetModules()
  })

  it('tells its subscribers and drops the dead credential', async () => {
    storeUser(store)
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    mod.reportSessionLost(TOKEN)

    expect(seen).toEqual(['lost'])
    // Cleared, not just flagged: router/index.js and AppHeader both read this key,
    // so leaving a token the server refuses would keep rendering a signed-in shell.
    expect(store.has('user')).toBe(false)
    off()
  })

  it('announces it as a logout so existing listeners react', async () => {
    storeUser(store)

    mod.reportSessionLost(TOKEN)

    // AppHeader already reloads its user on 'userLogout'. A second event type for
    // the same fact would mean teaching every such listener twice.
    expect(window.dispatchEvent).toHaveBeenCalledTimes(1)
    expect(window.dispatchEvent.mock.calls[0][0].type).toBe('userLogout')
  })

  it('keeps the reader language and other preferences', async () => {
    storeUser(store)
    store.set('app_language', 'ar')
    store.set('printTableFontSize', '14')

    mod.reportSessionLost(TOKEN)

    // A token the server rejected is not a sign-out the user asked for. A blanket
    // localStorage.clear() would reset their language for it.
    expect(store.get('app_language')).toBe('ar')
    expect(store.get('printTableFontSize')).toBe('14')
  })

  it('reports once for the burst of requests that die together', async () => {
    storeUser(store)
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    // Revoking a token kills every request already on the wire, and each comes
    // back with the same envelope. Without the dedupe the user gets one modal per
    // outstanding call.
    mod.reportSessionLost(TOKEN)
    mod.reportSessionLost(TOKEN)
    mod.reportSessionLost(TOKEN)

    expect(seen).toEqual(['lost'])
    off()
  })

  it('stays quiet when a usable session is still stored', async () => {
    storeUser(store)
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    // No token went out, yet a valid one is stored: the caller deliberately went
    // anonymous (the share link, a pre-login probe). Reporting here would evict a
    // live session over a request that was never going to use it.
    mod.reportSessionLost(null)

    expect(seen).toEqual([])
    expect(store.has('user')).toBe(true)
    off()
  })

  it('reports a gated request that carried no credential at all', async () => {
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    // The case the native alert("User not authenticated") calls covered, and the
    // reason the router guard was the only other net for it. The server rejects a
    // gated action sent with no token exactly as it rejects an unknown one, so this
    // is reported rather than left to the next navigation.
    mod.reportSessionLost(null)

    expect(seen).toEqual(['lost'])
    expect(store.has('user')).toBe(false)
    off()
  })

  it('reports the second sign-out, not just the first', async () => {
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    // Both are keyed by the no-token sentinel, so re-arming on login is the only
    // thing that makes a repeat reportable. armSessionLost() is what LoginView calls
    // once a token has actually been issued.
    mod.reportSessionLost(null)
    storeUser(store)
    mod.armSessionLost()
    store.delete('user')
    mod.reportSessionLost(null)

    expect(seen).toEqual(['lost', 'lost'])
    off()
  })

  it('does not let a burst of anonymous calls stack modals', async () => {
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    mod.reportSessionLost(null)
    mod.reportSessionLost(null)
    mod.reportSessionLost(undefined)

    expect(seen).toEqual(['lost'])
    off()
  })

  it('ignores a response that a newer login has already made obsolete', async () => {
    store.set('user', JSON.stringify({ id: 1, token: OTHER_TOKEN }))
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    mod.reportSessionLost(TOKEN)

    // The user re-logged-in while the old request was still in flight, so the token
    // that failed is no longer the one in localStorage. Evicting the fresh session
    // and interrupting them mid-work over an already-stale answer is the worse bug.
    expect(seen).toEqual([])
    expect(JSON.parse(store.get('user')).token).toBe(OTHER_TOKEN)
    off()
  })

  it('reports again once a new session is issued', async () => {
    storeUser(store)
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    mod.reportSessionLost(TOKEN)
    storeUser(store, OTHER_TOKEN)
    mod.reportSessionLost(OTHER_TOKEN)

    // Deduping by token value, not by "has reported at all", so signing in again
    // restores the ability to detect the next expiry.
    expect(seen).toEqual(['lost', 'lost'])
    off()
  })

  it('does not let one broken subscriber silence the others', async () => {
    storeUser(store)
    const seen = []
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {})
    const offBad = mod.onSessionLost(() => {
      throw new Error('subscriber exploded')
    })
    const offGood = mod.onSessionLost(() => seen.push('lost'))

    expect(() => mod.reportSessionLost(TOKEN)).not.toThrow()
    expect(seen).toEqual(['lost'])
    offBad()
    offGood()
    spy.mockRestore()
  })

  it('unsubscribes', async () => {
    storeUser(store)
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))
    off()

    mod.reportSessionLost(TOKEN)

    expect(seen).toEqual([])
  })

  it('sessionLostError reports and carries the code', async () => {
    const seen = []
    const off = mod.onSessionLost(() => seen.push('lost'))

    const err = mod.sessionLostError()

    // Both halves matter: the report opens the modal, and `code` is what lets
    // apiErrorText()/colorErrorText() render it translated instead of leaking the
    // raw string these ten guards used to throw.
    expect(seen).toEqual(['lost'])
    expect(err.code).toBe('not_authenticated')
    off()
  })
})

describe('callApi reports a dead session without changing its envelope', () => {
  let store

  beforeEach(() => {
    store = stubBrowser()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
    vi.restoreAllMocks()
    vi.resetModules()
  })

  // useApi.js loads per-server config from db_code.json before any request, so a
  // bare mock leaves callApi rejecting on config and never reaches the API.
  function stubFetchApiResult(apiResult) {
    const fetchMock = vi.fn(async (url) => {
      if (String(url).includes('db_code.json')) {
        return json({ db_name: 'merhab', uplod_path: 'mig_files' })
      }
      return json(apiResult)
    })
    vi.stubGlobal('fetch', fetchMock)
    return fetchMock
  }

  it('opens the modal path when the gate rejects the token', async () => {
    storeUser(store)
    stubFetchApiResult(rejected())
    const lost = await import('./useSessionLost')
    lost.armSessionLost()
    const seen = []
    const off = lost.onSessionLost(() => seen.push('lost'))
    const { useApi } = await import('./useApi')
    const api = useApi()

    // The caller still gets the payload it always got, so any screen that wants to
    // render its own wording still can - the report is additive.
    await expect(api.callApi({ action: 'get_cars', token: TOKEN })).resolves.toEqual(rejected())

    expect(seen).toEqual(['lost'])
    off()
  })

  it('clears the stored user, so the router lets /login through', async () => {
    storeUser(store)
    stubFetchApiResult(rejected())
    const lost = await import('./useSessionLost')
    lost.armSessionLost()
    const off = lost.onSessionLost(() => {})
    const { useApi } = await import('./useApi')

    await useApi().callApi({ action: 'get_cars' })

    // The beforeEach guard redirects to /login when there is no `user` key, and
    // redirects *away* from /login when there is one. Leaving the dead token in
    // place would bounce the user straight back to the page that just failed.
    expect(store.has('user')).toBe(false)
    off()
  })

  it('does not fire on an ordinary action failure', async () => {
    stubFetchApiResult({ success: false, code: 'color_exists', meta: { field: 'hexa' } })
    const lost = await import('./useSessionLost')
    lost.armSessionLost()
    const seen = []
    const off = lost.onSessionLost(() => seen.push('lost'))
    const { useApi } = await import('./useApi')

    await useApi().callApi({ action: 'create_color', token: TOKEN })

    expect(seen).toEqual([])
    off()
  })

  it('does not fire for the login action itself', async () => {
    stubFetchApiResult(rejected())
    const lost = await import('./useSessionLost')
    lost.armSessionLost()
    const seen = []
    const off = lost.onSessionLost(() => seen.push('lost'))
    const { useApi } = await import('./useApi')

    await useApi().callApi({ action: 'login', username: 'x', password: 'y' })

    expect(seen).toEqual([])
    off()
  })

  it('reports a signed-out tab, because the server rejects it the same way', async () => {
    stubFetchApiResult(rejected())
    const lost = await import('./useSessionLost')
    lost.armSessionLost()
    const seen = []
    const off = lost.onSessionLost(() => seen.push('lost'))
    const { useApi } = await import('./useApi')

    await useApi().callApi({ action: 'get_cars' })

    // No token was sent. The server rejects a gated action with no credential using
    // the same code as one with a bad credential, so this is a lost session and the
    // modal opens - rather than waiting for the router guard to notice on the next
    // navigation. This is the case the native alert() calls used to cover.
    expect(seen).toEqual(['lost'])
    off()
  })
})

describe('the modal is reachable from the shell and leads to login', () => {
  it('is mounted once, in App.vue', () => {
    // Mounted per-view is the pattern that made MessageBox invisible for the ~198
    // views that do not render it: an error with no owner shows up as nothing.
    expect(APP_VUE).toContain(
      "import SessionExpiredModal from '@/components/SessionExpiredModal.vue'",
    )
    expect(APP_VUE).toMatch(/<SessionExpiredModal\s*\/>/)
    expect(APP_VUE.match(/<SessionExpiredModal/g)).toHaveLength(1)
  })

  it('subscribes to the one detector rather than watching calls itself', () => {
    expect(MODAL).toContain("from '../composables/useSessionLost'")
    expect(MODAL).toContain('onSessionLost(')
  })

  it('sends the user to the login route when they act', () => {
    expect(MODAL).toContain("router.replace('/login')")
    // replace, not push: the screen behind is unreachable now, and Back into it
    // only fails again.
    expect(MODAL).not.toContain("router.push('/login')")
  })

  it('cannot be dismissed, because a dead token fails everything', () => {
    // Escape, backdrop click and a close button would each leave the user on a page
    // whose every read and every save fails the same way.
    expect(MODAL).not.toMatch(/@click(?:\.self)?="?handleClose/)
    expect(MODAL).not.toMatch(/event\.key === 'Escape'/)
    expect(MODAL).not.toMatch(/close-btn|@click="close"/)
  })

  it('stays off the routes that have no session to lose', () => {
    // /login is where this modal is going, so re-arming there is a trap; the share
    // link runs on its own credential.
    expect(MODAL).toContain("route.path === '/login'")
    expect(MODAL).toContain("route.name === 'client-details'")
  })

  it('exports no dismiss/cancel emit that a caller could wire up', () => {
    expect(MODAL).not.toMatch(/defineEmits/)
  })
})

describe('the modal tells the user what to do, in every locale', () => {
  const LOCALES = ['en', 'ar', 'fr', 'zh']

  it.each(LOCALES)('%s has every key the modal renders', (locale) => {
    const messages = readLocale(locale).sessionExpired
    expect(messages).toBeDefined()
    // The whole point of the change: a bare error code told the reader nothing
    // about what happened or what to do next.
    for (const key of ['title', 'message', 'hint', 'action']) {
      expect(typeof messages[key]).toBe('string')
      expect(messages[key].length).toBeGreaterThan(0)
    }
    // `action` is the button that leads to login, so it has to read as an action,
    // not as an acknowledgement.
    expect(messages.action).not.toMatch(/^(ok|got it|fermé)$/i)
  })

  it('names the next step, not just the failure', () => {
    const messages = readLocale('en').sessionExpired
    expect(messages.message.toLowerCase()).toMatch(/sign in/)
    expect(messages.hint.toLowerCase()).toMatch(/sign in/)
  })
})

describe('the client-side guards no longer dead-end on a raw string', () => {
  const CAR_STOCK = read('../components/car-stock/CarStockTable.vue')
  const LOGIN_VIEW = read('../views/LoginView.vue')

  it('no helper throws a bare "User not authenticated"', () => {
    // Ten helpers refused to attempt their call when getCurrentUser() was empty and
    // threw an untranslated string. The reader got no explanation and no way to act.
    expect(USE_API_SOURCE).not.toMatch(/throw new Error\('User not authenticated'\)/)
    expect(USE_API_SOURCE.match(/throw sessionLostError\(\)/g)).toHaveLength(10)
    expect(USE_API_SOURCE).toContain("from './useSessionLost'")
  })

  it('the car stock screen opens the modal instead of a native alert', () => {
    // Three guards in the largest view in the app, each of which fired a blocking
    // browser dialog telling the reader a fact they could do nothing about.
    expect(CAR_STOCK).not.toMatch(/alert\(t\('carStock\.user_(not_)?authenticated/)
    // Counted from code lines only - two of the guards explain themselves in a
    // comment that names the call.
    const code = CAR_STOCK.split('\n').filter((line) => !line.trim().startsWith('//'))
    expect(code.join('\n').match(/reportSessionLost\(null\)/g)).toHaveLength(3)
    expect(CAR_STOCK).toContain("from '../../composables/useSessionLost'")
  })

  it('a completed login re-arms detection', () => {
    // Without this the second sign-out of a reader who never held a token reads as a
    // duplicate of the first: both are keyed by the no-token sentinel.
    expect(LOGIN_VIEW).toContain('armSessionLost()')
    expect(LOGIN_VIEW).toMatch(/localStorage\.setItem\('user'[\s\S]{0,400}armSessionLost\(\)/)
  })
})

describe('isSessionLostError and sessionErrorMessage', () => {
  it('recognises both routes to the auth gate', async () => {
    const { isSessionLostError, sessionLostError, AUTH_FAILURE_CODE } = await import(
      './useSessionLost'
    )

    // sessionLostError() is the ten client-side guards. apiFailure() - built by the
    // helpers in useApi.js from a { code } envelope - is a rejected token. Both end
    // up in the same catch blocks, so both have to be recognised.
    expect(isSessionLostError(sessionLostError())).toBe(true)
    const fromEnvelope = new Error('not_authenticated')
    fromEnvelope.code = AUTH_FAILURE_CODE
    expect(isSessionLostError(fromEnvelope)).toBe(true)
  })

  it('leaves an ordinary failure alone', async () => {
    const { isSessionLostError } = await import('./useSessionLost')

    expect(isSessionLostError(new Error('File already exists'))).toBe(false)
    expect(isSessionLostError({ code: 'color_exists' })).toBe(false)
    expect(isSessionLostError(null)).toBe(false)
    expect(isSessionLostError(undefined)).toBe(false)
    // No code at all: helpers that throw a bare `new Error(result.error)` drop it,
    // and those keep showing their own message rather than silently going quiet.
    expect(isSessionLostError(new Error('not_authenticated'))).toBe(false)
  })

  it('blanks the message the modal already owns, and only that one', async () => {
    const { sessionErrorMessage, sessionLostError } = await import('./useSessionLost')

    // '' rather than a translated string: the modal is non-dismissable and sits above
    // everything, so anything these blocks set is invisible anyway. Both call sites
    // render with v-if="error", so '' hides it with no branching at the call site.
    expect(sessionErrorMessage(sessionLostError(), 'Failed to check in file')).toBe('')

    expect(sessionErrorMessage(new Error('Network error'), 'Failed to check in file')).toBe(
      'Network error',
    )
    expect(sessionErrorMessage({}, 'Failed to check in file')).toBe('Failed to check in file')
    expect(sessionErrorMessage(null, 'Failed to check in file')).toBe('Failed to check in file')
  })
})

describe('no caller renders a second message for a session the modal owns', () => {
  const CAR_FILES = read('../components/car-stock/CarFilesManagement.vue')
  const TRACKING = read('../components/car-stock/PhysicalCopyTracking.vue')
  const CAR_STOCK = read('../components/car-stock/CarStockTable.vue')

  it('the file screens route every caught message through the helper', () => {
    // These two hold all ten of the client-side guards. They used to assign
    // err.message directly, which put the untranslated string 'User not authenticated'
    // on screen and described a failure the modal had already explained.
    for (const source of [CAR_FILES, TRACKING]) {
      expect(source).not.toMatch(/error\.value = err\.message/)
      expect(source).toContain("from '../../composables/useSessionLost'")
    }
    expect(CAR_FILES.match(/sessionErrorMessage\(err,/g).length).toBeGreaterThanOrEqual(15)
    expect(TRACKING.match(/sessionErrorMessage\(err,/g).length).toBeGreaterThanOrEqual(6)
  })

  it('the batch transfer stops instead of firing one doomed request per file', () => {
    // Every remaining file fails identically once the token is gone, so the loop
    // used to send one request per file and collect one raw string per failure.
    expect(CAR_STOCK).toMatch(
      /catch \(err\) \{\n(?:.*\n)*?\s*if \(isSessionLostError\(err\)\) throw err/,
    )
  })

  it('the batch transfer alerts nothing when the session is gone', () => {
    // A native dialog renders in the browser chrome, above a teleported modal, so
    // alert() here would stack a second popup on the thing meant to replace it.
    expect(CAR_STOCK).toMatch(
      /catch \(err\) \{\n\s*\/\/ Re-thrown[\s\S]*?if \(isSessionLostError\(err\)\) return\n\s*alert\(/,
    )
    expect(CAR_STOCK).toContain("from '../../composables/useSessionLost'")
  })
})

// A window that actually has an alert(), which the plain stubBrowser() above
// deliberately does not: these cases are about the gate, so they need the global to
// exist. Returns the localStorage backing map so a case can adjust the stored
// session - reportSessionLost() deliberately ignores a report whose token is not the
// one that failed, so a case that wants a report has to be holding that token.
function stubAlertBrowser() {
  const store = new Map()
  const alertMock = vi.fn()
  vi.stubGlobal('window', {
    location: {
      href: 'https://world-automobile.com/cars/',
      origin: 'https://world-automobile.com',
    },
    dispatchEvent: vi.fn(),
    alert: alertMock,
  })
  vi.stubGlobal('localStorage', {
    getItem: (k) => (store.has(k) ? store.get(k) : null),
    setItem: (k, v) => store.set(k, String(v)),
    removeItem: (k) => store.delete(k),
    clear: () => store.clear(),
  })
  store.set('user', JSON.stringify({ id: 1, username: 'merhab', token: TOKEN }))
  return { store, alertMock }
}

describe('the alert gate', () => {
  // ~63 call sites render a caught failure through the global alert(), and a native
  // dialog is drawn by the browser in its own chrome - above page content, so no
  // z-index can hide it. Left alone, a dead session produced a browser popup stacked
  // on the modal meant to replace it, usually blaming the action rather than the
  // credential. These tests pin the gate that makes the modal the only dialog.

  let store
  let nativeAlert

  beforeEach(async () => {
    ;({ store, alertMock: nativeAlert } = stubAlertBrowser())
    // Both pieces of module state have to be reset between cases: the gate, or an
    // armed case silences the next one's alert, and the dedupe record, or a token
    // already reported in an earlier case reads as a duplicate and never reports at
    // all. armSessionLost() is what LoginView calls for exactly this reason.
    const { releaseSessionAlertGate, armSessionLost } = await import('./useSessionLost')
    releaseSessionAlertGate()
    armSessionLost()
  })

  afterEach(async () => {
    const { releaseSessionAlertGate } = await import('./useSessionLost')
    releaseSessionAlertGate()
    vi.unstubAllGlobals()
  })

  // Stand in for SessionExpiredModal: it arms from inside its session-lost
  // listener, which is the wiring under test.
  function armViaListener() {
    return import('./useSessionLost').then(({ onSessionLost, armSessionAlertGate }) => {
      const stop = onSessionLost(() => armSessionAlertGate())
      return () => stop()
    })
  }

  it('swallows the alert of the very request that failed', async () => {
    const { sessionLostError } = await import('./useSessionLost')
    const stop = await armViaListener()
    const debug = vi.spyOn(console, 'debug').mockImplementation(() => {})
    // The client guards only fire when there is no local user at all, so that is the
    // state to reproduce. (reportSessionLost(null) against a stored token returns
    // early on purpose - see the deliberate-anonymous-request branch.)
    store.delete('user')

    // The shape of a real failure: the request comes back dead, the caller's guard
    // throws, and the caller's catch does what 63 of them do.
    const caller = async () => {
      try {
        await Promise.resolve()
        throw sessionLostError()
      } catch (err) {
        window.alert('error updating stock: ' + err.message)
      }
    }

    await caller()

    // Asserted on the mock behind the global, not on window.alert: by now the global
    // is the gate's no-op, and that no-op is deliberately a plain function rather
    // than a spy.
    expect(nativeAlert).not.toHaveBeenCalled()
    // Not silent, though: a swallowed alert is otherwise indistinguishable from a
    // code path that never ran.
    expect(debug).toHaveBeenCalled()
    stop()
    debug.mockRestore()
  })

  it('is armed before that caller resumes, not a tick later', async () => {
    const { reportSessionLost, isSessionAlertGateArmed } = await import('./useSessionLost')
    const stop = await armViaListener()

    let armedWhenCallerResumed = null
    await Promise.resolve()
      .then(() => reportSessionLost(TOKEN))
      .then(() => {
        // No await between the report and this line. This is the property that
        // makes the gate work at all: alert() is synchronous and blocking, so a gate
        // armed on a later microtask - or keyed off the modal being in the DOM -
        // would miss the alert fired by this continuation.
        armedWhenCallerResumed = isSessionAlertGateArmed()
      })

    expect(armedWhenCallerResumed).toBe(true)
    stop()
  })

  it('gives the real alert back on release', async () => {
    const { reportSessionLost, releaseSessionAlertGate, isSessionAlertGateArmed } = await import(
      './useSessionLost'
    )
    const stop = await armViaListener()

    reportSessionLost(TOKEN)
    expect(isSessionAlertGateArmed()).toBe(true)
    expect(window.alert).not.toBe(nativeAlert)

    releaseSessionAlertGate()
    expect(isSessionAlertGateArmed()).toBe(false)
    expect(window.alert).toBe(nativeAlert)
    window.alert('back')
    expect(window.alert).toHaveBeenCalledWith('back')
    stop()
  })

  it('leaves confirm() alone', async () => {
    const { reportSessionLost } = await import('./useSessionLost')
    const confirmSpy = vi.fn()
    window.confirm = confirmSpy
    const stop = await armViaListener()

    reportSessionLost(TOKEN)

    // ~87 sites use confirm() to ask the user something. Those are user intent, not
    // an error channel - silencing them would break real workflows, not tidy
    // anything up.
    expect(window.confirm).toBe(confirmSpy)
    expect(window.confirm('Delete this file?')).toBe(undefined)
    expect(confirmSpy).toHaveBeenCalledWith('Delete this file?')
    stop()
  })

  it('cannot be armed twice into a lost original', async () => {
    const { reportSessionLost, armSessionAlertGate } = await import('./useSessionLost')
    const stop = await armViaListener()

    reportSessionLost(TOKEN)
    armSessionAlertGate()
    armSessionAlertGate()

    // A second arm must not overwrite the saved original with the no-op, or the
    // release below would "restore" the gate and leave alerts dead for the session.
    const { releaseSessionAlertGate } = await import('./useSessionLost')
    releaseSessionAlertGate()
    expect(window.alert).toBe(nativeAlert)
    stop()
  })

  it('does not clobber a patch installed after it', async () => {
    const { reportSessionLost, releaseSessionAlertGate } = await import('./useSessionLost')
    const stop = await armViaListener()

    reportSessionLost(TOKEN)
    const thirdParty = vi.fn()
    window.alert = thirdParty

    releaseSessionAlertGate()

    // Something replaced window.alert while the gate was armed. That install is
    // newer than ours and outranks it; putting the native one back would undo it.
    expect(window.alert).toBe(thirdParty)
    stop()
  })

  it('survives a host with no window, or no alert', async () => {
    const { armSessionAlertGate, releaseSessionAlertGate, isSessionAlertGateArmed } = await import(
      './useSessionLost'
    )

    vi.stubGlobal('window', {})
    expect(() => armSessionAlertGate()).not.toThrow()
    expect(isSessionAlertGateArmed()).toBe(false)

    vi.stubGlobal('window', undefined)
    expect(() => armSessionAlertGate()).not.toThrow()
    expect(() => releaseSessionAlertGate()).not.toThrow()
    expect(isSessionAlertGateArmed()).toBe(false)
  })

  it('leaves no restore path to the modal component', () => {
    // The module cannot know when the reader is done with the dialog, so it does not
    // guess with a timeout - a modal left open for five minutes would silently start
    // letting alerts through again. The component owns all three exits instead, and
    // all three have to exist or a stuck modal silences alerts app-wide.
    expect(MODAL).toContain('armSessionAlertGate()')
    expect(MODAL).toMatch(/watch\(visible[\s\S]*?releaseSessionAlertGate\(\)/)
    expect(MODAL).toMatch(/onUnmounted\([\s\S]*?releaseSessionAlertGate\(\)/)
    // Armed from the listener, not from reportSessionLost: this component knows the
    // route, so it can decline to arm on a screen the modal does not apply to.
    expect(MODAL).toMatch(
      /onSessionLost\([\s\S]*?if \(!exemptRoute\.value\) armSessionAlertGate\(\)/,
    )
  })
})

describe('apiErrorText names the real cause of a rejected token', () => {
  beforeEach(() => {
    // useApi reads window while it initialises, so the import in the first case
    // below needs one in place before it runs.
    stubAlertBrowser()
  })

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('maps not_authenticated instead of letting the caller blame the action', async () => {
    const { apiErrorText } = await import('./useApi')

    // The call sites are all `alert(apiErrorText(t, result) || t('someActionError'))`.
    // With no mapping, apiErrorText returns '' and the reader is told "error updating
    // stock" for what is an expired credential.
    const t = (key) => key
    expect(apiErrorText(t, rejected())).toBe('sessionExpired.message')

    // Unknown codes still return '' so the caller's own fallback is unchanged.
    expect(apiErrorText(t, { success: false, code: 'db_unavailable' })).toBe('')
    expect(apiErrorText(t, { success: false })).toBe('')
  })

  it('reuses the modal wording rather than inventing a near-duplicate', () => {
    // One event, one description. A second string for the same thing is how the two
    // dialogs drift apart in the first place.
    for (const locale of ['en', 'ar', 'fr', 'zh']) {
      const messages = readLocale(locale)
      expect(messages.sessionExpired.message).toBeTruthy()
      expect(messages.colorsView.errors.notAuthenticated).toBeTruthy()
    }
  })
})

describe('the wiring is where the comments claim it is', () => {
  it('uses the detector rather than reimplementing the code check', () => {
    expect(USE_API_SOURCE).toContain("from './useSessionLost'")
    // colorErrorText() maps the same code to a translated string, and that stays -
    // what must not come back is a second place that *decides* the session is over.
    expect(USE_API_SOURCE).not.toMatch(/[=!]==?\s*'not_authenticated'/)
    expect(USE_API_SOURCE).not.toMatch(/'not_authenticated'\s*===?/)
  })

  it('both auth-checked transports report, so uploads fail the same way', () => {
    // upload.php authenticates through the same require_api_user() guard. Without
    // the hook there, attaching a document to an expired session failed as a bare
    // "Upload failed".
    const reports = USE_API_SOURCE.match(/reportSessionLost\(/g) || []
    expect(reports.length).toBe(2)
    expect(USE_API_SOURCE).toContain('isSessionLost(result)')
    expect(USE_API_SOURCE).toContain('isSessionLost(result, data.action)')
  })

  it('does not retry a rejected token', () => {
    // The 403/429 backoff above is for rate limits and bot checks. Re-sending a
    // token the server has already refused only delays the same answer.
    const retry = USE_API_SOURCE.indexOf('retryCount < 3')
    const report = USE_API_SOURCE.indexOf('reportSessionLost(')
    expect(report).toBeGreaterThan(retry)
  })

  it('the detector module has no Vue import, so it is testable on its own', () => {
    expect(SESSION_LOST_SOURCE).not.toMatch(/from 'vue'/)
  })
})
