// Browser is the go-rod implementation of the wa.Client contract — the Go twin
// of src/whatsapp.ts WhatsAppClient. It attaches to the long-lived Chrome over
// CDP (spawning it detached when needed) and drives WhatsApp Web with the same
// human-like heuristics as the Node original.
package wa

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/proto"
	"github.com/ysmood/gson"

	"whatsapp-agent-tool/internal/channel/whatsapp/dom"
	"whatsapp-agent-tool/internal/chrome"
	"whatsapp-agent-tool/internal/config"
)

const defaultElementTimeout = 20 * time.Second

// Browser drives WhatsApp Web through Chrome's CDP endpoint.
type Browser struct {
	mu          sync.Mutex // serialises tool calls (the MCP server runs them concurrently)
	browser     *rod.Browser
	page        *rod.Page
	initialized bool
	transport   string // "cdp" | "launch" | "none"
	bornAt      time.Time
	headless    bool
}

// NewBrowser returns an unattached driver.
func NewBrowser() *Browser { return &Browser{bornAt: time.Now(), transport: "none"} }

var _ Client = (*Browser)(nil)

// --------------------------------------------------------------------- init

// Init attaches to Chrome or spawns it (idempotent).
func (b *Browser) Init() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.init()
}

func (b *Browser) init() error {
	if b.initialized && b.page != nil {
		return nil
	}
	b.page = nil
	b.browser = nil
	b.initialized = false

	b.headless = chrome.DefaultHeadless()
	transport := strings.ToLower(config.Env("WA_TRANSPORT"))
	if transport == "" {
		transport = "auto"
	}

	attached := false
	if transport != "launch" {
		attached = b.attachToRunningChrome()
	}
	if !attached && transport == "auto" {
		spawned, err := b.spawnAndAttach(b.headless)
		if err != nil {
			return err
		}
		attached = spawned
	}
	if !attached {
		if err := b.launchPersistent(b.headless); err != nil {
			return err
		}
	}
	if b.page == nil {
		return errors.New("could not obtain a WhatsApp Web page")
	}

	state, _ := b.waitForAppReady(120 * time.Second)
	if state == "blocked" {
		return errors.New("WhatsApp Web refused the bundled Chromium. Install/update Google Chrome (it is required).")
	}
	if state != "logged_in" {
		if b.headless {
			return errors.New("not linked yet. Run `WA_HEADLESS=0 whatsapp-tool keepalive start` in a terminal and scan the QR code with your phone.")
		}
		fmt.Fprintln(os.Stderr, "[whatsapp-mcp] QR code shown. On your phone: WhatsApp > Linked devices > Link a device.")
		if !b.waitForLoginInteractive(600 * time.Second) {
			return errors.New("linking failed or timed out. Re-run the link flow (visible window) and try again.")
		}
	}

	b.waitForStable()
	b.initialized = true
	pid := any("external")
	if st := chrome.ReadState(); st != nil {
		pid = st.Pid
	}
	endpoint, _ := chrome.CdpEndpoint()
	fmt.Fprintf(os.Stderr, "[whatsapp-mcp] ready (transport=%s, cdp=%s, chrome_pid=%v)\n", b.transport, endpoint, pid)
	return nil
}

func (b *Browser) attachToRunningChrome() bool {
	port, err := chrome.CdpPort()
	if err != nil {
		return false
	}
	if !chrome.CdpAliveOn(port, time.Second) {
		return false
	}
	ws, err := chrome.CdpWebSocketURL(port, time.Second)
	if err != nil {
		return false
	}
	b.browser = rod.New().ControlURL(ws)
	if err := b.browser.Connect(); err != nil {
		fmt.Fprintf(os.Stderr, "[whatsapp-mcp] CDP connect failed: %v\n", err)
		return false
	}
	b.transport = "cdp"

	for attempt := 1; attempt <= 4; attempt++ {
		if err := b.useWhatsAppPage(); err == nil {
			return true
		} else if attempt == 4 {
			fmt.Fprintf(os.Stderr, "[whatsapp-mcp] CDP attached but the page is not usable: %v\n", err)
		}
		time.Sleep(time.Duration(attempt) * 1500 * time.Millisecond)
	}
	return false
}

// spawnAndAttach spawns a detached Chrome on our port and attaches. Returns an
// error (never falling through to a legacy launch) when another Chrome already
// holds the profile and its debug port is not reachable — exactly like Node's
// WhatsAppClient.spawnAndAttach, which throws in that case.
func (b *Browser) spawnAndAttach(headless bool) (bool, error) {
	if holder := chrome.ProfileLockedByOther(0); holder != 0 {
		fmt.Fprintf(os.Stderr, "[whatsapp-mcp] profile busy: another Chrome (pid %d) is using %s.\n", holder, chrome.Profile())
		port, _ := chrome.CdpPort()
		if !chrome.CdpAliveOn(port, 600*time.Millisecond) {
			fmt.Fprintf(os.Stderr, "[whatsapp-mcp] %s is locked by another Chrome (npm run link?). Close that window and retry.\n", chrome.Profile())
			return false, fmt.Errorf("%s is locked by another Chrome (npm run link?). Close that window and retry.", chrome.Profile())
		}
		return b.attachToRunningChrome(), nil
	}

	if _, err := chrome.SpawnChrome(headless, dom.LoginURL); err != nil {
		fmt.Fprintf(os.Stderr, "[whatsapp-mcp] could not spawn Chrome: %v\n", err)
		return false, nil
	}
	if !chrome.WaitForCdp(30 * time.Second) {
		fmt.Fprintln(os.Stderr, "[whatsapp-mcp] Chrome started but never exposed its debug port.")
		return false, nil
	}
	return b.attachToRunningChrome(), nil
}

