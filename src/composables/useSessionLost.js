// Composable/module: one place that decides "the token is dead, show the user the
// session-expired modal".
//
// Why this is needed at all: api/lib/auth.php answers a rejected token with
// `{ success: false, code: 'not_authenticated', error: 'not_authenticated' }` and
// **HTTP 200** — deliberately, because useApi retries any 403/429 three times with
// backoff. So there is no transport-level signal to hang a global handler off, and
// every one of ~200 call sites handled it on its own: a native alert(), an inline
// string, or nothing at all. The user saw "not_authenticated" or a red line of
// text in a corner with no idea they had to go and log in again.
//
// The API's own documented reason for keeping the status at 200 is "the client
// distinguishes the outcome from the `code` field either way" — which is exactly
// what this module does, in one place, instead of two hundred.
//
// This mirrors the pattern already proven for the database manager in
// useDbManagerApi.js (isDbManagerSessionLost / onDbManagerSessionLost /
// reportSessionLost), deliberately: one dead-token concept, one implementation.

// The machine-readable code api/lib/auth.php emits from api_auth_fail() and
// require_api_user(). Kept as a named constant because it is asserted against the
// PHP source in the spec - a rename on either side has to fail there.
export const AUTH_FAILURE_CODE = 'not_authenticated'

// Public actions in api/api.php never reach require_api_user(), so they cannot
// produce this code. `login` is the important one: a wrong password on the login
// form is a credential problem to show inline, never a reason to tell someone
// their session just ended while they are standing on the page that issues one.
//
// Exported for tests rather than inlined, so the list and the server's list can be
// compared instead of being trusted.
export const SESSION_LOSS_EXEMPT_ACTIONS = [
  'ping',
  'login',
  'change_password_with_credentials',
  'get_client_share_data',
  'get_db_version',
]

/**
 * Is this result the auth gate rejecting us, rather than an ordinary action failure?
 *
 * An action that legitimately fails says something about the action
 * (`color_exists`, `qty_below_committed`); this code means the credential itself is
 * gone - expired, revoked, the token rotated by a password change, the server
 * database swapped out from under a running tab. Different event, different
 * consequence, so it must not share a code path with a normal error.
 *
 * `action` is the action name of the request that came back, and is used only to
 * keep the exempt/public actions above from ever being read as a lost session.
 *
 * @param {{success?: boolean, code?: string}|null} result
 * @param {string} [action] the action name of the originating request
 * @returns {boolean}
 */
export function isSessionLost(result, action = '') {
  if (!result || result.success !== false) return false
  if (result.code !== AUTH_FAILURE_CODE) return false
  if (action && SESSION_LOSS_EXEMPT_ACTIONS.includes(action)) return false
  return true
}

const listeners = new Set()

// Dedupe key for a request that went out with no credential at all. There is no
// token to key on there, and every anonymous gated call answers the same way, so
// without a sentinel one modal would be queued per in-flight request.
const NO_TOKEN = '<no-token>'

// The token we last reported as dead. Several requests are usually in flight when
// a token dies, and every one of them comes back with the same envelope; without
// this the listeners would fire once per outstanding call and the user would stack
// N modals. Keyed by token value so a genuinely new session is always reported.
let reportedToken = null

// Subscribe to session loss. Returns an unsubscribe function.
export function onSessionLost(listener) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

/**
 * Forget the dedupe record, so the next failure is reported.
 *
 * Called when a login issues a new session. Without it, the second sign-out of a
 * session that had none would be swallowed: both are keyed NO_TOKEN, so the second
 * would read as a duplicate of the first. Re-arming on login is what makes the
 * absence of a token a reportable event more than once per signed-out period.
 */
export function armSessionLost() {
  reportedToken = null
}

// Read just the token out of the stored user, tolerating a corrupt value.
function readStoredToken() {
  try {
    const raw = localStorage.getItem('user')
    if (!raw) return null
    return JSON.parse(raw)?.token ?? null
  } catch {
    // A corrupt `user` key is indistinguishable from no session, and the modal
    // below sends the reader to login either way.
    return null
  }
}

