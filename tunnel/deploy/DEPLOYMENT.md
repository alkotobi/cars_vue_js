# Deployment Guide — merhab.com

This document describes how to deploy the tunnel server and client for the
`merhab.com` domain using the subdomain routing model:

- Tunnel clients are identified by subdomain: `<id>.merhab.com`.
- Public URLs always use HTTPS: `https://<id>.merhab.com/`.
- The tunnel server listens on plain HTTP internally (port 83).
- TLS is terminated by nginx.

The same binaries also work for local development on macOS and Windows.

---

## Architecture Overview

High‑level data flow:

```text
Browser
  │
  ▼
nginx :443  (TLS termination, SNI: merhab.com / *.merhab.com)
  │
  ▼
tunnel-server :83  (plain HTTP)
    ▲          ▲
    │          │ HTTP proxy frames
    │
    │  TCP (no TLS, no nginx)
    │
    └───────── tunnel-client  (edge / local machine)
                 │
                 │ HTTP
                 ▼
             dist/ via Go file server on 127.0.0.1:<random>
```

Key points:

- **nginx** handles all TLS and HTTP/2. The Go tunnel server only speaks plain HTTP.
- **Port 83**: internal HTTP listener. Only nginx should talk to this port.
- **Port 3000**: control channel for tunnel clients. Clients connect directly
  (no nginx, raw TCP).
- **Port 8083**: admin API. Listens on localhost; must never be exposed externally.

Targets:

- Up to **1,000,000** registered tunnel clients per node.
- **50,000–100,000** concurrent active proxy requests per node.
- Horizontal sharding across many nodes using consistent hashing on tunnel ID
  (see “Horizontal Scaling” below).

---

## Firewall Rules

On the tunnel server node, configure the firewall so only the intended ports
are reachable externally.

Recommended rules:

| Port | Protocol | Source    | Purpose                         |
|------|----------|-----------|---------------------------------|
| 80   | TCP      | 0.0.0.0/0 | HTTP → HTTPS redirect          |
| 443  | TCP      | 0.0.0.0/0 | HTTPS tunnel traffic           |
| 3000 | TCP      | 0.0.0.0/0 | Client control channel (direct)|
| 83   | TCP      | 127.0.0.1 | Internal only — nginx → server |
| 8083 | TCP      | 127.0.0.1 | Admin API — local only         |

Example using **ufw**:

```bash
ufw allow 80/tcp
ufw allow 443/tcp
ufw allow 3000/tcp       # tunnel clients connect here directly

ufw deny 83/tcp          # block direct access; nginx only
ufw deny 8083/tcp        # block external admin access
```

Double‑check from a remote host:

```bash
nc -vz merhab.com 83      # should fail (connection refused or filtered)
nc -vz merhab.com 8083    # should fail
nc -vz merhab.com 3000    # should succeed
```

---

## DNS Setup

At your DNS provider, create the following records:

```text
A    merhab.com       163.245.222.142
A    www.merhab.com   163.245.222.142
A    *.merhab.com     163.245.222.142
```

Verify:

```bash
dig A merhab.com +short
dig A cars.merhab.com +short
```

Both should return `163.245.222.142`.

---

## TLS Certificate (Let’s Encrypt)

Use a wildcard certificate for `merhab.com` and `*.merhab.com` so all tunnel
subdomains are covered.

Example using **certbot** with a DNS challenge:

```bash
apt update
apt install certbot

certbot certonly \
  --manual \
  --preferred-challenges dns \
  -d merhab.com \
  -d "*.merhab.com"
```

Certbot will print instructions like:

```text
Please create a TXT record:
_acme-challenge.merhab.com → <random-token>
```

Add that TXT record at your DNS provider, wait 60–120 seconds for DNS to
propagate, then press Enter in the certbot terminal.

After success, the certificate will be available at:

- `/etc/letsencrypt/live/merhab.com/fullchain.pem`
- `/etc/letsencrypt/live/merhab.com/privkey.pem`

### Auto‑renewal

Add a cron entry to renew certificates and reload nginx:

```bash
cat <<'EOF' >/etc/cron.d/certbot
0 3 * * * root certbot renew --quiet && nginx -s reload
EOF
```

---

## Reverse proxy (nginx or Apache2)

The tunnel server speaks plain HTTP on port 83; TLS must be terminated by your
web server. You can use **nginx** or **Apache2** — the server itself is unchanged.

### nginx

The file `deploy/nginx.conf.example` contains a complete nginx vhost for
`merhab.com` and `*.merhab.com`:

- HTTP → HTTPS redirect with ACME HTTP‑01 support.
- HTTPS for `merhab.com` and `www.merhab.com` pointing to `127.0.0.1:83`.
- HTTPS for `*.merhab.com` also pointing to `127.0.0.1:83`, with:
  - `Host` preserved (`cars.merhab.com`).
  - `X-Forwarded-*` headers set.
  - WebSocket upgrade headers configured.
  - `proxy_buffering off` so responses stream directly.

