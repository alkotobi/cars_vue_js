// Package chrome is the Go twin of src/chrome.ts — Chrome lifecycle for the
// persistent WhatsApp session.
//
// Chrome must outlive this process: we spawn it detached with
// --remote-debugging-port and attach over CDP (~200-500 ms) instead of paying a
// full cold boot per task. Every account (WA_PROFILE) gets its own profile
// directory, CDP port and state file.
package chrome

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
)

const (
	portRangeStart = 9222
	portRangeEnd   = 9252
	loginURL       = "https://web.whatsapp.com/"
)

// Profile returns the active account name (WA_PROFILE, defaulting to
// "whatsapp-profile").
func Profile() string { return config.WAProfile() }

// ProfileIsDefault mirrors chrome.ts PROFILE_IS_DEFAULT.
func ProfileIsDefault() bool { return config.ProfileIsDefault() }

// DataDir is the Chrome user-data-dir for this account.
func DataDir() string { return filepath.Join(config.Root(), "data", Profile()) }

// DefaultHeadless is opt-in (WA_HEADLESS=1|true|yes). Default is a visible
// window so the user can always see and take over the browser.
func DefaultHeadless() bool {
	switch strings.ToLower(config.Env("WA_HEADLESS")) {
	case "1", "true", "yes":
		return true
	}
	return false
}

// StateFile records the pid/port of the Chrome we spawned (outside the profile).
func StateFile() string {
	if ProfileIsDefault() {
		return filepath.Join(config.Root(), "data", "whatsapp-cdp.json")
	}
	return filepath.Join(config.Root(), "data", Profile()+"-cdp.json")
}

// ChromeState mirrors the chrome.ts ChromeState.
type ChromeState struct {
	Pid       int   `json:"pid"`
	Port      int   `json:"port"`
	StartedAt int64 `json:"startedAt"`
	Headless  bool  `json:"headless"`
}

var cachedPort int

// CdpPort is the CDP port that belongs to OUR Chrome. Order: WA_CDP_PORT -> the
// port in our state file (if alive) -> first free port in 9222-9252. We never
// blindly attach to 9222: driving a foreign browser would read/send from the
// wrong WhatsApp account.
func CdpPort() (int, error) {
	if cachedPort != 0 {
		return cachedPort, nil
	}
	if forced := config.Env("WA_CDP_PORT"); forced != "" {
		p, err := strconv.Atoi(forced)
		if err != nil || p == 0 {
			p = portRangeStart
		}
		cachedPort = p
		return cachedPort, nil
	}
	if st := ReadState(); st != nil && st.Port != 0 {
		if CdpAliveOn(st.Port, 500*time.Millisecond) && StatePortBelongsToUs(st) {
			// Reuse the state-file port only when the Chrome answering CDP on it is
			// really OUR profile's browser (same user-data-dir). A live endpoint is
			// not enough: another profile (profile switching, keepalive, a leftover
			// session) may have claimed the port, and SpawnChrome reusing it would
			// collide with that browser instead of attaching — which is how profiles
			// ended up killing and relaunching each other. The free-port scan below
			// can never DISCOVER an existing Chrome's port, so dropping a live one
			// here would strand a genuinely restarted-in-place browser; ownership is
			// still checked in that case (StatePortBelongsToUs goes to the OS) so we
			// keep trusting the port only when it is provably ours.
			cachedPort = st.Port
			return cachedPort, nil
		}
	}
	for port := portRangeStart; port <= portRangeEnd; port++ {
		if portFree(port) {
			cachedPort = port
			return cachedPort, nil
		}
	}
	return 0, fmt.Errorf("no free CDP port in %d-%d. Set WA_CDP_PORT", portRangeStart, portRangeEnd)
}

// ResetPortCache forgets the cached port (after a stop or re-resolve).
func ResetPortCache() { cachedPort = 0 }

// CdpEndpoint is http://127.0.0.1:<port>.
func CdpEndpoint() (string, error) {
	port, err := CdpPort()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://127.0.0.1:%d", port), nil
}

func portFree(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return false
	}
	_ = ln.Close()
	return true
}

