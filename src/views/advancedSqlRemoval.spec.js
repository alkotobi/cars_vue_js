import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'

// The Advanced SQL screen is removed, and these tests exist to stop it coming
// back by accident.
//
// It was a page that let an admin type arbitrary SQL into a textarea and POST it
// to `execute_multi_sql`. Two details made that worse than it looked:
//   - executeQuery()/executeMultiQuery() prepare and execute the statement BEFORE
//     reading its first six characters, so the SELECT/INSERT/UPDATE/DELETE switch
//     only shaped the response. A DROP TABLE, or an UPDATE against users, took
//     effect and then answered "Invalid query type".
//   - execute_multi_sql applied no semicolon check (execute_sql has a weak one),
//     so a single request could run a batch of statements.
//
// The screen was gated by the router and again in its own onMounted, but both are
// client-side. api.php has no global authentication, so the endpoint was reachable
// by anyone who could reach the server - `api/.htaccess` sets CORS and PHP limits
// and nothing else.
//
// Deleting a feature is easy to undo by accident during a refactor, so the route,
// the endpoint, the helper and the table are each pinned here.
//
// This file later grew to cover the whole caller-supplied-SQL surface, because
// execute_sql survived that change and was the next thing to go.

const API_PHP = readFileSync(new URL('../../api/api.php', import.meta.url), 'utf8')
const ROUTER = readFileSync(new URL('../router/index.js', import.meta.url), 'utf8')
const SETUP_SQL = readFileSync(new URL('../../api/setup.sql', import.meta.url), 'utf8')
const USE_API = readFileSync(new URL('../composables/useApi.js', import.meta.url), 'utf8')

const exists = (path) => {
  try {
    readFileSync(path)
    return true
  } catch {
    return false
  }
}

describe('Advanced SQL is gone', () => {
  it('no longer has a view', () => {
    expect(exists(new URL('../views/AdvancedSqlView.vue', import.meta.url))).toBe(false)
  })

  it('has no route', () => {
    expect(ROUTER).not.toMatch(/advanced-sql/i)
    expect(ROUTER).not.toContain('AdvancedSqlView')
  })

  it('has no way to navigate to it', () => {
    // The Params sidebar was the only entry point; the header only had a title
    // string for the route name.
    const params = readFileSync(new URL('../views/ParamsView.vue', import.meta.url), 'utf8')
    const header = readFileSync(new URL('../components/AppHeader.vue', import.meta.url), 'utf8')
    const app = readFileSync(new URL('../App.vue', import.meta.url), 'utf8')

    for (const [name, source] of [
      ['ParamsView', params],
      ['AppHeader', header],
      ['App.vue', app],
    ]) {
      expect(source, `${name} still references advanced sql`).not.toMatch(/advanced[-_]?sql/i)
    }
  })

  it('has no locale strings', () => {
    for (const loc of ['en', 'ar', 'fr', 'zh']) {
      const messages = readFileSync(new URL(`../locales/${loc}.json`, import.meta.url), 'utf8')
      expect(messages, `${loc}.json still has navigation.advancedSql`).not.toContain(
        '"advancedSql"',
      )
    }
  })

  it('has no multi-statement endpoint and no helper behind it', () => {
    expect(API_PHP).not.toMatch(/execute_multi_sql/)
    // executeMultiQuery() split on ';' and ran every statement, with no
    // allow-list and no type check that ran before execution.
    expect(API_PHP).not.toContain('executeMultiQuery')
  })

  it('has no table', () => {
    expect(SETUP_SQL).not.toMatch(/`adv_sql`/)
    const install = readFileSync(new URL('../../api/install.php', import.meta.url), 'utf8')
    expect(install).not.toContain('adv_sql')

    // The migration that drops it on existing databases.
    const migration = readFileSync(
      new URL('../../api/migrations/030_drop_adv_sql.sql', import.meta.url),
      'utf8',
    )
    expect(migration).toMatch(/DROP TABLE IF EXISTS `adv_sql`/)
  })
})