To enable:

```bash
ln -s /etc/nginx/sites-available/tunnel-merhab /etc/nginx/sites-enabled/tunnel-merhab
nginx -t
systemctl reload nginx
```

Adjust paths if your distribution uses a different nginx layout.

### Apache2

The file `deploy/apache2.conf.example` does the same for Apache2:

- Enable modules: `sudo a2enmod ssl rewrite proxy proxy_http proxy_wstunnel headers`
- Copy the example to `/etc/apache2/sites-available/tunnel-merhab.conf`
- Enable site: `sudo a2ensite tunnel-merhab`
- Test and reload: `sudo apache2ctl configtest && sudo systemctl reload apache2`

Cert paths and firewall rules are the same as for nginx (port 83 internal only, 443 and 3000 open as described above).

---

## Server Installation (tunnel-server)

Build the server binary on a build machine or directly on the server.

```bash
cd /path/to/tunnel
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o tunnel-server ./cmd/tunnel-server
```

Copy the binary to `/usr/local/bin`:

```bash
cp tunnel-server /usr/local/bin/tunnel-server
chmod 755 /usr/local/bin/tunnel-server
```

Create the systemd unit `/etc/systemd/system/tunnel-server.service` from
`deploy/systemd/tunnel-server.service` (referenced here; content not repeated).

Reload and start:

```bash
systemctl daemon-reload
systemctl enable tunnel-server
systemctl start tunnel-server
systemctl status tunnel-server
```

Check health via nginx:

```bash
curl -s https://merhab.com/api/health
```

Expected response:

```json
{"status":"ok","version":"2.0.0",...}
```

---

## Client Usage (tunnel-client)

### Serve a built Vue/React app (dist/ directory)

On your development machine:

```bash
cd /path/to/app
npm run build   # produces dist/

# Run the tunnel client pointing at dist/
./tunnel-client \
  -domain merhab.com \
  -control 3000 \
  -local ./dist \
  -subdomain cars
```

What happens:

- `tunnel-client` starts a local HTTP file server on `127.0.0.1:<random>` that
  serves the `dist/` directory (via `DistServer`).
- It connects to `merhab.com:3000` over raw TCP (no nginx).
- The tunnel server registers the subdomain `cars` and returns:
  `https://cars.merhab.com/` in the `register_ack`.

You can then open:

```text
https://cars.merhab.com/
```

from any browser and see the app.

### Serve an existing local HTTP server

If you already have a local HTTP service running (for example, a Go API on
`http://localhost:3000`), you can tunnel it directly:

```bash
./tunnel-client \
  -domain merhab.com \
  -control 3000 \
  -local http://localhost:3000 \
  -subdomain myapi
```

Now `https://myapi.merhab.com/` proxies to your local service.

---

## Vue App Build Configuration

With subdomain routing, the Vue app does **not** need a special base path:

- The app is always served from the root of the subdomain:
  `https://cars.merhab.com/`.
- Static assets live under `/` in the dist output (e.g. `/index.js`,
  `/index.css`).

Minimal `vite.config.js` for production:

```ts
// vite.config.js
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    outDir: 'dist',
  },
})
```

Build and tunnel:

```bash
npm run build
./tunnel-client -domain merhab.com -local ./dist -subdomain cars
# Open https://cars.merhab.com/
```

---

## OS Tuning for 1M Connections

To support a large number of concurrent connections and file descriptors, tune
the kernel and limits accordingly.

`/etc/sysctl.d/99-tunnel.conf`:

```conf
net.ipv4.tcp_max_syn_backlog = 65536
net.core.somaxconn            = 65536
net.ipv4.ip_local_port_range  = 1024 65535
net.ipv4.tcp_tw_reuse         = 1
fs.file-max                   = 4000000
```

Apply:

```bash
sysctl -p /etc/sysctl.d/99-tunnel.conf
```

`/etc/security/limits.conf`:

```conf
* soft nofile 1000000
* hard nofile 1000000
```

Log out and back in (or restart) to apply `nofile` limits.

---

## Horizontal Scaling

Beyond **50k** concurrent active proxy requests per node, you will eventually
need to run multiple tunnel‑server instances.

High‑level approach:

- Run multiple tunnel‑server nodes, each with its own control port and HTTP
  listener (83) behind an L4/L7 load balancer.
- Use a consistent hash of the tunnel ID (subdomain) to choose a node.
- Ensure both HTTP (nginx vhosts) and control connections use the **same**
  hash, so:
  - `cars.merhab.com` and the control channel for `cars` always land on the
    same node.
- For a future multi‑node registry:
  - Back the registry with Redis or etcd so nodes see a shared view of
    registered clients.
  - Or shard by tunnel ID such that a given ID only ever registers on a
    specific node (“home node”).

Concrete multi‑node wiring is left for a later iteration; this document
captures the single‑node deployment and the scaling direction.

