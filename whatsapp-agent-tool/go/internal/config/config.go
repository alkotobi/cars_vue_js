// Package config centralises environment/.env reading — the Go twin of src/config/index.ts.
package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// C is the package-level config, initialised once at import from the environment
// and the <RootDir>/.env file. Real environment values beat .env values.
var C Config

type DB struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type Cars struct {
	BaseURL  string
	Tenant   string
	APIToken string
}

type Config struct {
	RootDir          string
	DB               DB
	Cars             Cars
	WAProfile        string
	ProfileIsDefault bool
}

// WAProfile returns the account slug (WA_PROFILE || "whatsapp-profile"),
// mirroring src/chrome.ts PROFILE. It steers data dir, CDP state and queue files.
func WAProfile() string { return C.WAProfile }

// ProfileIsDefault mirrors src/chrome.ts PROFILE_IS_DEFAULT (!WA_PROFILE).
func ProfileIsDefault() bool { return C.ProfileIsDefault }

// DataDir returns the RUN prefix for per-account files: "data" (default) or
// "data/<profile>" — mirroring src/chrome.ts DATA_DIR/PROFILE_DIR.
func DataDir() string {
	if C.ProfileIsDefault {
		return "data"
	}
	return "data/" + C.WAProfile
}

// dotenv holds KEY=VALUE from <RootDir>/.env, but only for keys the real
// environment does not already set (Node: `if env undefined` wins semantics).
var dotenv = make(map[string]string)

// Env reads the real environment first, then the .env overlay.
func Env(key string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return dotenv[key]
}

func rootDir() string {
	if r := os.Getenv("WA_ROOT_DIR"); r != "" {
		return r
	}
	// Walk upward from the cwd to the repo root (the dir that has .env).
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if fi, ferr := os.Stat(filepath.Join(dir, ".env")); ferr == nil && !fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return dir
		}
		dir = parent
	}
}

func loadDotenv(dir string) {
	raw, err := os.ReadFile(filepath.Join(dir, ".env"))
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(raw), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		eq := strings.Index(trimmed, "=")
		if eq < 0 {
			continue
		}
		key := strings.TrimSpace(trimmed[:eq])
		val := strings.TrimSpace(trimmed[eq+1:])
		if len(val) >= 2 {
			if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
				val = val[1 : len(val)-1]
			}
		}
		if _, ok := os.LookupEnv(key); !ok {
			dotenv[key] = val
		}
	}
}

// dbFromURL parses mysql://user:pass@host:port/db (also accepts mysql2://).
func dbFromURL(rawURL string) DB {
	if !strings.Contains(rawURL, "://") {
		rawURL = "mysql://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return DB{}
	}
	host := u.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}
	port := 3306
	if p := u.Port(); p != "" {
		if n, perr := strconv.Atoi(p); perr == nil {
			port = n
		}
	}
	user := ""
	pass := ""
	if u.User != nil {
		user, _ = url.PathUnescape(u.User.Username())
		pass, _ = url.PathUnescape(pw(u.User))
	}
	db := strings.Trim(strings.TrimPrefix(u.Path, "/"), "/")
	if db == "" {
		db = "ai"
	}
	return DB{Host: host, Port: port, User: user, Password: pass, Database: db}
}

func pw(ui *url.Userinfo) string {
	if ui == nil {
		return ""
	}
	p, _ := ui.Password()
	return p
}

// dbConfig mirrors src/config/index.ts dbConfig(): a single URL wins, else the DB_* vars.
func dbConfig() DB {
	u := Env("AI_DB_URL")
	if u == "" {
		u = Env("DATABASE_URL")
	}
	if u != "" {
		return dbFromURL(u)
	}
	port := 3306
	if p := Env("DB_PORT"); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			port = n
		}
	}
	return DB{
		Host:     Env("DB_HOST"),
		Port:     port,
		User:     Env("DB_USER"),
		Password: Env("DB_PASSWORD"),
		Database: Env("DB_NAME"),
	}
}

// carsConfig mirrors src/config/index.ts carsConfig().
func carsConfig() Cars {
	return Cars{
		BaseURL:  strings.TrimRight(strings.TrimSpace(Env("CARS_BASE_URL")), "/"),
		Tenant:   strings.TrimSpace(Env("CARS_TENANT")),
		APIToken: Env("CARS_API_TOKEN"),
	}
}

func init() {
	root := rootDir()
	loadDotenv(root)
	waProfile := Env("WA_PROFILE")
	if waProfile == "" {
		waProfile = "whatsapp-profile"
	}
	C = Config{
		RootDir:          root,
		DB:               dbConfig(),
		Cars:             carsConfig(),
		WAProfile:        waProfile,
		ProfileIsDefault: Env("WA_PROFILE") == "",
	}
}

// Root returns the repo root, anchored like Node's ROOT_DIR (env override or cwd).
func Root() string { return C.RootDir }
