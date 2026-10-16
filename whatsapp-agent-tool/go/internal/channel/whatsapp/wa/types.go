// Package wa is the Go twin of src/whatsapp.ts — the browser-backed WhatsApp
// Web client. Stage 5 implements the actual rod/chromedp driver; this file
// carries the client CONTRACT plus the wire types the MCP layer serialises
// (shapes mirror src/whatsapp/types.ts and the whatsapp_* tool payloads).
package wa

// ChatInfo mirrors src/whatsapp/types.ts ChatInfo (screen rows).
type ChatInfo struct {
	Name        string `json:"name"`
	Preview     string `json:"preview,omitempty"`
	Time        string `json:"time,omitempty"`
	Unread      bool   `json:"unread"` // Node always emits this boolean
	UnreadCount string `json:"unreadCount,omitempty"`
}

// ChatSweepResult mirrors ChatSweepResult (whatsapp_get_chats newest/unread).
// unreadTotal and note appear only when meaningful, so they are pointers.
type ChatSweepResult struct {
	Chats       []ChatInfo `json:"chats"`
	Collected   int        `json:"collected"`
	Requested   int        `json:"requested"`
	Scrolls     int        `json:"scrolls"`
	ReachedEnd  bool       `json:"reachedEnd"`
	Filter      string     `json:"filter"` // "all" | "unread"
	UnreadTotal *int       `json:"unreadTotal,omitempty"`
	ElapsedMs   int64      `json:"elapsedMs"`
	Note        *string    `json:"note,omitempty"`
}

// MessageType mirrors the union in src/whatsapp/types.ts.
type MessageType string

const (
	MessageText    MessageType = "text"
	MessageVoice   MessageType = "voice"
	MessageAudio   MessageType = "audio"
	MessageImage   MessageType = "image"
	MessageVideo   MessageType = "video"
	MessageCall    MessageType = "call"
	MessageMissed  MessageType = "missed_call"
	MessageSystem  MessageType = "system"
	MessageUnknown MessageType = "unknown"
)

// Message mirrors Message (whatsapp_get_last_messages rows).
type Message struct {
	From string      `json:"from"` // "me" | "them" | "unknown"
	Text string      `json:"text"`
	Time string      `json:"time,omitempty"`
	Type MessageType `json:"type,omitempty"`
}

// SessionInfo mirrors WhatsAppClient.sessionInfo().
type SessionInfo struct {
	Profile         string `json:"profile"`
	Transport       string `json:"transport"` // "cdp" | "launch" | "none"
	AttachedOverCdp bool   `json:"attachedOverCdp"`
	Initialized     bool   `json:"initialized"`
	CdpEndpoint     string `json:"cdpEndpoint"`
	CdpPort         int    `json:"cdpPort"`
	ChromePid       any    `json:"chromePid"` // int | null
	ChromeRunning   bool   `json:"chromeRunning"`
	ClientUptimeSec int64  `json:"clientUptimeSec"`
	PageUrl         any    `json:"pageUrl"` // string | null
	Note            string `json:"note"`
}

// Client is the WhatsApp Web driver the CLIs and MCP layer talk to. It models
// the public surface of src/whatsapp.ts WhatsAppClient. Implemented by the
// browser layer (Stage 5, go-rod); until then the command returns
// ErrNotImplemented from every browser method.
type Client interface {
	// Init attaches to Chrome or launches it (idempotent).
	Init() error
	// RestoreAllFilter resets the sidebar to the All filter (an Unread filter
	// hides the "ai questions" group — see pending.ts Gotcha 22).
	RestoreAllFilter() error
	IsLoggedIn() (bool, error)
	GetChats(limit int) ([]ChatInfo, error)
	GetLastChats(limit int, timeoutMs int) (ChatSweepResult, error)
	GetLastUnreadChats(limit int, timeoutMs int) (ChatSweepResult, error)
	OpenChat(nameOrPhone string) (bool, error)
	SendMessage(text string) (bool, error)
	GetLastMessages(limit int) ([]Message, error)
	MarkAsRead() (bool, error)
	SessionInfo() (SessionInfo, error)
	// IsAttached reports whether Chrome survives this process (CDP transport).
	IsAttached() bool
	// Close detaches from Chrome (keeps it alive on the CDP transport).
	Close() error
}

// ErrNotImplemented is returned by driver stubs until Stage 5 lands.
var ErrNotImplemented = &notImplementedError{}

type notImplementedError struct{}

func (*notImplementedError) Error() string {
	return "browser layer not implemented yet (stage 5) — the WhatsApp_* driving tools come online there"
}