// launchPersistent is the legacy in-process path (WA_TRANSPORT=launch): we spawn
// our own Chrome and kill it on Close, so it dies with this process.
func (b *Browser) launchPersistent(headless bool) error {
	if holder := chrome.ProfileLockedByOther(0); holder != 0 {
		return fmt.Errorf("%s is locked by another Chrome (pid %d). Close that window and retry.", chrome.Profile(), holder)
	}
	if _, err := chrome.SpawnChrome(headless, dom.LoginURL); err != nil {
		return fmt.Errorf("could not start Google Chrome: %w. Check that no other Chrome is using %s", err, chrome.Profile())
	}
	if !chrome.WaitForCdp(30 * time.Second) {
		return errors.New("Chrome started but never exposed its debug port")
	}
	if !b.attachToRunningChrome() {
		return errors.New("attached Chrome did not expose a usable WhatsApp page")
	}
	b.transport = "launch"
	return nil
}

// useWhatsAppPage reuses the open WhatsApp tab, navigating only when absent.
func (b *Browser) useWhatsAppPage() error {
	if b.browser == nil {
		return errors.New("no browser")
	}
	pages, err := b.browser.Pages()
	if err != nil {
		return err
	}

	var best *rod.Page
	for _, p := range pages {
		info, ierr := p.Info()
		if ierr != nil || !strings.Contains(info.URL, "web.whatsapp.com") {
			continue
		}
		if st, serr := b.loginStateOn(p); serr == nil && st == "logged_in" {
			best = p
			break
		}
		if best == nil {
			best = p
		}
	}
	if best == nil {
		for _, p := range pages {
			if best == nil {
				best = p
			}
		}
	}
	if best == nil {
		np, cerr := b.browser.Page(proto.TargetCreateTarget{URL: dom.LoginURL})
		if cerr != nil {
			return cerr
		}
		best = np
	}
	b.page = best
	_ = proto.PageBringToFront{}.Call(best)

	state, _ := b.loginStateOn(best)
	if state == "loading" {
		b.waitForAppReady(120 * time.Second)
	}
	if !strings.Contains(b.pageURL(), "web.whatsapp.com") {
		_ = best.Navigate(dom.LoginURL)
		b.waitForAppReady(120 * time.Second)
	}
	return nil
}

func (b *Browser) pageURL() string {
	if b.page == nil {
		return ""
	}
	info, err := b.page.Info()
	if err != nil {
		return ""
	}
	return info.URL
}

func (b *Browser) loginState() string {
	if b.page == nil {
		return "loading"
	}
	st, err := b.loginStateOn(b.page)
	if err != nil {
		return "loading"
	}
	return st
}

func (b *Browser) loginStateOn(p *rod.Page) (string, error) {
	js := `() => {
      const txt = document.body?.innerText || "";
      const chatList = !!document.querySelector('#pane-side, [data-testid="chat-list"], div[aria-label="Search"], [role="grid"]');
      if (txt.includes("Chats") || chatList || txt.includes("Search")) return "logged_in";
      if (txt.includes("works with Google Chrome") || txt.includes("Update your browser")) return "blocked";
      if (txt.includes("Scan to log in") || txt.includes("Scan QR code") || txt.includes("Log in with phone number") || txt.includes("Link a device")) return "qr";
      if (txt.includes("6-digit PIN") || txt.includes("Two-step verification") || txt.includes("Enter your phone number")) return "pin";
      return "loading";
    }`
	j, err := evalOn(p, js)
	if err != nil {
		return "loading", err
	}
	return j.Str(), nil
}

func (b *Browser) waitForAppReady(timeout time.Duration) (string, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state := b.loginState()
		if state != "loading" {
			return state, nil
		}
		time.Sleep(2 * time.Second)
	}
	return "loading", nil
}

func (b *Browser) waitForLoginInteractive(timeout time.Duration) bool {
	if b.page == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	last := ""
	for time.Now().Before(deadline) {
		state := b.loginState()
		if state != last {
			fmt.Fprintf(os.Stderr, "[whatsapp-mcp] page state: %s\n", state)
			last = state
		}
		if state == "logged_in" {
			b.waitForStable()
			return true
		}
		if state == "pin" {
			fmt.Fprintln(os.Stderr, "[whatsapp-mcp] 2-step verification prompt detected - type your PIN in the Chrome window.")
		}
		if state == "blocked" {
			return false
		}
		time.Sleep(2 * time.Second)
	}
	return false
}

func (b *Browser) isQrVisible() bool {
	st := b.loginState()
	return st == "qr" || st == "pin"
}

func (b *Browser) waitForStable() {
	time.Sleep(time.Duration(dom.Jitter(550, 0.35)) * time.Millisecond)
}

// ensureInit heals a lost page in place, else re-inits from scratch.
func (b *Browser) ensureInit() error {
	if b.initialized && b.page != nil {
		return nil
	}
	if b.initialized && b.browser != nil {
		if err := b.useWhatsAppPage(); err == nil && b.page != nil {
			b.initialized = true
			return nil
		}
	}
	b.initialized = false
	return b.init()
}

// ---------------------------------------------------------------- public API

// IsLoggedIn reports whether the account is authenticated.
func (b *Browser) IsLoggedIn() (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.isLoggedIn()
}

func (b *Browser) isLoggedIn() (bool, error) {
	if err := b.ensureInit(); err != nil {
		return false, err
	}
	if b.page == nil {
		return false, nil
	}
	return !b.isQrVisible(), nil
}

// GetChats returns the rendered sidebar rows (screen mode).
func (b *Browser) GetChats(limit int) ([]ChatInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.getChats(limit)
}

func (b *Browser) getChats(limit int) ([]ChatInfo, error) {
	if err := b.ensureInit(); err != nil {
		return nil, err
	}
	if b.page == nil {
		return []ChatInfo{}, nil
	}
	b.clearChatSearch()
	chats := b.readRenderedChats()
	if limit < len(chats) {
		chats = chats[:limit]
	}
	return chats, nil
}

