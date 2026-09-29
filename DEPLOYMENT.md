# Cars Management System — Deployment Runbook

Battle-tested guide from the live deploy to **https://world-automobile.com/cars**
(VPS `163.245.214.125`, domain `world-automobile.com`). Follow exactly; every command
was executed and verified against production.

> **Secrets live ONLY in server files.** DB passwords are in `api/config.local.php`,
> which is git-ignored and never committed. `api/config.php` is a tracked loader
> that resolves credentials from `config.local.php`, falling back to the
> `DB_HOST` / `DB_USER` / `DB_PASS` / `DB_NAME` environment variables.
> (`merhab_root` — MariaDB user, hosts `localhost` + `127.0.0.1`). App + db-manager
> admin login is `admin / 123`, stored as bcrypt in the DB. `db_code.json` →
> `db_9a7f4e0b2fa8134e0ea0`. Do NOT paste credentials into chats/docs.

---

## 1. Architecture

```
nginx :443 (world-automobile.com)          ← or a raw IP
  root /var/www/world-automobile.com
  └─ /<FOLDER>/  → Vue 3 SPA (dist build, SPA fallback to /<FOLDER>/index.html)
      └─ /<FOLDER>/api/ → PHP (api.php, db_manager_api.php, invitations.php)
```

`<FOLDER>` is supplied per client and may be any name. The app discovers it at
runtime, so one `dist/` serves every deployment:

  - **Mount point** — `src/utils/basePath.js` reads `import.meta.url`, not
    `location.pathname` (see §8.4); Vite `base: './'`. The served HTML gets a
    `<base href="/<FOLDER>/">` injected by nginx (§6.1) — the build ships none,
    because the folder is per-client.
- **API base** — `resolveApiBaseUrl()` in the same file: `<origin><mount>api`, so
  `https://host/<folder>/api`. Same rule in `index.html` for pre-boot fetches.
- **Router** — `createWebHistory(getBasePath())`, so deep links survive reloads.
- **Only nginx is folder-specific** — generate it, don't hand-edit (§6).
- **DB-manager** uses `api/db_manager_api.php` → DB `merhab_databases`.
- **3 databases**: `merhab_cars` (app), `merhab_databases` (manager), `merhab_invitations`.

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
GRANT ALL PRIVILEGES ON merhab_invitations.* TO 'merhab_root'@'localhost';
GRANT ALL PRIVILEGES ON merhab_invitations.* TO 'merhab_root'@'127.0.0.1';
FLUSH PRIVILEGES;
```

### 4.2 Create DBs

```sql
CREATE DATABASE IF NOT EXISTS merhab_cars        DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS merhab_databases   DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
CREATE DATABASE IF NOT EXISTS merhab_invitations DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci;
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
- `dbs` MUST match the deployed `db_code.json`:
  ```sql
  INSERT INTO dbs (db_code, db_name, files_dir, js_dir, is_created)
  VALUES ('db_9a7f4e0b2fa8134e0ea0', 'merhab_cars', 'files', '', 1);
  ```
  (`db_code.json` on server = `{"db_code":"db_9a7f4e0b2fa8134e0ea0"}` — keep in sync!)

### 4.5 Invitations DB — just create, no schema needed.

---

## 5. Build + deploy

One command, per client. It builds, refuses a non-portable build, uploads, writes
the per-server files, and verifies the result:

  ```bash
  ./deploy/deploy.sh <folder> <host> <db_code> [user@server] [webroot]
  # e.g.
  ./deploy/deploy.sh cars world-automobile.com db_9a7f4e0b2fa8134e0ea0

  # ALWAYS preview first. Builds and verifies, lists every file that would be
  # transferred, asserts no *.local.php is in the list, and writes nothing.
  ./deploy/deploy.sh cars world-automobile.com db_9a7f4e0b2fa8134e0ea0 --dry-run
  ```
  
  `db_code` is **this client's** database. Everything else — the `dist/` build — is
  shared between clients.
  
  `--dry-run` matters because this script writes to a live server. A deployment once
  went out without being intended, and once shipped `db_manager_config.local.php`,
  breaking the db-manager with "Access denied for user 'root'". The flag may appear
  anywhere in the arguments and is never mistaken for a positional.

<details>
<summary>Manual equivalent</summary>