// ChromePath resolves Google Chrome. WhatsApp Web refuses bundled Chromium.
func ChromePath() (string, error) {
	if override := config.Env("WA_CHROME_PATH"); override != "" {
		if _, err := os.Stat(override); err == nil {
			return override, nil
		}
	}
	candidates := []string{
		"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		"/Applications/Google Chrome Beta.app/Contents/MacOS/Google Chrome Beta",
		"/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
		"/usr/bin/google-chrome",
		"/usr/bin/google-chrome-stable",
		"/opt/google/chrome/chrome",
		"/usr/bin/chromium",
		"/usr/bin/chromium-browser",
		`C:\Program Files\Google\Chrome\Application\chrome.exe`,
		`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c, nil
		}
	}
	return "", fmt.Errorf("Google Chrome not found. Install Google Chrome, or set WA_CHROME_PATH to the binary")
}

// ReadState reads the recorded Chrome state, or nil.
func ReadState() *ChromeState {
	raw, err := os.ReadFile(StateFile())
	if err != nil {
		return nil
	}
	var st ChromeState
	if err := json.Unmarshal(raw, &st); err != nil {
		return nil
	}
	return &st
}

// WriteState persists the spawned Chrome's state.
func WriteState(st ChromeState) error {
	if err := os.MkdirAll(filepath.Dir(StateFile()), 0o755); err != nil {
		return err
	}
	data, err := jsonx.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StateFile(), data, 0o644)
}

// ClearState removes the state file (best effort).
func ClearState() { _ = os.Remove(StateFile()) }

// PidAlive reports whether a process exists.
func PidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

// listeningPid returns the pid of the process LISTENING on local port (0 when
// nothing listens or the OS query fails). Used to prove which profile owns a
// CDP port before we reuse it.
func listeningPid(port int) int {
	out, err := exec.Command("lsof", "-nP", fmt.Sprintf("-iTCP:%d", port), "-sTCP:LISTEN", "-t").Output()
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if pid, err := strconv.Atoi(line); err == nil && pid > 0 {
			return pid
		}
	}
	return 0
}

// cmdlineContains reports whether the running process pid's command line
// contains need (e.g. our --user-data-dir path). Only paths we can read are
// supported; on an unknown platform it falls back to "the process is alive".
func cmdlineContains(pid int, need string) bool {
	if pid <= 0 {
		return false
	}
	if need == "" {
		return PidAlive(pid)
	}
	var cmd string
	switch runtime.GOOS {
	case "darwin", "freebsd", "openbsd", "netbsd":
		out, err := exec.Command("ps", "-ww", "-p", strconv.Itoa(pid), "-o", "command=").Output()
		if err != nil {
			return false
		}
		cmd = string(out)
	case "linux":
		raw, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
		if err != nil {
			return false
		}
		cmd = strings.ReplaceAll(string(raw), "\x00", " ")
	default:
		return PidAlive(pid)
	}
	return strings.Contains(cmd, need)
}

// StatePortBelongsToUs reports whether the CDP endpoint on st.Port is this
// profile's Chrome. A live recorded pid is verified against our user-data-dir;
// a stale pid (Chrome restarted in place on the same port) is verified by
// asking the OS who actually LISTENS on the port and checking that process
// instead. When ownership cannot be proven we return false so the caller picks
// a fresh free port rather than colliding with another profile's browser.
func StatePortBelongsToUs(st *ChromeState) bool {
	if st.Pid > 0 && PidAlive(st.Pid) {
		return cmdlineContains(st.Pid, DataDir())
	}
	if p := listeningPid(st.Port); p > 0 {
		return cmdlineContains(p, DataDir())
	}
	return false
}

// ProfileLockedByOther returns the pid of another Chrome holding the profile
// lock (e.g. a running `npm run link`), or 0.
func ProfileLockedByOther(ourPid int) int {
	lock := filepath.Join(DataDir(), "SingletonLock")
	target, err := os.Readlink(lock)
	if err != nil {
		return 0
	}
	parts := strings.Split(target, "-")
	pid, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil || pid <= 0 {
		return 0
	}
	if ourPid > 0 && pid == ourPid {
		return 0
	}
	if PidAlive(pid) {
		return pid
	}
	return 0
}

func fetchVersion(port int, timeout time.Duration) (string, bool) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(fmt.Sprintf("http://127.0.0.1:%d/json/version", port))
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", false
	}
	var info struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", false
	}
	return info.WebSocketDebuggerURL, info.WebSocketDebuggerURL != ""
}

// CdpAliveOn reports whether a CDP-debuggable Chrome listens on port.
func CdpAliveOn(port int, timeout time.Duration) bool {
	_, ok := fetchVersion(port, timeout)
	return ok
}

// CdpWebSocketURL resolves the CDP websocket endpoint on port (for the driver).
func CdpWebSocketURL(port int, timeout time.Duration) (string, error) {
	ws, ok := fetchVersion(port, timeout)
	if !ok {
		return "", fmt.Errorf("no CDP endpoint on port %d", port)
	}
	return ws, nil
}

// CdpAlive reports whether our port has a CDP-debuggable Chrome.
func CdpAlive(timeout time.Duration) bool {
	port, err := CdpPort()
	if err != nil {
		return false
	}
	return CdpAliveOn(port, timeout)
}

// WaitForCdp polls our port until a CDP endpoint answers or the timeout hits.
func WaitForCdp(timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if CdpAlive(600 * time.Millisecond) {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// SpawnResult mirrors chrome.ts SpawnResult.
type SpawnResult struct {
	Pid      int
	Port     int
	Headless bool
}

// SpawnChrome spawns Google Chrome detached, opening url. The child is detached
// from our session so it keeps running after we exit.
func SpawnChrome(headless bool, url string) (*SpawnResult, error) {
	if err := os.MkdirAll(DataDir(), 0o755); err != nil {
		return nil, err
	}
	port, err := CdpPort()
	if err != nil {
		return nil, err
	}
	bin, err := ChromePath()
	if err != nil {
		return nil, err
	}
	if url == "" {
		url = loginURL
	}
	args := []string{
		fmt.Sprintf("--user-data-dir=%s", DataDir()),
		fmt.Sprintf("--remote-debugging-port=%d", port),
		// Chrome >=111 rejects CDP websocket connections carrying an Origin header.
		"--remote-allow-origins=*",
		"--no-first-run",
		"--no-default-browser-check",
		"--disable-blink-features=AutomationControlled",
		"--no-sandbox",
		"--disable-setuid-sandbox",
		"--disable-features=Translate,OptimizationHints,MediaRouter",
		"--window-size=1440,900",
	}
	if headless {
		args = append(args, "--headless=new")
	}
	args = append(args, url)

	cmd := exec.Command(bin, args...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	// Setsid detaches the child into its own session so it survives our exit.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	pid := cmd.Process.Pid
	// Release the handle; we never Wait (the process outlives us on purpose).
	_ = cmd.Process.Release()

	_ = WriteState(ChromeState{Pid: pid, Port: port, StartedAt: time.Now().UnixMilli(), Headless: headless})
	cachedPort = port
	return &SpawnResult{Pid: pid, Port: port, Headless: headless}, nil
}

// StopChrome kills the Chrome we spawned (for `keepalive:stop`).
func StopChrome() (stopped bool, pid int, how string) {
	st := ReadState()
	port := 0
	if st != nil {
		port = st.Port
	}
	if port == 0 {
		port, _ = CdpPort()
	}
	if st == nil || st.Pid <= 0 || !PidAlive(st.Pid) {
		ClearState()
		ResetPortCache()
		return false, 0, "no running Chrome for this profile"
	}
	pid = st.Pid
	_ = syscall.Kill(pid, syscall.SIGTERM)
	for i := 0; i < 20 && PidAlive(pid); i++ {
		time.Sleep(250 * time.Millisecond)
	}
	if PidAlive(pid) {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		time.Sleep(500 * time.Millisecond)
	}
	ClearState()
	ResetPortCache()
	return !PidAlive(pid), pid, fmt.Sprintf("sent SIGTERM/SIGKILL to pid %d (port %d)", pid, port)
}