// RestoreAllFilter resets the sidebar to the All filter.
func (b *Browser) RestoreAllFilter() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.restoreAllFilter()
}

func (b *Browser) restoreAllFilter() error {
	if err := b.ensureInit(); err != nil {
		return err
	}
	if b.page == nil {
		return nil
	}
	if b.filterLooksActive("all") {
		return nil
	}
	if ok, _ := b.setChatFilter("all"); ok {
		b.waitForListSettled(8000)
	}
	return nil
}

// GetLastChats walks the sidebar (human-like) and collects up to limit chats.
func (b *Browser) GetLastChats(limit, timeoutMs int) (ChatSweepResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.getLastChats(limit, timeoutMs, "all")
}

// GetLastUnreadChats is the same walk with the Unread filter chip active.
func (b *Browser) GetLastUnreadChats(limit, timeoutMs int) (ChatSweepResult, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.getLastChats(limit, timeoutMs, "unread")
}

func (b *Browser) getLastChats(limit, timeoutMs int, filter string) (ChatSweepResult, error) {
	if err := b.ensureInit(); err != nil {
		return ChatSweepResult{}, err
	}
	started := time.Now()
	if b.page == nil {
		return ChatSweepResult{Chats: []ChatInfo{}, Requested: limit, Filter: filter, Note: strptr("no page")}, nil
	}
	b.clearChatSearch()

	if timeoutMs < 5000 {
		timeoutMs = 5000
	}
	if timeoutMs > 600000 {
		timeoutMs = 600000
	}
	timeout := time.Duration(timeoutMs) * time.Millisecond

	var unreadTotal *int
	appliedOK, total := b.setChatFilter(filter)
	if filter == "unread" {
		unreadTotal = total
	}
	if !appliedOK {
		return ChatSweepResult{
			Chats: []ChatInfo{}, Requested: limit, Filter: filter, UnreadTotal: unreadTotal,
			ElapsedMs: msSince(started),
			Note:      strptr(fmt.Sprintf(`could not activate WhatsApp's %s filter (data-testid="filter-button" not found).`, filter)),
		}, nil
	}

	b.growSidebarViewport()
	b.waitForListSettled(8000)
	b.scrollListToTop(120)

	seen := map[string]ChatInfo{}
	order := []string{}
	absorb := func(rendered []ChatInfo) {
		for _, c := range rendered {
			if _, ok := seen[c.Name]; !ok {
				seen[c.Name] = c
				order = append(order, c.Name)
			} else if c.Unread && !seen[c.Name].Unread {
				seen[c.Name] = c
			}
		}
	}

	scrolls := 0
	bottomHolds := 0
	reachedEnd := false
	var rendered []ChatInfo

	for time.Since(started) < timeout {
		rendered = b.readRenderedChats()
		absorb(rendered)
		if len(order) >= limit {
			break
		}
		step := b.humanScrollStep()
		scrolls++
		if step.atBottom && step.moved < 4 {
			bottomHolds++
			if bottomHolds >= 3 {
				absorb(b.readRenderedChats())
				reachedEnd = true
				break
			}
		} else {
			bottomHolds = 0
		}
	}
	if filter == "unread" {
		b.setChatFilter("all")
	}

	out := make([]ChatInfo, 0, len(order))
	for _, n := range order {
		out = append(out, seen[n])
	}
	if len(out) > limit {
		out = out[:limit]
	}
	collected := len(order)
	if collected > limit {
		collected = limit
	}
	return ChatSweepResult{
		Chats:       out,
		Collected:   collected,
		Requested:   limit,
		Scrolls:     scrolls,
		ReachedEnd:  reachedEnd,
		Filter:      filter,
		UnreadTotal: unreadTotal,
		ElapsedMs:   msSince(started),
		Note:        dom.BuildSweepNote(len(order), limit, reachedEnd, filter, unreadTotal),
	}, nil
}

// OpenChat opens a chat by name/phone. Ports WhatsAppClient.openChat.
func (b *Browser) OpenChat(nameOrPhone string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.openChat(nameOrPhone)
}

func (b *Browser) openChat(nameOrPhone string) (bool, error) {
	if err := b.ensureInit(); err != nil {
		return false, err
	}
	if b.page == nil {
		return false, nil
	}
	query := strings.TrimSpace(nameOrPhone)
	if query == "" {
		return false, nil
	}
	before := b.getCurrentChatName()
	if strings.Contains(dom.NormalizeTitle(before), dom.NormalizeTitle(query)) {
		b.clearChatSearch()
		return true, nil
	}

	for attempt := 0; attempt < 3; attempt++ {
		r := b.tryOpenChat(query, before)
		if r == "ok" {
			return true, nil
		}
		if r == "miss" {
			break
		}
		b.clearChatSearch()
		time.Sleep(400 * time.Millisecond)
	}
	b.clearChatSearch()
	return false, nil
}

