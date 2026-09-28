# Migration Guide — Path-Based → Subdomain-Based Routing

This guide documents the changes introduced by Agents 1–10 to move from
path-based routing (e.g. `/cars/...`) to **subdomain-based routing**
(`cars.merhab.com`) for the tunnel server and client.

---

## What Changed

### Routing

- **Before**
  - Tunnel ID extracted from `r.URL.Path`:
    - `/cars/src/main.js` → tunnel ID = `cars`, backend path = `/src/main.js`.
  - Public URLs included a path prefix: e.g.
    - `http://IP:83/cars/`.

- **After**
  - Tunnel ID comes **only** from the `Host` header:
    - `Host: cars.merhab.com` → tunnel ID = `cars`.
  - `r.URL.Path` is forwarded verbatim to the local app
    (via `r.URL.RequestURI()` in frames), including query strings.
  - Base domain requests:
    - `Host: merhab.com` or `Host: www.merhab.com`
      → landing JSON: `{"status":"ok","service":"tunnel","version":"2.0.0"}`.
  - Foreign domains (not ending in `.merhab.com`) return **404**.

Key helpers in `internal/config`:

- `Config.SubdomainFrom(host string) (string, bool)`:
  - Strips port (`net.SplitHostPort`).
  - Lowercases the host.
  - Accepts only single-label subdomains directly under `BaseDomain`:
    - `"cars.merhab.com"` → `"cars", true`.
    - `"cars.other.merhab.com"` → `"" , false`.
    - `"Cars.MERHAB.COM:443"` → `"cars", true`.
- `Config.IsBaseDomain(host string) bool`:
  - `true` for `merhab.com` and `www.merhab.com` (with or without `:443`).
  - `false` for subdomains such as `cars.merhab.com`.

### Public URLs

- **Before**: `register_ack` JSON:

  ```json
  {"type":"register_ack","domain":"cars","publicUrl":"http://IP:83/cars/"}
  ```

- **After**:

  ```json
  {"type":"register_ack","domain":"cars","publicUrl":"https://cars.merhab.com/"}
  ```

  - No IP address in the URL.
  - No port (nginx terminates TLS on 443).
  - No path prefix (routing is by **subdomain**, not by URL path).

`Config.PublicURL(subdomain)` generates these URLs and is used by
`internal/control` when sending `register_ack`.

### Local Serving (DistServer)

- A new helper `DistServer` in `internal/client/distserver.go`:
  - Serves a `dist/` directory over HTTP on `127.0.0.1:<random>` using
    Go’s `http.FileServer`.
  - SPA fallback:
    - If a requested path does not map to an on-disk file, `index.html` is
      served. This supports Vue Router / React Router history mode.
  - Binds to `127.0.0.1` only; never exposed externally.
  - MIME types:
    - Handled by Go’s `mime.TypeByExtension`, which correctly sets
      `application/javascript`, `text/css`, `application/wasm`, etc.

The tunnel client auto-detects `-local`:

- If `-local` starts with `http://` or `https://`:
  - Used directly as the local forwarding target.
- Else:
  - Treated as a directory path; `DistServer` is started and its URL is used
    as the local target.

---

## DNS Changes Required

At your DNS provider, create:

```text
A    merhab.com       163.245.222.142
A    www.merhab.com   163.245.222.142
A    *.merhab.com     163.245.222.142
```

Verification:

```bash
dig A merhab.com +short
dig A cars.merhab.com +short
```

Both should resolve to `163.245.222.142`.

---

## nginx Changes Required

The old example nginx config that fronted a single HTTP entrypoint has been
replaced with a full vhost in `deploy/nginx.conf.example`:

- HTTP (80):
  - Redirects all `merhab.com` and `*.merhab.com` traffic to HTTPS, except
    for ACME HTTP‑01 challenges under `/.well-known/acme-challenge/`.
- HTTPS (443):
  - `merhab.com` and `www.merhab.com` → `127.0.0.1:83`.
  - `*.merhab.com` (`[a-z0-9-]` only) → `127.0.0.1:83`.
  - `Host` is preserved; Go code uses `Host` to determine tunnel ID.
  - WebSocket upgrade headers are set (future‑proofing).
  - `proxy_buffering off` so large/streaming responses are not buffered.

