import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'

// The DB manager was 100% dead, and the tests passed anyway.
//
// db_manager_api.php gated every action behind the tenant app's admin token -
// require_api_user($appConn, $postData) - but every caller in the UI talks to it
// with a hand-rolled fetch() against a different database entirely, where no app
// token is ever sent. The DB manager has its own `login` table in merhab_databases
// with its own session in localStorage, and LoginSignup.vue stored that session's
// response under `db_manager_user`. Nobody was bridging the two, so login itself
// was behind the gate it could not satisfy: not one action could run, including
// the one that would have issued a credential.
//
// The failure was invisible to the suite because it was entirely structural - no
// test called the endpoint. So these tests assert on the shape of the contract
// rather than on behaviour: that the gate names a token the callers actually send,
// and that every db_manager_api.php call site carries it.

const DB_MANAGER_API = readFileSync(
  new URL('../../api/db_manager_api.php', import.meta.url),
  'utf8',
)
const API_PHP = readFileSync(new URL('../../api/api.php', import.meta.url), 'utf8')
const SETUP_SQL = readFileSync(new URL('../../api/setup.sql', import.meta.url), 'utf8')
const EXPORT_SQL = readFileSync(new URL('../../api/export.sql', import.meta.url), 'utf8')
const MIGRATION_031 = readFileSync(
  new URL('../../api/migrations/031_login_api_token.sql', import.meta.url),
  'utf8',
)
const DBM_API = readFileSync(new URL('../composables/useDbManagerApi.js', import.meta.url), 'utf8')
const DB_MANAGER_VIEW = readFileSync(new URL('./DbManagerView.vue', import.meta.url), 'utf8')

// Every .vue file that talks to db_manager_api.php.
const DB_MANAGER_CALLERS = [
  '../components/db-manager/Databases.vue',
  '../components/db-manager/UpdateDbStructure.vue',
  '../components/params/GeneralSettingsForm.vue',
]

const read = (rel) => readFileSync(new URL(rel, import.meta.url), 'utf8')

// The `action: '...'` lines in a file, paired with whether a token is sent
// alongside. Pairing them is the point: a test that only counted tokens, or only
// counted actions, would pass on a file where every action had a token but one.
const dbManagerBodies = (source) => {
  const bodies = []
  const lines = source.split('\n')
  for (let i = 0; i < lines.length; i++) {
    const match = lines[i].match(/^\s*action: '([a-z_]+)',\s*$/)
    if (!match) continue
    const window = lines.slice(i, i + 4).join('\n')
    bodies.push({ action: match[1], hasToken: /token: getDbManagerToken\(\)/.test(window) })
  }
  return bodies
}

describe('the DB manager authenticates against its own login table', () => {
  it('no longer gates on the tenant app token', () => {
    // This was the bug: an app credential the DB manager's callers never have.
    expect(DB_MANAGER_API).not.toMatch(/require_api_user/)
  })

  it('gates on login.api_token instead', () => {
    expect(DB_MANAGER_API).toMatch(/api_token/)
    expect(DB_MANAGER_API).toMatch(/function dbm_token_user/)
  })

  it('resolves the caller from that token, in the gate itself', () => {
    // Asserting only that the helper is defined somewhere in the file was not
    // enough: stubbing out the call site kept every test green while leaving no
    // authentication at all. So this looks inside the gate block.
    const gate = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf('if (!in_array($action, DBM_PUBLIC_ACTIONS'),
      DB_MANAGER_API.indexOf('switch ($action)'),
    )
    expect(gate).toMatch(/dbm_token_user\(\$conn, \$token\)/)
    expect(gate).toMatch(/=== null/)
  })

  it('reads the token out of the request body it was sent in', () => {
    const gate = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf('if (!in_array($action, DBM_PUBLIC_ACTIONS'),
      DB_MANAGER_API.indexOf('switch ($action)'),
    )
    expect(gate).toMatch(/\$inputData\['token'\]/)
  })

  it('leaves login, signup and the public code lookup reachable', () => {
    // login has to be reachable, or the manager can never obtain the token the
    // gate wants. That circularity is the whole reason this allowlist exists.
    const list = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf('const DBM_PUBLIC_ACTIONS'),
      DB_MANAGER_API.indexOf(']', DB_MANAGER_API.indexOf('const DBM_PUBLIC_ACTIONS')),
    )
    expect(list).toMatch(/'login'/)
    expect(list).toMatch(/'signup'/)
    expect(list).toMatch(/'get_database_by_code'/)
  })

  it('does not let the caller mark its own request public', () => {
    expect(DB_MANAGER_API).not.toMatch(/\$inputData\['public'\]/)
  })

  it('rejects an unauthenticated call with an envelope rather than a fatal', () => {
    expect(DB_MANAGER_API).toMatch(/'success'\s*=>\s*false/)
  })

  it('checks public actions before dispatching', () => {
    const gate = DB_MANAGER_API.indexOf('DBM_PUBLIC_ACTIONS')
    const dispatch = DB_MANAGER_API.indexOf('switch ($action)')
    expect(gate).toBeGreaterThan(-1)
    expect(gate).toBeLessThan(dispatch)
  })
})

