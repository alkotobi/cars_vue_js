// Package identity ports src/ai-identity.ts — a unique, irreplicable name for
// the assistant, minted BY CODE (CSPRNG) and persisted per WhatsApp profile.
package identity

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"whatsapp-agent-tool/internal/config"
	"whatsapp-agent-tool/internal/jsonx"
)

// File resolves exactly like Node's IDENTITY_FILE — data/ai-identity.json for the
// default account, data/<profile>-ai-identity.json otherwise.
func File(profile string, profileIsDefault bool) string {
	name := "ai-identity.json"
	if !profileIsDefault {
		name = fmt.Sprintf("%s-ai-identity.json", profile)
	}
	return filepath.Join(config.Root(), "data", name)
}

type Identity struct {
	Name      string `json:"name"`
	Token     string `json:"token"`
	Profile   string `json:"profile"`
	CreatedAt string `json:"createdAt"`
}

const ADJECTIVES = `amber arctic bold brave bright brisk calm clever cobalt cosmic crimson crisp daring deft eager ember fleet frosty gentle golden grand hardy hidden ivory jade keen lively lucid lunar mellow mighty noble nimble north opal prime quiet rapid royal rustic sable sharp silent silver sleek solar spry stark steel sunny swift tender tidal vivid warm wild wise young zesty azure coral dawn dusk fern`

const NOUNS = `falcon heron otter lynx raven cobra ibex puma wolf fox hawk owl panda tiger eagle bison comet lantern compass anchor harbor cedar willow maple orbit ember quartz onyx garnet topaz delta nebula canyon river summit glacier meadow forest desert aurora phoenix griffin dragon sphinx hydra beacon prism vector falconer cypher nomad sailor pilot ranger hunter keeper scholar artisan weaver smith mason scribe pioneer sentinel`

var adjectives = splitWords(ADJECTIVES)
var nouns = splitWords(NOUNS)

func splitWords(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		j := i
		for j < len(s) && s[j] != ' ' {
			j++
		}
		out = append(out, s[i:j])
		i = j + 1
	}
	return out
}

func cryptoPick(n int) int {
	if n <= 0 {
		return 0
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		// crypto/rand never fails on supported platforms; panic loudly rather
		// than fall back to a predictable source (identity must stay irreplicable).
		panic("identity: crypto/rand unavailable: " + err.Error())
	}
	return int(v.Int64())
}

// Generate mints a fresh identity: `<adjective>-<noun>-<12 hex>`.
func Generate(profile string) Identity {
	tokenBytes := make([]byte, 6)
	if _, err := rand.Read(tokenBytes); err != nil {
		panic("identity: crypto/rand.Read failed: " + err.Error())
	}
	token := hex.EncodeToString(tokenBytes) // 48-bit, unguessable
	name := adjectives[cryptoPick(len(adjectives))] + "-" + nouns[cryptoPick(len(nouns))] + "-" + token
	return Identity{Name: name, Token: token, Profile: profile, CreatedAt: time.Now().UTC().Format("2006-01-02T15:04:05.000Z07:00")}
}

func isValid(v Identity) bool {
	return v.Name != "" && v.Token != "" && v.Profile != "" && v.CreatedAt != ""
}

func readFrom(file string) (Identity, bool) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return Identity{}, false
	}
	var id Identity
	if err := json.Unmarshal(raw, &id); err != nil || !isValid(id) {
		return Identity{}, false
	}
	return id, true
}

func writeTo(file string, id Identity) error {
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	data, err := jsonx.MarshalIndent(id, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(file, append(data, '\n'), 0o644)
}

var cached Identity
var cachedFresh bool

// Ensure reads the persisted identity or mints + saves one on first run.
// Fresh is true only when it created the identity.
func Ensure() (Identity, bool) {
	if cached.Name == "" {
		file := File(config.WAProfile(), config.ProfileIsDefault())
		if id, ok := readFrom(file); ok {
			cached = id
		} else {
			cached = Generate(config.C.WAProfile)
			if err := writeTo(file, cached); err != nil {
				panic("identity: cannot persist to " + file + ": " + err.Error())
			}
			cachedFresh = true
		}
	}
	return cached, cachedFresh
}

// Get returns the identity, minting it if needed.
func Get() Identity {
	id, _ := Ensure()
	return id
}

// Reset re-mints and overwrites the persisted identity (explicit, code-only).
func Reset() Identity {
	id := Generate(config.C.WAProfile)
	file := File(config.WAProfile(), config.ProfileIsDefault())
	if err := writeTo(file, id); err != nil {
		panic("identity: cannot persist to " + file + ": " + err.Error())
	}
	cached = id
	cachedFresh = true
	return id
}
