// Package carshttp is the HTTP client for the cars app (a SEPARATE system from
// the shared `ai` database). We reach a tenant's api.php; the AI user
// authenticates and a token is cached in the OS credential store
// (internal/keyringx), never on disk. The login envelope also carries the
// user's role_name and the role's permission names (role_permissions +
// permissions tables server-side) — those are cached so the tools follow the
// app's role->permissions model instead of just holding a token.
package carshttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
	"whatsapp-agent-tool/internal/keyringx"
)

// ApiError mirrors CarsApiError.
type ApiError struct {
	Code       string
	Message    string
	HTTPStatus int
}

func (e *ApiError) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Message
}

const requestTimeout = 20 * time.Second

// httpClient is the shared API client. Keep-alives are OFF on purpose: a nil
// Transport would fall back to http.DefaultTransport, whose pooled connections
// the tenant's nginx closes while the chat-poll daemon sits idle between
// cycles, so the next POST reuses a dead connection and fails with "EOF".
var httpClient = &http.Client{
	Timeout:   requestTimeout,
	Transport: &http.Transport{DisableKeepAlives: true},
}

func assertConfigured() error {
	if config.C.Cars.BaseURL == "" || config.C.Cars.Tenant == "" {
		return &ApiError{Code: "not_configured", Message: "CARS_BASE_URL and CARS_TENANT must be set in .env"}
	}
	return nil
}

func tenantEndpoint() string {
	return config.C.Cars.BaseURL + "/" + url.PathEscape(config.C.Cars.Tenant) + "/api/api.php"
}

// postJSON POSTs a JSON body to the tenant API and parses the response.
func postJSON(body map[string]any) (*apiResponse, error) {
	if err := assertConfigured(); err != nil {
		return nil, err
	}
	endpoint := tenantEndpoint()
	payload, err := jsonx.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Cards-Tenant", config.C.Cars.Tenant)
	req.Header.Set("CARDS-TENANT", config.C.Cars.Tenant)
	res, err := httpClient.Do(req)
	if err != nil {
		return nil, &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, &ApiError{Code: "network_error", Message: "cannot reach " + endpoint + ": " + err.Error()}
	}
	var resp apiResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, &ApiError{Code: "bad_response", Message: fmt.Sprintf("non-JSON response (HTTP %d)", res.StatusCode)}
	}
	return &resp, nil
}

// apiResponse holds the cars API envelope. Data is kept as raw JSON so row
// key order and number formatting pass through unchanged.
//
// LastInsertID/AffectedRows are RawMessage because api.php does not type them:
// an INSERT answers `lastInsertId` straight from PDO::lastInsertId(), which is a
// JSON string, while other paths cast to int. Keeping the raw token lets a
// single parser accept both (see flexInt64 in write.go).
type apiResponse struct {
	Success      bool            `json:"success"`
	Token        string          `json:"token"`
	Code         string          `json:"code"`
	Error        string          `json:"error"`
	Data         json.RawMessage `json:"data"`
	User         *LoginUser      `json:"user"`
	LastInsertID json.RawMessage `json:"lastInsertId"`
	AffectedRows json.RawMessage `json:"affectedRows"`
}

// LoginUser mirrors the `user` object the app returns on login: the account
// plus its role and that role's permission list (auth.php joins permissions
// and role_permissions by role_id).
type LoginUser struct {
	ID          int          `json:"id"`
	Username    string       `json:"username"`
	RoleID      int          `json:"role_id"`
	RoleName    string       `json:"role_name"`
	Permissions []Permission `json:"permissions"`
}

// Permission is one entry of the role's permission list.
type Permission struct {
	Name        string `json:"permission_name"`
	Description string `json:"description"`
}

// LoginResult is what a successful login gives us: the fresh token plus the
// user's role and permissions.
type LoginResult struct {
	Token string
	User  LoginUser
}

// doLogin logs in and returns the token plus the user's role/permissions.
func doLogin(username, password string) (*LoginResult, error) {
	resp, err := postJSON(map[string]any{"action": "login", "username": username, "password": password})
	if err != nil {
		return nil, err
	}
	if !resp.Success || resp.Token == "" {
		code := resp.Code
		if code == "" {
			code = "login_failed"
		}
		return nil, &ApiError{Code: code, Message: "the AI user login was rejected"}
	}
	res := &LoginResult{Token: resp.Token}
	if resp.User != nil {
		res.User = *resp.User
	}
	return res, nil
}

// Login logs in as the AI user and returns a fresh token (does not touch the store).
func Login(username, password string) (string, error) {
	res, err := doLogin(username, password)
	if err != nil {
		return "", err
	}
	return res.Token, nil
}

// cacheRoleAndPermissions stores the role name and the permission names the
// login returned, so later commands (stock) can enforce the app's model.
func cacheRoleAndPermissions(u LoginUser) error {
	if err := keyringx.SetRole(u.RoleName); err != nil {
		return err
	}
	// The user's id is cached too so write commands can stamp id_user (the app
	// records the acting user on every payment/row it creates).
	if u.ID != 0 {
		if err := keyringx.SetUserID(strconv.Itoa(u.ID)); err != nil {
			return err
		}
	}
	names := make([]string, 0, len(u.Permissions))
	for _, p := range u.Permissions {
		names = append(names, p.Name)
	}
	raw, err := json.Marshal(names)
	if err != nil {
		return err
	}
	return keyringx.SetPermissions(string(raw))
}

