// Package statex is the Go twin of src/state.ts — hot-file management for
// ai_questions.json + unanswered_chats.json, plus trim/archive/promote.
package statex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/timex"
)

// Chat-watching: the paths below mirror src/state.ts `hot()` — byte-for-byte
// for the default account, per-profile under data/ otherwise.
func QuestionsPath() string { return hot("ai_questions.json", "questions.json") }
func ChatsPath() string     { return hot("unanswered_chats.json", "chats.json") }
func ArchivePath() string {
	if config.ProfileIsDefault() {
		return filepath.Join(config.Root(), "data", "state-archive.json")
	}
	return filepath.Join(config.Root(), "data", config.WAProfile()+"-archive.json")
}

// RunFile returns data/<basename> (default) or data/<profile>-<basename>
// (per-account run reports like pending.json / triage.json).
func RunFile(basename string) string {
	return hot("data/"+basename, basename)
}

func hot(rootFile, customName string) string {
	if config.ProfileIsDefault() {
		return filepath.Join(config.Root(), rootFile)
	}
	return filepath.Join(config.Root(), "data", config.WAProfile()+"-"+customName)
}

const maxAgeMs = int64(60 * 24 * 60 * 60 * 1000) // 60 days

type QuestionContext struct {
	ChatKey         string   `json:"chatKey"`
	Needed          []string `json:"needed"`
	Reason          *string  `json:"reason,omitempty"`
	AskedAt         string   `json:"askedAt"`
	SentToGroup     bool     `json:"sentToGroup"`
	GroupLastScanAt *string  `json:"groupLastScanAt,omitempty"`
	AnsweredBy      *string  `json:"answeredBy,omitempty"`
	AnsweredAt      *string  `json:"answeredAt,omitempty"`
	AnswerText      *string  `json:"answerText,omitempty"`
	AnswerNote      *string  `json:"answerNote,omitempty"`
	Answered        bool     `json:"answered"`
	LastMessage     *string  `json:"lastMessage,omitempty"`
	LastMessageType *string  `json:"lastMessageType,omitempty"`
	LastMessageAt   *string  `json:"lastMessageAt,omitempty"`
	ProfileId       *int     `json:"profileId,omitempty"`
}

type AiQuestion struct {
	Id              int              `json:"id"`
	Question        string           `json:"question"`
	AskedAt         string           `json:"askedAt"`
	Answered        bool             `json:"answered"`
	AnsweredAt      *string          `json:"answeredAt,omitempty"`
	Answer          *string          `json:"answer,omitempty"`
	AnsweredBy      *string          `json:"answeredBy,omitempty"`
	SentToGroup     *bool            `json:"sentToGroup,omitempty"`
	SentToGroupAt   *string          `json:"sentToGroupAt,omitempty"`
	GroupLastScanAt *string          `json:"groupLastScanAt,omitempty"`
	Context         *QuestionContext `json:"context,omitempty"`
	ProfileId       *int             `json:"profileId,omitempty"`
}

