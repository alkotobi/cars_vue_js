# AGENTS.md

Repo-wide instructions for agents working in this repository.
(`whatsapp-agent-tool/AGENTS.md` governs the Go tool; this file governs the app/deployment.)

## Deployment: ALWAYS push to BOTH servers

This project runs on **two production servers**. They must always run the **same
revision**. Any change that applies to production must be pushed to **both**, never
just one.

| Server | IP | Domain | Web root | Shared API | PHP-FPM |
| ------ | -- | ------ | -------- | ---------- | ------- |
| 1 | `163.245.222.142` | merhab.com / www.merhab.com | `/var/www/html` | `/var/www/html/api/api.php` | `php8.3-fpm` |
| 2 | `163.245.214.125` | world-automobile.com | `/var/www/world-automobile.com` | `/var/www/world-automobile.com/api/api.php` | `php8.5-fpm` |

`api/api.php` is the **shared** backend symlinked/copied by every tenant on each
server, so a single edit must land on both hosts.

Rules:

1. **Push every update to both servers.** If you change `api/api.php` (or any shared
   backend file), deploy it to server 1 *and* server 2 in the same task.
2. **Never leave them divergent.** If one server is unreachable or unfixed, say so
   explicitly and call it out as out-of-sync — do not silently update only one.
3. **Back up first.** Copy the existing file to a timestamped `.bak.<YYYYMMDDHHMMSS>`
   before overwriting.
4. **Preserve ownership/permissions.** `chown www-data:www-data` and `chmod 644` after
   copying, then `php -l` and reload that server's PHP-FPM.
5. **Verify and report.** After deploying, print the deployed file's `md5` on **each**
   server and confirm both match the local file.
6. **Never let a deploy change upload-folder ownership.** Every tenant's upload tree
   (`<webroot>/<tenant>/files/…`) must be owned by `www-data:www-data` and be group
   writable (dirs `2775`, files `664`). A deploy that `rsync -a`'s from a Mac stamps
   the folders with the local uid `501`/`staff`, and PHP-FPM (`www-data`) can no
   longer write — the app then fails with **`Upload directory is not writable`**
   (thrown at `api/upload.php` when `is_writable()` is false) on every client/id
   upload. This already broke `merhab.com` (server 1, tenants `mig`/`mig_26`).
   Deploy with `rsync -rlptD --no-owner --no-group …` (do **not** use `-a`/`-o`/`-g`),
   or re-own the upload dirs after syncing:

   ```bash
   chown -R www-data:www-data /var/www/<root>/<tenant>/files
   find /var/www/<root>/<tenant>/files -type d -exec chmod 2775 {} +
   find /var/www/<root>/<tenant>/files -type f -exec chmod 664 {} +
   ```

   After any deploy, prove every tenant can write before calling it done:

   ```bash
   for d in $(find /var/www -type d -name files 2>/dev/null); do
     sudo -u www-data test -w "$d" && echo "OK  $d" || echo "BAD $d"
   done
   ```

Access uses the production `root` SSH credentials already used in this repo
(see `DEPLOYMENT.md` and the existing tooling). Do not commit new secrets.

## Known non-issues / leftovers (do not confuse with the bug above)

- Server 1: `/var/www/html/files` is `root:root 755` — it is the non-tenant root,
  not a tenant upload dir; no live request resolves there.
- Server 2: `/var/www/world-automobile.com/test_tenant_files/` (legacy sibling
  layout, `root:root`) and `/var/www/h.com/cars` (orphan copy, uid `501`) are not
  referenced by any enabled nginx config. Leave them alone; they are not the
  `Upload directory is not writable` cause.
