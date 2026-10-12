// Package keyringx is the OS-native credential store for the cars app AI user
// via github.com/zalando/go-keyring (no cgo): Windows Credential Manager,
// macOS Keychain, Linux Secret Service.
// Fail-closed: an unreadable store is an error, never a plaintext fallback.
// A missing entry is a null string, not an error.
//
// The token lives here (never on disk). The AI user's role name and
// permission names are cached here too, refreshed on every fresh login, so
// the stock command can follow the cars app's role->permissions model.
package keyringx

import (
	"errors"

	"github.com/zalando/go-keyring"
)

const (
	SecretUsername    = "cars-ai-username"
	SecretPassword    = "cars-ai-password"
	SecretToken       = "cars-ai-token"
	SecretRole        = "cars-ai-role"
	SecretPermissions = "cars-ai-permissions"
	SecretUserID      = "cars-ai-user-id"
)

const service = "whatsapp-agent-tool"

// Error mirrors SecretStoreError in the Node source.
type Error struct {
	Account string
	Err     error
}

func (e *Error) Error() string {
	msg := "credential store unavailable for " + e.Account + ": "
	if e.Err == nil {
		return msg + "unknown error"
	}
	return msg + e.Err.Error()
}

func (e *Error) Unwrap() error { return e.Err }

// get returns "" when the entry does not exist (Node null); store errors are
// wrapped and returned as an error.
func get(account string) (string, error) {
	v, err := keyring.Get(service, account)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return "", nil
		}
		return "", &Error{Account: account, Err: err}
	}
	return v, nil
}

func set(account, value string) error {
	if err := keyring.Set(service, account, value); err != nil {
		return &Error{Account: account, Err: err}
	}
	return nil
}

func del(account string) (bool, error) {
	if err := keyring.Delete(service, account); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return false, nil
		}
		return false, &Error{Account: account, Err: err}
	}
	return true, nil
}

func GetUsername() (string, error) { return get(SecretUsername) }
func GetPassword() (string, error) { return get(SecretPassword) }
func GetToken() (string, error)    { return get(SecretToken) }
func GetRole() (string, error)     { return get(SecretRole) }
func GetUserID() (string, error)   { return get(SecretUserID) }
func GetPermissions() (string, error) {
	return get(SecretPermissions)
}

func SetUsername(v string) error { return set(SecretUsername, v) }
func SetPassword(v string) error { return set(SecretPassword, v) }
func SetToken(v string) error    { return set(SecretToken, v) }
func SetRole(v string) error     { return set(SecretRole, v) }
func SetUserID(v string) error   { return set(SecretUserID, v) }
func SetPermissions(v string) error {
	return set(SecretPermissions, v)
}

func DeleteUsername() (bool, error)    { return del(SecretUsername) }
func DeletePassword() (bool, error)    { return del(SecretPassword) }
func DeleteToken() (bool, error)       { return del(SecretToken) }
func DeleteRole() (bool, error)        { return del(SecretRole) }
func DeleteUserID() (bool, error)      { return del(SecretUserID) }
func DeletePermissions() (bool, error) { return del(SecretPermissions) }