describe('db-manager credentials are issued and revoked server-side', () => {
  it('mints a token at login', () => {
    expect(DB_MANAGER_API).toMatch(/function dbm_new_token/)
    expect(DB_MANAGER_API).toMatch(/random_bytes\(32\)/)
    expect(DB_MANAGER_API).toMatch(/UPDATE login SET api_token/)
  })

  it('returns the token so the client has something to send back', () => {
    // Silent on this, the gate would reject every subsequent call.
    expect(DB_MANAGER_API).toMatch(/'token'\s*=>\s*\$token/)
  })

  it('gives one generic failure for both unknown user and wrong password', () => {
    // Otherwise the response distinguishes the two and becomes a username oracle.
    const login = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf("case 'login':"),
      DB_MANAGER_API.indexOf("case 'logout':"),
    )
    // Three distinct replies used to be an oracle for "does this user exist and is
    // it enabled?" against the table holding host-level credentials. Compared
    // against the messages the case actually emits - the comment above it names all
    // three variants, so comments are stripped first.
    const emitted = login
      .split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n')

    expect(emitted).toMatch(/'Invalid username or password'/)
    expect(emitted).not.toMatch(/not found|does not exist|deactivated|inactive|disabled/i)
  })

  it('compares against a dummy hash when the user does not exist', () => {
    // Skipping the comparison for unknown users makes the failure instant, which
    // is a timing oracle even when the message is generic.
    expect(DB_MANAGER_API).toMatch(/const DBM_DUMMY_HASH = '\$2y\$/)
    // The unknown user falls back to the dummy hash, so password_verify still runs
    // on a real bcrypt compare and takes the same time.
    expect(DB_MANAGER_API).toMatch(/\$userData\['pass'\] \?\? DBM_DUMMY_HASH/)
  })

  it('rotates the token on each login', () => {
    const login = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf("case 'login':"),
      DB_MANAGER_API.indexOf("case 'logout':"),
    )
    const verify = login.indexOf('password_verify')
    const rotate = login.indexOf('UPDATE login SET api_token')
    expect(verify).toBeGreaterThan(-1)
    expect(rotate).toBeGreaterThan(verify)
  })

  it('revokes the token on logout', () => {
    expect(DB_MANAGER_API).toMatch(/case 'logout':/)
    expect(DB_MANAGER_API).toMatch(/SET api_token = NULL/)
  })

  it('requires authentication to reach logout', () => {
    // Not in the public list: an unauthenticated logout would be a no-op that
    // looks like it worked.
    const list = DB_MANAGER_API.slice(
      DB_MANAGER_API.indexOf('const DBM_PUBLIC_ACTIONS'),
      DB_MANAGER_API.indexOf(']', DB_MANAGER_API.indexOf('const DBM_PUBLIC_ACTIONS')),
    )
    expect(list).not.toMatch(/'logout'/)
  })
})