func (b *Browser) tryOpenChat(query, before string) string {
	if b.page == nil {
		return "retry"
	}
	nQuery := dom.NormalizeTitle(query)
	dbg := func(args ...any) {
		if config.Env("WA_DEBUG") == "1" {
			fmt.Fprintln(os.Stderr, append([]any{"[openChat]"}, args...)...)
		}
	}

	preTitles := b.sidebarRowTitles()
	searchSel := `[data-testid="chat-list-search-container"] input, [data-testid="chat-list-search-container"] [contenteditable="true"]`

	typed := false
	for tries := 0; tries < 2 && !typed; tries++ {
		b.clickSelector(searchSel, 5*time.Second)
		time.Sleep(300 * time.Millisecond)
		b.clearSelector(searchSel)
		b.typeText(query, 40, 90)
		time.Sleep(400 * time.Millisecond)
		boxText := b.selectorValue(searchSel)
		typed = strings.Contains(dom.NormalizeTitle(boxText), nQuery)
		dbg(fmt.Sprintf("try=%d boxText=%q typed=%v", tries, boxText, typed))
		if !typed {
			composeSel := `[data-testid="conversation-compose-box-input"]`
			draft := strings.TrimSpace(b.selectorText(composeSel))
			if dom.NormalizeTitle(draft) == nQuery {
				b.clearSelector(composeSel)
			}
			b.clearChatSearch()
			time.Sleep(500 * time.Millisecond)
		}
	}
	if !typed {
		return "retry"
	}

	rows := preTitles
	matchedIdx := -1
	shrank := false
	deadline := time.Now().Add(6 * time.Second)
	for time.Now().Before(deadline) {
		rows = b.sidebarRowTitles()
		if len(rows) < len(preTitles) {
			shrank = true
		}
		matchedIdx = indexWhere(rows, nQuery)
		dbg(fmt.Sprintf("rows=%d matchedIdx=%d shrank=%v", len(rows), matchedIdx, shrank))
		if matchedIdx >= 0 || (shrank && len(rows) == 0) {
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	dbg(fmt.Sprintf("pre=%d rows=%d matchedIdx=%d shrank=%v titles=%v", len(preTitles), len(rows), matchedIdx, shrank, rows))

	if matchedIdx < 0 && shrank && len(rows) == 0 {
		time.Sleep(700 * time.Millisecond)
		rows = b.sidebarRowTitles()
		if len(rows) == 0 {
			return "miss"
		}
		matchedIdx = indexWhere(rows, nQuery)
	}
	if matchedIdx < 0 && (!shrank || len(rows) == 0) {
		return "retry"
	}
	useIdx := matchedIdx
	if useIdx < 0 {
		useIdx = 0
	}

	click := func(idx int) bool {
		if !b.clickRowAt(idx) {
			return false
		}
		return b.conversationReady(6 * time.Second)
	}

	if !click(useIdx) {
		again := b.sidebarRowTitles()
		idx := indexWhere(again, nQuery)
		if idx < 0 || !click(idx) {
			return "retry"
		}
	}
	b.clearChatSearch()
	b.waitForStable()

	after := b.getCurrentChatName()
	nAfter := dom.NormalizeTitle(after)
	dbg(fmt.Sprintf("verify query=%q before=%q after=%q matched=%v", nQuery, dom.NormalizeTitle(before), nAfter, matchedIdx >= 0))
	if nAfter == "" {
		return "retry"
	}
	if strings.Contains(nAfter, nQuery) {
		return "ok"
	}
	if matchedIdx < 0 && nAfter != dom.NormalizeTitle(before) {
		return "ok"
	}
	return "retry"
}

// SendMessage types text into the open conversation and presses Enter.
func (b *Browser) SendMessage(text string) (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sendMessage(text)
}

func (b *Browser) sendMessage(text string) (bool, error) {
	if awaitingConfirmation() {
		return false, errors.New("send blocked by WA_SEND_DISABLED (dry-run)")
	}
	if err := b.ensureInit(); err != nil {
		return false, err
	}
	if b.page == nil {
		return false, nil
	}
	if strings.TrimSpace(text) == "" {
		return false, nil
	}

	compose := `[data-testid="conversation-compose-box-input"]`
	if !b.clickSelector(compose, 8*time.Second) {
		time.Sleep(time.Duration(dom.Jitter(1200, 0.3)) * time.Millisecond)
		b.focusSelector(compose)
		time.Sleep(time.Duration(dom.Jitter(400, 0.3)) * time.Millisecond)
	}

	b.waitForStable()
	b.clearSelector(compose)
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for i, line := range lines {
		if i > 0 {
			b.pressShiftEnter()
		}
		b.typeText(line, 20, 60)
	}
	time.Sleep(time.Duration(dom.Jitter(150, 0.4)) * time.Millisecond)
	b.pressEnter()
	time.Sleep(time.Duration(dom.Jitter(700, 0.4)) * time.Millisecond)
	return true, nil
}

// GetLastMessages reads the open conversation's recent messages.
func (b *Browser) GetLastMessages(limit int) ([]Message, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.getLastMessages(limit)
}

func (b *Browser) getLastMessages(limit int) ([]Message, error) {
	if err := b.ensureInit(); err != nil {
		return nil, err
	}
	if b.page == nil {
		return []Message{}, nil
	}

	js := `() => {
      const panel = document.querySelector('[data-testid="conversation-panel-messages"]');
      const panelRect = panel?.getBoundingClientRect();
      const dir = panel ? getComputedStyle(panel).direction : "ltr";
      const center = panelRect ? panelRect.x + panelRect.width / 2 : 0;
      const containers = Array.from(document.querySelectorAll('[data-testid="conversation-panel-messages"] [data-testid="msg-container"]'));
      const els = containers.length > 0 ? containers : Array.from(document.querySelectorAll('[data-testid="msg-container"]'));
      return {
        dir,
        hasPanel: !!panelRect && panelRect.width > 0,
        center,
        items: els.map((e) => {
          const r = e.getBoundingClientRect();
          const st = e.querySelector('[data-testid="selectable-text"]');
          const mt = e.querySelector('[data-testid="msg-meta"]');
          // Media/system rows carry no selectable text; detect them in the DOM
          // so a voice note or a missed call is never dropped from the read.
          const has = (sel) => !!e.querySelector(sel);
          return {
            text: st ? (st.innerText || "") : "",
            meta: mt ? (mt.innerText || "") : "",
            audio: has('[data-testid="audio-player"],[data-testid="msg-audio"],[data-icon="audio-play"],[data-icon="audio-pause"],audio'),
            image: has('[data-testid="image-thumb"],[data-testid="media-viewer"],img:not([data-testid*="emoji"])'),
            video: has('video,[data-testid="video-player"],[data-icon="video-play"]'),
            system: has('[data-testid="system-message"],[data-testid="message-system"],[data-testid="conversation-info-message"]'),
            tailOut: !!e.querySelector('[data-testid="tail-out"]'),
            tailIn: !!e.querySelector('[data-testid="tail-in"]'),
            bubbleCenter: r.x + r.width / 2,
            width: r.width,
          };
        }),
      };
    }`
	raw, err := b.eval(js)
	if err != nil {
		return []Message{}, nil
	}
	panel := dom.RawPanel{
		Dir:      raw.Get("dir").Str(),
		HasPanel: raw.Get("hasPanel").Bool(),
		Center:   raw.Get("center").Num(),
	}
	items := raw.Get("items").Arr()
	messages := make([]Message, 0, len(items))
	for _, it := range items {
		text := strings.TrimSpace(dom.CleanText(it.Get("text").Str()))
		audio := it.Get("audio").Bool()
		image := it.Get("image").Bool()
		video := it.Get("video").Bool()
		system := it.Get("system").Bool()
		if text == "" && !audio && !image && !video && !system {
			continue // neither text nor media — nothing to read
		}
		m := dom.RawMessage{
			Text:         text,
			Meta:         it.Get("meta").Str(),
			TailOut:      it.Get("tailOut").Bool(),
			TailIn:       it.Get("tailIn").Bool(),
			BubbleCenter: it.Get("bubbleCenter").Num(),
			Width:        it.Get("width").Num(),
		}
		from := "them"
		if dom.IsOutgoing(m, panel) {
			from = "me"
		}
		meta := strings.TrimSpace(dom.CleanText(m.Meta))
		extra := dom.MessageExtra{HasAudio: &audio, HasImage: &image, HasVideo: &video, IsSystem: &system}
		kind := dom.InferMessageType(text, &extra)
		if kind == "audio" {
			kind = "voice" // WhatsApp audio bubbles are voice notes (mic icon); audio files arrive as documents
		}
		if text == "" {
			text = mediaPlaceholder(kind)
		}
		messages = append(messages, Message{
			From: from,
			Text: text,
			Time: meta,
			Type: MessageType(kind),
		})
	}
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	return messages, nil
}

// mediaPlaceholder labels a bubble that carries no selectable text (voice note,
// photo, call row) so reports and prompts still show what the message IS.
func mediaPlaceholder(kind string) string {
	switch kind {
	case "voice":
		return "[voice message]"
	case "audio":
		return "[audio]"
	case "image":
		return "[image]"
	case "video":
		return "[video]"
	case "call":
		return "[call]"
	case "missed_call":
		return "[missed call]"
	case "system":
		return "[system message]"
	default:
		return "[message]"
	}
}

// MarkAsRead is a best-effort read receipt (open the header, then Escape).
func (b *Browser) MarkAsRead() (bool, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.markAsRead()
}

func (b *Browser) markAsRead() (bool, error) {
	if err := b.ensureInit(); err != nil {
		return false, err
	}
	if b.page == nil {
		return false, nil
	}
	header := `[data-testid="conversation-info-header-chat-title"]`
	el, err := b.page.Timeout(3 * time.Second).Element(header)
	if err != nil {
		return false, nil
	}
	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return false, nil
	}
	time.Sleep(400 * time.Millisecond)
	_ = b.page.Keyboard.Type(input.Escape)
	time.Sleep(time.Duration(dom.Jitter(300, 0.3)) * time.Millisecond)
	return true, nil
}

// SessionInfo reports how we are talking to Chrome.
func (b *Browser) SessionInfo() (SessionInfo, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sessionInfo()
}

func (b *Browser) sessionInfo() (SessionInfo, error) {
	port, _ := chrome.CdpPort()
	state := chrome.ReadState()
	var pid any
	if state != nil {
		pid = state.Pid
	}
	var note string
	switch b.transport {
	case "cdp":
		note = "Attached over CDP: Chrome survives this process, next task reuses it."
	case "launch":
		note = "Chrome launched by this process (WA_TRANSPORT=launch): it dies with this process."
	default:
		note = "No browser attached yet - the next tool call attaches to Chrome or starts it."
	}
	var pageURL any
	if u := b.pageURL(); u != "" {
		pageURL = u
	}
	return SessionInfo{
		Profile:         chrome.Profile(),
		Transport:       b.transport,
		AttachedOverCdp: b.transport == "cdp",
		Initialized:     b.initialized,
		CdpEndpoint:     fmt.Sprintf("http://127.0.0.1:%d", port),
		CdpPort:         port,
		ChromePid:       pid,
		ChromeRunning:   chrome.CdpAliveOn(port, 600*time.Millisecond),
		ClientUptimeSec: int64(time.Since(b.bornAt).Seconds()),
		PageUrl:         pageURL,
		Note:            note,
	}, nil
}

// IsAttached reports whether Chrome survives this process (CDP transport).
func (b *Browser) IsAttached() bool { return b.transport == "cdp" }

// Close detaches from Chrome (keeps it alive on the CDP transport).
func (b *Browser) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.close()
}