The tunnel server itself **does not** speak TLS; nginx terminates TLS and
talks plain HTTP to `127.0.0.1:83`.

See `deploy/DEPLOYMENT.md` and `deploy/nginx.conf.example` for full config and
Certbot instructions.

---

## Server Flag Changes

`cmd/tunnel-server/main.go` now uses:

- `-domain` (default: `merhab.com`):
  - Sets `cfg.BaseDomain`.
- `-port` (default: `83`):
  - Internal HTTP listener (nginx proxies here).
- `-control` (default: `3000`):
  - Control TCP port; clients connect directly.
- `-admin` (default: `8083`):
  - Admin API on `127.0.0.1` only.

The old `-host` flag has been removed.

Startup banner shows:

```text
Domain:   *.merhab.com
HTTP:     0.0.0.0:83    (nginx proxies here)
Control:  0.0.0.0:3000  (clients connect directly)
Admin:    127.0.0.1:8083
```

And logs firewall reminders:

- `firewall: ensure port 3000 is open (control channel)`
- `firewall: ensure port 83 is NOT exposed (nginx only)`
- `firewall: ensure port 8083 is blocked externally (admin only)`

---

## Client Flag Changes

`cmd/tunnel-client/main.go` now uses:

- `-domain` (default: `merhab.com`):
  - Base domain for public URLs and control host.
- `-control` (default: `3000`):
  - Control TCP port; client connects to `<domain>:<control>`.
- `-local` (required):
  - Either a URL (`http://...`, `https://...`) or a directory path (e.g. `./dist`).
- `-subdomain`:
  - Tunnel subdomain; `-subdomain cars` → `cars.merhab.com`.
- `-log-level`, `-version` unchanged.

The old `-server` and `-port` flags are gone.

On startup, the client prints a banner similar to:

```text
Tunnel Client v2.0.0
Domain:    cars.merhab.com
Control:   merhab.com:3000
Local:     http://127.0.0.1:52341
Note: control connects directly, not nginx
Waiting for registration...
```

After registration, the client logs:

```text
✓ Tunnel active: https://cars.merhab.com/
```

---

## Vue App Changes (vite.config.js)

With subdomain routing, the Vue app no longer needs a custom `base` path:

- The app is always served from `/` at `https://cars.merhab.com/`.
- Vite’s default `base: '/'` is correct.

Minimal production config:

```ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
  },
})
```

Build and run:

```bash
npm run build
./tunnel-client -domain merhab.com -local ./dist -subdomain cars
```

Static assets in `dist/` are served with correct MIME types via Go’s
`http.FileServer` and streamed through the tunnel.

---

## Verification Commands

### 1. Environment and tests

```bash
cd /path/to/tunnel
go version         # should be go1.22.2 (toolchain go1.22.2)
go mod tidy

CGO_ENABLED=0 go test ./...
```

For CI on Linux, optionally:

```bash
go test -race ./internal/...
```

(On macOS, `-race` may hit `dyld missing LC_UUID` depending on the local toolchain.)

### 2. Routing and URLs

```bash
curl -s https://merhab.com/api/health

curl -s -H 'Host: cars.merhab.com' http://127.0.0.1:83/api/health

curl -s -H 'Host: merhab.com' http://127.0.0.1:83/
curl -s -H 'Host: www.merhab.com' http://127.0.0.1:83/

curl -s -o /dev/null -w '%{http_code}\n' -H 'Host: evil.com' http://127.0.0.1:83/
```

Expected:

- Base domain → `200` with landing JSON.
- Foreign domain → `404`.

### 3. Asset MIME types

After building and tunnelling a Vue app:

```bash
curl -s https://cars.merhab.com/                # index.html
curl -s -o /dev/null -w '%{content_type}\n' \
  https://cars.merhab.com/assets/index*.js
```

Expected `Content-Type` for JS: `application/javascript`.

### 4. Public URLs in register_ack

Enable debug logging and check client or server logs for:

```json
{"type":"register_ack","domain":"cars","publicUrl":"https://cars.merhab.com/"}
```

There should be no port number or path prefix in `publicUrl`.

---

This migration shifts the tunnel architecture from path-based routing to a
clean, DNS- and TLS‑aligned subdomain model, simplifying deployment and making
it easier to reason about public URLs, caching, and security boundaries.