/**
 * Drop the dead credential from localStorage.
 *
 * Only the `user` key, not localStorage.clear(). A token that no longer works is
 * not the same event as a sign-out, and wiping everything would throw away the
 * reader's chosen language (`app_language`) and the print font sizes for a session
 * they never asked to end. The server-side revoke already happened - this is only
 * the local half, and it runs even if the caller never acts on the modal.
 */
function clearStoredUser() {
  try {
    localStorage.removeItem('user')
  } catch {
    // Private-mode storage failures are not worth failing over: the modal is
    // about to send them to the login page anyway.
  }
}

/**
 * React to a dead session: clear the stale credential, then tell the UI.
 *
 * Clearing before notifying is the ordering that matters. router/index.js reads
 * `localStorage.user` to decide whether a route needs auth, and AppHeader's
 * getUser() reads the same key, so a listener that ran first and cached a user
 * object would keep rendering a header for a session that no longer exists.
 *
 * `token` is the token the failing request actually carried, or null when it
 * carried none at all - which is itself reportable. Both come back from the server
 * as the same code, because a gated action with no token is rejected exactly like a
 * gated action with a token the server does not recognise. That is the case the
 * native `alert('User not authenticated')` calls used to cover one screen at a
 * time, and the reason the router guard was the only other net for it.
 *
 * @param {string|null} token the token sent with the failing request, or null
 */
export function reportSessionLost(token) {
  const key = token || NO_TOKEN
  if (key === reportedToken) return

  const storedToken = readStoredToken()

  if (token) {
    // The stored token still being the one that failed is the guard against a stale
    // in-flight response: the reader re-logged-in while an old request was still on
    // the wire, so localStorage now holds a fresh, valid token. Evicting that - and
    // interrupting them mid-work over a response that was already obsolete is
    // worse than doing nothing.
    if (storedToken !== token) return
  } else if (storedToken) {
    // No token went out, yet a usable one is stored: the caller deliberately went
    // anonymous (the share link, a pre-login probe) and the session is fine. Only a
    // genuine absence of one is a lost session.
    return
  }

  reportedToken = key
  clearStoredUser()

  // Same event LogoutButton announces, so AppHeader and anything else already
  // listening for it does not have to learn a second one.
  try {
    window.dispatchEvent(new CustomEvent('userLogout'))
  } catch {
    // No window (tests, SSR) is not a reason to skip the listeners below.
  }

  for (const listener of listeners) {
    try {
      listener()
    } catch (err) {
      // One broken subscriber must not stop the others being told.
      console.error('session-lost listener threw:', err)
    }
  }
}

/**
 * The error a client-side "there is no local user" guard should throw.
 *
 * Ten helpers in useApi.js used to refuse to even attempt their call when
 * getCurrentUser() came back empty, throwing a bare Error whose message was the
 * untranslated string 'User not authenticated'. Reporting here is what turns those
 * into the same blocking, actionable modal every other failure gets, and attaching
 * `code` is what lets the existing apiErrorText()/colorErrorText() translators
 * render it in the reader's language instead of leaking the raw string.
 *
 * The message stays English because this module has no translator. The modal is
 * what the reader actually sees - it is non-dismissable - so this text only reaches
 * a console, or a caller that is already handling the code itself.
 *
 * @returns {Error}
 */
export function sessionLostError() {
  reportSessionLost(null)
  const err = new Error('User not authenticated')
  err.code = AUTH_FAILURE_CODE
  return err
}

// ---------------------------------------------------------------------------
// Native alert() suppression while the session-expired modal owns the screen.
//
// Why this exists: ~63 call sites render a caught API failure through the global
// alert() (see noErrorAlerts.spec.js, which keeps that number from growing). None of
// them knew about the dead session, so each produced its own message - and a native
// dialog is rendered by the browser in its own chrome, ABOVE page content. A
// teleported Vue modal cannot outrank it, however high its z-index. So the reader
// got a browser popup stacked on the dialog that was supposed to explain the
// problem, and often the popup was the more misleading of the two: the usual shape is
// alert(apiErrorText(t, result) || t('someActionError')), and with no apiErrorText
// mapping for the auth code that fallback blames the action ("error updating stock")
// for what is actually an expired credential.
//
// This is a deliberately global fix to a local problem, and the reason it is a
// contained module rather than something sprinkled through 63 files: the decision
// "the session-expired modal is the only dialog the reader needs" can be made once,
// here, instead of re-decided - and re-decided wrongly - at every call site.
//
// Two things it deliberately does NOT do:
//   - confirm() is untouched. Those ~87 sites are user-intent dialogs, not an error
//     channel; swallowing them would break real workflows.
//   - It is not keyed off whether the modal is in the DOM. alert() is synchronous and
//     blocking, so a DOM check races Vue's scheduler flush and would miss the case
//     this exists for: the alert firing in the same tick as the report, from the
//     continuation of the very request that failed.
//
// Patched lazily at arm time rather than captured at module load, because tests
// replace `window` wholesale between cases and a module-load capture would be stale.
// ---------------------------------------------------------------------------