func (b *Browser) close() error {
	if b.transport == "cdp" {
		fmt.Fprintln(os.Stderr, "[whatsapp-mcp] detaching from Chrome (the browser keeps running).")
		// Do NOT call b.browser.Close(): rod's Close sends the CDP "Browser.close"
		// command, which quits Chrome — killing the very browser this transport
		// exists to keep alive. We only drop our connection; the OS closes the
		// socket when this process exits and the detached Chrome survives.
	} else if b.transport == "launch" {
		if b.browser != nil {
			_ = b.browser.Close()
		}
		chrome.StopChrome()
	}
	b.page = nil
	b.browser = nil
	b.transport = "none"
	b.initialized = false
	return nil
}

// ------------------------------------------------------------------ helpers

func (b *Browser) eval(js string, args ...any) (gson.JSON, error) {
	if b.page == nil {
		return gson.JSON{}, errors.New("no page")
	}
	return evalOn(b.page, js, args...)
}

func evalOn(p *rod.Page, js string, args ...any) (gson.JSON, error) {
	obj, err := p.Timeout(defaultElementTimeout).Eval(js, args...)
	if err != nil {
		return gson.JSON{}, err
	}
	return obj.Value, nil
}

// readRenderedChats scrapes every materialised sidebar row.
func (b *Browser) readRenderedChats() []ChatInfo {
	if b.page == nil {
		return nil
	}
	js := `() => {
      const rows = Array.from(document.querySelectorAll('[data-testid="chat-list"] [data-testid^="list-item-"]'));
      return rows.map((row) => ({
        title: row.querySelector('[data-testid="cell-frame-title"]')?.innerText || "",
        secondary: row.querySelector('[data-testid="cell-frame-secondary"]')?.innerText || "",
        unread: !!row.querySelector('[data-testid="icon-unread-count"]'),
        unreadCount: row.querySelector('[data-testid="icon-unread-count"]')?.innerText || "",
      }));
    }`
	raw, err := b.eval(js)
	if err != nil {
		return nil
	}
	chats := []ChatInfo{}
	for _, row := range raw.Arr() {
		name := dom.ChatNameFrom(row.Get("title").Str())
		if name == "" {
			continue
		}
		sec := dom.Lines(row.Get("secondary").Str())
		var tv string
		preview := strings.Join(sec, " ")
		if last := lastOrEmpty(sec); last != "" && len(sec) > 1 && dom.TimeReTest(last) {
			tv = last
			preview = strings.Join(sec[:len(sec)-1], " ")
		}
		preview = dom.TrimPreviewNumber(preview)
		ci := ChatInfo{Name: name, Time: tv, Unread: row.Get("unread").Bool()}
		if preview != "" {
			ci.Preview = preview
		}
		if ci.Unread {
			if uc := strings.TrimSpace(row.Get("unreadCount").Str()); uc != "" {
				ci.UnreadCount = uc
			}
		}
		chats = append(chats, ci)
	}
	return chats
}

