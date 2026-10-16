package chatbot

// Group is one conversation the bot may answer.
type Group struct {
	ID   int64
	Name string
}

// Message is one message of a conversation, shaped so the prompt can show who
// wrote what. It is the loop's channel-agnostic wire type: the JSON tags match
// the cars-app `chat_messages` row, and any other channel's Transport produces
// the same shape. bot.go and prompt.go depend on this type — never on a
// transport's own models — which is what lets one loop serve several channels.
type Message struct {
	ID         int64  `json:"id"`
	GroupID    int64  `json:"id_chat_group"`
	Text       string `json:"message"`
	Time       string `json:"time"`
	SenderName string `json:"sender_username"`
	SenderRole string `json:"sender_role"`
	SenderType string `json:"sender_type"`
}

// Transport is the channel-specific half of the poll loop: everything that
// talks to the destination. PollOnce owns only coordination — the single-writer
// lease, the per-group cursors, the cycle report and the alarm — so a second
// autonomous channel reuses the whole loop by implementing this one interface
// and passing it to PollOnce. Implementations live beside their channel
// (carsTransport in transport_cars.go is the cars-app one).
//
// The WhatsApp channel deliberately does NOT implement it. It has no Go polling
// loop at all: the operator's opencode session drives it through the whatsapp_*
// tools, and AGENTS.md §0 keeps it human-gated. Autonomy is a per-channel
// policy, not a property every channel must have.
type Transport interface {
	// Identity is the account the bot posts as; ok=false means no id could be
	// resolved and the operator must seed credentials.
	Identity() (int64, bool, error)

	// Groups lists the conversations this identity may answer.
	Groups(userID int64) ([]Group, error)

	// Perms is the capability set granted inside one conversation. A missing or
	// unreadable configuration must be fully restricted, never permissive.
	Perms(groupID int64) (Perms, error)

	// Read returns the messages of a conversation newer than afterID, oldest
	// first, excluding the bot's own posts.
	Read(groupID, afterID, excludeUserID int64) ([]Message, error)

	// History returns the last limit messages of a conversation (any sender) —
	// context for the model. Callers tolerate an error and then prompt without
	// history, so an implementation may return (nil, err) freely.
	History(groupID int64, limit int) ([]Message, error)

	// MaxID is the newest message id of a conversation (0 when empty); it seeds
	// a conversation the bot has never seen before.
	MaxID(groupID int64) (int64, error)

	// Latest is the newest message of a conversation (ok=false when empty); its
	// timestamp decides whether a newly discovered conversation is answered or
	// seeded past.
	Latest(groupID int64) (Message, bool, error)

	// Send posts the bot's reply into the conversation.
	Send(groupID, userID int64, text string) error
}
