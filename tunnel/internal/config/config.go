package config

import (
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"
)

// Capacity constants: per-node limits and initial allocations.
const (
	MaxRegisteredClients int = 1_000_000 // idle control connections per node
	MaxActiveProxyReqs   int = 50_000    // concurrent HTTP/WS proxy requests per node
	InitialRegistryCap   int = 1_024
	ClientHashBuckets    int = 131_072 // must be power of 2; for future sharded map
)

// Timeouts for control, proxy, and connection lifecycle.
const (
	HeartbeatTimeout    = 120 * time.Second
	ProxyRequestTimeout = 120 * time.Second
	ClientIdleTimeout   = 300 * time.Second
	WSIdleTimeout       = 600 * time.Second
	DialTimeout         = 10 * time.Second
	ShutdownGracePeriod = 30 * time.Second
)

// BaseDomain is the root domain. Tunnel subdomains are <id>.BaseDomain.
// Never hardcode this string anywhere else — always reference this constant.
const BaseDomain = "merhab.com"

// PublicScheme is the scheme used in public URLs shown to clients.
// TLS is terminated by nginx; the server itself speaks plain HTTP.
const PublicScheme = "https"

// Default ports for HTTP, control, and admin listeners.
const (
	// DefaultHTTPPort is the internal HTTP port nginx proxies to.
	// nginx listens on 443, terminates TLS, forwards to this port.
	DefaultHTTPPort = 83
	// DefaultControlPort is the port tunnel clients connect to for the control channel.
	// This port must be open in the firewall and NOT behind nginx.
	DefaultControlPort = 3000
	DefaultAdminPort   = 8083
)

// Proxy limits: per-client and global concurrency, header and domain size caps.
const (
	MaxPendingPerClient int = 512
	MaxPendingGlobal    int = 256 * 1024
	MaxDomainLen        int = 255
	MaxHeaderCount      int = 64
	MaxHeaderKeyLen     int = 128
	MaxHeaderValLen     int = 4096
)

// Backpressure: copy buffer and flow-control watermarks.
const (
	CopyBufferSize int = 32 * 1024  // 32 KiB io.Copy buffer
	HighWatermark  int = 256 * 1024 // pause upstream reads above this
	LowWatermark   int = 64 * 1024  // resume upstream reads below this
)

// Frame protocol: header size, max payload, and control line length.
const (
	FrameHeaderSize int = 13              // type(1) + id(8) + length(4)
	FrameMaxPayload int = 4 * 1024 * 1024 // 4 MiB
	ControlLineMax  int = 64 * 1024
)

// Version is the tunnel binary version string for /health and metrics.
const Version = "2.0.0"

// Config holds runtime configuration parsed from flags or environment.
// Server and client use different subsets; Validate() checks the active role.
type Config struct {
	// Server
	HTTPPort     int
	ControlPort  int
	AdminPort    int
	BaseDomain   string
	PublicScheme string
	Workers      int // 0 = GOMAXPROCS

	// Client
	ServerAddr   string
	ControlHost  string
	LocalURL     string
	Subdomain    string
	ReconnectMax time.Duration // cap for exponential backoff

	// Observability
	LogLevel     string
	PProfEnabled bool
}

// ServerDefaults returns a Config with server defaults filled in.
func ServerDefaults() Config {
	return Config{
		BaseDomain:   BaseDomain,
		PublicScheme: PublicScheme,
		HTTPPort:     DefaultHTTPPort,
		ControlPort:  DefaultControlPort,
		AdminPort:    DefaultAdminPort,
		Workers:      0,
		LogLevel:     "info",
		PProfEnabled: false,
	}
}

// ClientDefaults returns a Config with client defaults filled in.
func ClientDefaults() Config {
	return Config{
		ServerAddr:   "",
		ControlHost:  BaseDomain,
		LocalURL:     "http://127.0.0.1:5173",
		Subdomain:    "",
		ReconnectMax: 5 * time.Minute,
		LogLevel:     "info",
		PProfEnabled: false,
	}
}