func (b *Browser) sidebarRowTitles() []string {
	if b.page == nil {
		return nil
	}
	js := `() => Array.from(document.querySelectorAll('[data-testid="chat-list"] [data-testid^="list-item-"]')).map((r) => ((r.querySelector('[data-testid="cell-frame-title"]')?.textContent) || "").trim())`
	raw, err := b.eval(js)
	if err != nil {
		return nil
	}
	out := []string{}
	for _, v := range raw.Arr() {
		out = append(out, v.Str())
	}
	return out
}

func (b *Browser) getCurrentChatName() string {
	if b.page == nil {
		return ""
	}
	raw, err := b.eval(`() => {
      const el = document.querySelector('[data-testid="conversation-info-header-chat-title"]');
      return el ? (el.textContent || "") : "";
    }`)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(raw.Str())
}

// clearChatSearch empties the chat-list search box.
func (b *Browser) clearChatSearch() {
	if b.page == nil {
		return
	}
	sel := `[data-testid="chat-list-search-container"] input, [data-testid="chat-list-search-container"] [contenteditable="true"]`
	current := b.selectorValue(sel)
	if strings.TrimSpace(current) == "" {
		return
	}
	b.clickSelector(sel, 3*time.Second)
	if !b.clearSelector(sel) {
		_ = b.page.Keyboard.Type(input.Backspace)
	}
	time.Sleep(400 * time.Millisecond)
}

// clickSelector clicks the first match of sel; returns whether it clicked.
func (b *Browser) clickSelector(sel string, timeout time.Duration) bool {
	if b.page == nil {
		return false
	}
	el, err := b.page.Timeout(timeout).Element(sel)
	if err != nil {
		return false
	}
	return el.Click(proto.InputMouseButtonLeft, 1) == nil
}

func (b *Browser) focusSelector(sel string) {
	if b.page == nil {
		return
	}
	if el, err := b.page.Timeout(5 * time.Second).Element(sel); err == nil {
		_ = el.Focus()
	}
}

func (b *Browser) selectorText(sel string) string {
	if b.page == nil {
		return ""
	}
	el, err := b.page.Timeout(2 * time.Second).Element(sel)
	if err != nil {
		return ""
	}
	t, err := el.Text()
	if err != nil {
		return ""
	}
	return t
}

// selectorValue reads the value of an input/contenteditable (for the search box).
func (b *Browser) selectorValue(sel string) string {
	if b.page == nil {
		return ""
	}
	el, err := b.page.Timeout(2 * time.Second).Element(sel)
	if err != nil {
		return ""
	}
	v, _ := el.Eval(`() => this.value ?? this.textContent ?? ""`)
	if v == nil {
		return ""
	}
	return v.Value.Str()
}

// clearSelector empties an input/contenteditable and fires an input event.
func (b *Browser) clearSelector(sel string) bool {
	if b.page == nil {
		return false
	}
	el, err := b.page.Timeout(3 * time.Second).Element(sel)
	if err != nil {
		return false
	}
	_, err = el.Eval(`() => {
      this.focus();
      if (this.tagName === 'INPUT' || this.tagName === 'TEXTAREA') {
        const setter = Object.getOwnPropertyDescriptor(this.constructor.prototype, 'value')?.set;
        if (setter) setter.call(this, ''); else this.value = '';
        this.dispatchEvent(new Event('input', { bubbles: true }));
      } else {
        this.textContent = '';
        this.dispatchEvent(new InputEvent('input', { bubbles: true }));
      }
      return true;
    }`)
	return err == nil
}