// execute_sql outlived the screen by a release, and was worse than the screen it
// belonged to: no auth at all, and one of its four callers was in ClientDetailsView,
// a route the router exempts from authentication for share-token clients. Arbitrary
// SQL was reachable from a shared link with nobody logged in.
describe('execute_sql is gone', () => {
  it('has no endpoint', () => {
    expect(API_PHP).not.toMatch(/case\s+'execute_sql'/)
  })

  it('has no callers left in the app', () => {
    for (const file of [
      '../components/containers/GoogleMapPopup.vue',
      '../components/containers/ContainersRefList.vue',
      '../views/ClientDetailsView.vue',
    ]) {
      const source = readFileSync(new URL(file, import.meta.url), 'utf8')
      expect(source, `${file} still calls execute_sql`).not.toMatch(/action:\s*'execute_sql'/)
    }
  })

  it('replaced the map popup write with a token-gated named action', () => {
    const containers = readFileSync(
      new URL('../../api/actions/containers.php', import.meta.url),
      'utf8',
    )
    expect(containers).toMatch(/function handle_save_container_tracking/)
    // id_user came from the token, not from a literal 1 in the component.
    expect(containers).toMatch(/\$user\['id'\]/)
    expect(API_PHP).toMatch(/case\s+'save_container_tracking'/)
  })

  it('replaced the two tracking reads with one named action', () => {
    const list = readFileSync(
      new URL('../components/containers/ContainersRefList.vue', import.meta.url),
      'utf8',
    )
    // The N+1 is gone: one call, not a call per container.
    expect(list).toMatch(/await getContainerTracking\(\)/)
    expect(list).not.toMatch(/Promise\.all\(trackingPromises\)/)
    expect(API_PHP).toMatch(/case\s+'get_container_tracking'/)
  })
})

// Closing execute_sql is not enough on its own: the bare {query} endpoint still ran
// any statement it was given. It now requires a token verified server-side.
describe('the SQL passthrough requires a session', () => {
  it('gates the passthrough on a server-verified token', () => {
    // The gate has to sit before executeQuery() runs, not after.
    const gateIndex = API_PHP.indexOf('require_api_user($passthroughConn, $postData)')
    expect(gateIndex).toBeGreaterThan(-1)

    const passthroughIndex = API_PHP.indexOf('// Regular query handling')
    expect(gateIndex).toBeGreaterThan(passthroughIndex)
  })

  it('loads the auth helpers it needs', () => {
    expect(API_PHP).toMatch(/require_once __DIR__ \. '\/lib\/auth\.php'/)
  })

  it('stops trusting the client for user_id and is_admin on this path', () => {
    // Scoped to the passthrough. The payment_confirmed branch used to read
    // `user_id` and `is_admin` off the payload: send is_admin: true and the
    // hasPermission() check was skipped, so any caller could confirm payments.
    const passthrough = API_PHP.slice(API_PHP.indexOf('// Regular query handling'))
    expect(passthrough).not.toMatch(/\$postData\['is_admin'\]/)
    expect(passthrough).not.toMatch(/\$postData\['user_id'\]/)
    // Both now come from the token that was verified above.
    expect(passthrough).toMatch(/\$passthroughUser\['id'\]/)
    expect(passthrough).toMatch(/\$passthroughUser\['role_id'\]/)
  })

  it('does not trust a client-supplied database name', () => {
    expect(API_PHP).not.toMatch(/\$postData\['dbname'\]/)
    // It resolves the database from the per-deployment db_code.json instead.
    expect(API_PHP).toMatch(/function resolveDbNameFromCode/)
    expect(USE_API).not.toMatch(/dbname: db_name/)
  })

  it('sends the token by default rather than only when asked', () => {
    expect(USE_API).not.toMatch(/data\.requiresAuth \? null : localStorage/)
  })
})

