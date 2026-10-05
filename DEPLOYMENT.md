# Cars Management System — Deployment Runbook

Battle-tested guide from the live deploy to **https://world-automobile.com/cars**
(VPS `163.245.214.125`, domain `world-automobile.com`). Follow exactly; every command
was executed and verified against production.

> **Secrets live ONLY in server files.** DB passwords are in `api/config.local.php`,
> which is git-ignored and never committed. `api/config.php` is a tracked loader
> that resolves credentials from `config.local.php`, falling back to the
> `DB_HOST` / `DB_USER` / `DB_PASS` environment variables.
> (`merhab_root` — MariaDB user, hosts `localhost` + `127.0.0.1`). App + db-manager
> admin login is `admin / 123`, stored as bcrypt in the DB. Do NOT paste credentials
> into chats/docs.
>
> **`db_name` must stay empty in `api/config.local.php`.** `DB_HOST`/`DB_USER`/
> `DB_PASS` are connection settings and are needed; the database is not one of them.
> Which database a request is served from is resolved per request from the client in
> the URL, and `api/api.php:getDbConfig()` **refuses** (`db_unavailable`) anything that
> resolves to no client at all. It used to fall back to `db_name` instead, which meant
> a request arriving with no client in it — the repository root, or any path that is
> not a client folder — quietly opened whichever single database that setting named.
> On this machine that was `merhab_cars`: a live tenant with real users, password
> hashes and `api_tokens`. `db_name` is now read only by `api/install.php`, which has
> to be told which database to create and refuses without it.

---

## 1. Architecture

```
nginx :443 (world-automobile.com)          ← or a raw IP
  root /var/www
  ├─ /dist/      ONE build, shared by every client
  ├─ /api/       ONE codebase, shared by every client
  └─ /<CLIENT>/
       ├─ /<CLIENT>/          → alias to /dist/, SPA fallback
       ├─ /<CLIENT>/files/…   → that client's uploads, and nothing else
       └─ /<CLIENT>/api/…     → alias to /api/, with CARDS_TENANT=<CLIENT>
```

`<CLIENT>` is the client's **database name**, unchanged, and it is also its folder.
One name identifies the database, the folder, the URL prefix and — with
`base_domain` set — the subdomain. `api_valid_tenant()` (`api/lib/appdb.php`) accepts
`^[a-z][a-z0-9_]{0,63}$` and refuses `api` and `dist`, which are the shared folders
at the same level; the provisioning library enforces the identical set.

A client folder holds `files/` and nothing else. There is no per-client build and no
per-client `api/`, so a release is two rsyncs and every client moves together:

```
dist/ → /var/www/dist/     api/ → /var/www/api/
```

Two ways to serve the same client, both generated from the `dbs` registry:

| Form | URL | Base injected |
| --- | --- | --- |
| Path | `https://host/<client>/…` | `<base href="/<client>/">` |
| Subdomain | `https://<client>.host/…` | none — the app is already at the root |

  - **Mount point** — `src/utils/basePath.js` reads `import.meta.url`, not
    `location.pathname` (see §8.4); Vite `base: './'`. In path form the served HTML
    gets a `<base>` injected by nginx (§6.1); in subdomain form it is not, because
    there is no prefix to add.
- **API base** — `resolveApiBaseUrl()` in the same file: `<origin><mount>api`, so
  `https://host/<client>/api`. Same rule in `index.html` for pre-boot fetches.
- **Router** — `createWebHistory(getBasePath())`, so deep links survive reloads.
- **Which client is this?** `api/lib/appdb.php`. nginx sets `CARDS_TENANT` in each
  location block, and that value wins over everything else. Failing that, the first
  path segment; failing that, the first host label **compared against the configured
  `base_domain`**, never guessed. `/api/api.php` at the webroot root names no client
  and resolves to none, so it is refused rather than guessed at (§6).
- **DB-manager** uses `api/db_manager_api.php` → DB `merhab_databases`.
- **2 databases**: `merhab_cars` (app), `merhab_databases` (manager). The third,
  `merhab_invitations`, went with the invitations feature (§4.5).

---

## 2. SSH access + deployment key

```bash
alias deployssh="ssh -i ~/.ssh/cars_deploy root@163.245.214.125"
```

Pubkey (`cars_deploy.pub`) is in `root@…:~/.ssh/authorized_keys`. Rotate the keypair
on this machine; on the server just replace the line. Always use `www-data` for
`api/config.local.php` owner so PHP-FPM can read it:
`chown www-data:www-data api/config.local.php`. Create it on a new server with
`cp api/config.example.php api/config.local.php`, then fill in the credentials.

Git deployment key for pushes: `~/.ssh/cars_gh_ekotobi` (ed25519), registered as a
**write** deploy key on `github.com/ekotobi/cars_vue_js.git` via `gh api`. Remote:

```bash
github.com-ekotobi:ekotobi/cars_vue_js.git
# ~/.ssh/config
Host github.com-ekotobi
  HostName github.com
  User git
  IdentityFile ~/.ssh/cars_gh_ekotobi
  IdentitiesOnly yes
```

---

## 3. Server stack (verified)

- Ubuntu; nginx 1.28.3; PHP 8.5.4 + php8.5-fpm (socket `unix:/run/php/php8.5-fpm.sock`).
- MariaDB 11.8.6, unix socket auth for root; app user `merhab_root`@`localhost`+`127.0.0.1`.
- Web root `/var/www/world-automobile.com`; vhost `/etc/nginx/sites-available/default`.
- Certbot cert exists for `world-automobile.com`; upload size overridden in php.ini
  (nginx `client_max_body_size 10G`).

---

## 4. Databases

### 4.1 Create user + DBs (as root via socket)

```sql
CREATE USER 'merhab_root'@'localhost' IDENTIFIED BY '<from api/config.local.php>';
CREATE USER 'merhab_root'@'127.0.0.1' IDENTIFIED BY '<same>';
GRANT ALL PRIVILEGES ON merhab_cars.*        TO 'merhab_root'@'localhost';
GRANT ALL PRIVILEGES ON merhab_cars.*        TO 'merhab_root'@'127.0.0.1';
GRANT ALL PRIVILEGES ON merhab_databases.*   TO 'merhab_root'@'localhost';
GRANT ALL PRIVILEGES ON merhab_databases.*   TO 'merhab_root'@'127.0.0.1';
FLUSH PRIVILEGES;
```

### 4.2 Create DBs

```sql
CREATE DATABASE IF NOT EXISTS merhab_cars        DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS merhab_databases   DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
```

### 4.3 App schema (MariaDB — REQUIRED fixes)

`api/setup.sql` is MySQL-8 oriented. For MariaDB apply this or you get
`ERROR 150` / collation errors:

1. **99 that9 that0 replace:** `sed -i 's/utf8mb4_0900_ai_ci/utf8mb4_general_ci/g' setup.sql`
2. **FK ordering:** fixed. `users` is declared before `car_name_media`, `upgrades`,
   `car_apgrades` and `sell_bill`, so the file now has no forward references and
   applies in a single pass (55 tables, 0 FK errors). The old
   `order_schema.py` top-sort is no longer needed.