func (b *Browser) conversationReady(timeout time.Duration) bool {
	if b.page == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		raw, err := b.eval(`() => {
          const title = document.querySelector('[data-testid="conversation-info-header-chat-title"]');
          const compose = document.querySelector('[data-testid="conversation-compose-box-input"]');
          return !!(title && compose);
        }`)
		if err == nil && raw.Bool() {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}

// clickRowAt scrolls the nth chat row into view and clicks its centre.
func (b *Browser) clickRowAt(idx int) bool {
	if b.page == nil {
		return false
	}
	raw, err := b.eval(`(i) => {
      const rs = Array.from(document.querySelectorAll('[data-testid="chat-list"] [data-testid^="list-item-"]'));
      const row = rs[i];
      if (!row) return null;
      row.scrollIntoView({ block: "center" });
      const r = row.getBoundingClientRect();
      return { x: r.x + r.width / 2, y: r.y + r.height / 2 };
    }`, idx)
	if err != nil || raw.Nil() {
		return false
	}
	x, y := raw.Get("x").Num(), raw.Get("y").Num()
	if err := b.page.Mouse.MoveTo(proto.Point{X: x, Y: y}); err != nil {
		return false
	}
	return b.page.Mouse.Click(proto.InputMouseButtonLeft, 1) == nil
}

// typeText dispatches real key events per rune (works for Arabic too).
func (b *Browser) typeText(s string, minDelay, maxDelay int) {
	if b.page == nil {
		return
	}
	for _, r := range s {
		t := string(r)
		_ = proto.InputDispatchKeyEvent{
			Type:           proto.InputDispatchKeyEventTypeKeyDown,
			Key:            t,
			Text:           t,
			UnmodifiedText: t,
		}.Call(b.page)
		time.Sleep(time.Duration(dom.Rand(minDelay, maxDelay)) * time.Millisecond)
		_ = proto.InputDispatchKeyEvent{
			Type: proto.InputDispatchKeyEventTypeKeyUp,
			Key:  t,
		}.Call(b.page)
	}
}

func (b *Browser) pressEnter() {
	if b.page == nil {
		return
	}
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyDown, Key: "\r", Code: "Enter", WindowsVirtualKeyCode: 13}.Call(b.page)
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyUp, Key: "\r", Code: "Enter", WindowsVirtualKeyCode: 13}.Call(b.page)
}

func (b *Browser) pressShiftEnter() {
	if b.page == nil {
		return
	}
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyDown, Key: "Shift", Code: "ShiftLeft", WindowsVirtualKeyCode: 16, Modifiers: 8}.Call(b.page)
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyDown, Key: "\r", Code: "Enter", WindowsVirtualKeyCode: 13, Modifiers: 8}.Call(b.page)
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyUp, Key: "\r", Code: "Enter", WindowsVirtualKeyCode: 13, Modifiers: 8}.Call(b.page)
	_ = proto.InputDispatchKeyEvent{Type: proto.InputDispatchKeyEventTypeKeyUp, Key: "Shift", Code: "ShiftLeft", WindowsVirtualKeyCode: 16}.Call(b.page)
}

// --------------------------------------------------------------- sidebar walk

type scrollStep struct {
	atBottom bool
	moved    float64
}

type listGeo struct {
	x, y, width, height, top, scrollHeight, renderedRows float64
}

func (b *Browser) listGeometry() *listGeo {
	if b.page == nil {
		return nil
	}
	raw, err := b.eval(`() => {
      const list = document.querySelector('[data-testid="chat-list"]');
      if (!list) return null;
      let scroller = list;
      let el = list;
      while (el && el !== document.body) {
        const oy = getComputedStyle(el).overflowY;
        if (oy === "auto" || oy === "scroll") { scroller = el; break; }
        el = el.parentElement;
      }
      const s = scroller || list;
      const r = s.getBoundingClientRect();
      return {
        x: r.x, y: r.y, width: r.width, height: r.height,
        top: s.scrollTop, scrollHeight: s.scrollHeight,
        renderedRows: document.querySelectorAll('[data-testid="chat-list"] [data-testid^="list-item-"]').length,
      };
    }`)
	if err != nil || raw.Nil() {
		return nil
	}
	return &listGeo{
		x: raw.Get("x").Num(), y: raw.Get("y").Num(),
		width: raw.Get("width").Num(), height: raw.Get("height").Num(),
		top: raw.Get("top").Num(), scrollHeight: raw.Get("scrollHeight").Num(),
		renderedRows: raw.Get("renderedRows").Num(),
	}
}

func (b *Browser) humanScrollStep() scrollStep {
	g := b.listGeometry()
	if g == nil {
		return scrollStep{atBottom: true}
	}
	topBefore := g.top

	_ = b.page.Mouse.MoveTo(proto.Point{
		X: g.x + float64(dom.Rand(20, maxInt(21, int(g.width-20)))),
		Y: g.y + float64(dom.Rand(15, maxInt(16, int(g.height/2)))),
	})

	flick := rand.Float64() < 0.18
	var factor float64
	if flick {
		factor = float64(dom.Rand(110, 145)) / 100
	} else {
		factor = float64(dom.Rand(55, 105)) / 100
	}
	_ = b.page.Mouse.Scroll(0, float64(int(g.height*factor+0.5)), 1)

	wait := dom.Jitter(dom.Rand(380, 950), 0.45)
	if rand.Float64() < 0.12 {
		wait += dom.Rand(600, 1600)
	}
	time.Sleep(time.Duration(wait) * time.Millisecond)

	if rand.Float64() < 0.25 {
		_ = b.page.Mouse.Scroll(0, -float64(dom.Rand(30, 140)), 1)
		time.Sleep(time.Duration(dom.Jitter(dom.Rand(120, 320), 0.4)) * time.Millisecond)
	}

	after := b.listGeometry()
	if after == nil {
		return scrollStep{atBottom: true}
	}
	return scrollStep{
		atBottom: after.top+after.height >= after.scrollHeight-12,
		moved:    after.top - topBefore,
	}
}

