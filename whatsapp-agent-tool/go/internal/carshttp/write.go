// Write support for the cars app's `query` passthrough. The app sends INSERTs
// over the same action=query endpoint it uses for SELECTs (api.php runs them
// with executeQuery() and answers `{success, lastInsertId}`), so this is the
// one write channel the tool has. Kept separate from the read path so it is
// obvious at the call site that a command mutates data.
package carshttp

import (
	"bytes"
	"encoding/json"
	"strconv"
)

// WriteResult is what a successful mutation answers: the new row id (INSERT)
// or the number of rows touched (UPDATE/DELETE). Both are 0 when the statement
// did not report them, which is not by itself an error.
type WriteResult struct {
	LastInsertID int64
	AffectedRows int64
}

// ExecAsAI runs a parameterised write as the AI user, with the same token retry
// as QueryAsAI. Only fixed, parameterised statements are passed here — the
// caller never builds SQL from user input.
func ExecAsAI(queryString string, params []any) (*WriteResult, error) {
	return withToken(func(token string) (*WriteResult, error) {
		resp, err := postJSON(map[string]any{"query": queryString, "params": params, "token": token})
		if err != nil {
			return nil, err
		}
		if resp.Success {
			return &WriteResult{
				LastInsertID: flexInt64(resp.LastInsertID),
				AffectedRows: flexInt64(resp.AffectedRows),
			}, nil
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

// flexInt64 reads an integer that the server may have sent either bare (12) or
// quoted ("12"). api.php's INSERT branch forwards PDO::lastInsertId() uncast,
// which is a string, while UPDATE/DELETE send an int — the parser accepts both
// rather than depending on which shape a given endpoint happens to use.
func flexInt64(raw json.RawMessage) int64 {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || string(raw) == "null" {
		return 0
	}
	var n int64
	if err := json.Unmarshal(raw, &n); err == nil {
		return n
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if parsed, perr := strconv.ParseInt(s, 10, 64); perr == nil {
			return parsed
		}
	}
	return 0
}
