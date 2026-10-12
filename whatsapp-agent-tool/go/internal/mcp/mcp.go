// Package mcp is the Go twin of src/index.ts — the stdio MCP server exposing
// the whatsapp_* tools. It answers with the same text payloads as Node (compact
// JSON for the tools, pretty for session_info/identity), so an MCP client
// cannot tell the ports apart.
//
// Stage 4: whatsapp_identity is fully offline (code-minted, file-backed) and
// works now; the browser-backed tools call into the wa.Client contract, whose
// stub returns ErrNotImplemented until Stage 5 wires the rod driver.
package mcp

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/identity"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/wa"
)

// Server owns the MCP server and the WhatsApp driver.
type Server struct {
	srv *server.MCPServer
	wa  wa.Client
}

// New builds the MCP server and registers every tool (mirrors src/index.ts).
func New(cli wa.Client) *Server {
	s := &Server{wa: cli}
	s.srv = server.NewMCPServer("whatsapp-agent-tool", "0.1.0")
	s.register()
	return s
}

func replyText(payload string) *mcp.CallToolResult { return mcp.NewToolResultText(payload) }

// replyErr mirrors Node's `Error: ${err.message}` + isError:true.
func replyErr(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError("Error: " + err.Error())
}

func compact(v any) string {
	b, _ := jsonx.Marshal(v)
	return string(b)
}

func pretty(v any) string {
	b, _ := jsonx.MarshalIndent(v, "", "  ")
	return string(b)
}