3. **`car_apgrades.id_upgrade` sign mismatch:** fixed. `setup.sql` now declares
   `int unsigned`, matching `upgrades.id`. No `ALTER TABLE` needed on new installs;
   existing databases keep the manual fix from their original deploy.

Then seed: `roles` (admin/user/SELLER), `users` (admin + `role_id` 1, password bcrypt `123`),
`versions` = **25** (matches `useVersionCheck.js`). Insert the `dbs` row only in the
**manager** DB (below).

### 4.4 Manager DB (`merhab_databases`)

Tables: `login` + `dbs`. Seed:
- `login`: admin / 123 with same bcrypt hash as app users, `active=1`.
- `dbs` is the **client list** — the source the nginx config is generated from. One
  row per client, keyed by its name, which is also its database name:
  ```sql
  INSERT INTO dbs (db_code, db_name, files_dir, js_dir, is_created)
  VALUES (NULL, 'merhab_cars', '/merhab_cars/files', '/merhab_cars', 1);
  ```

  **`db_name` must equal the folder and the URL prefix**, unchanged. It is used as
  the MySQL database name, the folder under the webroot and the first URL segment,
  and nothing normalises between them — that is the point. Lowercase letters, digits
  and underscores only (`^[a-z][a-z0-9_]{0,63}$`); `api` and `dist` are refused
  because they are the shared folders. A dash or a dot used to be normalised to an
  underscore, which produced a folder `acme-motors` holding a database
  `acme_motors` — two names for one client, and every rename, backup and support
  question then needs the rule to connect them.

  **`files_dir` is the client's uploads inside its own folder**, and the recorded
  value is not used to build paths — it is derived from the name and a row that
  disagrees is reported as a mismatch the Provision screen can fix. Deriving it is
  what removed the old sibling `<client>_files/` arrangement, where `files_dir` was
  the only thing recording the layout and a leading slash silently dropped meant
  uploads resolved to `<client>/<client>_files/…`: they appeared to succeed and every
  stored file then 404'd.

  `js_dir` is the client's URL prefix, `/<client>` — not `/`, and not a path to
  anything. It is recorded so the DB manager and the renderer agree on the URL, and
  so a hand-edited row is caught rather than silently ignored.

  `is_created = 1` is what makes the renderer emit a location block for the client. A
  row without it is a draft: it has no database and no folder, so a block for it would
  hand out a URL that 500s.

#### Migration 031: `login.api_token`

`db_manager_api.php` authenticates against this table's `api_token`, not the app's
`users` table — the DB manager is a separate trust domain guarding host-level
operations. `login` needs one more column:

```bash
mysql -u USER -p merhab_databases < api/migrations/031_login_api_token.sql
```

**Target the manager DB (`merhab_databases`), not `merhab_cars`** — this is the one
migration in the set whose DB name differs from the others. Run it before deploying
frontend code that logs into the DB manager, or every login attempt returns HTTP 500
from the `UPDATE login SET api_token` in the login case. Existing installs need it;
fresh installs get the column from `api/setup.sql`.

Seeded rows have `api_token = NULL`, which is the correct signed-out state — the
first successful login mints a token. To revoke a session by hand:

```sql
UPDATE login SET api_token = NULL WHERE username = 'admin';
```

#### What the Create button actually does

Worth being explicit, because it is not what the label suggests and it is a
common source of confusion:

- **It does not run `api/migrations/*.sql`.** It replays `api/setup.sql` in full.
  That is deliberate: a fresh tenant is built from one file rather than from 30-odd
  migrations, so `setup.sql` has to stay in sync with the migration set on its own.
  If `setup.sql` and the migrations disagree, the migrations are the truth for
  existing databases and `setup.sql` is the truth for new ones.
- **It creates the database.** `db_name` is free text from the create form, so the
  button issues `CREATE DATABASE IF NOT EXISTS` before applying anything. Names go
  through `tenant_assert_valid_db_name()` first, which is the same check the runtime
  resolver makes: lowercase letters, digits and underscores, starting with a letter,
  at most 64 characters, and not `api` or `dist`. An unvalidated `;` would truncate
  the PDO DSN's `dbname`, and a name `appdb.php` refuses would provision a client
  whose own site cannot find its database.
- **It does not provide the app.** The folder it builds holds `files/` and nothing
  else. The build is `/var/www/dist/` and the API is `/var/www/api/`, both shared and
  both deployed by `deploy/deploy.sh`. The frontend derives its API base as
  `<origin><basePath>api` (`src/utils/basePath.js`) and nginx maps that to the shared
  `api/`, so a client whose nginx block has not been rendered 404s every call — the
  symptom is a wall of `POST <client>/api/api.php 404` in the console, not a PHP
  error.
- **The version number is not validated.** The modal's value is written straight to
  `versions`; nothing confirms `setup.sql` really corresponds to it. `db_updates` in
  the manager DB is what drives later upgrades, and it ships empty, so Update
  Structure has nothing to apply until migrations are registered there.
- **A silently-dropped statement is the failure mode to watch for.** The statement
  splitter used to keep only `CREATE TABLE` / `INSERT`, so the `ALTER TABLE` that
  installs the `buy_details.id_car_name` foreign key was filtered out with no error
  and the action still reported success — every fresh database came back missing the
  constraint. `src/views/dbManagerAuth.spec.js` now asserts the keep-list covers every
  construct `setup.sql` uses.

### 4.5 Invitations — removed.

The invitations feature is gone: `api/invitations.php`, `src/views/InvitationsView.vue`,
`src/components/InvitationsTable.vue`, the `/invitations` route, its dashboard button and
the `invitations.*` / `dashboard.invitations` strings in all four locale files. A fresh
deployment needs none of it, and the third database is not part of the setup above.

The `merhab_invitations` database itself is **not** dropped by that — it is left in place
with its rows, so this is reversible. Drop it by hand once you are sure:

```bash
mysqldump -u root -p merhab_invitations > backups/merhab_invitations_$(date +%F).sql
mysql -u root -p -e "DROP DATABASE merhab_invitations;"
```

**Already-deployed folders still carry the old file.** `deploy/deploy.sh` deliberately
rsyncs *without* `--delete`, so `invitations.php` (and the stale `InvitationsView` chunk)
survives in `<webroot>/<FOLDER>/api/` and keeps answering until removed by hand:

```bash
rm -f /var/www/<host>/<FOLDER>/api/invitations.php
rm -f /var/www/<host>/<FOLDER>/InvitationsView.*.js /var/www/<host>/<FOLDER>/InvitationsView.*.css
```

Until then it is only reachable as an authenticated admin — `require_app_admin()` is gone
from `api/lib/appdb.php` with its only caller, so a stale copy calls an undefined
function and returns 500 rather than serving data.

---

## 5. Build + deploy

