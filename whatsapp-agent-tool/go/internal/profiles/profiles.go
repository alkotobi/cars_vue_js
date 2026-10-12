// Package profiles is the Go twin of src/profiles.ts — the account registry
// (written by code, never by the model). The `name` field IS the WA_PROFILE value.
package profiles

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/timex"
)

// Ref mirrors the optional chat cache on a profile entry.
type ChatRef struct {
	ChatName      string  `json:"chatName"`
	ChatKey       *string `json:"chatKey,omitempty"`
	LastSeenAt    *string `json:"lastSeenAt,omitempty"`
	LastMessageAt *string `json:"lastMessageAt,omitempty"`
}

type Entry struct {
	Id            int       `json:"id"`
	Name          string    `json:"name"`
	Number        string    `json:"number"`
	Description   string    `json:"description"`
	Active        bool      `json:"active"`
	LinkedAt      *string   `json:"linkedAt,omitempty"`
	LastAiReplyAt *string   `json:"lastAiReplyAt,omitempty"`
	Chats         []ChatRef `json:"chats,omitempty"`
}

type File struct {
	NextId   int     `json:"nextId"`
	Profiles []Entry `json:"profiles"`
}

// ProfilesPath returns the profiles.json path at the repo root.
func ProfilesPath() string { return filepath.Join(config.Root(), "profiles.json") }

func Load() File {
	raw, err := os.ReadFile(ProfilesPath())
	if err != nil {
		return File{NextId: 1, Profiles: []Entry{}}
	}
	f := File{NextId: 1, Profiles: []Entry{}}
	if err := json.Unmarshal(raw, &f); err != nil {
		return File{NextId: 1, Profiles: []Entry{}}
	}
	if f.Profiles == nil {
		f.Profiles = []Entry{}
	}
	if f.NextId <= 0 {
		maxID := 0
		for _, e := range f.Profiles {
			if e.Id > maxID {
				maxID = e.Id
			}
		}
		f.NextId = maxID + 1
	}
	return f
}

func Save(f File) error {
	data, err := jsonx.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ProfilesPath(), append(data, '\n'), 0o644)
}

func Find(name string) (Entry, bool) {
	pf := Load()
	for _, e := range pf.Profiles {
		if e.Name == name {
			return e, true
		}
	}
	return Entry{}, false
}

// Add registers a freshly linked profile; errors if the name is taken.
func Add(name, number, description string) (Entry, error) {
	pf := Load()
	for _, e := range pf.Profiles {
		if e.Name == name {
			return Entry{}, fmt.Errorf("profile %s is already registered in %s", name, ProfilesPath())
		}
	}
	now := timex.NowIso()
	entry := Entry{
		Id:            pf.NextId,
		Name:          name,
		Number:        number,
		Description:   description,
		Active:        true,
		LinkedAt:      &now,
		LastAiReplyAt: nil,
	}
	pf.NextId += 1
	pf.Profiles = append(pf.Profiles, entry)
	if err := Save(pf); err != nil {
		return Entry{}, err
	}
	return entry, nil
}

// MarkAiReply stamps the "last time AI replied" field.
func MarkAiReply(name string) (Entry, error) {
	return mutate(name, func(e *Entry) {
		now := timex.NowIso()
		e.LastAiReplyAt = &now
	})
}

// SetActive flips the active flag.
func SetActive(name string, active bool) (Entry, error) {
	return mutate(name, func(e *Entry) { e.Active = active })
}

func mutate(name string, change func(*Entry)) (Entry, error) {
	pf := Load()
	for i := range pf.Profiles {
		if pf.Profiles[i].Name == name {
			change(&pf.Profiles[i])
			if err := Save(pf); err != nil {
				return Entry{}, err
			}
			return pf.Profiles[i], nil
		}
	}
	return Entry{}, errors.New("profile " + name + " not found in " + ProfilesPath())
}