```bash
npm ci && npm run build
grep -rE 'world-automobile|localhost:8000' dist --include='*.js'   # must print nothing

TARGET=/var/www/world-automobile.com/cars
# --delete is deliberately NOT used. It would delete anything else in the folder
# (client uploads, per-server config) that is not in dist/. Old hashed assets are
# harmless; a wiped folder is not. Prune them deliberately instead:
#   ssh $HOST "cd $TARGET && ls -dt index.*.js 2>/dev/null | tail -n +4 | xargs -r rm -f"
  # db_code.json and EVERY *.local.php are PER SERVER and must be excluded, or this
  # machine's database and credentials overwrite the client's. Do not enumerate the
  # credential files by name -- a new one will be missed, and that is exactly how
  # db_manager_config.local.php was shipped once.
  rsync -az --exclude 'db_code.json'      -e "ssh -i ~/.ssh/cars_deploy" dist/ root@$HOST:$TARGET/
  rsync -az --exclude '*.local.php'       -e "ssh -i ~/.ssh/cars_deploy" api/  root@$HOST:$TARGET/api/
printf '{ "db_code": "%s" }\n' "$DB_CODE" | ssh -i ~/.ssh/cars_deploy root@$HOST \
  "cat > $TARGET/db_code.json && chown www-data:www-data $TARGET/db_code.json"
```

</details>

### Per-server vs per-build

| File | Scope | Why |
|---|---|---|
| `dist/` (hashed assets) | **shared** | folder- and host-agnostic by construction |
| `<folder>/db_code.json` | **per server** | names the DB this deployment talks to |
| `api/config.local.php` | **per server** | DB password; git-ignored, never deployed |
| nginx vhost | **per server** | the only place the folder name appears |

`db_code.json` is fetched at runtime from `<mount>db_code.json`, so it is *not*
compiled in. The `remove-db-code-json` Vite plugin deletes it from `dist/` on every
build — the copy in `public/` is a local `npm run dev` default only. Without that, a
stale `db_code` would point a new client at this machine's database. A `404` on
`db_code.json` now reports the fix in the error text instead of a bare status.

`deploy/deploy.sh` enforces this: it fails the deploy if `dist/db_code.json` exists
or if the build's *code* (comments stripped) contains a hardcoded host, a `:8000`
dev port, a `192.168.*` guess, or an absolute URL containing the folder name.

`db_code.json` must match the `dbs` row in the **manager** DB (see §4.4). It is
resolved by PHP from `api/../<js_dir>/db_code.json`; an empty `js_dir` means the app
root, which is what the app fetches from and what the deploy guide seeds.

Writable dirs (created, owned by www-data):
`files/` (subdirs buy_pi, sell_pi, documents, ids, payments_swift, banks_logos,
letter_head, logo, uploads, chat_files), `backups/`, `mig_files/`.

> `api/upload.php` default base dir is `mig_files`.

---

## 6. nginx app block (PHP MUST be nested — key fix!)

**The folder name is supplied per client and is not baked into the build.** The app
resolves its own mount point at runtime, so a single `dist/` works under any folder.
Only nginx needs to know the name — and it is generated, never hand-edited:

```bash
./deploy/render-nginx.sh <folder> <domain_or_ip> [webroot] [php_sock] --write
# e.g. ./deploy/render-nginx.sh cars world-automobile.com /var/www/world-automobile.com
# writes deploy/out/nginx-cars.conf; review, then install:
cp deploy/out/nginx-cars.conf /etc/nginx/sites-available/default
nginx -t && systemctl reload nginx
```

The script validates the folder name against a strict allowlist
(`^[A-Za-z0-9_][A-Za-z0-9._-]*$`) and rejects anything that could terminate the token
and inject a directive. This matters: the hardening rules below are the only thing
between a public client and `install.php` / `drop_all_tables.sql`.

Rendered output for `cars` (identical to the hand-written original, plus the `.sql` rule):

```nginx
location = /cars { return 301 /cars/; }        # no trailing slash -> blank page

location = /cars/api/install.php        { deny all; }
location = /cars/api/drop_all_tables.sql { deny all; }

  location ^~ /cars/ {
      # $uri/ is deliberately absent — see the note after this block.
      try_files $uri /cars/index.html;

      location = /cars/index.html {              # inject <base>, see §6.1
          sub_filter_once on;
          sub_filter_types text/html;
          sub_filter '<meta charset="UTF-8">' '<base href="/cars/"><meta charset="UTF-8">';
      }

      location ~* \.sql$ { deny all; }           # schema dumps — MUST be nested
    location ~ \.php$ {                        # nested — evaluated inside ^~
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.5-fpm.sock;
        fastcgi_send_timeout 600;
        fastcgi_read_timeout 600;
    }

    location ~ /\. { deny all; }               # hide .htaccess etc.
}
```

  Why the structure is what it is:
  
  - **`$uri/` is not in `try_files`.** Route paths can collide with real directories.
    The route `cars` is `/cars`, and `/cars/cars/` is a real directory on the server
    holding `logo.png` and `gml2.png`. With `$uri/`, nginx matched that directory, found
    no `index.html` inside it, and returned **403 Forbidden** instead of the app. It
    looks intermittent because client-side navigation never issues a request for the
    URL — only a refresh or a pasted link reaches `try_files`. Dropping `$uri/` means
    a path that is a directory but not a route falls through to the SPA and the router
    decides. Static assets are matched by the first `$uri` test and are unaffected.
  - **PHP nested inside `^~`.** A bare `location ^~ /cars/ { try_files … }` **swallows
    `.php`** and serves the source as text (405 on POST). This is the failure that cost
    the original deploy.