One command, **per server** — not per client. It builds, refuses a non-portable build,
uploads the two shared directories, checks the per-server credential file exists, and
verifies the result:

  ```bash
  ./deploy/deploy.sh <host> [user@server] [webroot]
  # e.g.
  ./deploy/deploy.sh world-automobile.com
  ./deploy/deploy.sh world-automobile.com root@163.245.214.125 /var/www

  # ALWAYS preview first. Builds and verifies, lists every file that would be
  # transferred, asserts no *.local.php is in the list, and writes nothing.
  ./deploy/deploy.sh world-automobile.com --dry-run
  ```

  There is no client argument, and that is the point: a release is `dist/` and `api/`
  and nothing else. Every client on the server moves to the new version at once, and
  there is no per-client rsync that can be run for some clients and forgotten for
  others — which is how two clients ended up on different code with nothing in the
  output saying so. Client folders are never written to by this script at all; they
  hold only uploads.

  `--dry-run` matters because this script writes to a live server. A deployment once
  went out without being intended, and once shipped `db_manager_config.local.php`,
  breaking the db-manager with "Access denied for user 'root'". The flag may appear
  anywhere in the arguments and is never mistaken for a positional.

  After deploying, re-render nginx **only if the client list changed** (§6A). A new
  release does not change the URL layout, so this is not part of every deploy.

<details>
<summary>Manual equivalent</summary>

```bash
npm ci && npm run build
grep -rE 'world-automobile|localhost:8000' dist --include='*.js'   # must print nothing

WEBROOT=/var/www
SSH="ssh -i ~/.ssh/cars_deploy root@$HOST"

  # --delete IS used on both, and is safe because neither destination holds anything
  # that is not part of the deployment. Without it, a feature deleted from src/ stayed
  # reachable on every server forever: removing the invitations feature left
  # InvitationsView.*.js and invitations.php in each webroot, still served.
  rsync -az --delete --exclude '.DS_Store'  -e "ssh -i ~/.ssh/cars_deploy" dist/ root@$HOST:$WEBROOT/dist/
  # EVERY *.local.php is PER SERVER and must be excluded, or this machine's credentials
  # overwrite the server's. Do not enumerate the credential files by name -- a new one
  # will be missed, and that is exactly how db_manager_config.local.php was shipped
  # once. --exclude also protects them from --delete, which never removes an excluded
  # file from the destination.
  rsync -az --exclude '*.local.php'        -e "ssh -i ~/.ssh/cars_deploy" api/  root@$HOST:$WEBROOT/api/
```

Nothing is written into `<webroot>/<client>/`. A client folder holds `files/`, and a
`--delete` that reached one would delete that client's invoices — which is exactly
what the old per-client rsync had to guard against with an exclude, and what putting
the build in its own folder makes unnecessary.

</details>

### Per-server vs per-build

| File | Scope | Why |
|---|---|---|
| `dist/` (hashed assets) | **shared** | one build serves every client; folder- and host-agnostic by construction |
| `api/` (the codebase) | **shared** | one set of endpoints, one set of credentials |
| `api/config.local.php` | **per server** | registry DB password; git-ignored, never deployed |
| `base_domain` in `api/config.php` | **per server** | the domain subdomain tenants are served on; empty means path form only |
| nginx vhost | **per server** | generated from the `dbs` registry; the only place client names appear |

`api/config.local.php` also holds the optional `ai_base_url` / `ai_api_key` /
`ai_model` keys behind the supplier credibility button (any OpenAI-compatible
endpoint: OpenAI, DeepSeek, OpenRouter, Ollama). They are per-server for the same
reason the DB password is: one build serves many clients, and the key belongs to
the machine holding it. Leave them empty and the feature reports "not configured
on this server" rather than failing. The same values can come from `AI_API_KEY`,
`AI_BASE_URL` and `AI_MODEL`.

`ai_base_url` is the API **root** (`https://openrouter.ai/api/v1`), not the full
endpoint: `api/lib/ai_client.php` appends `/chat/completions` itself.

### Switching the credibility feature off

Two independent switches, and it takes both to hide the feature completely:

| Switch | Where | Effect |
|---|---|---|
| `CREDIBILITY_ENABLED` | `src/lib/featureFlags.js` (compile-time) | hides the column, the row-menu entry, the modal; makes no API call |
| `credibility_enabled` | `api/config.local.php` (per server, default `1`) | every credibility endpoint answers `credibility_disabled` |

The server-side one is the guarantee. It runs inside the handlers after the admin
check, so a cached bundle, a stale tab or a direct call to `api.php` cannot spend
money on a model the operator has switched off — which the UI flag alone cannot
prevent, because a flag in a bundle is only a request not to render a button.

Set `'credibility_enabled' => '0'` per server. The default is `1` so that
upgrading a working server never switches the feature off by surprise; the value
also comes from `CREDIBILITY_ENABLED` in the environment, and anything
unrecognised falls back to the default rather than guessing, so a typo cannot
silently enable it. Neither switch deletes anything: the component, composable,
endpoint, styles and translations all stay in place and tested.

To verify a server is really off, with an admin token and no model call:

```sh
curl -s -X POST https://<host>/<folder>/api/api.php \
  -H 'Content-Type: application/json' \
  -d '{"action":"get_supplier_credibility_latest","token":"<admin api_token>"}'
# {"success":false,"code":"credibility_disabled","error":"credibility_disabled"}
```

The endpoint reads a JSON body (`action`, not `query` — `query` is the separate
raw-SQL entry point) and the token is the `api_token` column on `users`.

A check blocks until the model answers, so expect the request to take as long as
the slowest model you would point it at — 10-30s is normal, and a free model
sharing a rate-limited pool can be slower still. `ai_client.php` raises PHP's own
execution limit for the call, but if your host overrides that per request (nginx
`fastcgi_read_timeout`, or a hard PHP-FPM pool limit) the user gets a gateway
error instead of a clean "the model is busy" message. Both `AI_TIMEOUT` (default
60) and the button's own timeout are worth raising on slow models.

Free and shared models are rate-limited upstream, which surfaces as
`ai_rate_limited` with a retry button; a key the provider refuses surfaces as
`ai_auth_failed`, which is deliberately not retryable.

**There is no `db_code.json` any more.** The database is chosen by the URL: the
client's name is its database name, resolved by `api/lib/appdb.php` from
`CARDS_TENANT` (set by nginx), else the first path segment, else the host label
compared against `base_domain`. Nothing is fetched at runtime and nothing is compiled
in, so there is no file that can be stale, missing, or ship inside a build pointing a
client at the wrong database. `deploy/deploy.sh` still fails if a
`dist/db_code.json` appears, as a check that the file has not come back.

The client's name must be a valid tenant name (`^[a-z][a-z0-9_]{0,63}$`, not `api` or
`dist`) — the screen, the provisioning library and the runtime resolver all enforce
the same set, and a name they cannot agree on provisions a client whose own site
cannot find its database.

Writable dirs (created, owned by www-data), **per client, inside the client folder**:
`<client>/files/` (subdirs buy_pi, sell_pi, documents, ids, payments_swift,
banks_logos, letter_head, logo, uploads, chat_files).

