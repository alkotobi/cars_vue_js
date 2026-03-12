Tunnel deployment is designed for Linux production nodes while still supporting local development on macOS and Windows.

High-level targets:

- Up to 1,000,000 registered tunnel clients per node.
- 50,000–100,000 concurrent active proxy requests per node.
- Horizontal sharding across many nodes using consistent hashing on tunnel ID.

Sharding model (overview only, to be refined by later agents):

- A fronting L4/L7 load balancer receives inbound traffic for the tunnel server.
- A stable hash of the tunnel ID selects the backend node responsible for that tunnel.
- Clients reconnect to the same node using the same tunnel ID, keeping registry lookups local.

Systemd unit files:

- `deploy/systemd/tunnel-server.service` runs the Linux tunnel server binary.
- `deploy/systemd/tunnel-client.service` runs a long-lived client on edge hosts when needed.

For networking and Nginx examples, see `deploy/nginx.conf.example`.
