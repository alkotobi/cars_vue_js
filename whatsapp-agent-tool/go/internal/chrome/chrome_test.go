package chrome

import (
	"net"
	"os"
	"testing"
)

// TestCmdlineContainsThisProcess verifies cmdlineContains sees our own process's
// command line (the mechanism the ownership guard relies on).
func TestCmdlineContainsThisProcess(t *testing.T) {
	if len(os.Args) == 0 {
		t.Skip("no argv")
	}
	if !cmdlineContains(os.Getpid(), os.Args[0]) {
		t.Fatalf("expected own cmdline to contain %q", os.Args[0])
	}
	if cmdlineContains(os.Getpid(), "THIS-SENTINEL-SHOULD-NOT-EXIST-9f3a1b") {
		t.Fatal("sentinel unexpectedly found in own cmdline")
	}
}

// TestListeningPidFindsOwnListener verifies listeningPid identifies the process
// bound to a local TCP port (again, the mechanism behind the ownership guard).
func TestListeningPidFindsOwnListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	if pid := listeningPid(port); pid != os.Getpid() {
		t.Fatalf("listeningPid(%d) = %d, want %d", port, pid, os.Getpid())
	}
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	if pid := listeningPid(port); pid != 0 {
		t.Fatalf("listeningPid(%d) after close = %d, want 0", port, pid)
	}
}

// TestStatePortBelongsToUsRejectsForeignListener proves the guard refuses to
// reuse a port whose owner is not our profile's browser. Our own live pid is a
// valid listener that is provably NOT Chrome with our DataDir.
func TestStatePortBelongsToUsRejectsForeignListener(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port
	if StatePortBelongsToUs(&ChromeState{Pid: os.Getpid(), Port: port}) {
		t.Fatal("foreign (non-profile) listener must be rejected by StatePortBelongsToUs")
	}
}