> Nothing writes to a client folder except uploads. A `--delete` that reached one
> would delete that client's documents, so `deploy/deploy.sh` never targets them.

---

## 6. nginx app block (PHP MUST be nested — key fix!)

**There is no per-client nginx config to write.** One generated file holds every
client, both URL forms, and it is produced from the `dbs` registry — never hand-edited,
because a hand-edited block is the state in which a deleted client's data stays
reachable at a URL nobody remembers owning:

```bash
# as root, through the installed wrapper (see §6A for the one-time setup)
/usr/local/bin/cars-nginx-render --check     # print what would be installed
/usr/local/bin/cars-nginx-render --write     # install it, then nginx -t and reload
```

For one client the render is the same as for ten, and the file looks like this (one
`location` per client inside a single `server`, plus optional subdomain vhosts outside
it):

```nginx
# ---- merhab_cars ----
location /merhab_cars/ {
    alias /var/www/dist/;                          # the ONE build
    try_files $uri /merhab_cars/index.html;
}
location = /merhab_cars/index.html {
    alias /var/www/dist/index.html;
    sub_filter_once on;
    sub_filter '<meta charset="UTF-8">' '<base href="/merhab_cars/"><meta charset="UTF-8">';
}
location ^~ /merhab_cars/files/ {
    root /var/www;                                 # → /var/www/merhab_cars/files/…
    try_files $uri =404;                           # =404, not the SPA shell
}
location ~ ^/merhab_cars/api/(?<cars_api_merhab_cars_[0-9a-f]{8}>/.*\.php)$ {
    set $cars_api_merhab_cars_[0-9a-f]{8} $1;
    alias /var/www/api/$2;
    include snippets/fastcgi-php.conf;
    fastcgi_param CARDS_TENANT merhab_cars;        # the tenant, not a guess
    fastcgi_pass unix:/run/php/php8.5-fpm.sock;
}
location ~ ^/merhab_cars/api/.*\.(sql|sh|env)$ { deny all; }
location ~ ^/merhab_cars/files/.*\.php$       { deny all; }
```

  Two decisions in there are load-bearing and easy to undo by accident:

  - **`alias`, not `root`, for the shared build.** `root` would resolve
    `/merhab_cars/sell-bills/index.html` against the *document path*, miss every file
    (the build is not under `/merhab_cars/` on disk), and hand back the shell for
    every asset.
  - **The build location is a plain prefix, NOT `^~`.** `^~` stops nginx considering
    regex locations, so `/merhab_cars/api/*.php` would fall through to the SPA fallback
    and return HTML for a POST — a MIME-type error in the browser console with nothing
    in the access log to explain it.

The renderer validates each client name against the same allowlist the app resolves
(`^[a-z][a-z0-9_]{0,63}$`, plus a refusal of `api` and `dist`) and refuses to render at
all on a name that could terminate an nginx token. This matters: the hardening rules
are the only thing between a public client and `install.php` / `setup.sql`.

The full per-block output, and why each rule is where it is:

  - **A directory is not in `try_files`.** Route paths can collide with real
    directories — `/cars/cars/` is a real directory holding `logo.png`. With `$uri/`,
    nginx matched that directory, found no `index.html` inside it, and returned **403
    Forbidden** instead of the app. It looks intermittent because client-side
    navigation never issues a request for the URL — only a refresh or a pasted link
    reaches `try_files`. Dropping it means a path that is a directory but not a route
    falls through to the SPA and the router decides. Static assets are matched by the
    first test and are unaffected.
  - **The API is a regex location, and the build's is not `^~`.** `^~` beats regex
    locations, so a `^~` build block would swallow `/cars/api/*.php` and return HTML
    for a POST — a MIME error in the console with nothing in the access log. The API
    location therefore uses a named capture, which also keeps the per-client variable
    names distinct; nginx refuses to start on a duplicate variable, so two clients
    would take every site on the machine down.
  - **`CARDS_TENANT` is set, not derived.** `api/lib/appdb.php` trusts it above
    everything else. The alternative — guessing the tenant from `SCRIPT_NAME` — broke
    the moment the shared `api/` existed, because the script path has no client in it.
  - **The `.sql` deny sits next to the API block, not at server level.** A
    `~* \.sql$` rule at server level was in the vhost on production while all three
    dumps still returned **200**, because the `^~` prefix meant the regex was never
    evaluated. Deny rules that appear correct and do nothing are worse than none.
  - **`/api` at the webroot root is 404.** A request there names no client, so serving
    it would mean guessing which database it meant. Same reason `/` is 404: a client
    that has not been provisioned has no folder to be served from.
  - **What `.sql` blocks.** `export.sql`, `setup.sql`, `recreate_selection_tables.sql`
    and `migrations/*.sql` were served as plain text — the full DB schema, public. No
    feature regresses: `db_manager_api.php` reads `setup.sql` from disk via `__DIR__`,
    and backups are generated and served by `backup.php`.

After re-rendering, **verify rather than assume** — a deny rule in the file does not
mean the request is denied:

  ```bash
  C=cars   # any provisioned client
  for f in api/install.php api/setup.sql api/export.sql; do
    printf '%-24s %s\n' "$f" "$(curl -s -o /dev/null -w '%{http_code}' https://host/$C/$f)"
  done   # every one must be 403/404

  curl -s -o /dev/null -w '%{http_code}\n' https://host/api/api.php    # must be 404: no client named
  curl -s -o /dev/null -w '%{http_code}\n' https://host/nosuchclient/  # must be 404
  curl -s -o /dev/null -w '%{http_code}\n' https://host/$C/files/nope.png  # 404, not the shell
  ```

  ### 6.1 The `<base>` injection (deep links)
  
  `index.html` references its entry script relatively (`./index.<hash>.js`) and ships
  **no** `<base>`, because the folder is per-client and unknown at build time. A
  relative URL resolves against the *document*, which is correct at depth 1 only.
  
  On `/cars/sell-bills/5` the browser therefore requested
  `/cars/sell-bills/index.<hash>.js`, missed, and `try_files` handed back
  `index.html` with `Content-Type: text/html`:
  
  ```
  Failed to load module script: Expected a JavaScript-or-Wasm module script but the
  server responded with a MIME type of "text/html".
  ```
  
  This broke **every route with a path parameter** and predates the folder-agnostic
  work. The page returned 200 throughout, so it presented as a blank tab rather than
  an error, and no build or config change appeared to cause it.
  
  nginx injects the base when serving the shell, so one `dist/` still serves any
  client:
  
  ```nginx
  location = /cars/index.html {
      sub_filter_once on;
      sub_filter '<meta charset="UTF-8">' '<base href="/cars/"><meta charset="UTF-8">';
  }
  ```
  
  Needs `ngx_http_sub_module` (stock Ubuntu/Debian: present). The anchor is
  `<meta charset>`, present in every build; `sub_filter_once` keeps it to the first
  match so comment prose about `<base>` is untouched. `sub_filter_types` is not set:
  its only purpose is to add to the default `text/html`, and naming it produced a
  "duplicate MIME type" warning on every `nginx -t`.
  
  **Subdomain form gets no injection.** At `https://cars.example.com/` the app is
  already at the root, so `<base href="/">` would be right and `<base href="/cars/">`
  would 404 every asset. The renderer emits the `location = /index.html` block only
  for the path form.

  **Consequence:** `index.html` is load-bearing for this fix. Removing the
  `<meta charset>` anchor, or serving the app from a vhost without this block, breaks
  every deep link. If you ever see the MIME error, check for `<base>` in the served
  HTML first:
  
  ```bash
  curl -s https://host/cars/sell-bills/5 | grep -o '<base[^>]*>'   # must print one
  curl -s -o /dev/null -w '%{http_code} %{content_type}\n' \
    https://host/cars/index.<hash>.js                                  # must be application/javascript

  # subdomain form: must print nothing
  curl -s https://cars.example.com/sell-bills/5 | grep -o '<base[^>]*>'
  ```
  
  This also fixes the `index.html` bootstrap (`getApiBase()` resolves `./api` against
  the document). Without `<base>`, `/cars/cars/` produced a 404 for
  `/cars/cars/api/db_manager_api.php` because the real API is at `/cars/api/`.
