| File                                   | Package   | Role                                   | Agent # | Key exports |
| -------------------------------------- | --------- | -------------------------------------- | ------- | ----------- |
| go.mod                                 | module    | Root Go module definition              | 0       | (deps only) |
| Makefile                               | -         | Build, test, lint entry points         | 0       | (targets)   |
| cmd/tunnel-server/main.go             | main      | Server entry point: flags, wiring, proxy+control+admin, graceful shutdown | 11 | main        |
| cmd/tunnel-client/main.go             | main      | Client entry point (skeleton)          | 0       | main        |
| internal/config/doc.go                | config    | Package docs for configuration         | 0       | (none)      |
| internal/config/config.go            | config    | Constants, Config, defaults, Validate | 1       | Config, ServerDefaults, ClientDefaults, Validate, Version, consts |
| internal/config/config_test.go      | config_test | Unit tests for config                 | 1       | (tests)     |
| internal/log/doc.go                   | log       | Package docs for logging wrapper       | 0       | (none yet)  |
| internal/log/log.go                    | log       | Structured slog-based logger           | 2       | Logger, New, SetDefault, Debug/Info/Warn/Error/Fatal |
| internal/registry/doc.go              | registry  | Package docs for tunnel registry       | 0       | (none)      |
| internal/registry/registry.go        | registry  | Client storage, O(1) lookup by domain  | 4       | Client, Registrar, New, NewWithMax, GenerateDomain, Err* |
| internal/registry/registry_test.go    | registry_test | Unit tests for registry             | 4       | (tests)     |
| internal/frame/doc.go                 | frame     | Package docs for binary frame protocol | 0       | (none yet)  |
| internal/control/doc.go               | control   | Package docs for control plane         | 0       | (none)      |
| internal/control/control.go          | control   | Handler, ServeConn, demux, JSON, frameDispatch | 5 | Handler, NewHandler, ServeConn |
| internal/control/control_test.go     | control_test | Unit tests for control (register, heartbeat, frame, DoS, ctx) | 5 | (tests) |
| internal/proxy/dispatcher.go         | proxy     | Dispatcher interface for DeliverFrame | 5       | Dispatcher  |
| internal/proxy/proxy.go               | proxy     | Proxy, ServeHTTP, DeliverFrame, pendingMap, WSUpgrader | 6 | Proxy, New, NewWithSemSize, SetWSUpgrader |
| internal/proxy/proxy_test.go         | proxy_test | Unit tests for proxy (happy path, MIME, 404/502/503/504, path strip, body streaming, concurrent -race) | 6 | (tests) |
| internal/wsproxy/dispatcher.go       | wsproxy   | Dispatcher interface for DeliverFrame | 5       | Dispatcher  |
| internal/proxy/doc.go                 | proxy     | Package docs for HTTP proxying         | 0       | (none yet)  |
| internal/wsproxy/doc.go               | wsproxy   | Package docs for WebSocket proxying    | 0       | (none)      |
| internal/wsproxy/wsproxy.go           | wsproxy   | WSProxy, HandleUpgrade, DeliverFrame, sessionMap, idleWatcher | 7 | WSProxy, New |
| internal/wsproxy/wsproxy_test.go      | wsproxy_test | Unit tests for wsproxy (unknown ID, 101, reject non-101) | 7 | (tests) |
| internal/client/doc.go                | client    | Package docs for tunnel client          | 0       | (none)      |
| internal/client/client.go             | client    | TunnelClient, Run, connect, demux, backoff, graceful shutdown | 8 | TunnelClient, New |
| internal/client/forwarder.go          | client    | Forwarder, HandleRequest, DeliverBody, pipe body, http.Client reuse | 8 | Forwarder, NewForwarder, SetSendFrame |
| internal/client/client_test.go       | client_test | Unit tests for Run cancel, New         | 8       | (tests)     |
| internal/client/forwarder_test.go     | client_test | Unit tests for forwarder (success, path strip, 502, no-op, double-slash) | 8 | (tests)     |
| internal/admin/doc.go                 | admin     | Package docs for admin server          | 0       | (none)      |
| internal/admin/admin.go               | admin     | Server, /api/health, /api/stats, /api/clients, disconnect, pprof (opt-in) | 10 | Server, New, Start, Shutdown |
| internal/admin/admin_test.go          | admin_test | Unit tests for health, stats, clients, disconnect, pprof disabled | 10 | (tests)     |
| internal/ratelimit/doc.go             | ratelimit | Package docs for rate limiting         | 0       | (none)      |
| internal/ratelimit/ratelimit.go       | ratelimit | Limiter, shards, Allow, Cleanup, normaliseAddr, FNV-1a | 9 | Limiter, NewLimiter, Allow, Cleanup, SetCleanupStaleForTest |
| internal/ratelimit/ratelimit_test.go  | ratelimit_test | Unit tests + BenchmarkAllow (under limit, burst, refill, concurrent -race, different IPs, IPv6 /48, cleanup) | 9 | (tests) |
| internal/client/distserver.go          | client    | DistServer helper serving dist/ via local HTTP for tunnelling | 12 | DistServer, NewDistServer |
| internal/client/distserver_test.go     | client    | Unit tests for DistServer (files, SPA fallback, invalid dir) | 12 | (tests)     |

## Concurrency model

High-level goroutine diagram (to be refined by later agents):

  tunnel-server:
    - 1 goroutine: accept loop for client control connections
    - 1 goroutine per tunnel client: control connection handler
    - 1 goroutine per active proxy request: browser ↔ client data pump

  tunnel-client:
    - 1 goroutine: control connection to server
    - 1 goroutine per active local HTTP request: local app ↔ server data pump

All long-lived goroutines have a clear owner and are cancelled via context trees rooted in main().

## Wire protocol

The control plane will use JSON messages over a persistent connection to register tunnels, exchange heartbeats, and signal new proxy requests.
The data plane will use a binary frame protocol with length-prefixed frames, stream IDs, and explicit close semantics to multiplex many HTTP streams over a single persistent connection between server and client.

Details of the control JSON schema and frame layout will be specified in `internal/control` and `internal/frame` by later agents.

## Scalability model

- Up to 1,000,000 registered tunnel clients per server node (idle goroutines are cheap; registry is sharded logically by tunnel ID).
- 50,000–100,000 concurrent active proxy requests per node, enforced via per-client and global semaphores in the proxy layer.
- Horizontal sharding across nodes is achieved by consistent hashing on tunnel ID and fronting the nodes with an L4/L7 load balancer (documented in `deploy/DEPLOYMENT.md`).