- **The `.sql` deny must be nested too.** `^~` takes precedence over a regex location,
  so a `~* \.sql$` rule declared at *server* level is never evaluated and the files are
  served anyway. This was done by mistake on production: the rule was in the vhost and
  all three dumps still returned **200**. Deny rules that appear correct and do nothing
  are worse than none — the `.sql` rule sits inside the `^~` block.
- **`location = /cars` 301s to `/cars/`.** Without the trailing slash the relative
  base resolves to `/` and every asset 404s (blank page).
- **What `.sql` blocks.** `export.sql`, `setup.sql`, `recreate_selection_tables.sql`
  and `migrations/*.sql` were served as plain text — the full DB schema, public. No
  feature regresses: `db_manager_api.php` reads `setup.sql` from disk via `__DIR__`,
  and backups are generated and served by `backup.php`.

After changing the vhost, **verify rather than assume** — a deny rule in the file does
not mean the request is denied:

  ```bash
  for f in api/install.php api/setup.sql api/export.sql; do
    printf '%-24s %s\n' "$f" "$(curl -s -o /dev/null -w '%{http_code}' https://host/cars/$f)"
  done   # every one must be 403/404
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
  
  nginx injects the base when serving the shell, so one `dist/` still serves any folder:
  
  ```nginx
  location = /cars/index.html {
      sub_filter_once on;
      sub_filter_types text/html;
      sub_filter '<meta charset="UTF-8">' '<base href="/cars/"><meta charset="UTF-8">';
  }
  ```
  
  Needs `ngx_http_sub_module` (stock Ubuntu/Debian: present). The anchor is
  `<meta charset>`, present in every build; `sub_filter_once` keeps it to the first
  match so comment prose about `<base>` is untouched.
  
  **Consequence:** `index.html` is now load-bearing for this fix. Removing the
  `<meta charset>` anchor, or serving the app from a vhost without this block, breaks
  every deep link. If you ever see the MIME error, check for `<base>` in the served
  HTML first:
  
  ```bash
  curl -s https://host/cars/sell-bills/5 | grep -o '<base[^>]*>'   # must print one
  curl -s -o /dev/null -w '%{http_code} %{content_type}\n' \
    https://host/cars/index.<hash>.js                                  # must be application/javascript
  ```
  
  This also fixes the `index.html` bootstrap (`getApiBase()` resolves `./api` against
  the document). Without `<base>`, `/cars/cars/` produced a 404 for
  `/cars/cars/api/db_manager_api.php` because the real API is at `/cars/api/`.
  
### Deploying on a raw IP

`https://IP/...` needs a certificate with that IP as a SAN — Let's Encrypt does not
issue for bare IPs, and the browser will warn otherwise. For an HTTP-only IP deploy,
edit the rendered file: make the TLS block `listen 80;` only and drop the two
`ssl_*` lines. The app itself is unaffected.

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
4. Deep links work because nginx falls back to `/<folder>/index.html`; keep the `^~`
   + nested `\.php$` structure or PHP breaks.
5. **Never hardcode a hostname or folder in the app.** Five copies of an API-URL
   helper had drifted (`useApi`, `GeneralSettingsForm`, and three db-manager
   components); some omitted the mount point entirely, producing `https://host/api`
   which 404s under any folder, and all treated any `192.168.*` host as a dev box,
   pointing at `:8000` where a deployed server has nothing listening. They all now
   call `resolveApiBaseUrl()`. Add new callers to that function, never a new copy —
   and run the `grep` in §5 after building.
6. `db_code.json` is **per server**, never part of the build (see §5). It must match
   the `dbs` row in the manager DB. An empty `js_dir` means the app root; the
   db-manager's read/write actions used to reject an empty `js_dir` as "not
   configured", which made the file uneditable in production — `resolveDbCodeJsonPath()`
   in `db_manager_api.php` now treats empty as app root and still refuses any
   `js_dir` that escapes it.
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
  12. `deploy.sh` writes to a **live server**. Always run `--dry-run` first, and never
     `rsync --delete`: the target folder holds client uploads and per-server config
     that are not in `dist/`.

---

## 9. Rollback / safety

- Vhost backup: `/etc/nginx/sites-available/default.bak.<timestamp>` (made before each change).
- Per-client isolation: each client gets its own folder, its own nginx config, its own
  `db_code.json` and its own DB credentials in `api/config.local.php`. The `dist/`
  build is shared; nothing else is.
- Full DB dumps: `mysqldump merhab_cars` → `api/backups/`.
- `git push origin feature-from-e9dcfaf` → `github.com/ekotobi/cars_vue_js.git`.
