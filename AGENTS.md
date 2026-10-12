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

Access uses the production `root` SSH credentials already used in this repo
(see `DEPLOYMENT.md` and the existing tooling). Do not commit new secrets.
