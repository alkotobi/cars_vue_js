# Security findings

Open, not yet fixed. Recorded here so they are tracked rather than living only in a
chat log. Nothing in this document has been remediated; do not read it as a
description of the current state of a fix.

Found 2026-10-03 while reviewing `src/views/CarNamesView.vue`. Every finding
below is in the shared API layer, not in that view, so fixing the view did not
address any of them.

## Severity: critical

`api/api.php` executes caller-supplied SQL with **no authentication and no
permission check**, and sends `Access-Control-Allow-Origin: *`. This is the
default code path: any POST that omits `action` reaches it.

- `api/api.php:2` — `header('Access-Control-Allow-Origin: *');`
- `api/api.php:2717` — `// Regular query handling`, reached when no `action` matched
- `api/api.php:2725` — `$query = $postData['query'];`
- `api/api.php:58` — `executeQuery()` runs `$conn->prepare($sql); $stmt->execute($params)`

`$postData['token']` is never read anywhere in `api.php`. The token that
`useApi` sends on every call is transmitted and then ignored.

PDO blocks stacked statements by default, so this is any *single* statement. That
is still enough to read or destroy the entire database:

- `SELECT * FROM users` returns every row, including `password` hashes and
  `api_token` values.
- `DELETE FROM users`, `UPDATE users SET password = ...`,
  `DROP TABLE cars_stock` all execute.

Because the response header is a wildcard, this is reachable cross-origin from any
page a logged-in user's browser visits.

### Client-side admin checks protect nothing

Views gate destructive buttons on a value read from `localStorage`. For example
`src/views/CarNamesView.vue` computes `isAdmin` from `localStorage.getItem('user')`
and hides Delete buttons when it is false. The hidden request can simply be sent
directly. This pattern is used across the app, not just that view.

## Severity: high — permission checks are bypassable

Where permission checks do exist, they are gated on a client-supplied flag:

```php
$isAdmin = isset($postData['is_admin']) ? (bool)$postData['is_admin'] : false;
...
if (!$isAdmin && !hasPermission($conn, $userId, '<some permission>')) { /* 403 */ }
```

Because `$isAdmin` comes from the request body, sending `is_admin: 1` skips
`hasPermission()` entirely. The flag is never derived from the token or the
database. Affected sites:

| Line | Guard | Permission bypassed |
|------|-------|----------------------|
| `api/api.php:1014` → `:1023` | `$isAdmin &&` | `can_upload_car_files` (car files listing) |
| `api/api.php:1014` → `:1114` | `$isAdmin &&` | permission filtering when listing files |
| `api/api.php:1268` | `$isAdmin` | admin-only file delete |
| `api/api.php:2219` → `:2228` | `$isAdmin &&` | `can_upload_car_files` (physical transfer) |
| `api/api.php:2741` → `:2750` | `$isAdmin &&` | `can_confirm_payment` |

## Severity: medium — DDL runs while reporting failure

`executeQuery()` decides its response from the first six characters of the
statement:

```php
$queryType = strtoupper(substr(trim($sql), 0, 6));
switch ($queryType) {
    case 'SELECT': ...
    case 'INSERT': ...
    case 'UPDATE': case 'DELETE': ...
    default: return ['success' => false, 'error' => 'Invalid query type'];
}
```

`DROP`, `ALTER`, `TRUNCATE` and `CREATE` hit `default`, so the statement **has
already run** by the time the function reports `Invalid query type`. A caller can
drop a table and be told it failed. Any retry logic built on that response will
happily run it again.

## Dead code, and a schema comment that is wrong

`api/lib/auth.php` implements the correct approach:

- `api/lib/auth.php:21` — `api_token_user($conn, $token)`
- `api/lib/auth.php:58` — `require_api_admin($conn, array $postData)`

The only file that requires it is `api/actions/supplier_credibility.php:26`, so
these functions are unreachable everywhere else.

`api/migrations/021_users_api_token.sql:27` adds the column with the comment:

> `Token sent by useApi; checked server-side by lib/auth.php`

and `api/setup.sql:224` repeats it. That is accurate only for the supplier
credibility action. Anyone reading the schema will reasonably conclude the token
is checked on every request. It is not. Correct the comment when this is fixed.

## Remediation order

1. **Require a valid token on the default path.** Reject the request unless
   `api_token_user()` resolves it to a user. This is the single change that closes
   the critical finding.
2. **Stop trusting `is_admin` from the request body.** Derive it from the resolved
   user, or drop it and rely on `hasPermission()` alone. Fix all five sites.
3. **Reject non-DML statements** in `executeQuery()` before preparing them, rather
   than after.
4. **Replace the raw-query path with named server actions**, each with its own
   permission check — the shape `api/actions/supplier_credibility.php` already
   uses. 28 of 32 views currently post raw SQL, so this is the bulk of the work
   and should follow 1–3 rather than precede them.
5. **Restrict CORS** to the app's own origin instead of `*`.
6. **Rotate every credential** that was ever committed, pasted, or left in a
   remote URL. As of this writing the GitHub PAT and the GitLab token in
   `.git/config` are both exposed and still valid.

Steps 1 and 2 are small and stop the bleeding. Step 4 is the real fix and should
be planned as its own piece of work.

## Reproducing

Read-only checks against a scratch database — do not run the DELETE or DROP
statements against anything you care about.

Confirmed on a local instance on 2026-10-03. All three returned `200` with data:

```
# 1. No token at all
$ curl -s -X POST http://127.0.0.1:8123/api.php -H 'Content-Type: application/json' \
    -d '{"query":"SELECT COUNT(*) AS n FROM users","dbname":"merhab_cars"}'
{"success":true,"data":[{"n":1}]}

# 2. Bogus token — byte-identical, so the token is not consulted
$ curl -s -X POST http://127.0.0.1:8123/api.php -H 'Content-Type: application/json' \
    -d '{"query":"SELECT COUNT(*) AS n FROM users","token":"not-a-real-token","dbname":"merhab_cars"}'
{"success":true,"data":[{"n":1}]}

# 3. Credential material, still no token
$ curl -s -X POST http://127.0.0.1:8123/api.php -H 'Content-Type: application/json' \
    -d '{"query":"SELECT id, LEFT(password,12) AS pw, api_token FROM users LIMIT 2","dbname":"merhab_cars"}'
{"success":true,"data":[{"id":1,"pw":"$2y$10$xk9Kh","api_token":"<64 hex chars>"}]}
```

Response 3 is the severity argument: a `$2y$10$` bcrypt hash and a live
`api_token` are readable by anyone who can reach the endpoint. The `api_token`
returned by that request was captured, which is on its own a reason to rotate it
rather than merely note it.

To reproduce the other findings:

```bash
# is_admin is self-asserted: send is_admin:1 to skip hasPermission()
# at api/api.php:2750 (payment confirmation) and :1023 / :2228 (car files).

# DDL reports failure after succeeding — check the table is actually gone:
curl -s -X POST http://127.0.0.1:8123/api.php -H 'Content-Type: application/json' \
  -d '{"query":"DROP TABLE some_scratch_table","dbname":"merhab_cars"}'
# -> {"success":false,"error":"Invalid query type"}   ... and the table is gone.
```

## Related

`CarNamesView.vue` was hardened separately: deletes that would orphan rows are
now refused with the referencing records listed, failures surface in a dialog
instead of the console, and brand/car-name validation is enforced. That work is
data-integrity and UX only. **It does not mitigate anything in this document**,
because the browser is not a trust boundary here.