// Validate returns an error if any required field is missing or out of range.
// Call after applying defaults and flags; for server mode only server fields
// are checked; for client mode only client fields are checked. Pass serverMode true
// when validating server config, false for client config.
func (c Config) Validate(serverMode bool) error {
	if serverMode {
		return c.validateServer()
	}
	return c.validateClient()
}

func (c Config) validateServer() error {
	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		return fmt.Errorf("config: HTTPPort %d out of range [1, 65535]", c.HTTPPort)
	}
	if c.ControlPort < 1 || c.ControlPort > 65535 {
		return fmt.Errorf("config: ControlPort %d out of range [1, 65535]", c.ControlPort)
	}
	if c.AdminPort < 1 || c.AdminPort > 65535 {
		return fmt.Errorf("config: AdminPort %d out of range [1, 65535]", c.AdminPort)
	}
	if c.Workers < 0 {
		return fmt.Errorf("config: Workers %d must be >= 0", c.Workers)
	}
	if strings.TrimSpace(c.BaseDomain) == "" {
		return fmt.Errorf("config: BaseDomain must not be empty")
	}
	if !strings.Contains(c.BaseDomain, ".") {
		return fmt.Errorf("config: BaseDomain %q must contain at least one dot", c.BaseDomain)
	}
	switch c.PublicScheme {
	case "http", "https":
	default:
		return fmt.Errorf("config: PublicScheme %q must be http or https", c.PublicScheme)
	}
	return validateLogLevel(c.LogLevel)
}

func (c Config) validateClient() error {
	if c.LocalURL == "" {
		return fmt.Errorf("config: LocalURL is required")
	}
	// LocalURL must be parseable and have a host (e.g. http://127.0.0.1:5173).
	u, err := url.Parse(c.LocalURL)
	if err != nil {
		return fmt.Errorf("config: LocalURL invalid: %w", err)
	}
	if u.Host == "" {
		return fmt.Errorf("config: LocalURL must include host (e.g. http://127.0.0.1:5173)")
	}
	if len(c.Subdomain) > MaxDomainLen {
		return fmt.Errorf("config: Subdomain length %d exceeds MaxDomainLen %d", len(c.Subdomain), MaxDomainLen)
	}
	if c.ReconnectMax < 0 {
		return fmt.Errorf("config: ReconnectMax must be >= 0")
	}
	return validateLogLevel(c.LogLevel)
}

func validateLogLevel(level string) error {
	switch level {
	case "debug", "info", "warn", "error", "":
		return nil
	default:
		return fmt.Errorf("config: LogLevel %q must be one of debug, info, warn, error", level)
	}
}

// PublicURL returns the full public URL for a tunnel subdomain.
// Example: PublicURL("cars") → "https://cars.merhab.com/".
func (c Config) PublicURL(subdomain string) string {
	scheme := c.PublicScheme
	if scheme == "" {
		scheme = PublicScheme
	}
	domain := c.BaseDomain
	if domain == "" {
		domain = BaseDomain
	}
	return fmt.Sprintf("%s://%s.%s/", scheme, subdomain, domain)
}

// IsBaseDomain returns true if host is the base domain or www subdomain.
// These requests are NOT tunnel lookups.
func (c Config) IsBaseDomain(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	base := c.BaseDomain
	if base == "" {
		base = BaseDomain
	}
	return host == base || host == "www."+base
}

// SubdomainFrom extracts the tunnel ID from a Host header value.
// "cars.merhab.com" → "cars", true
// "merhab.com"      → "", false  (base domain, not a tunnel)
// "other.com"       → "", false  (foreign domain)
func (c Config) SubdomainFrom(host string) (string, bool) {
	host = strings.ToLower(strings.TrimSpace(host))
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	base := c.BaseDomain
	if base == "" {
		base = BaseDomain
	}
	suffix := "." + base
	if !strings.HasSuffix(host, suffix) {
		return "", false
	}
	sub := strings.TrimSuffix(host, suffix)
	if sub == "" || sub == "www" {
		return "", false
	}
	// Reject nested subdomains like "cars.other.merhab.com" — only a single
	// label directly under BaseDomain is considered a tunnel ID.
	if strings.Contains(sub, ".") {
		return "", false
	}
	return sub, true
}
