package main

import (
	"errors"
	"fmt"
	"os"
	"whatsapp-agent-tool/internal/webfetch"
)

// cmdWeb is the chat agent's network reader. `web fetch` retrieves one page
// through the Go-enforced source allowlist (internal/webfetch) — the agent's
// own opencode webfetch is disabled because opencode 1.x cannot scope it to a
// domain list. It is read-only and ungated: the sources are public registries
// and car-spec sites, never cost or supplier data.
func cmdWeb(args []string) {
	usage := `usage: whatsapp-tool web <fetch <url> [--max-chars N] | list>`
	if len(args) == 0 {
		failErr(usage)
	}
	switch args[0] {
	case "list":
		ok(pair("action", "web-list"), pair("count", len(webfetch.Allowlist)), pair("domains", webfetch.Allowlist))
	case "fetch":
		rest := args[1:]
		positionals := positionalArgs(rest, map[string]bool{"--max-chars": true})
		if len(positionals) == 0 {
			failErr(`web fetch needs a URL: whatsapp-tool web fetch "<url>" [--max-chars N]`)
		}
		opts := webfetch.Options{}
		if v := flag(rest, "max-chars"); v != "" {
			fmt.Sscanf(v, "%d", &opts.MaxChars)
		}
		res, err := webfetch.Fetch(positionals[0], opts)
		if err != nil {
			var we *webfetch.Error
			if errors.As(err, &we) {
				failErr(we.Message, pair("code", we.Code))
			}
			failErr(err.Error())
		}
		o := ordered{Pairs: []orderedPair{
			pair("status", "OK"),
			pair("action", "web-fetch"),
			pair("url", res.URL),
			pair("finalUrl", res.FinalURL),
			pair("httpStatus", res.StatusCode),
			pair("contentType", res.ContentType),
			pair("title", res.Title),
			pair("chars", len([]rune(res.Text))),
			pair("truncated", res.Truncated),
			pair("text", res.Text),
		}}
		printCompact(&o)
		os.Exit(0)
	default:
		failErr(usage)
	}
}
