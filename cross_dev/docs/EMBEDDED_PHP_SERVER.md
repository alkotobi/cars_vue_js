# Embedding a Portable PHP Server in CrossDev

## Goal

Run the Cars API (PHP in `api/`) inside the CrossDev desktop app so that:

- No separate PHP/XAMPP install is required.
- The app can work offline (API runs on localhost).
- One portable bundle (app + PHP + API) per platform.

## Approach: Bundle PHP CLI + Built-in Server

PHP’s built-in web server is suitable for this:

```text
php -S 127.0.0.1:PORT -t /path/to/api
```

- **Port**: Fixed port (e.g. `8765`) or chosen at runtime (find a free port).
- **Document root**: Path to the bundled `api/` folder (next to the app or inside the bundle).

## What to Bundle

| Platform   | PHP binary / runtime                                      | API folder      |
|-----------|------------------------------------------------------------|-----------------|
| Windows   | Portable `php.exe` (e.g. from [windows.php.net](https://windows.php.net/download/)) in `Resources/php/` or next to `.exe` | Copy of `api/` (with config, sqlite, etc. as needed) |
| macOS     | PHP binary in `.app/Contents/Resources/php/` or via embedded runtime | Same            |
| Linux     | `php` binary in `lib/` or `resources/` next to the app    | Same            |
| iOS       | Not standard; consider a small HTTP server in C++/ObjC that talks to SQLite, or a pre-packaged backend. PHP on iOS is uncommon. |

Optional: bundle SQLite (or use PHP’s SQLite extension) so the embedded API has a local DB.

### macOS: Portable PHP with MySQL (pdo_mysql / mysqli)

To get a **single-file, portable PHP CLI with MySQL support** on macOS (no Homebrew PHP or system PHP required), use [static-php-cli](https://static-php.dev/) to build a static binary that includes `pdo_mysql`, `mysqli`, and common extensions (curl, openssl, mbstring, etc.).

**Option A – Use the project script (recommended)**

From the repo root:

```bash
chmod +x cross_dev/scripts/build-php-macos.sh
./cross_dev/scripts/build-php-macos.sh
```

- **Requirements**: macOS, Xcode Command Line Tools (`xcode-select --install`), `curl`.
- **Output**: `cross_dev/bundle/php-macos/php` (and optional `php.ini`).
- **First run**: downloads the static-php-cli binary and PHP/dependency sources, then compiles. Subsequent runs reuse the build; delete `cross_dev/build-php-static/` to start clean.
- **Custom output dir**: `./cross_dev/scripts/build-php-macos.sh /path/to/output`

**Option B – Manual build with static-php-cli**

1. Download the `spc` binary for your Mac arch from [static-php.dev – Manual build](https://static-php.dev/en/guide/manual-build.html) (e.g. `spc-macos-aarch64` or `spc-macos-x86_64`), then `chmod +x spc`.
2. Download sources: `./spc download --for-extensions=mysqlnd,pdo_mysql,mysqli,curl,openssl,mbstring,phar,posix,zlib --with-php=8.3`
3. Build CLI: `./spc build mysqlnd,pdo_mysql,mysqli,curl,openssl,mbstring,phar,posix,zlib --build-cli`
4. Use the binary at `buildroot/bin/php` (copy it into your app bundle or `cross_dev/bundle/php-macos/`).

Then run the embedded server with:

```bash
./php -S 127.0.0.1:8765 -t /path/to/api
```

## CrossDev (C++) Side

### 1. Startup: start the PHP server

- In `AppRunner` or a small “embedded server” module (e.g. `EmbeddedPhpServer` or an app-specific handler):
  - Resolve paths:
    - **PHP binary**: e.g. `getExecutableDir() + "/php"` (or `php.exe` on Windows).
    - **API doc root**: e.g. `getExecutableDir() + "/api"` or a path inside the app bundle.
  - Choose a port (fixed, e.g. `8765`, or use a helper to find a free port).
  - Spawn process (platform-specific):
    - **Windows**: `CreateProcess` or `_popen` with `php -S 127.0.0.1:8765 -t "C:\path\to\api"`.
    - **macOS/Linux**: `fork` + `exec` or `posix_spawn` with `php -S 127.0.0.1:8765 -t /path/to/api`.
  - Store the port (and optionally the process handle) so the WebView/frontend can use it.
  - Optional: wait until `127.0.0.1:PORT` responds (e.g. simple TCP connect or HTTP GET) before loading the UI.

### 2. Shutdown: stop the PHP server

- On app exit, kill the spawned PHP process (e.g. by process handle on Windows, or by PID on macOS/Linux).

### 3. Tell the frontend the API base URL

The Vue app (e.g. `useApi.js`) must use `http://127.0.0.1:PORT/api` when running inside CrossDev. Options:

- **A. Inject before load**  
  When loading the main window URL (e.g. `file:///path/to/index.html`), inject a small script that runs first and sets a global, e.g.:

  ```js
  window.__CROSS_DEV_API_BASE__ = 'http://127.0.0.1:8765/api';
  ```

  Then in `useApi.js`, if `window.__CROSS_DEV_API_BASE__` is set, use it as the API base URL (and do not use `hostname`/`localhost`/merhab.com for that case).

- **B. Query param**  
  Load the app with a query param, e.g. `file:///path/index.html?api=http://127.0.0.1:8765/api`, and in `useApi.js` parse `location.search` and use that as the API base when present.

- **C. Native → JS message**  
  After the page loads, send a message from C++ (e.g. via the existing message router / bridge) with the API base URL; the frontend stores it and uses it for all API calls.

Option A or B keeps the API base available before the first `useApi()` call.

## Vue / useApi.js Changes

- If `window.__CROSS_DEV_API_BASE__` is set (or an API URL is present in the query string), use it as the base for `API_BASE_URL` instead of `localhost:8000` or `merhab.com`.
- Ensure CORS is not an issue (same-origin is not possible with file://; the embedded API must allow `Origin: null` or the exact file:// origin if the browser sends it, or you load the app from `http://127.0.0.1:ANOTHER_PORT` so the API is same-origin).

## Optional: Serve the Vue app from the same PHP server

- Build the Vue app (e.g. `dist/`) and point PHP’s document root to a folder that contains both:
  - `api/` (or a router that forwards `/api/*` to your PHP entrypoint).
  - `index.html` and assets for the SPA.
- Then open `http://127.0.0.1:8765/` in the WebView. No `file://` and same-origin API calls.

This might require a small PHP router or a separate port for the SPA (e.g. PHP on 8765 for API only, and still use file:// for the app, or serve everything on one port).

## Summary

| Step | Action |
|------|--------|
| 1 | Bundle portable PHP binary + `api/` folder per platform. |
| 2 | On app start, spawn `php -S 127.0.0.1:PORT -t <api_path>`. |
| 3 | Pass the API base URL to the frontend (inject global or query param). |
| 4 | In `useApi.js`, prefer the embedded API URL when set. |
| 5 | On app exit, kill the PHP process. |

This gives you a portable PHP server embedded in CrossDev for Windows/macOS/Linux. iOS, if needed, would be handled separately (e.g. no PHP, use a small native HTTP + DB layer instead).
