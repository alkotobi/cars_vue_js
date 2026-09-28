Tunnel is a lightweight, Cloudflare-Tunnel-style reverse tunnel written in Go 1.22.2 to expose your local Cars Vue.js app to the internet via a tunnel server running on 163.245.222.142.

Architecture (high level):

  Browser 
    │
    ▼
[tunnel-server @ 163.245.222.142] ⇄ [tunnel-client on dev machine] ─HTTP─> [local Cars app]

Quickstart (development):

  1. Install Go 1.22.2 and ensure `go` is on your PATH.
  2. From the `tunnel/` directory, run:
       make setup
       make fmt
       make test
       make build
  3. Start the tunnel server (implementation will be added by later agents):
       ./cmd/tunnel-server/tunnel-server
  4. Start the tunnel client on your local machine:
       ./cmd/tunnel-client/tunnel-client
  5. Point your browser at the public URL exposed by the server.

For production deployment details, see `deploy/DEPLOYMENT.md`.
