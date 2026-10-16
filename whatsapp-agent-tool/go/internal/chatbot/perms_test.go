package chatbot

import (
	"testing"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/groupauth"
)

func TestPermsEnvValue(t *testing.T) {
	cases := []struct {
		name string
		p    Perms
		want string
	}{
		{"none is empty (still restricted)", Perms{}, ""},
		{"edit only", Perms{Edit: true}, groupauth.Edit},
		{"suppliers only", Perms{Suppliers: true}, groupauth.Suppliers},
		{"cost only", Perms{Cost: true}, groupauth.Cost},
		{"all three, stable order", Perms{Edit: true, Suppliers: true, Cost: true},
			groupauth.Edit + "," + groupauth.Suppliers + "," + groupauth.Cost},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.p.EnvValue(); got != tc.want {
				t.Fatalf("EnvValue() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestPermsEnvValueUsesConstValues pins that EnvValue emits the CLI's exact
// capability constants (not free-form strings) — the two sides must agree.
func TestPermsEnvValueUsesConstValues(t *testing.T) {
	got := Perms{Edit: true, Cost: true}.EnvValue()
	if got != groupauth.Edit+","+groupauth.Cost {
		t.Fatalf("EnvValue() = %q, want %q", got, groupauth.Edit+","+groupauth.Cost)
	}
}

func TestPermsAgentName(t *testing.T) {
	old := config.C.Chatbot.Agent
	defer func() { config.C.Chatbot.Agent = old }()
	config.C.Chatbot.Agent = "chatbot"

	if got := (Perms{}).AgentName(); got != "chatbot" {
		t.Fatalf("no-edit AgentName() = %q, want chatbot", got)
	}
	if got := (Perms{Edit: true}).AgentName(); got != "chatbot-editor" {
		t.Fatalf("edit AgentName() = %q, want chatbot-editor", got)
	}
}