func (s *Server) register() {
	s.srv.AddTool(mcp.NewTool("whatsapp_is_logged_in",
		mcp.WithToolTitle("Check if WhatsApp Web is logged in"),
		mcp.WithDescription("Returns true if authenticated and not showing QR code. If false, you need to scan QR on first run (headed mode recommended)."),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		loggedIn, err := s.wa.IsLoggedIn()
		if err != nil {
			return replyErr(err), nil
		}
		return replyText(compact(map[string]bool{"loggedIn": loggedIn})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_get_chats",
		mcp.WithToolTitle("List chats (screen / newest / unread)"),
		mcp.WithDescription("List chats from the sidebar. mode: \"screen\" = only the ~15 rows WhatsApp has rendered (fast, default); \"newest\" = scroll the sidebar like a person to collect the last `limit` chats; \"unread\" = switch to WhatsApp's unread filter and collect them (also returns unreadTotal = the counter on the filter chip). Use newest/unread whenever you need more than the first screen."),
		mcp.WithNumber("limit", mcp.Description("How many chats (default 15 screen / 50 newest-unread; screen caps at 50)"),
			mcp.Min(1), mcp.Max(1000)),
		mcp.WithString("mode", mcp.Description("screen | newest | unread (default screen)"),
			mcp.Enum("screen", "newest", "unread")),
		mcp.WithNumber("timeoutMs", mcp.Description("newest/unread only: give up after this long (default 120000)"),
			mcp.Min(5000), mcp.Max(600000)),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		mode := strArg(args, "mode", "screen")
		switch mode {
		case "newest":
			res, err := s.wa.GetLastChats(intArg(args, "limit", 50), intArg(args, "timeoutMs", 120000))
			if err != nil {
				return replyErr(err), nil
			}
			return replyText(compact(struct {
				Mode string `json:"mode"`
				wa.ChatSweepResult
			}{mode, res})), nil
		case "unread":
			res, err := s.wa.GetLastUnreadChats(intArg(args, "limit", 50), intArg(args, "timeoutMs", 120000))
			if err != nil {
				return replyErr(err), nil
			}
			return replyText(compact(struct {
				Mode string `json:"mode"`
				wa.ChatSweepResult
			}{mode, res})), nil
		default:
			chats, err := s.wa.GetChats(minInt(intArg(args, "limit", 15), 50))
			if err != nil {
				return replyErr(err), nil
			}
			// Node: JSON.stringify({ mode, chats }) — mode key first.
			return replyText(compact(struct {
				Mode  string        `json:"mode"`
				Chats []wa.ChatInfo `json:"chats"`
			}{"screen", chats})), nil
		}
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_open_chat",
		mcp.WithToolTitle("Open a chat"),
		mcp.WithDescription("Open chat by contact name, group name, or phone number."),
		mcp.WithString("nameOrPhone", mcp.Required(), mcp.Description("Contact/group name or phone number")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ok, err := s.wa.OpenChat(strArg(req.GetArguments(), "nameOrPhone", ""))
		if err != nil {
			return replyErr(err), nil
		}
		return replyText(compact(map[string]bool{"success": ok})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_send_message",
		mcp.WithToolTitle("Send a message"),
		mcp.WithDescription("Send text message to currently open chat."),
		mcp.WithString("text", mcp.Required(), mcp.Description("Message text to send")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ok, err := s.wa.SendMessage(strArg(req.GetArguments(), "text", ""))
		if err != nil {
			return replyErr(err), nil
		}
		return replyText(compact(map[string]bool{"success": ok})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_get_last_messages",
		mcp.WithToolTitle("Get last messages in current chat"),
		mcp.WithDescription("Read recent messages from the currently open conversation."),
		mcp.WithNumber("limit", mcp.Description("Number of messages to return (default 10)"),
			mcp.Min(1), mcp.Max(50)),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		msgs, err := s.wa.GetLastMessages(intArg(req.GetArguments(), "limit", 10))
		if err != nil {
			return replyErr(err), nil
		}
		return replyText(compact(map[string]any{"messages": msgs})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_mark_as_read",
		mcp.WithToolTitle("Mark current chat as read"),
		mcp.WithDescription("Mark the currently open chat as read (best effort)."),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ok, err := s.wa.MarkAsRead()
		if err != nil {
			return replyErr(err), nil
		}
		return replyText(compact(map[string]bool{"success": ok})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_session_info",
		mcp.WithToolTitle("Browser session diagnostics"),
		mcp.WithDescription("How this server is talking to Chrome: profile, transport (cdp/launch), CDP port, Chrome pid, whether the browser survives this process. Use it when calls feel slow or to check the persistent session is alive."),
	), func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		info, err := s.wa.SessionInfo()
		if err != nil {
			return replyErr(err), nil
		}
		// Node: JSON.stringify({ ...info, identity: getIdentity() }, null, 2)
		return replyText(pretty(struct {
			wa.SessionInfo
			Identity identity.Identity `json:"identity"`
		}{info, identity.Get()})), nil
	})

	s.srv.AddTool(mcp.NewTool("whatsapp_identity",
		mcp.WithToolTitle("Get this assistant's AI identity"),
		mcp.WithDescription("Returns the unique, irreplicable name the tool assigned to this assistant by code (not chosen by AI). Persisted per WhatsApp profile and stable across sessions. Pass force=true only to re-mint a fresh name."),
		mcp.WithBoolean("force", mcp.Description("Re-mint a new identity (code-only). Default false.")),
	), func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var id identity.Identity
		if force, _ := boolArg(req.GetArguments(), "force"); force {
			id = identity.Reset()
		} else {
			id, _ = identity.Ensure()
		}
		return replyText(pretty(id)), nil
	})
}

// Serve hangs on the stdio transport until the client disconnects.
func (s *Server) Serve() error { return server.ServeStdio(s.srv) }

// Start logs the Node startup stderr line (never stdout — that is the JSON-RPC
// channel) and mints/loads the identity first, exactly like main().
func Start(cli wa.Client) error {
	id, fresh := identity.Ensure()
	fmt.Fprintf(os.Stderr, "[whatsapp-mcp] Server started (profile=%s) identity=%s%s\n",
		config.WAProfile(), id.Name, freshNote(fresh))
	s := New(cli)
	return s.Serve()
}

func freshNote(fresh bool) string {
	if fresh {
		return fmt.Sprintf(" (new, saved to %s)", identity.File(config.WAProfile(), config.ProfileIsDefault()))
	}
	return " (loaded)"
}

// ------------------------------------------------------------------- args

func strArg(args map[string]any, key, def string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

func intArg(args map[string]any, key string, def int) int {
	if v, ok := args[key]; ok {
		switch t := v.(type) {
		case float64:
			return int(t)
		case int:
			return t
		case int64:
			return int(t)
		case uint64:
			return int(t)
		case string:
			if n, err := strconv.Atoi(t); err == nil {
				return n
			}
		}
	}
	return def
}

func boolArg(args map[string]any, key string) (bool, bool) {
	if v, ok := args[key]; ok {
		if b, ok := v.(bool); ok {
			return b, true
		}
	}
	return false, false
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