### Deploying on a raw IP

`https://IP/...` needs a certificate with that IP as a SAN. Let's Encrypt does issue
for bare IPs now, but only from certbot >= 5.4 — older clients ask for a `dns-01`
challenge, which cannot be satisfied for an address, and fail with a message that
reads like a network problem:

```bash
certbot --version          # must be >= 5.4
sudo certbot certonly --standalone -d 163.245.214.125     # or --webroot with the ACME block
sudo certbot install --cert-name 163.245.214.125 --nginx  # writes cert_dir/ and renews
```

`cert_dir` in the deployment config has to match the resulting
`/etc/letsencrypt/live/<name>/`, or nginx starts and then fails every TLS handshake
with a missing-certificate error. The rendered config ships the ACME `location`
needed for `--webroot` renewal. For an HTTP-only deploy, make the TLS block
`listen 80;` only and drop the two `ssl_*` lines; the app is unaffected either way.

---

## 6A. Many clients on one server: one database and one folder each

This is the arrangement the server actually uses, and the only one (§6 is the same
renderer with one client). `163.245.214.125/<client>/`, each client with its own
database and its own `files/`, while `api/` and `dist/` are **shared**:

```
/var/www/api/                     one codebase
/var/www/dist/                    one build
/var/www/<client>/files/…         that client's uploads
```

What is deliberately **not** per client, and what each used to be:

| Used to be | Now | Why |
|---|---|---|
| `<client>/api/` | `/var/www/api/` | one release updates every client; a missed copy was how two clients ran different code |
| `<client>/` (the build) | `/var/www/dist/` | same reason, and a `--delete` here cannot reach uploads |
| `<client>_files/` (sibling) | `<client>/files/` | the client folder is now the whole client, so a backup or a listing needs no naming rule |
| `<client>/db_code.json` | the URL itself | nothing to keep in sync, and nothing to leak into a build |

Two consequences that are easy to get wrong:

- **`api/` and `dist/` being shared is a fact about the filesystem, not an option.**
  The provisioning library probes for it (`tenant_has_shared_api()`) and can be told
  outright with `shared_api: true`. If that probe is wrong every client silently gets
  a private API and the next deploy updates only one of them.
- **The client's name in the URL is its identity, and it is the database name.**
  Nothing is read from disk to decide which database a request means, so the same
  `api/` serves every client correctly and there is no file that can be stale,
  missing or copied.

### One-time root setup

```bash
# 1. The config. Start from the example and edit; it is the single source of truth.
sudo install -o root -g root -m 0640 deploy/cars-deploy.example.json /etc/cars-deploy.json
sudoedit /etc/cars-deploy.json     # template_database, webroot, php_socket, server_name,
                                   # cert_dir, registry credentials

# 2. The renderer, deliberately OUTSIDE the web root: root executes it.
sudo install -d -o root -g root -m 0755 /usr/local/lib/cars
sudo install -o root -g root -m 0644 deploy/render-nginx-multitenant.php /usr/local/lib/cars/
sudo install -o root -g root -m 0644 deploy/nginx-multitenant.conf.template /usr/local/lib/cars/

# 3. The wrapper the sudoers rule names.
sudo install -o root -g root -m 0755 deploy/cars-nginx-render /usr/local/bin/cars-nginx-render

# 4. The renderer reads the name validator from the app's own lib/. That file is
#    served from the web root, so it has to be root-owned and unwritable — see below.
sudo chown root:root /var/www/api/lib/tenant-provision.php
sudo chmod 0644 /var/www/api/lib/tenant-provision.php

# 5. The sudoers rule, then verify it: a malformed file breaks ALL sudo on the box.
sudo install -o root -g root -m 0440 deploy/sudoers-cars-nginx.example /etc/sudoers.d/cars-nginx
sudo visudo -c

# 6. Enable the generated config. Once, by hand — it is generated from then on.
sudo ln -s /etc/nginx/sites-available/cars-multitenant.conf /etc/nginx/sites-enabled/
sudo cars-nginx-render --check        # review; installs nothing
sudo cars-nginx-render                # install, `nginx -t`, reload
```

**Why step 4 is not optional.** `cars-nginx-render` runs as root and executes the
renderer, and the renderer `require`s `api/lib/tenant-provision.php`. That file lives
under `/var/www/api/` — the web root — so in any setup where the app is deployed by a
non-root user, or where `api/lib/` ended up group-writable, the web user can rewrite
a file that root then loads. That is a root shell, and no amount of hardening on the
renderer prevents it. The wrapper now checks the ownership of everything the renderer
reads (it asks the renderer which files, rather than keeping a second list that could
drift) and refuses to run otherwise, so a missed `chown` shows up as a refused
render rather than as a silent escalation. `deploy/deploy.sh` rsyncs `api/` as root
and leaves modes alone, so it preserves this — but if you ever `chown -R www-data
/var/www/api`, re-run step 4.

### The template database

Every client is built from one reference-only database: schema, the reference rows
(brands, roles, permissions, defaults, shipping lines), and the admin account — and
**no client data**. Build it once:

```bash
php api/lib/tenant-provision.php --build-template cars_template --seed-source <clean db>
```

It applies `setup.sql` (which already contains the reference INSERTs), applies the
migrations, and then verifies the result across all 38 business tables before
reporting `ok`. It refuses to run against an existing name without `--force`, because
the usual `--seed-source` is a development database that also holds client data, and
"rebuild the template" run against one of those is how one client's records become
another client's dropdown options. `--seed-source` defaults to the database
`api/config.php` names, which on this server is a live tenant — pass it explicitly.

### Adding a client

In the DB manager: **Provision** on the registry row. The dialog shows two lists
first — what the *server* is missing (one-time root steps) and what *this client* is
missing — and the Provision button stays disabled until the server list is clear,
rather than failing halfway with a permissions error. It creates the database, applies
the schema and migrations, copies the reference rows, sets the admin password to
`123`, clears any copied session token, and creates the `files/` upload folders.

