// Package jsonx marshals exactly like Node's JSON.stringify: no HTML escaping.
//
// encoding/json escapes <, > and & to \u003c / \u003e / \u0026 by default,
// but JSON.stringify leaves them literal. State files, FAQ rows and car names
// routinely contain "<" / "&", so every byte-parity output must go through here
// (SetEscapeHTML(false)). Encoder.Encode appends a newline that Marshal does
// not, so it is trimmed to keep the two functions interchangeable with the
// stdlib names they replace.
package jsonx

import (
	"bytes"
	"encoding/json"
)

// Marshal is json.Marshal without HTML escaping and without a trailing newline.
func Marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// MarshalIndent is json.MarshalIndent without HTML escaping and without a
// trailing newline.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}
