package chatbot

import (
	"strings"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/groupauth"
)

// Perms is the capability set of the chat group a reply is being generated for.
// The bot loads it from the ai DB (internal/db/groupperms) before every reply
// and threads it through Generate → runOpenCode, where it becomes the child's
// WBOT_PERMS environment variable — the CLI then enforces it.
type Perms struct {
	Edit      bool
	Suppliers bool
	Cost      bool
}

// EnvValue is the WBOT_PERMS value: the comma-separated granted capability
// names. The bot ALWAYS sets the variable (possibly empty), which is what makes
// the CLI treat the run as restricted; absent means operator/daemon.
func (p Perms) EnvValue() string {
	var names []string
	if p.Edit {
		names = append(names, groupauth.Edit)
	}
	if p.Suppliers {
		names = append(names, groupauth.Suppliers)
	}
	if p.Cost {
		names = append(names, groupauth.Cost)
	}
	return strings.Join(names, ",")
}

// AgentName is the opencode agent to run: the editor agent when the group may
// give edit instructions (it may write documentation *.md files in the repo),
// otherwise the plain read-only chatbot agent.
func (p Perms) AgentName() string {
	if p.Edit {
		return config.C.Chatbot.Agent + "-editor"
	}
	return config.C.Chatbot.Agent
}