It does **not** copy a build or write `db_code.json`: the folder name binds the
database and `dist/` is shared, so there is nothing per client to write. The run
reports whether the shared build is present instead of writing one — a client that is
otherwise complete and 404s is now a legible state rather than a silent one.

**Re-deploy build** on the same dialog copies the current build into the *shared*
`dist/`, so it updates every client on the server. It is not gated on a selected row,
because there is no per-client build to deploy.

**Apply nginx** is a separate button on purpose. It is the only step that needs root,
and the only one whose effects reach past this client, so it is never fired as a side
effect of setting a client up. It writes a new
`/etc/nginx/sites-available/cars-multitenant.conf`, runs `nginx -t`, restores the
previous file if the test fails, and reloads. The client list is generated from the
`dbs` table, so **the config is only as current as the last render** — add or delete a
client and the URL does not change until you re-run it. Deleting a client without
re-rendering leaves their data reachable at a URL nobody remembers owning.

`--check` prints exactly what would be installed, which is how a change gets reviewed
before it reaches `/etc`.

---

## 7. php.ini upload block (NOT in FPM pool)

Append to `/etc/php/8.5/fpm/php.ini` (NOT www.conf — `php_value[]` fails on PHP 8.5
with `value is NULL for a ZEND_INI_PARSER_ENTRY`):

```
upload_max_filesize = 100M
post_max_size = 110M
memory_limit = 512M
max_execution_time = 600
max_input_time = 600
```

---

## 8. Known gotchas learned in this deploy

1. `insert_user` in `api.php` required **count($params) === 9** but the form sends **7**
   → HTTP 400 "Invalid parameters". Fixed to `=== 7` (password still param[2]).
2. Never merge `setup.sql` verbatim into MariaDB (see §4.3).
3. `execute_sql` / `executeQuery` only allow SELECT/INSERT/UPDATE/DELETE; DESCRIBE →
   "Invalid query type".
4. Deep links work because the build's location falls back to
   `/<client>/index.html`; keep the plain-prefix build block and the regex API block
   separate. Merging them under `^~` breaks PHP in a way that returns HTML for a POST
   rather than an error (§6).
5. **Never hardcode a hostname or client name in the app.** Five copies of an API-URL
   helper had drifted (`useApi`, `GeneralSettingsForm`, and three db-manager
   components); some omitted the mount point entirely, producing `https://host/api`
   which 404s under any client, and all treated any `192.168.*` host as a dev box,
   pointing at `:8000` where a deployed server has nothing listening. They all now
   call `resolveApiBaseUrl()`. Add new callers to that function, never a new copy —
   and run the `grep` in §5 after building.
6. **Never let a client name be normalised.** `acme-motors` → `acme_motors` is the
   kind of convenience that produces a folder and a database that do not share a name,
   so a rename, a backup or a support question cannot connect them without knowing the
   rule. Refuse the name instead (see §4.4); `app_valid_tenant()` and
   `tenant_assert_valid_db_name()` enforce the same set for the same reason.
7. The portability gate strips `//` and `/* */` comments before scanning `dist/`. Both
   forms of comment contain the example hosts the source explains in prose, so a
   literal scan fails every legitimate build. When editing the gate, note that a naive
   `s://.*::` also eats the `//` in `https://` and hides real URLs — hence the
   `[^:]` guard.
  8. **Never exclude credentials by naming one file.** `--exclude config.local.php` is
     not enough: `db_manager_config.local.php` also holds this machine's DB credentials,
     and a deploy shipped it to production, breaking the db-manager with "Access denied
     for user 'root'". Always `--exclude '*.local.php'`, and `deploy/deploy.sh` asserts
     via `rsync --dry-run` that no `.local.php` is in the transfer list.
  9. **A route can collide with a real directory on the server.** The route `cars` is
     `/cars`, and `/cars/cars/` also exists as a directory holding `logo.png` and
     `gml2.png`. With `$uri/` in `try_files`, nginx matched the directory, found no
     `index.html` inside it, and returned **403 Forbidden** instead of the app. It
     looked intermittent because client-side navigation never requests the URL — only a
     refresh or a pasted link reaches `try_files`. `$uri/` is therefore omitted; see §6.
  10. **A missing module-scope binding does not fail the build.** `isLocalhost` was
     deleted during a refactor while four call sites still referenced it. The module
     parsed and imported fine; it only threw `ReferenceError: isLocalhost is not defined`
     once a request actually ran, so every API call, the cookie verification, the
     version check and the alerts failed while the page still returned 200. A
     declaration-only test would not have caught it — `src/composables/useApi.spec.js`
     drives `callApi()` and reproduces the exact error.
  11. `rsync` prints **nothing** in dry-run without `-v`. A `--dry-run` built on plain
     `rsync -n` reported "(no changes)" for a 132-file deploy, which is worse than no
     dry-run at all. Always pass `-v` (and an `--out-format`) when previewing.
  12. `deploy.sh` writes to a **live server**. Always run `--dry-run` first. It now uses
     `rsync --delete`, which is safe *only* because the destinations are `dist/` and
     `api/` and nothing else: the per-client rsync it replaced had to exclude `files/`
     from `--delete`, and a typo in that exclude list deleted a client's invoices.
     Never point either rsync at a client folder.
 13. **`alias` is not `root` for the shared build.** With `root`, `/<client>/sell-bills/index.<hash>.js`
     resolves against the document path, the file is not there (the build is not under
     `/<client>/` on disk), and every asset gets the HTML shell instead — a blank tab
     with no error. The same applies to uploads in subdomain form, where the path is
     `/files/…` but the folder is `<webroot>/<client>/files/`.
 14. **`CARDS_TENANT` beats everything, and everything else fails closed.** nginx sets
     it per location block; `api/lib/appdb.php` trusts it first, then the first path
     segment, then the host label *compared against the configured `base_domain`*. The
     hostname is chosen by whoever sent the request, so a request to an unrelated host
     must resolve to no client rather than to the first label it happens to have. That
     is why `/api/api.php` at the webroot root is 404 and not a live tenant.

---

## 9. Rollback / safety

- Vhost backup: `/etc/nginx/sites-available/default.bak.<timestamp>` (made before each change).
- Per-client isolation: each client gets its own database and its own `files/`, and
  nginx maps `/<client>/…` at that one client only. `api/`, `dist/` and
  `api/config.local.php` are per server and shared by design — which means a release
  is all-or-nothing across clients, and rolling one back rolls back all of them.
- Full DB dumps: `mysqldump merhab_cars` → `api/backups/`.
- `git push origin feature-from-e9dcfaf` → `github.com/ekotobi/cars_vue_js.git`.

> **Unresolved security findings.** `api/api.php` executes caller-supplied SQL with no
> authentication and no permission check, under `Access-Control-Allow-Origin: *`, and the
> client-side `is_admin` flag bypasses every permission check in that file. See
> [`SECURITY.md`](SECURITY.md) for the details, affected line numbers and remediation
> order. Nothing there has been fixed yet.

