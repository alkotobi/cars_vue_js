// Package webfetch is the chat agent's only network reader. It fetches a page
// ONLY from a fixed allowlist of company-credibility and car-specification
// sites, so a fetched page can never come from an arbitrary host.
//
// opencode 1.x has no per-URL webfetch permission (the `webfetch` tool takes a
// single allow/deny action, unlike `bash`/`edit`), so the agent's own webfetch
// is disabled and the allowlist is enforced HERE in Go instead: the agent is
// told to read pages through `whatsapp-tool web fetch <url>`. Every hop is
// checked — the requested URL and each redirect target — so an allowlisted page
// cannot bounce us onto an arbitrary host.
package webfetch

import (
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// Allowlist is the fixed set of source domains the agent may read. A host is
// allowed when it equals one of these domains or is a subdomain of one (so
// "www.gsxt.gov.cn" and "gsxt.gov.cn" both match "gsxt.gov.cn", but
// "gsxt.gov.cn.evil.com" does not). Edit this list to add a source; keep the
// chat agent's body text (chatbot/agent/chatbot.md) in step with it.
var Allowlist = []string{
	// company credibility — official registries / credit / courts
	"gsxt.gov.cn",
	"creditchina.gov.cn",
	"wenshu.court.gov.cn",
	"zxgk.court.gov.cn",
	"samr.gov.cn",
	"cr.gov.hk",
	"e-services.cr.gov.hk",
	"icris.cr.gov.hk",
	"info.gov.hk",
	"tid.gov.hk",
	// company credibility — aggregators & trade directories
	"opencorporates.com",
	"tianyancha.com",
	"qcc.com",
	"aiqicha.baidu.com",
	"chinverify.com",
	"chinesecheck.com",
	"companieshouse.hk",
	"alibaba.com",
	"globalsources.com",
	"made-in-china.com",
	"hktdc.com",
	"1688.com",
	// car specifications
	"autohome.com.cn",
	"dcdapp.com",
	"yiche.com",
	"car.info",
	"wikipedia.org",
	"carsguide.com.au",
	"carwow.co.uk",
}

// Allowed reports whether host (with any port stripped) is on the allowlist.
// Matching is case-insensitive and covers subdomains.
func Allowed(host string) bool {
	host = strings.ToLower(strings.TrimSpace(host))
	if i := strings.LastIndexByte(host, ':'); i >= 0 {
		host = host[:i] // drop an explicit :port or IPv6 bracket tail
	}
	host = strings.Trim(host, "[]")
	if host == "" {
		return false
	}
	for _, d := range Allowlist {
		if host == d || strings.HasSuffix(host, "."+d) {
			return true
		}
	}
	return false
}

// Options bounds one fetch. Zero fields take the defaults below.
type Options struct {
	Timeout  time.Duration
	MaxBytes int64
	MaxChars int
}

func (o *Options) applyDefaults() {
	if o.Timeout <= 0 {
		o.Timeout = 20 * time.Second
	}
	if o.MaxBytes <= 0 {
		o.MaxBytes = 1 << 20 // 1 MiB
	}
	if o.MaxChars <= 0 {
		o.MaxChars = 12000
	}
}

// Result is a fetched, plain-text page.
type Result struct {
	URL         string `json:"url"`
	FinalURL    string `json:"finalUrl"`
	StatusCode  int    `json:"httpStatus"`
	ContentType string `json:"contentType"`
	Title       string `json:"title"`
	Text        string `json:"text"`
	Bytes       int    `json:"bytes"`
	Truncated   bool   `json:"truncated"`
}

// Error carries a machine-readable code so the CLI and the model can react
// (blocked_domain vs fetch_failed vs unsupported_content …).
type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

const userAgent = "whatsapp-tool-webfetch/1.0 (chat assistant; allowlisted sources)"

// Fetch retrieves rawURL if its host (and every redirect host) is on the
// allowlist, and returns the page as plain text.
func Fetch(rawURL string, opts Options) (*Result, error) {
	return fetch(rawURL, opts, Allowed)
}

// fetch is Fetch with an injectable allowlist check, so tests can drive real
// HTTP servers (on 127.0.0.1) without touching the production allowlist.
func fetch(rawURL string, opts Options, allowed func(string) bool) (*Result, error) {
	opts.applyDefaults()

	target := strings.TrimSpace(rawURL)
	u, err := url.Parse(target)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil, &Error{Code: "bad_url", Message: "provide an absolute http(s) URL"}
	}
	if !allowed(u.Hostname()) {
		return nil, &Error{Code: "blocked_domain", Message: fmt.Sprintf("domain %q is not on the source allowlist (run `whatsapp-tool web list` to see the allowed domains)", u.Hostname())}
	}

	client := &http.Client{
		Timeout: opts.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return &Error{Code: "too_many_redirects", Message: "too many redirects"}
			}
			if !allowed(req.URL.Hostname()) {
				return &Error{Code: "blocked_redirect", Message: fmt.Sprintf("redirect to a non-allowlisted domain blocked: %q", req.URL.Hostname())}
			}
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, &Error{Code: "bad_url", Message: err.Error()}
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,text/plain;q=0.9,*/*;q=0.5")

	resp, err := client.Do(req)
	if err != nil {
		var we *Error
		if errors.As(err, &we) {
			return nil, we
		}
		return nil, &Error{Code: "fetch_failed", Message: err.Error()}
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, opts.MaxBytes+1))
	if err != nil {
		return nil, &Error{Code: "read_failed", Message: err.Error()}
	}
	truncated := int64(len(raw)) > opts.MaxBytes
	if truncated {
		raw = raw[:opts.MaxBytes]
	}

	ct := strings.TrimSpace(resp.Header.Get("Content-Type"))
	if !isTextual(ct) {
		return nil, &Error{Code: "unsupported_content", Message: fmt.Sprintf("the page is not text (Content-Type %q) — only HTML/plain-text pages can be read", ct)}
	}

	var title, text string
	if isHTML(ct) || looksLikeHTML(raw) {
		title, text = htmlToText(string(raw))
	} else {
		title, text = "", normalizeText(string(raw))
	}
	text = truncateRunes(text, opts.MaxChars)

	return &Result{
		URL:         target,
		FinalURL:    resp.Request.URL.String(),
		StatusCode:  resp.StatusCode,
		ContentType: ct,
		Title:       title,
		Text:        text,
		Bytes:       len(raw),
		Truncated:   truncated,
	}, nil
}

func isTextual(ct string) bool {
	ct = strings.ToLower(ct)
	switch {
	case ct == "":
		return true
	case strings.HasPrefix(ct, "text/"),
		strings.Contains(ct, "html"),
		strings.Contains(ct, "xml"),
		strings.Contains(ct, "json"):
		return true
	default:
		return false
	}
}

func isHTML(ct string) bool {
	ct = strings.ToLower(ct)
	return strings.Contains(ct, "html") || strings.Contains(ct, "xml")
}

func looksLikeHTML(b []byte) bool {
	head := strings.ToLower(string(b))
	if len(head) > 4096 {
		head = head[:4096]
	}
	return strings.Contains(head, "<html") || strings.Contains(head, "<!doctype html") || strings.Contains(head, "<body") || strings.Contains(head, "<div")
}

var (
	reScript = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	reStyle  = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	reTitle  = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
	reBreak  = regexp.MustCompile(`(?i)<(br|hr|/p|/div|/li|/tr|/h[1-6]|/section|/article|/table|/ul|/ol|/header|/footer)\b[^>]*>`)
	reTag    = regexp.MustCompile(`(?s)<[^>]*>`)
	reInline = regexp.MustCompile(`[ \t\r\f\v\x{00a0}]+`)
)

// htmlToText returns the page title and a whitespace-normalized plain-text
// rendering of the body: script/style are dropped, block tags become line
// breaks, the remaining tags are stripped and HTML entities are decoded.
func htmlToText(s string) (title, text string) {
	if m := reTitle.FindStringSubmatch(s); m != nil {
		title = normalizeText(html.UnescapeString(reTag.ReplaceAllString(m[1], " ")))
	}
	s = reScript.ReplaceAllString(s, " ")
	s = reStyle.ReplaceAllString(s, " ")
	s = reBreak.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return title, normalizeText(s)
}

// normalizeText collapses runs of spaces to one, trims each line and drops
// blank lines — a compact, model-friendly block of text.
func normalizeText(s string) string {
	s = reInline.ReplaceAllString(s, " ")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			out = append(out, line)
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

func truncateRunes(s string, max int) string {
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	return string([]rune(s)[:max])
}