func (b *Browser) waitForListSettled(maxMs int) {
	deadline := time.Now().Add(time.Duration(maxMs) * time.Millisecond)
	lastHeight := -1.0
	stable := 0
	for time.Now().Before(deadline) {
		var height float64
		if g := b.listGeometry(); g != nil {
			height = g.scrollHeight
		}
		if height > 0 && height == lastHeight {
			stable++
			if stable >= 2 {
				return
			}
		} else {
			stable = 0
		}
		lastHeight = height
		time.Sleep(time.Duration(dom.Jitter(450, 0.3)) * time.Millisecond)
	}
}

func (b *Browser) scrollListToTop(maxSteps int) {
	if b.page == nil {
		return
	}
	stalls := 0
	for i := 0; i < maxSteps; i++ {
		g := b.listGeometry()
		if g == nil || g.top <= 2 || g.scrollHeight <= g.height+4 {
			return
		}
		_ = b.page.Mouse.MoveTo(proto.Point{
			X: g.x + float64(dom.Rand(20, maxInt(21, int(g.width-20)))),
			Y: g.y + float64(dom.Rand(15, maxInt(16, int(g.height/2)))),
		})
		before := g.top
		_ = b.page.Mouse.Scroll(0, -float64(int(g.height*float64(dom.Rand(2, 3)))), 1)
		time.Sleep(time.Duration(dom.Jitter(dom.Rand(180, 420), 0.4)) * time.Millisecond)

		after := b.listGeometry()
		if after == nil || after.top >= before-2 {
			stalls++
			if stalls >= 2 {
				return
			}
		} else {
			stalls = 0
		}
	}
}

func (b *Browser) growSidebarViewport() {
	g := b.listGeometry()
	if g == nil || g.height > 700 {
		return
	}
	if b.page == nil {
		return
	}
	size, err := b.eval(`() => ({ w: window.innerWidth, h: window.innerHeight })`)
	if err == nil && size.Get("h").Num() >= 1200 {
		return
	}
	_ = b.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: 1600, Height: 1300, DeviceScaleFactor: 1,
	})
	time.Sleep(time.Duration(dom.Jitter(500, 0.3)) * time.Millisecond)
}

func (b *Browser) setChatFilter(which string) (bool, *int) {
	if b.page == nil {
		return false, nil
	}
	raw, err := b.eval(`() => Array.from(document.querySelectorAll('[data-testid="filter-button"]')).map((e) => (e.innerText || "").replace(/\s+/g, " ").trim())`)
	if err != nil || len(raw.Arr()) == 0 {
		return false, nil
	}
	state := []string{}
	for _, v := range raw.Arr() {
		state = append(state, v.Str())
	}

	index := -1
	for i, label := range state {
		if which == "unread" {
			if dom.MatchUnread(label) {
				index = i
				break
			}
		} else if dom.MatchAll(label) {
			index = i
			break
		}
	}
	if index < 0 {
		if which == "unread" {
			index = minInt(1, len(state)-1)
		} else {
			index = 0
		}
	}
	total := dom.ParseArabicDigits(state[index])

	b.clickFilterButton(index)
	time.Sleep(time.Duration(dom.Jitter(900, 0.35)) * time.Millisecond)

	ok := b.filterLooksActive(which)
	if !ok {
		fallback := 0
		if which == "unread" {
			fallback = minInt(1, len(state)-1)
		}
		if fallback != index {
			b.clickFilterButton(fallback)
			time.Sleep(time.Duration(dom.Jitter(900, 0.35)) * time.Millisecond)
			ok = b.filterLooksActive(which)
		}
	}
	return ok, total
}

func (b *Browser) clickFilterButton(index int) {
	// WhatsApp wraps each sidebar filter chip in a div[data-testid="filter-button"]
	// (role="none"); the real control is an inner button[role="tab"]. Clicking the
	// wrapper is a no-op (the filter never switches), so click the inner button.
	_, _ = b.eval(`(i) => {
      const chip = document.querySelectorAll('[data-testid="filter-button"]')[i];
      if (!chip) return false;
      const btn = chip.querySelector('button[role="tab"]') || chip;
      btn.click();
      return true;
    }`, index)
}

func (b *Browser) filterLooksActive(which string) bool {
	if b.page == nil {
		return false
	}
	raw, err := b.eval(`() => {
      const list = Array.from(document.querySelectorAll('[data-testid="chat-list"] [data-testid^="list-item-"]'));
      return {
        total: list.length,
        unread: list.filter((r) => !!r.querySelector('[data-testid="icon-unread-count"]')).length,
      };
    }`)
	if err != nil {
		return false
	}
	total := int(raw.Get("total").Num())
	unread := int(raw.Get("unread").Num())
	if total == 0 {
		return false
	}
	if which == "unread" {
		return unread == total
	}
	return unread < total
}

// ----------------------------------------------------------------- tiny util

func strptr(s string) *string { return &s }

func msSince(t time.Time) int64 { return time.Since(t).Milliseconds() }

func lastOrEmpty(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[len(s)-1]
}

func indexWhere(rows []string, nQuery string) int {
	for i, t := range rows {
		if strings.Contains(dom.NormalizeTitle(t), nQuery) {
			return i
		}
	}
	return -1
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// awaitingConfirmation gates real sends for dry-run smoke tests.
func awaitingConfirmation() bool {
	switch strings.ToLower(config.Env("WA_SEND_DISABLED")) {
	case "1", "true", "yes":
		return true
	}
	return false
}
