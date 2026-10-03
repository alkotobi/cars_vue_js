import { readJsonResponse } from '@/utils/readJsonResponse'
import { getBasePath, resolveApiBaseUrl } from '@/utils/basePath'

// The one way to talk to db_manager_api.php.
//
// Every call used to be a hand-rolled fetch() with the token attached by hand, and
// that had two consequences worth not repeating:
//
//   1. Forgetting the token produced a 401-shaped failure with no hint about why -
//      which is exactly how the DB manager ended up 100% unreachable behind its own
//      auth gate while 177 tests passed.
//   2. Nothing noticed when a token stopped being valid. The gate answers
//      {success: false, message: 'Not authenticated...'}, each call site rendered
//      that as an inline error, and DbManagerView decided "logged in" from the mere
//      presence of a localStorage key - so a revoked token (logout in another tab,
//      or api_token cleared by hand) left a permanently broken shell instead of the
//      login form.
//
// Both are now structural rather than per-call-site discipline: the token goes on
// every request that should carry one, and the auth-failure envelope is detected in
// one place that reports it to whoever is listening.
//
// This module is deliberately separate from useApi.js. useApi.js is the *tenant*
// app's API client and authenticates against the users table; the DB manager is a
// separate trust domain with its own realm, and folding the two together is how the
// wrong token ended up wired up in the first place.

// The gate's rejection message, from the auth block in db_manager_api.php. Matched
// by prefix so a wording tweak to the tail ("...Sign in to the database manager.")
// does not silently stop session-loss detection.
const AUTH_FAILURE_PREFIX = 'Not authenticated.'

// Resolve the API URL the same way everywhere: the API ships inside the app folder,
// so it has to include the mount point or every /folder deploy 404s. components
// that already receive an api-base-url prop can pass it in instead.
export function getDbManagerApiUrl(baseUrlOverride) {
  const base =
    baseUrlOverride ||
    resolveApiBaseUrl({
      override: import.meta.env?.VITE_API_BASE_URL,
      protocol: window.location.protocol,
      hostname: window.location.hostname,
      port: window.location.port,
      basePath: getBasePath(),
      isDev: import.meta.env.DEV,
    })

  return `${base}/db_manager_api.php`
}

// Read the DB-manager token out of localStorage.
//
// Deliberately NOT getStoredToken() from useApi.js. The database manager
// authenticates against the registry's `login` table, which guards host-level
// operations, rather than the tenant app's `users` table. LoginSignup.vue writes
// the login response's `data` object wholesale under `db_manager_user`, and that
// object is {id, user, token}.
//
// Tolerant of a corrupt value for the same reason getStoredToken() is: a partial
// write would otherwise throw during navigation and take the app down with it.
export function getDbManagerToken() {
  try {
    const raw = localStorage.getItem('db_manager_user')
    if (!raw) return null
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed.token === 'string' ? parsed.token : null
  } catch {
    return null
  }
}

// Persist a completed login. Stores the whole {id, user, token} object, matching
// what the login response puts in `data`.
export function setDbManagerSession(session) {
  localStorage.setItem('db_manager_user', JSON.stringify(session))
}

// Drop the stored DB-manager token.
//
// The server-side revoke is what actually matters - DbManagerSidebar.vue posts
// `logout` so login.api_token is nulled rather than merely hidden. This is the
// local half, and it runs even if that request failed, because a stale local
// credential must not outlive a sign-out the user asked for.
export function clearDbManagerToken() {
  try {
    localStorage.removeItem('db_manager_user')
  } catch {
    // Private-mode storage failures are not worth failing a logout over.
  }
}

// Is this result the gate rejecting us, rather than an ordinary action failure?
//
// Exported for tests. An action that legitimately fails says something about the
// action; this message means the session is gone, which is a different event with a
// different consequence, so the two must not share a code path.
export function isDbManagerSessionLost(result) {
  return (
    !!result &&
    result.success === false &&
    typeof result.message === 'string' &&
    result.message.startsWith(AUTH_FAILURE_PREFIX)
  )
}

const listeners = new Set()

// The token we last reported as dead. Several requests are usually in flight when a
// token is revoked, and every one of them comes back with the same envelope; without
// this the listeners would fire once per outstanding call. Keyed by token value so a
// genuinely new session is always reported, and a failure before login (where there
// is no token) never reports at all.
let reportedToken = null

// Subscribe to session loss. Returns an unsubscribe function.
export function onDbManagerSessionLost(listener) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

// React to a dead token: clear the stale credential, then tell the UI.
//
// Clearing before notifying matters. DbManagerView decides whether to show the login
// form or the panels from this same localStorage key, so a listener that ran first
// and cached `isLoggedIn = true` would keep rendering the panels over a token that
// no longer exists.
function reportSessionLost(token) {
  if (!token || token === reportedToken) return
  reportedToken = token
  clearDbManagerToken()

  for (const listener of listeners) {
    try {
      listener()
    } catch (err) {
      // One broken subscriber must not stop the others being told.
      console.error('db-manager session-lost listener threw:', err)
    }
  }
}

// Call a db_manager_api.php action and parse the JSON envelope.
//
// `anonymous: true` is for login and signup only: they run before there is a token
// to send, and a failure there is a credential problem to show on the login form,
// never a lost session.
export async function dbManagerRequest(action, payload = {}, options = {}) {
  const url = getDbManagerApiUrl(options.baseUrl)
  const token = options.anonymous ? null : getDbManagerToken()

  const body = { action, ...payload }
  if (token) {
    body.token = token
  }

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })

  const result = await readJsonResponse(response, url)

  if (token && isDbManagerSessionLost(result)) {
    reportSessionLost(token)
  }

  return result
}

// Same, but hands back the untouched Response for actions that stream a file rather
// than returning JSON (backup_databases serves .sql / .zip / .octet-stream).
//
// The auth check only applies when the response is JSON - i.e. when the gate refused
// us - and it reads a clone so the caller can still consume the original body. The
// clone is only taken on the JSON branch, so a multi-megabyte dump is never buffered
// twice.
//
// A network error here propagates without reporting session loss: an unreachable
// server says nothing about whether the token is still good, and signing the user
// out on a flaky connection would be worse than showing the error.
export async function dbManagerRequestRaw(action, payload = {}, options = {}) {
  const url = getDbManagerApiUrl(options.baseUrl)
  const token = options.anonymous ? null : getDbManagerToken()

  const body = { action, ...payload }
  if (token) {
    body.token = token
  }

  const response = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  })

  const contentType = response.headers.get('Content-Type') || ''
  if (token && contentType.includes('application/json')) {
    try {
      const result = await readJsonResponse(response.clone(), url)
      if (isDbManagerSessionLost(result)) {
        reportSessionLost(token)
      }
    } catch {
      // Unparseable JSON is not the auth envelope; the caller gets the Response and
      // reports the real problem.
    }
  }

  return response
}
