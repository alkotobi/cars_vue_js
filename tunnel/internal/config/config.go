package config

import (
	"fmt"
	"net/url"
	"time"
)

// Capacity constants: per-node limits and initial allocations.
const (
	MaxRegisteredClients int = 1_000_000 // idle control connections per node
	MaxActiveProxyReqs   int = 50_000     // concurrent HTTP/WS proxy requests per node
	InitialRegistryCap   int = 1_024
	ClientHashBuckets    int = 131_072   // must be power of 2; for future sharded map
)

// Timeouts for control, proxy, and connection lifecycle.
const (
	HeartbeatTimeout    = 120 * time.Second
	ProxyRequestTimeout = 60 * time.Second
	ClientIdleTimeout   = 300 * time.Second
	WSIdleTimeout       = 600 * time.Second
	DialTimeout         = 10 * time.Second
	ShutdownGracePeriod = 30 * time.Second
)

// Default ports for HTTP, control, and admin listeners.
const (
	DefaultHTTPPort    = 83
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
	CopyBufferSize  int = 32 * 1024  // 32 KiB io.Copy buffer
	HighWatermark   int = 256 * 1024  // pause upstream reads above this
	LowWatermark    int = 64 * 1024   // resume upstream reads below this
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
	HTTPPort    int
	ControlPort int
	AdminPort   int
	PublicHost  string
	Workers     int // 0 = GOMAXPROCS

	// Client
	ServerAddr   string
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
		HTTPPort:    DefaultHTTPPort,
		ControlPort: DefaultControlPort,
		AdminPort:   DefaultAdminPort,
		PublicHost:  "",
		Workers:     0,
		LogLevel:    "info",
		PProfEnabled: false,
	}
}

// ClientDefaults returns a Config with client defaults filled in.
func ClientDefaults() Config {
	return Config{
		ServerAddr:    "",
		LocalURL:      "http://127.0.0.1:5173",
		Subdomain:     "",
		ReconnectMax:  5 * time.Minute,
		LogLevel:      "info",
		PProfEnabled:  false,
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
	return validateLogLevel(c.LogLevel)
}

func (c Config) validateClient() error {
	if c.ServerAddr == "" {
		return fmt.Errorf("config: ServerAddr is required")
	}
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