describe('the token has somewhere to live', () => {
  it('is a column on login', () => {
    expect(MIGRATION_031).toMatch(/login/)
    expect(MIGRATION_031).toMatch(/api_token/)
  })

  it('is indexed, because every gated request looks up by it', () => {
    expect(MIGRATION_031).toMatch(/INDEX `idx_login_api_token` \(`api_token`\)/)
  })

  it('is present for fresh installs too', () => {
    expect(SETUP_SQL).toMatch(/api_token/)
    expect(EXPORT_SQL).toMatch(/api_token/)
  })

  it('is idempotent, since it is run against live databases', () => {
    // MariaDB has no ADD COLUMN IF NOT EXISTS, so this checks INFORMATION_SCHEMA
    // before each ALTER - re-running against an existing server has to be a no-op,
    // not "Duplicate column name".
    expect(MIGRATION_031).toMatch(/INFORMATION_SCHEMA\.COLUMNS/)
    expect(MIGRATION_031).toMatch(/INFORMATION_SCHEMA\.STATISTICS/)
    expect(MIGRATION_031).toMatch(/PREPARE stmt FROM/)
  })

  it('targets whatever database it is pointed at, not a hard-coded one', () => {
    // Deployments run this against merhab_databases while every other migration in
    // the set targets merhab_cars, so a hard-coded name would be a footgun.
    expect(MIGRATION_031).toMatch(/SET @db_name = DATABASE\(\)/)
    const sql = MIGRATION_031.split('\n')
      .filter((line) => !line.trim().startsWith('--'))
      .join('\n')
    expect(sql).not.toMatch(/merhab_databases|merhab_cars/)
  })
})