// The original window.alert, or null while the gate is not armed. Doubles as the
// "already armed" flag.
let originalNativeAlert = null

function suppressedAlert() {
  // Not silent: a suppressed alert is a bug report in itself, and without this
  // there is no way to tell "the gate ate it" from "that code path never ran".
  console.debug(
    'session-lost: suppressed a native alert() behind the session-expired modal. ' +
      'This is expected while the modal is open.',
  )
}

export function isSessionAlertGateArmed() {
  return originalNativeAlert !== null
}

/**
 * Make window.alert a no-op until releaseSessionAlertGate().
 *
 * Called by SessionExpiredModal from its session-lost listener, not from
 * reportSessionLost() directly: the listener loop there is synchronous, so arming
 * here is still live before any calling continuation can reach an alert(), and this
 * component - unlike useSessionLost - knows the current route and can decline to arm
 * on a screen the modal does not apply to.
 *
 * Safe to call twice; the second call is a no-op so a burst of reports cannot stack
 * patches or lose the original.
 */
export function armSessionAlertGate() {
  // A missing alert is not a missing gate to report on: tests and non-DOM hosts
  // legitimately have no window, and some hardened browsers make it non-writable.
  if (typeof window === 'undefined' || typeof window.alert !== 'function') return
  if (originalNativeAlert !== null) return

  originalNativeAlert = window.alert
  try {
    window.alert = suppressedAlert
  } catch (err) {
    // Losing the gate is survivable - the modal still appears; worst case one alert
    // stacks on it - so a locked-down global must not break the report that led here.
    originalNativeAlert = null
    console.error('session-lost: could not suppress window.alert:', err)
  }
}

/**
 * Put the real window.alert back.
 *
 * Only undoes our own patch: if something else replaced window.alert while the gate
 * was armed, that install is newer than ours and outranks it.
 */
export function releaseSessionAlertGate() {
  const original = originalNativeAlert
  if (original === null) return
  originalNativeAlert = null

  if (typeof window === 'undefined' || window.alert !== suppressedAlert) return
  try {
    window.alert = original
  } catch (err) {
    console.error('session-lost: could not restore window.alert:', err)
  }
}

/**
 * Is this caught error the auth gate talking, rather than an action failing?
 *
 * Covers both routes to that answer: sessionLostError() above, and the apiFailure()
 * errors useApi.js builds from a `{ code }` envelope - so a helper that throws
 * apiFailure(result, ...) for a rejected token is recognised here too, not just the
 * ten client-side guards.
 *
 * @param {any} err
 * @returns {boolean}
 */
export function isSessionLostError(err) {
  return err?.code === AUTH_FAILURE_CODE
}

/**
 * The message a caller should render for a caught error - or '' when the
 * session-expired modal already owns it.
 *
 * Two things were wrong with `error.value = err.message` in the catch blocks this
 * replaces. The raw English string 'User not authenticated' leaked into the UI,
 * because those blocks had no translator; and they rendered *a second* explanation
 * of a failure the modal is already explaining, in a screen the reader cannot reach
 * while the modal is up.
 *
 * '' is the right answer rather than a translated string: the modal is
 * non-dismissable and sits above everything, so anything set here is invisible.
 * Both call sites render with `v-if="error"`, so an empty string hides the message
 * without any branching at the call site.
 *
 * @param {any} err the caught error
 * @param {string} fallback the caller's own generic message
 * @returns {string}
 */
export function sessionErrorMessage(err, fallback) {
  if (isSessionLostError(err)) return ''
  return err?.message || fallback
}