type PendingChat struct {
	ChatKey         string   `json:"chatKey"`
	ChatName        *string  `json:"chatName,omitempty"`
	Status          string   `json:"status,omitempty"`
	Reason          *string  `json:"reason,omitempty"`
	Needed          []string `json:"needed,omitempty"`
	LastMessage     *string  `json:"lastMessage,omitempty"`
	LastMessageType *string  `json:"lastMessageType,omitempty"`
	LastMessageAt   *string  `json:"lastMessageAt,omitempty"`
	AskedAt         *string  `json:"askedAt,omitempty"`
	QuestionIds     []int    `json:"questionIds,omitempty"`
	DataReceivedAt  *string  `json:"dataReceivedAt,omitempty"`
	RepliedAt       *string  `json:"repliedAt,omitempty"`
	Answered        *bool    `json:"answered,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
	ProfileId       *int     `json:"profileId,omitempty"`
}

type QuestionsFile struct {
	Description *string      `json:"description,omitempty"`
	NextId      int          `json:"nextId"`
	Questions   []AiQuestion `json:"questions"`
}

type ChatsFile struct {
	Description *string       `json:"description,omitempty"`
	Chats       []PendingChat `json:"chats"`
}

func readJson[T any](file string, fallback T) T {
	raw, err := os.ReadFile(file)
	if err != nil {
		return fallback
	}
	if err := json.Unmarshal(raw, &fallback); err != nil {
		return fallback
	}
	return fallback
}

func LoadQuestions() QuestionsFile {
	f := readJson(QuestionsPath(), QuestionsFile{NextId: 1, Questions: []AiQuestion{}})
	if f.Questions == nil {
		f.Questions = []AiQuestion{}
	}
	if f.NextId <= 0 {
		maxID := 0
		for _, q := range f.Questions {
			if q.Id > maxID {
				maxID = q.Id
			}
		}
		f.NextId = maxID + 1
	}
	return f
}

func LoadChats() ChatsFile {
	f := readJson(ChatsPath(), ChatsFile{Chats: []PendingChat{}})
	if f.Chats == nil {
		f.Chats = []PendingChat{}
	}
	return f
}

func SaveQuestions(f QuestionsFile) error {
	return writeJson(QuestionsPath(), f)
}

func SaveChats(f ChatsFile) error {
	return writeJson(ChatsPath(), f)
}

func writeJson(file string, v any) error {
	data, err := jsonx.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, append(data, '\n'), 0o644)
}

// GroupTag returns `[ID: profile#n]` (unique across accounts in the shared group).
func GroupTag(id int) string { return "[ID: " + config.WAProfile() + "#" + itoa(id) + "]" }

// GroupKey returns the bare `profile#n` key inside the tag.
func GroupKey(id int) string { return config.WAProfile() + "#" + itoa(id) }

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf []byte
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

// IsActiveStatus — a chat that still needs attention.
func IsActiveStatus(status string) bool {
	return status == "waiting_for_data" || status == "data_ready"
}

// TrimAndArchive moves dead weight out of hot files into the archive
// (nothing is ever hard-deleted). See src/state.ts for the trim rules.
func TrimAndArchive(qf *QuestionsFile, cf *ChatsFile) (archivedQuestions, archivedChats int) {
	activeLinked := map[int]bool{}
	for i := range cf.Chats {
		if IsActiveStatus(cf.Chats[i].Status) {
			for _, id := range cf.Chats[i].QuestionIds {
				activeLinked[id] = true
			}
		}
	}

	keepQ := []AiQuestion{}
	var archQ []AiQuestion
	for _, q := range qf.Questions {
		tooOld := timex.AgeMs(q.AskedAt) > maxAgeMs
		dead := q.Answered && !activeLinked[q.Id]
		if tooOld || dead {
			archQ = append(archQ, q)
		} else {
			keepQ = append(keepQ, q)
		}
	}

	keepC := []PendingChat{}
	var archC []PendingChat
	for _, c := range cf.Chats {
		// Node: ageMs(c.askedAt) — an absent askedAt parses as NaN -> Infinity -> archived.
		tooOld := true
		if c.AskedAt != nil && *c.AskedAt != "" {
			tooOld = timex.AgeMs(*c.AskedAt) > maxAgeMs
		}
		done := c.Status == "replied" || c.Status == "closed" || c.Status == "expired"
		if tooOld || done {
			archC = append(archC, c)
		} else {
			keepC = append(keepC, c)
		}
	}

	if len(archQ) > 0 || len(archC) > 0 {
		arch := readJson(ArchivePath(), Archive{})
		arch.ArchivedAt = timex.NowIso()
		arch.Questions = append(arch.Questions, archQ...)
		arch.Chats = append(arch.Chats, archC...)
		if err := os.MkdirAll(filepath.Dir(ArchivePath()), 0o755); err == nil {
			writeJson(ArchivePath(), arch)
			qf.Questions = keepQ
			cf.Chats = keepC
		}
	}
	return len(archQ), len(archC)
}

// Archive is the shape of data/state-archive.json.
type Archive struct {
	ArchivedAt string        `json:"archivedAt"`
	Questions  []AiQuestion  `json:"questions"`
	Chats      []PendingChat `json:"chats"`
}

// PromoteReady flips waiting_for_data → data_ready for chats whose linked
// questions are all answered with a real answer. Returns the promoted chats.
func PromoteReady(cf *ChatsFile, qf *QuestionsFile) []PendingChat {
	byID := map[int]AiQuestion{}
	for _, q := range qf.Questions {
		byID[q.Id] = q
	}
	var promoted []PendingChat
	for i := range cf.Chats {
		c := &cf.Chats[i]
		if c.Status != "waiting_for_data" {
			continue
		}
		ids := c.QuestionIds
		if len(ids) == 0 {
			continue
		}
		allAnswered := true
		for _, id := range ids {
			q, ok := byID[id]
			if !ok || !q.Answered || q.Answer == nil || strings.TrimSpace(*q.Answer) == "" {
				allAnswered = false
				break
			}
		}
		if !allAnswered {
			continue
		}
		c.Status = "data_ready"
		now := timex.NowIso()
		c.DataReceivedAt = &now
		promoted = append(promoted, *c)
	}
	return promoted
}