// The invariant changed once the calls were centralised. It used to be "every caller
// attaches the token by hand" - true, and true only until someone added a call site
// and forgot. It is now "no component hand-rolls a db-manager request at all", which
// cannot be satisfied by forgetting.
describe('db-manager requests go through one helper', () => {
  const CALLERS = [
    ...DB_MANAGER_CALLERS,
    '../components/db-manager/DbManagerSidebar.vue',
    '../components/db-manager/LoginSignup.vue',
  ]

  for (const file of CALLERS) {
    it(`${file} does not hand-roll a fetch to the endpoint`, () => {
      // This is the assertion that would have caught the original outage. The token
      // is attached inside the helper now, so a component that builds its own request
      // body has opted out of both the token and the session-loss handling.
      expect(read(file)).not.toMatch(/fetch\(`?[^)]*db_manager_api\.php/)
    })
  }

  it('the helper attaches the token to every request it makes', () => {
    expect(DBM_API).toMatch(/body\.token = token/)
  })

  it('reads it from the DB-manager session, not the app session', () => {
    // getStoredToken() reads `user`. Mixing the two would send an app token the
    // registry rejects, which is the same dead end in a new place.
    const fn = DBM_API.slice(
      DBM_API.indexOf('export function getDbManagerToken'),
      DBM_API.indexOf('export function setDbManagerSession'),
    )
    expect(fn).toMatch(/db_manager_user/)
    expect(fn).not.toMatch(/localStorage\.getItem\('user'\)/)
  })

  it('sends no token for login and signup, which run before there is one', () => {
    // They must be marked anonymous, and the helper must honour that: a token sent
    // here would imply the gate is optional.
    expect(DBM_API).toMatch(/anonymous: true/)
    expect(DBM_API).toMatch(/options\.anonymous \? null : getDbManagerToken\(\)/)

    const loginSignup = read('../components/db-manager/LoginSignup.vue')
    expect(loginSignup).toMatch(/anonymous: true/)
    expect(loginSignup).not.toMatch(/getDbManagerToken/)
  })
})

describe('no db-manager action travels in a query string', () => {
  // A GET puts the token in the access log, which is a credential on disk.
  it('no caller requests db_manager_api.php with an action= parameter', () => {
    for (const file of DB_MANAGER_CALLERS) {
      expect(read(file)).not.toMatch(/db_manager_api\.php\?/)
    }
  })
})

describe('logout ends the session on the server', () => {
  it('the sidebar revokes server-side and clears locally', () => {
    const sidebar = read('../components/db-manager/DbManagerSidebar.vue')
    expect(sidebar).toMatch(/dbManagerRequest\('logout'\)/)
    expect(sidebar).toMatch(/clearDbManagerToken/)
  })

  it('clears locally even when the request fails', () => {
    // Otherwise a network blip leaves the credential the user asked to be rid of,
    // and refusing to log out locally strands the user in a session they cannot end.
    const sidebar = read('../components/db-manager/DbManagerSidebar.vue')
    const fn = sidebar.slice(sidebar.indexOf('const logout'))
    expect(fn).toMatch(/finally/)
  })
})

describe('protected-file overwrite cannot fail open', () => {
  // Overwriting config.php or db_manager_config.php is how you take over an
  // install, so the confirmation has to appear whenever the server cannot
  // positively say the files are absent.
  const uploads = read('../components/db-manager/Databases.vue')
  const check = uploads.slice(uploads.indexOf('const findExistingProtectedFiles'))

  it('asks the DB manager, not the app API', () => {
    expect(uploads).toMatch(/dbManagerRequest\('check_api_files_exist'/)
    // and definitely not the tenant app's api.php, which no longer serves it. The
    // lookbehind matters: db_manager_api.php ends in api.php.
    expect(uploads).not.toMatch(/(?<!_)api\.php/)
  })

  it('warns about every candidate unless the server confirmed none exist', () => {
    // The old filter was `success && data`, so a refused or failed check fell
    // through to no prompt at all - overwriting silently exactly when the server
    // could not answer.
    expect(check).toMatch(/if \(!checkResult\.success/)
    const failSafe = check.slice(check.indexOf('if (!checkResult.success'))
    expect(failSafe).toMatch(/return fileNames/)
  })

  it('still treats an explicit per-file answer as authoritative', () => {
    expect(check).toMatch(/filter\(\(name\) => checkResult\.data\[name\] === true\)/)
  })

  it('prompts before uploading, not after', () => {
    const call = check.indexOf('await findExistingProtectedFiles(protectedFilesToUpload)')
    const uploadStart = check.indexOf('uploadingPhp.value = true')
    expect(call).toBeGreaterThan(-1)
    expect(uploadStart).toBeGreaterThan(call)
  })
})

// A token can stop being valid while the page is open: signed out in another tab, or
// api_token cleared by hand in the registry. DbManagerView decided "logged in" from
// the mere presence of a localStorage key, so that left a permanently broken shell -
// panels rendered, every request refused - instead of the login form. Only the server
// knows the token is dead, so the rejection has to be recognised as an event rather
// than rendered as an ordinary error.
describe('a revoked token drops back to the login form', () => {
  it('recognises the gate rejection specifically', () => {
    // Sharing a code path with ordinary action failures would log users out over a
    // failed backup or a bad version number.
    expect(DBM_API).toMatch(/export function isDbManagerSessionLost/)
    const fn = DBM_API.slice(
      DBM_API.indexOf('export function isDbManagerSessionLost'),
      DBM_API.indexOf('const listeners'),
    )
    expect(fn).toMatch(/success === false/)
    expect(fn).toMatch(/startsWith\(AUTH_FAILURE_PREFIX\)/)
  })

  it('matches on a prefix, so rewording the tail cannot silently disable it', () => {
    expect(DBM_API).toMatch(/AUTH_FAILURE_PREFIX = 'Not authenticated\.'/)
  })

  it('every helper request checks the envelope it gets back', () => {
    expect(DBM_API).toMatch(/isDbManagerSessionLost\(result\)/)
  })

  it('checks the streaming path too, via a clone', () => {
    // backup_databases returns a file, so there is no envelope to read in the normal
    // way - but the gate still answers JSON, and that has to be noticed.
    const raw = DBM_API.slice(DBM_API.indexOf('export async function dbManagerRequestRaw'))
    expect(raw).toMatch(/response\.clone\(\)/)
    expect(raw).toMatch(/isDbManagerSessionLost/)
  })

  it('does not clone the large body when it is a download', () => {
    // Cloning a multi-megabyte dump to inspect it would buffer it twice. The clone is
    // inside the JSON branch only.
    const raw = DBM_API.slice(DBM_API.indexOf('export async function dbManagerRequestRaw'))
    const jsonBranch = raw.indexOf("contentType.includes('application/json')")
    const clone = raw.indexOf('response.clone()')
    expect(jsonBranch).toBeGreaterThan(-1)
    expect(clone).toBeGreaterThan(jsonBranch)
  })

  it('clears the stale credential before telling anyone', () => {
    // DbManagerView derives isLoggedIn from that same key. A listener that ran first
    // and read storage would see the token still present.
    const fn = DBM_API.slice(
      DBM_API.indexOf('function reportSessionLost'),
      DBM_API.indexOf('export async function dbManagerRequest'),
    )
    expect(fn.indexOf('clearDbManagerToken()')).toBeLessThan(fn.indexOf('listener()'))
  })

  it('reports once per dead token, not once per in-flight request', () => {
    // Revoking a token while several requests are open makes all of them fail, and
    // re-rendering the login form per response would be its own flicker.
    expect(DBM_API).toMatch(/reportedToken/)
    const fn = DBM_API.slice(
      DBM_API.indexOf('function reportSessionLost'),
      DBM_API.indexOf('export async function dbManagerRequest'),
    )
    expect(fn).toMatch(/if \(!token \|\| token === reportedToken\) return/)
  })

  it('never reports when there was no token to lose', () => {
    // Otherwise a login attempt that happens to hit the same message could nuke a
    // session belonging to another tab.
    const fn = DBM_API.slice(
      DBM_API.indexOf('function reportSessionLost'),
      DBM_API.indexOf('export async function dbManagerRequest'),
    )
    expect(fn).toMatch(/if \(!token/)
  })

  it('the view unsubscribes on unmount', () => {
    // The view is behind a route guard. A surviving listener would keep writing to a
    // ref nothing renders, and would accumulate one per visit.
    expect(DB_MANAGER_VIEW).toMatch(/onUnmounted/)
    expect(DB_MANAGER_VIEW).toMatch(/onDbManagerSessionLost/)
    expect(DB_MANAGER_VIEW).toMatch(/stopListening\(\)/)
  })

  it('the view no longer treats a localStorage key as proof of a session', () => {
    // Comments stripped first: the replacement is described just above the code, and
    // the old expression appears there verbatim.
    const code = DB_MANAGER_VIEW.split('\n')
      .filter((line) => !line.trim().startsWith('//'))
      .join('\n')
    expect(code).not.toMatch(/localStorage\.getItem\('db_manager_user'\)/)
    expect(code).toMatch(/!!getDbManagerToken\(\)/)
  })

  it('returns to the login form on session loss and says why', () => {
    const fn = DB_MANAGER_VIEW.slice(DB_MANAGER_VIEW.indexOf('onDbManagerSessionLost(() =>'))
    expect(fn).toMatch(/isLoggedIn\.value = false/)
    expect(fn).toMatch(/sessionLostMessage\.value/)
  })

  it('clears the warning on a fresh sign-in', () => {
    // Otherwise a successful login is announced underneath a stale "session ended".
    expect(DB_MANAGER_VIEW).toMatch(/handleLoginSuccess/)
    const fn = DB_MANAGER_VIEW.slice(
      DB_MANAGER_VIEW.indexOf('const handleLoginSuccess'),
      DB_MANAGER_VIEW.indexOf('const handleLogout'),
    )
    expect(fn).toMatch(/sessionLostMessage\.value = ''/)
  })
})

describe('check_api_files_exist is reachable', () => {
  // It was moved out of api.php because it reads files outside any tenant's
  // scope - it inspects the install root - so it belongs behind the DB manager's
  // own auth rather than the app's, where it sat with no caller-authenticated
  // purpose at all.
  it('is served by db_manager_api.php', () => {
    expect(DB_MANAGER_API).toMatch(/case 'check_api_files_exist':/)
  })

  it('is no longer on the app API', () => {
    expect(API_PHP).not.toMatch(/case 'check_api_files_exist'/)
  })

  it('still resolves file names inside the install root', () => {
    const action = DB_MANAGER_API.slice(DB_MANAGER_API.indexOf("case 'check_api_files_exist':"))
    expect(action).toMatch(/realpath/)
    expect(action).toMatch(/file_exists/)
  })
})

describe('api.php survives a body it cannot use', () => {
  // $postData went straight into require_api_user(array $postData) and the switch,
  // so a scalar or null body - a truncated request, a bare `"x"`, an empty body -
  // raised a TypeError instead of a 4xx. Is this in a SQL file, incidentally.
  it('rejects a non-array body before anything else uses it', () => {
    expect(API_PHP).toMatch(/is_array\(\$postData\)/)
  })

  it('answers with a client error, not a fatal', () => {
    const guard = API_PHP.slice(API_PHP.indexOf('is_array($postData)'))
    expect(guard).toMatch(/400/)
  })

  it('guards before the action gate that is typed to take an array', () => {
    // Anchored on the gate, not on the first mention of require_api_user - that
    // string appears in a comment near the top of the file.
    const guard = API_PHP.indexOf('is_array($postData)')
    const gate = API_PHP.indexOf("in_array($postData['action'], PUBLIC_ACTIONS, true)")
    expect(guard).toBeGreaterThan(-1)
    expect(gate).toBeGreaterThan(-1)
    expect(guard).toBeLessThan(gate)
  })
})

// The Create button reported success while silently leaving out part of the
// schema.
//
// create_tables does not run api/migrations/*.sql. It replays api/setup.sql,
// which is the whole point: a fresh tenant is built from one file rather than
// from 30-odd migrations. But the statement splitter ended by keeping only
// CREATE TABLE / INSERT (later DROP TRIGGER and CREATE TRIGGER), and setup.sql
// carries one ALTER TABLE - the only thing in the file that puts a foreign key
// on buy_details.id_car_name, which it has to add afterwards because
// buy_details is created before the cars_names it references.
//
// ALTER was not on the list, so it was filtered out with no error. Every fresh
// database came back missing the constraint and its supporting index, the
// action still reported success, and nothing in the suite noticed. The tell
// was a freshly built tenant differing from a migrated one by exactly one
// foreign key.
describe('setup.sql is applied in full by create_tables', () => {
  const statements = DB_MANAGER_API.slice(
    DB_MANAGER_API.indexOf('function dbm_is_setup_statement('),
  )

  it('keeps every statement kind setup.sql actually contains', () => {
    // Each keyword here corresponds to a construct present in setup.sql. Adding a
    // construct to that file without adding it to this list silently drops it.
    for (const keyword of [
      'CREATE TABLE',
      'ALTER TABLE',
      'CREATE TRIGGER',
      'DROP TRIGGER',
      'INSERT',
    ]) {
      expect(statements).toContain(keyword)
    }
  })

  it('would not drop the foreign key setup.sql adds to buy_details', () => {
    // The regression guard proper: the ALTER that installs the constraint.
    expect(SETUP_SQL).toMatch(
      /ALTER TABLE `buy_details`[\s\S]*?ADD INDEX `idx_buy_details_id_car_name`/,
    )
    expect(SETUP_SQL).toMatch(/FOREIGN KEY \(`id_car_name`\) REFERENCES `cars_names`/)
    expect(statements).toMatch(/ALTER TABLE/)
  })

  it('splits on the DELIMITER the trigger blocks are fenced with', () => {
    // Without this the BEGIN...END bodies are cut at their inner semicolons.
    expect(DB_MANAGER_API).toMatch(/DELIMITER/)
    expect(SETUP_SQL).toMatch(/DELIMITER \$\$/)
  })

  it('keeps the raw driver text out of the create_tables response', () => {
    // Scoped to this action on purpose. Other actions still append
    // $e->getMessage() to their responses, which is pre-existing and untouched
    // here; claiming otherwise would be a test that passes only by being vague.
    const start = DB_MANAGER_API.indexOf("case 'create_tables':")
    const end = DB_MANAGER_API.indexOf('case ', start + 10)
    const action = DB_MANAGER_API.slice(start, end)

    expect(start).toBeGreaterThan(-1)
    expect(action).not.toMatch(/\$response\['message'\][^\n]*getMessage\(\)/)
    expect(action).not.toMatch(/\$errors\[\] = \$e->getMessage\(\)/)
    // It still has to be recorded somewhere.
    expect(action).toMatch(/error_log\(/)
  })

  it('creates the database when the registry names one that is not there yet', () => {
    // db_name is free text from the create form, so pressing Create on a new name
    // has to create the database rather than fail with "Unknown database".
    expect(DB_MANAGER_API).toMatch(/CREATE DATABASE IF NOT EXISTS/)
  })

  it('validates that name before it reaches a statement or a DSN', () => {
    expect(DB_MANAGER_API).toMatch(/function dbm_validate_database_name\(/)
    // The DSN is semicolon-delimited, so an unvalidated ';' would truncate dbname.
    expect(DB_MANAGER_API).toMatch(/dbm_validate_database_name/)
  })
})