## 10. Several clients on the local dev server

Locally one Vite server serves **every** provisioned client from the same build and the
same `api/`, each on its own database:

| URL | Database | Uploads |
| --- | --- | --- |
| `http://localhost:5173/merhab_cars/` | `merhab_cars` | `merhab_cars/files/` |
| `http://localhost:5173/<other-client>/` | `<other-client>` | `<other-client>/files/` |

Every client has HMR, because there is one build and one `api/` — the same arrangement
production uses. What Vite cannot do on its own is route a client prefix, so:

- `vite.config.js` discovers the client mounts by scanning for directories that contain
  `files/` — the condition `tenant-provision.php` satisfies, so a newly provisioned
  client is covered without editing a list. Each mount serves the **shared** `dist/`
  for every path under it (so client-side routes survive a refresh), injects
  `<base href="/<mount>/">` so the hashed assets resolve, and proxies `/<mount>/api` to
  `localhost:8000` **with the prefix intact** — stripping it is what makes a client
  talk to the wrong database.
- The root `/` answers **404**, never a redirect or an index. Production's rendered
  config has `location / { return 404; }` for the same reason: a 302 to a client's
  login page is a login form that cannot authenticate.
- The database is not configured in the app. `api/lib/appdb.php` resolves the client
  from `CARDS_TENANT` (set by the proxy), else the first path segment, and connects to
  the database of that name. Nothing is read from the client's folder, so there is no
  `db_code.json` for two clients to disagree about.
- `<client>/` folders are git-ignored and hold only `files/`.

Working on it:

```bash
# One-time: create the client database and its files/ folder (idempotent).
npm run mig27:setup

# src/ changes need nothing: every client is served from the one build.
# vite.config.js changes are not hot-reloaded:
npm run dev            # restart
```

`npm run mig27:setup -- --force` drops and rebuilds the client from `api/setup.sql` plus
the forward migrations in `api/migrations/`. It seeds only the rows a client cannot
start without (admin account, roles, permissions, lookup rows), never business data.
One migration, `031_login_api_token.sql`, targets `merhab_databases` rather than the
client, and is applied there instead.

Verifying two clients are really separate:

```bash
curl -s http://localhost:5173/<client>/ | grep '<base href="/<client>/">'
curl -s http://localhost:5173/<client>/api/db_manager_api.php \
  '?action=get_database_by_code&db_code=db_93036eb23669c0fd4c27'
# → {"db_name":"<client>","files_dir":"/<client>/files","js_dir":"/<client>"}

# Log in through each; the tokens must differ, and each lands in its own users table.
curl -s -X POST http://localhost:5173/<client>/api/api.php -H 'Content-Type: application/json' \
  -d '{"action":"login","username":"admin","password":"123"}'
```

Gotchas that cost time here, all of which apply to production too:

13. **Never rebuild a path from `files_dir`.** The registry records it so the DB
    manager and the renderer can agree, but the path comes from the client name and
    `app_db_files_dir()`, which returns `files` — a name **relative to the client
    folder**. Reconstructing it against the wrong base silently looks in
    `<root>/<client>/<client>/files` and every upload fails with "Invalid upload
    directory". `app_deployment_root()` is the one place that knows the base.
14. **A PHP variable read inside a function is not the global one.** `app_db_pdo()`
    read `$db_config` without `global`, so host/user/pass arrived as `null` and PDO
    failed with `Access denied for user ''@'localhost'` — an empty 500 from every
    endpoint that authenticates through `appdb.php` (`upload.php`, `backup.php`),
    on every deployment. The same applies to config files required
    *inside* a function: `require` there assigns into the function's scope, so
    `db_manager_config.php` is loaded at file scope instead.
15. **`__DIR__` is the file's own directory, so depth matters.** In `api/lib/appdb.php`
    the app folder is `dirname(__DIR__, 2)`; in `api/upload.php` it is `dirname(__DIR__)`.

## Supplier credibility: what it can and cannot tell you

> **Currently switched off, on both layers.** `CREDIBILITY_ENABLED` in
> `src/lib/featureFlags.js` is `false`, so the column, the row-menu entry, the modal
> and the background fetch are all hidden; `credibility_enabled` in
> `api/config.local.php` is `'0'`, so the endpoints refuse as well. Nothing was
> deleted: the component, composable, endpoint, styles and translations are all still
> in place and tested. See "Switching the credibility feature off" above for how to
> turn each layer back on. The UI flag is a compile-time constant on purpose - see the
> file for why - so it needs a rebuild and redeploy like any other change.

The check sends the model **the supplier's name and nothing else**, then asks two
questions: is this supplier credible, and does it have any court cases.

Answering the second question honestly is the whole difficulty. A language model
has no internet access, no company register and no court register, so asking it
to "do a full search" does not produce a search - it produces a confident,
specific, fabricated lawsuit against a real company. On a screen someone uses to
decide whether to send cars and money, that is the worst output this feature
could produce.

So the prompt requires the model to:

- never imply it looked anything up, and never say it did a search;
- reason from the name itself, which is real analysis (country, legal form,
  whether it names a legal entity or only a trading label);
- use its own memory only for companies it genuinely knows, and label that as
  recollection - the modal then marks the court answer as unverified;
- answer "I cannot check this and do not recall anything" when unsure, which
  counts as a complete answer.

Observed behaviour, verified against the live model:

| Supplier | Question 2 answer |
|---|---|
| Weifang Century Sovereign Automobile Sales Co., Ltd | "I have no knowledge of any court cases involving this company, and I cannot verify its legal history." |
| ChongQin huanyu | "I have no information about any court cases ... cannot recall any from my training data." |
| Toyota Motor Corporation | Answers from memory, labelled as recollection: recalls related to vehicle safety recalls and regulatory settlements. |

If real registry or litigation data is ever required, it has to come from a
provider that sells it, wired in as its own action with a citation. It cannot be
made reliable by a longer prompt here.

Existing databases need the two answer columns:

```bash
mysql -u USER -p DBNAME < api/migrations/024_court_records_answer.sql
```

It is idempotent, so re-running it on a server that already has the columns is a
no-op that prints `Column ... already exists`.

Until it is applied, all three credibility endpoints return
`db_schema_outdated`, rather than a PHP fatal on `Unknown column
'court_records'`. That check is an INFORMATION_SCHEMA lookup
(`supplier_credibility_require_schema()`), not a try/catch: the two read
endpoints never select these columns, so on a server missing 024 they would
answer `200` and the feature would look healthy until someone pressed Assess.
Verified by dropping the column and calling all three.

## Car names were not attached to any brand

`cars_names.id_brand` was NULL for all 35 default car names, so Car Models
(`src/views/CarNamesView.vue`) rendered a blank Brand on every row and nothing
could group or filter models by marque. Adding a car was unaffected, because
`CarStockForm.vue` loads names with a plain `SELECT id, car_name` and never joins
or filters on brand.

The brands were worked out from the model rather than from the name, because the
names are unreliable: `TIGO 3` is a Chery Tiggo, `CHERY TIGGO 3X` spells the
marque wrong, and the whole VW group is prefixed `T-` without ever saying VW.