// A gate with no way through for anonymous callers would break the share page and
// the login screen. Those are the only two, and both are named actions now - a
// client-sent "public" flag on the passthrough would have been the same bug as the
// is_admin flag above.
describe('the two public reads are their own actions', () => {
  it('the share page is one read-only action keyed by share_token', () => {
    const share = readFileSync(
      new URL('../../api/actions/client_share.php', import.meta.url),
      'utf8',
    )
    expect(API_PHP).toMatch(/case\s+'get_client_share_data'/)
    expect(share).toMatch(/function handle_get_client_share_data/)

    // Scoped to the client the token resolves to, never to caller input.
    expect(share).toMatch(/WHERE c\.share_token = \?/)
    // Read-only: no writes in a handler anyone can reach without a token. Comments
    // are stripped first, since this file's own prose names the statements it avoids.
    const code = share
      .split('\n')
      .map((line) => line.replace(/\/\/.*$/, ''))
      .join('\n')
    expect(code).not.toMatch(/\b(INSERT\s+INTO|UPDATE\s+\w|DELETE\s+FROM|DROP\s+TABLE)\b/i)
  })

  it('the version dialog has its own action', () => {
    // DatabaseVersionCheck is mounted unconditionally in App.vue, so it runs on
    // /login with no session and would 401 behind the gate.
    expect(API_PHP).toMatch(/case\s+'get_db_version'/)
    const versionCheck = readFileSync(
      new URL('../composables/useVersionCheck.js', import.meta.url),
      'utf8',
    )
    expect(versionCheck).toMatch(/await getDbVersion\(\)/)
    expect(versionCheck).not.toMatch(/query:\s*'SELECT version/)
  })

  it('no public view still posts raw SQL', () => {
    // The router exempts exactly these three paths. /settings makes no calls.
    for (const file of ['../views/ClientDetailsView.vue', '../composables/useVersionCheck.js']) {
      const source = readFileSync(new URL(file, import.meta.url), 'utf8')
      expect(source, `${file} posts raw SQL`).not.toMatch(/query:\s*`/)
      expect(source, `${file} posts raw SQL`).not.toMatch(/query:\s*'/)
    }
  })
})

// hash_password and insert_user were named actions, which made them look narrower
// than the passthrough. They were not: the action said which column to hash, the
// payload said which rows to write, and neither checked a token. hash_password took
// "UPDATE users SET password = ? WHERE username = ?" straight from the login form,
// so anyone could rewrite any account's password.
describe('user writes go through named actions', () => {
  const USERS_PHP = readFileSync(new URL('../../api/actions/users.php', import.meta.url), 'utf8')

  it('neither endpoint exists', () => {
    expect(API_PHP).not.toMatch(/case\s+'hash_password'/)
    expect(API_PHP).not.toMatch(/case\s+'insert_user'/)
  })

  it('no action runs a caller-supplied statement any more', () => {
    const dispatch = API_PHP.slice(API_PHP.indexOf("switch($postData['action'])"))
    expect(dispatch).not.toMatch(/executeQuery\(\$postData/)
  })

  it('the login screen no longer downloads password hashes', () => {
    const login = readFileSync(new URL('../views/LoginView.vue', import.meta.url), 'utf8')
    // Comments stripped: the file explains in prose what it used to send.
    const code = login
      .split('\n')
      .map((line) => line.replace(/\/\/.*$/, ''))
      .join('\n')

    expect(code).not.toMatch(/u\.password/)
    expect(code).not.toMatch(/action:\s*'verify_password'/)
    expect(code).toMatch(/action:\s*'change_password_with_credentials'/)
  })

  it('verify_password is gone, so a stolen hash cannot be brute-forced cheaply', () => {
    // It answered "does this password match this hash" for any pair, which made it
    // a free offline-cracking oracle for whoever downloaded a hash.
    expect(API_PHP).not.toMatch(/case\s+'verify_password'/)
  })

  it('changing a password targets a row the request cannot choose freely', () => {
    // change_own_password takes no username or user id at all.
    const own = USERS_PHP.slice(
      USERS_PHP.indexOf('function handle_change_own_password'),
      USERS_PHP.indexOf('function handle_set_user_password'),
    )
    expect(own).not.toMatch(/\$postData\['username'\]/)
    expect(own).not.toMatch(/\$postData\['user_id'\]/)
    expect(own).toMatch(/\$user\['id'\]/)

    // The admin variant takes an id and is admin-gated.
    expect(USERS_PHP).toMatch(/function handle_set_user_password[\s\S]*require_api_admin/)
  })

  it('rotates the api_token when a password changes', () => {
    expect(USERS_PHP).toMatch(/UPDATE users SET password = \?, api_token = \?/)
  })
})

// The allowlist closes the dozen-odd actions that read $postData['is_admin'] to
// decide what to do. Sending is_admin: true used to be all it took.
describe('the action switch is gated', () => {
  it('requires a token for everything not explicitly public', () => {
    const gate = API_PHP.indexOf("in_array($postData['action'], PUBLIC_ACTIONS, true)")
    const dispatch = API_PHP.indexOf("switch($postData['action'])")
    expect(gate).toBeGreaterThan(-1)
    expect(gate).toBeLessThan(dispatch)
  })

  it('lists public actions as names the server decides, not a request flag', () => {
    const list = API_PHP.slice(
      API_PHP.indexOf('const PUBLIC_ACTIONS'),
      API_PHP.indexOf(']', API_PHP.indexOf('const PUBLIC_ACTIONS')),
    )
    expect(list).toMatch(/'ping'/)
    expect(list).toMatch(/'login'/)
    expect(list).toMatch(/'change_password_with_credentials'/)
    expect(list).toMatch(/'get_client_share_data'/)
    expect(list).toMatch(/'get_db_version'/)

    // A client-settable "public" escape hatch would be the is_admin bug again.
    expect(API_PHP).not.toMatch(/\$postData\['public'\]/)
  })
})

describe('payment_confirmed write gate', () => {
  // Regression, worth pinning because the failure mode was so indirect.
  //
  // The gate was briefly reduced to "does the query mention payment_confirmed
  // anywhere", and in the same edit $queryUpper was deleted. That broke two things
  // at once. AlertsView.vue's unconfirmed-payment COUNT is a SELECT that merely
  // reads `sb.payment_confirmed = 0`, so it matched the gate, entered the block,
  // and hit `strpos($queryUpper, 'UPDATE')` on an undefined variable. PHP emitted
  // "Warning: Undefined variable" into the response body ahead of the JSON, so the
  // browser failed with `Unexpected token '<', "<br />\n<b>"...` and the actual
  // cause - a wrongly-classified read - was nowhere in the message.
  //
  // So: the verb has to be the first real token (leading comments stripped), and
  // the column has to appear somewhere.

  it('defines $queryUpper before the gate consumes it', () => {
    const assign = API_PHP.indexOf('$queryUpper = strtoupper(')
    expect(assign).toBeGreaterThan(-1)

    // First use after the assignment must come later in the file, otherwise some
    // other block was reading it before this point in the request lifecycle.
    const firstUse = API_PHP.indexOf('$queryUpper', assign + 1)
    expect(firstUse).toBeGreaterThan(assign)

    // Derived from comment-stripped SQL, which is what makes the prefix tests
    // below resistant to a leading `/* comment */`.
    expect(API_PHP).toMatch(
      /\$queryUpper = strtoupper\(api_strip_leading_sql\(trim\(\$query\)\)\)/,
    )
  })

  it('requires a write verb at the start of the statement, not just a mention', () => {
    // Anchored at the statement start so a read that merely names the column is
    // not gated. Compared as a literal because the PHP pattern is full of regex
    // metacharacters and escaping it again just to assert on it is unreadable.
    expect(API_PHP).toContain(
      "preg_match('/^(UPDATE|INSERT|REPLACE)\\b/', $queryUpper) === 1",
    )
    // The broken form: non-empty check only, so any SELECT naming the column gated.
    expect(API_PHP).not.toMatch(/api_strip_leading_sql\(trim\(\$query\)\) !== ''/)
  })

  it('requires the column to be named', () => {
    expect(API_PHP).toContain("strpos($queryUpper, 'PAYMENT_CONFIRMED') !== false")
  })

  it('does not narrow the gate back to UPDATE only', () => {
    // UPDATE-only let `INSERT INTO sell_bill (payment_confirmed) VALUES (1)` write
    // the column with no permission check. The gate runs before shape detection,
    // so widening the verbs does not disturb the simple-format rewrite below it.
    expect(API_PHP).toContain('UPDATE|INSERT|REPLACE')
  })
})
