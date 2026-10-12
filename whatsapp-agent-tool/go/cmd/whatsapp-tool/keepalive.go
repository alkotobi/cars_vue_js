package main

// keepalive — Go twin of src/keepalive.ts. Make sure the shared WhatsApp Chrome
// is up (spawned detached, so it survives this process), report status, or stop it.
//
//	whatsapp-tool keepalive start    # make sure Chrome is up, then wait
//	whatsapp-tool keepalive status   # is it up? is WhatsApp logged in?
//	whatsapp-tool keepalive stop     # kill the shared Chrome

import (
	"fmt"
	"os"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/chrome"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/wa"
)

func cmdKeepalive(cli wa.Client, args []string) {
	mode := "start"
	if len(args) > 0 && args[0] != "" {
		mode = strings.ToLower(args[0])
	}
	headless := chrome.DefaultHeadless()

	switch mode {
	case "start":
		keepaliveStart(cli, headless)
	case "status":
		keepaliveStatus(cli, headless)
	case "stop":
		keepaliveStop()
	default:
		fmt.Fprintf(os.Stderr, "[keepalive] unknown mode %q (use start | status | stop)\n", mode)
		os.Exit(1)
	}
}

func keepaliveStart(cli wa.Client, headless bool) {
	if chrome.CdpAlive(800 * time.Millisecond) {
		ep, _ := chrome.CdpEndpoint()
		fmt.Fprintf(os.Stderr, "[keepalive] Chrome already up on %s - reusing it.\n", ep)
	} else {
		if holder := chrome.ProfileLockedByOther(0); holder != 0 {
			fmt.Fprintf(os.Stderr, "[keepalive] profile is locked by another Chrome (pid %d). Close ./link_manual.sh first.\n", holder)
			os.Exit(1)
		}
		port, _ := chrome.CdpPort()
		fmt.Fprintf(os.Stderr, "[keepalive] starting Chrome on port %d (headless=%v)...\n", port, headless)
		res, err := chrome.SpawnChrome(headless, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, "[keepalive] fatal:", err)
			os.Exit(1)
		}
		if !chrome.WaitForCdp(30 * time.Second) {
			fmt.Fprintln(os.Stderr, "[keepalive] Chrome did not expose its debug port in time.")
			os.Exit(1)
		}
		ep, _ := chrome.CdpEndpoint()
		fmt.Fprintf(os.Stderr, "[keepalive] Chrome up, pid %d, endpoint %s\n", res.Pid, ep)
	}

	if err := cli.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "[keepalive] WhatsApp not ready: %v\n", err)
	} else if info, err := cli.SessionInfo(); err != nil {
		fmt.Fprintf(os.Stderr, "[keepalive] WhatsApp not ready: %v\n", err)
	} else if data, err := jsonx.Marshal(info); err == nil {
		fmt.Fprintf(os.Stderr, "[keepalive] WhatsApp ready: %s\n", data)
	}
	_ = cli.Close()

	fmt.Fprintln(os.Stderr, "[keepalive] Chrome keeps running. Ctrl+C here will NOT stop it.")
	fmt.Fprintln(os.Stderr, "[keepalive] stop it with: whatsapp-tool keepalive stop")
	select {}
}

func keepaliveStatus(cli wa.Client, headless bool) {
	alive := chrome.CdpAlive(800 * time.Millisecond)
	ep, _ := chrome.CdpEndpoint()
	port, _ := chrome.CdpPort()

	state := chrome.ReadState() // *ChromeState -> JSON object or null
	var locked any = nil
	if holder := chrome.ProfileLockedByOther(0); holder != 0 {
		locked = holder
	}
	printJSON(&ordered{Pairs: []orderedPair{
		pair("profile", chrome.Profile()),
		pair("chromeRunning", alive),
		pair("endpoint", ep),
		pair("port", port),
		pair("state", state),
		pair("profileLockedByOther", locked),
	}})
	if !alive {
		return
	}

	if err := cli.Init(); err != nil {
		printJSON(&ordered{Pairs: []orderedPair{pair("error", err.Error())}})
		return
	}
	defer func() { _ = cli.Close() }()
	info, err := cli.SessionInfo()
	if err != nil {
		printJSON(&ordered{Pairs: []orderedPair{pair("error", err.Error())}})
		return
	}
	loggedIn, _ := cli.IsLoggedIn()
	chats, _ := cli.GetChats(3)
	names := make([]string, 0, len(chats))
	for _, c := range chats {
		names = append(names, c.Name)
	}
	printJSON(&keepaliveReady{SessionInfo: info, LoggedIn: loggedIn, SampleChats: names})
}

// keepaliveReady mirrors Node's { ...sessionInfo, loggedIn, sampleChats }.
type keepaliveReady struct {
	wa.SessionInfo
	LoggedIn    bool     `json:"loggedIn"`
	SampleChats []string `json:"sampleChats"`
}

func keepaliveStop() {
	stopped, pid, how := chrome.StopChrome()
	pairs := []orderedPair{pair("stopped", stopped)}
	if pid > 0 {
		pairs = append(pairs, pair("pid", pid))
	}
	pairs = append(pairs, pair("how", how))
	printJSON(&ordered{Pairs: pairs})
	if stopped {
		os.Exit(0)
	}
	os.Exit(1)
}
