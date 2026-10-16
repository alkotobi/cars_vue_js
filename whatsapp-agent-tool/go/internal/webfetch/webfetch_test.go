package webfetch

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestAllowed(t *testing.T) {
	allow := []string{
		"gsxt.gov.cn",
		"www.gsxt.gov.cn",
		"GSXT.GOV.CN",
		"gsxt.gov.cn:443",
		"e-services.cr.gov.hk",
		"sub.e-services.cr.gov.hk",
		"autohome.com.cn",
		"car.info",
		"en.wikipedia.org",
	}
	for _, h := range allow {
		if !Allowed(h) {
			t.Errorf("Allowed(%q) = false, want true", h)
		}
	}
	deny := []string{
		"",
		"evil.com",
		"gsxt.gov.cn.evil.com",
		"notgsxt.gov.cn",
		"gov.cn",
		"car.info.evil.com",
		"1698.com",
	}
	for _, h := range deny {
		if Allowed(h) {
			t.Errorf("Allowed(%q) = true, want false", h)
		}
	}
}

func TestFetchBadURL(t *testing.T) {
	for _, in := range []string{"", "not a url", "ftp://gsxt.gov.cn/x", "gsxt.gov.cn"} {
		_, err := Fetch(in, Options{})
		var we *Error
		if !errors.As(err, &we) || we.Code != "bad_url" {
			t.Errorf("Fetch(%q): got %v, want bad_url", in, err)
		}
	}
}

// TestFetchBlockedDomain uses the production allowlist and must refuse an
// off-list host WITHOUT any network I/O.
func TestFetchBlockedDomain(t *testing.T) {
	_, err := Fetch("https://evil.example/x", Options{})
	var we *Error
	if !errors.As(err, &we) || we.Code != "blocked_domain" {
		t.Fatalf("got %v, want blocked_domain", err)
	}
}

const sampleHTML = `<!doctype html><html><head><title> GSXT &amp; Co </title>
<style>body{color:red}</style><script>alert(1)</script></head>
<body><h1>Enterprise</h1><p>Name: <b>Acme</b> Ltd.</p>
<ul><li>Reg no 123</li><li>Status 存续</li></ul></body></html>`

func TestFetchOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleHTML))
	}))
	defer srv.Close()

	res, err := fetch(srv.URL, Options{}, localOnly)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Errorf("status = %d, want 200", res.StatusCode)
	}
	if res.Title != "GSXT & Co" {
		t.Errorf("title = %q, want %q", res.Title, "GSXT & Co")
	}
	for _, want := range []string{"Enterprise", "Name: Acme Ltd.", "Reg no 123", "Status 存续"} {
		if !strings.Contains(res.Text, want) {
			t.Errorf("text missing %q:\n%s", want, res.Text)
		}
	}
	for _, bad := range []string{"<", ">", "alert(1)", "color:red"} {
		if strings.Contains(res.Text, bad) {
			t.Errorf("text should have no %q: %s", bad, res.Text)
		}
	}
}

func TestFetchTruncates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(strings.Repeat("abcdef ", 100)))
	}))
	defer srv.Close()

	res, err := fetch(srv.URL, Options{MaxChars: 10, MaxBytes: 8}, localOnly)
	if err != nil {
		t.Fatal(err)
	}
	if n := len([]rune(res.Text)); n > 10 {
		t.Errorf("text runes = %d, want <= 10", n)
	}
	if !res.Truncated {
		t.Error("Truncated = false, want true (MaxBytes exceeded)")
	}
	if res.Bytes != 8 {
		t.Errorf("Bytes = %d, want 8", res.Bytes)
	}
}

func TestFetchBlocksRedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://blocked.example/secret", http.StatusFound)
	}))
	defer srv.Close()

	_, err := fetch(srv.URL, Options{}, localOnly)
	var we *Error
	if !errors.As(err, &we) || we.Code != "blocked_redirect" {
		t.Fatalf("got %v, want blocked_redirect", err)
	}
}

func TestFetchUnsupportedContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4"))
	}))
	defer srv.Close()

	_, err := fetch(srv.URL, Options{}, localOnly)
	var we *Error
	if !errors.As(err, &we) || we.Code != "unsupported_content" {
		t.Fatalf("got %v, want unsupported_content", err)
	}
}

func TestFetchTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		_, _ = w.Write([]byte("late"))
	}))
	defer srv.Close()

	_, err := fetch(srv.URL, Options{Timeout: 20 * time.Millisecond}, localOnly)
	var we *Error
	if !errors.As(err, &we) || we.Code != "fetch_failed" {
		t.Fatalf("got %v, want fetch_failed (timeout)", err)
	}
}

// localOnly allows only the httptest server's host, so the tests exercise the
// real fetch/redirect pipeline without the production allowlist.
func localOnly(host string) bool { return host == "127.0.0.1" || host == "localhost" }

// TestAllowlistSelfConsistent guards the curated list: no blanks, no dupes,
// and every entry must match itself plus a subdomain of itself.
func TestAllowlistSelfConsistent(t *testing.T) {
	seen := map[string]bool{}
	for _, d := range Allowlist {
		if d == "" || strings.ContainsAny(d, " /:") {
			t.Errorf("bad allowlist entry %q", d)
		}
		if seen[d] {
			t.Errorf("duplicate allowlist entry %q", d)
		}
		seen[d] = true
		if !Allowed(d) {
			t.Errorf("Allowed(%q) = false for its own entry", d)
		}
		if !Allowed("www." + d) {
			t.Errorf("Allowed(%q) = false for a subdomain", "www."+d)
		}
	}
}