| Brand | Count | Names |
|---|---|---|
| VW | 10 | GOLF 300TSI R-LINE, GOLF R-LINE (FULL OPTION), T-CROSS, T-ROC ×2, TIGUAN L, TACOUA, THARU ×3 |
| CHERY | 8 | CHERY TIGO 7, CHERY TIGGO 3X, TIGO 3, COOLRAY ×5 |
| KIA | 4 | K3, KX1…, SELTOS LUXURY BLACK ROOF, SONET BLACK ROOF |
| MG | 2 | MG5 BASE AUTO, MG5 MAN |
| GEELY | 2 | EMGRAND ×2 |
| LIVAN | 2 | LIVAN AUTO, LIVAN MAN |
| JETTA | 1 | JETTA VS5 |
| JETOUR | 1 | DASHING PRO 1.6 DCT |
| AUDI / CHANGAN / FREIGHT / SKODA / PEUGEOT | 1 each | A3, CHANGAN CS75 PLUS, FREIGHT, KAMIQ GT, 2008 |

Three of these started out as guesses and have since been confirmed by the
operator, so the earlier "worth checking" caveat no longer applies:

- `TACOUA` and the three `THARU`s match no model that could be identified from
  the name. They were provisionally placed with VW because they cluster with the
  VW `T-` names and nothing else claims that prefix. **Confirmed: VW.**
- `DASHING PRO 1.6 DCT` could not be placed at all, and GEELY was a guess on
  lineage. **Confirmed: JETOUR**, which is what it now maps to.

`LIVAN AUTO` and `LIVAN MAN` are LIVAN cars, but there was no LIVAN row in
`brands`, so the migration creates one (`INSERT IGNORE`, brand is UNIQUE, id is
AUTO_INCREMENT, so it is a no-op where LIVAN already exists). That has to happen
*before* the update: the update joins `cars_names` to `brands` on the brand name,
so with no LIVAN row both names would have joined against nothing and stayed NULL
- silently, with no error to notice.

Two brand names were misspelled, and migration 026 renames both. They are
renames rather than new rows, so ids 3 and 6 are untouched and everything already
pointing at them keeps working:

- `CHERRY` → **CHERY**, the actual marque.
- `JETA` → **JETTA**, a brand in its own right sitting next to the unrelated real
  marque `JETOUR` (id 7).

The second one also moves a car. `JETTA VS5` was mapped to VW because Jetta is a
VW model and `JETA` looked like a mangled `Jetta` — but `JETA` was a mangled
`JETTA`, a separate brand, so the car belongs to JETTA and VW drops from 11 names
to 10. That is why migration 025 no longer lists it.

Each rename is guarded on the corrected name not already existing, because
`brands.brand` is UNIQUE and an unguarded UPDATE would abort with a duplicate-key
error and leave the second rename unapplied. Where both spellings are present the
migration skips rather than merging two marques; that needs a human.

Two car names are also misspelled, and unlike a brand these are referenced only
by id — `buy_details.id_car_name` and `car_name_media.car_name_id` — so correcting
them renames a row without touching anything else. Migration 027 fixes both:

- `CHERRY TIGO 7` → `CHERY TIGO 7`
- `SELTOS LUXERY BLACK ROOF` → `SELTOS LUXURY BLACK ROOF`

So migration 025 matches **both** spellings of these two. It previously matched
only the misspelled spelling on purpose, which meant the corrected spelling had
to stay wrong; and it would have been a silent no-op against the corrected
spelling, since an equality test that matches no rows raises nothing. Matching
both is what lets 025 and 027 run in either order.

Existing databases need both, in either order:

```bash
mysql -u USER -p DBNAME < api/migrations/025_car_names_brand.sql
mysql -u USER -p DBNAME < api/migrations/026_brand_renames.sql
mysql -u USER -p DBNAME < api/migrations/027_car_name_typos.sql
```

Order does not matter, and that is deliberate: 025 joins on
`IN ('CHERY','CHERRY')` and `IN ('JETTA','JETA')`, and matches both spellings of
the two misspelled car names, so it is correct whether or not the renames have
happened yet. A single-value join would match nothing on a renamed database and
leave the names NULL with no error — the same applies to a single-spelling
`IN` list for the car names. 026 and 027 are each guarded on the corrected value
not already existing, so they are safe to re-run and they no-op once applied.

Verified by running both orders against the same database and getting identical
results, and by re-running each several times: 13 brands, 35 names, none NULL.

### Migration 028: car-name delete guards in the database

CarModelsView refuses to delete a brand or car name that is still referenced, and
lists what is referencing it. That check runs in the browser, so it is a
convenience rather than a guarantee: sending `DELETE FROM cars_names WHERE id = ?`
straight to the API skips it, because `api/api.php` executes caller-supplied SQL
without authenticating the caller (see [SECURITY.md](SECURITY.md)).

Verified against a scratch copy with the browser check bypassed. A car name with
one `buy_details` row and two `car_name_media` rows deleted with no error at all:

```
DELETE FROM cars_names WHERE id = 4  ->  succeeded, no error
buy_details rows still pointing at id 4 afterwards: 1
car_name_media rows for id 4: 0   (cascaded away, files left on disk)
```

Migration 028 puts both guarantees in the schema:

- `buy_details.id_car_name` gains `ON DELETE RESTRICT`. It had no foreign key,
  which is why the delete above left a dangling row.
- `car_name_media.car_name_id` changes from `ON DELETE CASCADE` to `RESTRICT`, so
  deleting a car name can no longer destroy media rows and strand the files.

`RESTRICT` rather than `CASCADE` on purpose. Cascading destroys purchase history,
and the files on disk survive the rows either way, so the right answer is to
refuse the delete and let a human decide.

```bash
mysql -u USER -p DBNAME < api/migrations/028_car_name_delete_guards.sql
```

Only existing databases need this. `api/setup.sql` already declares both
constraints, so a fresh install gets them without running 028; running it against
a fresh install anyway is a no-op.

`api/setup.sql` builds `buy_details` before `cars_names`, and MySQL rejects a
foreign key whose referenced table does not exist yet, so the `buy_details`
constraint is added as an `ALTER TABLE` immediately after `cars_names` is created.
That ordering is the likely reason the column was never constrained in the first
place.

Before adding the foreign key it clears `buy_details.id_car_name` values that
point at a deleted car name, keeping the purchase row and dropping only a link
that was already dangling. The count of affected rows is printed by the migration
rather than done silently; on the database this was written for it was zero. The
migration reports the two constraints it owns when it finishes, discovers any
existing constraint from `information_schema` rather than dropping it by a
hardcoded name, and is safe to re-run.

`cars_names.id_brand` is still unconstrained, so deleting a brand remains a
browser-only check. Adding that foreign key needs a decision first: brands are
referenced by `cars_names`, and unlike car names there is no obvious case for
cascading a brand delete through every model under it.

A car name that nothing references still deletes normally, which is the point —
the constraint blocks the unsafe case only.