// CurrentUserID returns the AI user's id as cached at login. ok is false when no
// id is stored (e.g. the token came from CARS_API_TOKEN, or a session predating
// the id cache); callers then leave the id_user column unset.
func CurrentUserID() (id int, ok bool, err error) {
	raw, err := keyringx.GetUserID()
	if err != nil {
		return 0, false, err
	}
	if raw == "" {
		return 0, false, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, false, nil
	}
	return n, true, nil
}

// HasPermission reports whether the AI user's role is granted the named
// permission. known is false when no permission snapshot is stored (e.g. the
// token came from CARS_API_TOKEN); callers then skip enforcement.
func HasPermission(name string) (known, ok bool, err error) {
	raw, err := keyringx.GetPermissions()
	if err != nil {
		return false, false, err
	}
	if raw == "" {
		return false, false, nil
	}
	var names []string
	if err := json.Unmarshal([]byte(raw), &names); err != nil {
		return false, false, &ApiError{Code: "bad_response", Message: "stored permission list is invalid"}
	}
	for _, n := range names {
		if n == name {
			return true, true, nil
		}
	}
	return true, false, nil
}

// LoginAndCache logs in with the stored credentials and caches the token.
// Missing credentials -> ApiError{missing_credentials}.
func LoginAndCache() (string, error) {
	username, err := keyringx.GetUsername()
	if err != nil {
		return "", err
	}
	password, err := keyringx.GetPassword()
	if err != nil {
		return "", err
	}
	if username == "" || password == "" {
		return "", &ApiError{
			Code:    "missing_credentials",
			Message: "the AI user's credentials are not stored yet — run: whatsapp-tool stock creds --username <user>",
		}
	}
	res, err := doLogin(username, password)
	if err != nil {
		return "", err
	}
	if err := keyringx.SetToken(res.Token); err != nil {
		return "", err
	}
	if err := cacheRoleAndPermissions(res.User); err != nil {
		return "", err
	}
	return res.Token, nil
}

// ResolveToken mirrors resolveToken(): env token, then the OS store, then login.
func ResolveToken() (string, error) {
	if err := assertConfigured(); err != nil {
		return "", err
	}
	if config.C.Cars.APIToken != "" {
		return config.C.Cars.APIToken, nil
	}
	tok, err := keyringx.GetToken()
	if err != nil {
		return "", err
	}
	if tok != "" {
		return tok, nil
	}
	return LoginAndCache()
}

// withToken resolves a token, runs fn, and — when the reply is a single
// not_authenticated and no CARS_API_TOKEN was supplied — logs in once and
// retries. Shared by the read (QueryAsAI) and write (ExecAsAI) paths so token
// handling cannot drift between them.
func withToken[T any](run func(token string) (T, error)) (T, error) {
	var zero T
	token, err := ResolveToken()
	if err != nil {
		return zero, err
	}
	out, err := run(token)
	if err != nil {
		var ae *ApiError
		if errors.As(err, &ae) && ae.Code == "not_authenticated" && config.C.Cars.APIToken == "" {
			tkn, lerr := LoginAndCache()
			if lerr != nil {
				return zero, lerr
			}
			return run(tkn)
		}
		return zero, err
	}
	return out, nil
}

// QueryAsAI runs a parameterised SELECT as the AI user, re-logging in once if
// the stored/env token lapsed. Returns the server's rows verbatim (raw JSON so
// column order and number formatting are preserved byte-for-byte).
func QueryAsAI(queryString string, params []any) ([]json.RawMessage, error) {
	return withToken(func(token string) ([]json.RawMessage, error) {
		resp, err := postJSON(map[string]any{"query": queryString, "params": params, "token": token})
		if err != nil {
			return nil, err
		}
		if resp.Success {
			var rows []json.RawMessage
			if len(resp.Data) > 0 && string(resp.Data) != "null" {
				if err := json.Unmarshal(resp.Data, &rows); err != nil {
					return nil, &ApiError{Code: "bad_response", Message: "cannot reach " + tenantEndpoint() + ": " + "invalid data payload"}
				}
			}
			return rows, nil
		}
		if resp.Code == "not_authenticated" {
			return nil, &ApiError{Code: "not_authenticated", Message: "token rejected"}
		}
		msg := "query failed"
		if resp.Error != "" {
			msg = resp.Error
		}
		return nil, &ApiError{Code: "query_failed", Message: msg}
	})
}

// SaveAndVerifyCredentials persists the AI user's credentials after a verified
// login, then caches the token, role and permissions.
func SaveAndVerifyCredentials(username, password string) (string, error) {
	res, err := doLogin(username, password)
	if err != nil {
		return "", err
	}
	if err := keyringx.SetUsername(username); err != nil {
		return "", err
	}
	if err := keyringx.SetPassword(password); err != nil {
		return "", err
	}
	if err := keyringx.SetToken(res.Token); err != nil {
		return "", err
	}
	if err := cacheRoleAndPermissions(res.User); err != nil {
		return "", err
	}
	return res.Token, nil
}

var _ = strings.TrimSpace
