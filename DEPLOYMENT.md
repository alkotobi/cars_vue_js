# Cars Management System — Deployment Runbook

Battle-tested guide from the live deploy to **https://world-automobile.com/cars**
(VPS `163.245.214.125`, domain `world-automobile.com`). Follow exactly; every command
was executed and verified against production.

> **Secrets live ONLY in server files.** DB passwords are in `api/config.php`
> (`merhab_root` — MariaDB user, hosts `localhost` + `127.0.0.1`). App + db-manager
> admin login is `admin / 123`, stored as bcrypt in the DB. `db_code.json` →
> `db_9a7f4e0b2fa8134e0ea0`. Do NOT paste credentials into chats/docs.

---

## 1. Architecture

```
nginx :443 (world-automobile.com)
  root /var/www/world-automobile.com
  ├─ /cars/  → Vue 3 SPA (dist build, SPA fallback to /cars/index.html)
  │   └─ /cars/api/ → PHP (api.php, db_manager_api.php, invitations.php)
  └─ fallback → /cars/index.html for deep links
```

- **SPA** (Vue 3 + Vite): API base `https://world-automobile.com/cars/api`
  (set in `src/composables/useApi.js`, `index.html`, `src/components/**`,
  `src/views/**` — old `www.merhab.com/api` kept as comments).
- **DB-manager** uses `api/db_manager_api.php` → DB `merhab_databases`.
- **3 databases**: `merhab_cars` (app), `merhab_databases` (manager), `merhab_invitations`.

---

## 2. SSH access + deployment key

```bash
alias deployssh="ssh -i ~/.ssh/cars_deploy root@163.245.214.125"
```

Pubkey (`cars_deploy.pub`) is in `root@…:~/.ssh/authorized_keys`. Rotate the keypair
on this machine; on the server just replace the line. Always use `www-data` for
`api/config.php` owner so PHP-FPM can read it: `chown www-data:www-data api/config.php`.

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
CREATE USER 'merhab_root'@'localhost' IDENTIFIED BY '<from api/config.php>';
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

`api/setup.sql` is MySQL-8 oriented. For MariaDB apply ALL of these or you get
`ERROR 150` / collation errors:

1. **99 that9 that0 replace:** `sed -i 's/utf8mb4_0900_ai_ci/utf8mb4_general_ci/g' setup.sql`
2. **FK ordering:** MariaDB builds FKs inline and referencing tables must come after
   referenced ones. Reorder CREATE TABLEs **topologically by FK** (no forward refs).
   A python `order_schema.py` produced the working file — 55 tables, 0 FK errors.
3. **`car_apgrades.id_upgrade` sign mismatch:** setup.sql declares `int`, FK points at
   `upgrades.id` which is `int unsigned` → errno 150.
   ```sql
   ALTER TABLE merhab_cars.car_apgrades MODIFY id_upgrade INTEGER UNSIGNED NOT NULL;
   ```

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

```bash
cd <repo>
npm ci
npm run build          # API base already points at /cars/api (prod)
# dist/assets must reference world-automobile.com/cars/api

TARGET=/var/www/world-automobile.com/cars
rsync -az --delete -e "ssh -i ~/.ssh/cars_deploy" dist/  root@163.245.214.125:$TARGET/
rsync -az        -e "ssh -i ~/.ssh/cars_deploy" api/    root@163.245.214.125:$TARGET/api/
```

Writable dirs (created, owned by www-data):
`files/` (subdirs buy_pi, sell_pi, documents, ids, payments_swift, banks_logos,
letter_head, logo, uploads, chat_files), `backups/`, `mig_files/`.

> `api/upload.php` default base dir is `mig_files`.

---

## 6. nginx `/cars/` block (PHP MUST be nested — key fix!)

The naive `location ^~ /cars/ { try_files … /cars/index.html; }` **swallows .php**
(serves source as text, 405 on POST). The working vhost uses nested regex locations:

```nginx
location ^~ /cars/ {
    try_files $uri $uri/ /cars/index.html;

    location ~ \.php$ {                      # nested — evaluated inside ^~ block
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.5-fpm.sock;
        fastcgi_send_timeout 600;
        fastcgi_read_timeout 600;
    }

    location ~ /\. { deny all; }             # hide .htaccess etc.
}
```

Hardening in place (verified 403):
- `location = /cars/api/install.php { deny all; }`
- `location = /cars/api/drop_all_tables.sql { deny all; }`

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
4. Deep links work because nginx falls back to `/cars/index.html`; keep the `^~`
   + nested `\.php$` structure or PHP breaks.

---

## 9. Rollback / safety

- Vhost backup: `/etc/nginx/sites-available/default.bak.<timestamp>` (made before each change).
- Full DB dumps: `mysqldump merhab_cars` → `api/backups/`.
- `git push origin feature-from-e9dcfaf` → `github.com/ekotobi/cars_vue_js.git`.
