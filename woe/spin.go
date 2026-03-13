package woe

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

/*

The idea is that the wheel will be controlled by API webhooks.
Since the original model is a "Wheel of Expertise", that's the
metaphor used here in the code.

Right now this uses rand to get random numbers.
In the future, it could support sequential selection based on wheel physics.

So events like this can be triggered:

	/s/15w+fast+3x

or "spin a 15 sided wheel at least three times"
where the body contains the contents of the wheel

Authentication: A token is created for the wheel.
This may be tied to the version of the wheel that's being displayed,
as a way to synchronize browsers to wheels (header inspection).
And use authentication that synchronizes to this ID, i.e. HMAC

This is a single wheel. It doesn't track multiple wheels yet.
One wheel, one token, per process.

*/

type Payload struct {
	ID        string              `json:"id"`
	Version   string              `json:"version"`
	EventType string              `json:"event_type"`
	Timestamp time.Time           `json:"timestamp"`
	Data      map[string][]string `json:"data"`
}

func (p *Payload) validatePayload() error {
	if p.ID == "" {
		errors.New("missing id")
	}
	if p.Version == "" {
		errors.New("missing version")
	}
	if p.EventType == "" {
		errors.New("missing event_type")
	}
	if p.Timestamp.IsZero() {
		errors.New("missing timestamp")
	}
	if p.Data["entries"] == nil {
		return errors.New("missing entries")
	}
	return nil
}

type Wheel struct {
	mu         sync.Mutex
	Token      string       // Token for this wheel, currently a random UUID
	SpinID     string       // SpinID for this spin, created at producer
	SpunString string       // Current spun entry
	Entries    *[]string    // Each space of the wheel
	Server     *http.Server // Server for this wheel
	Mux        *mux.Router  // Router for this Server
}

func NewWheel(wd *[]string) (*Wheel, error) {
	token, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	wheel := &Wheel{
		Token:   token.String(),
		Entries: wd,
	}

	return wheel, nil
}

// SetupMux configures endpoint routing
func (we *Wheel) SetupMux() *mux.Router {
	r := mux.NewRouter()

	// No auth required on default wheel operations
	r.HandleFunc("/randomize", we.RandTermHandler) // Updates the default wheel entries
	r.HandleFunc("/spin", we.SpinHandler)          // Spins a default wheel

	// Auth required on configurable wheels
	r.HandleFunc("/s/{args}", we.SpinArgsHandler) // Spins a configurable wheel, auth required

	return r
}

// RandTermHandler kicks off a process to replace the default wheel entries
func (we *Wheel) RandTermHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("not implemented"))
}

// SpinHandler for simple webhook
func (we *Wheel) SpinHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(we.Spin(1)))

	slog.Info("Wheel spin",
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("path", r.URL.Path))
}

func (we *Wheel) validateSig(r *http.Request, body []byte) bool {
	sig := r.Header.Get("X-Hub-Signature")
	if sig == "" {
		slog.Error("signature header not found")
		return false
	}

	mac := hmac.New(sha256.New, []byte(we.Token))
	mac.Write(body)
	hexmac := hex.EncodeToString(mac.Sum(nil))
	expect := "sha256=" + hexmac

	// Constant time comparison instead of string equality prevents timing attacks.
	return hmac.Equal([]byte(sig), []byte(expect))
}

// SpinArgsHandler for webhook with args and a configuration body
func (we *Wheel) SpinArgsHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("Failed to read configuration")
		http.Error(w, "Failed to read configuration", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !we.validateSig(r, body) {
		slog.Error("Failed to validate authentication signature")
		http.Error(w, "authorization required", http.StatusUnauthorized)
		return
	}

	if !we.spinIDCacheMiss(w, r) {
		// No cache miss, retry ID exists
		// The client should be notified to stop retries
		w.WriteHeader(http.StatusOK)

		slog.Warn("Wheel SpinID found in cache")
		return
	}

	// Cache miss, it's a new spin!

	// TODO: This does not update the entries yet!!!
	// we.Entities needs to be updated with the webhook payload

	// Parse Payload
	payload := &Payload{}
	err = json.Unmarshal(body, payload)
	if err != nil {
		slog.Error("Failed to unmarshal payload", slog.Any("error", err))
		http.Error(w, "Failed to unmarshal payload", http.StatusBadRequest)
		return
	}

	// Validate Payload
	err = payload.validatePayload()
	if err != nil {
		slog.Error("Failed to validate payload", slog.Any("error", err))
		http.Error(w, "Failed to validate payload", http.StatusBadRequest)
		return
	}

	// Record Payload Data
	entries := payload.Data["entries"]

	// Write new entries, spin, and record the new spun entry
	we.mu.Lock()
	we.Entries = &entries
	spun := we.Spin(1)
	we.SpunString = spun
	we.mu.Unlock()

	w.Write([]byte(spun))

	slog.Info("Wheel spin",
		slog.String("entry", spun),
		slog.String("method", r.Method),
		slog.String("request", r.RequestURI),
		slog.String("remote_addr", r.RemoteAddr),
		slog.String("path", r.URL.Path))
}

// spinIDCacheMiss looks for the current SpinID in a cache kept in Wheel.SpinID
// if the SpinID exists, it's been successfully processed, and returns false.
// if the SpinID is new, it's a cache miss, so it is cached and returns true.
func (we *Wheel) spinIDCacheMiss(w http.ResponseWriter, r *http.Request) bool {
	id := r.Header.Get("X-Spin-ID")

	if id == "" {
		slog.Error("Missing X-Spin-ID")
		http.Error(w, "Missing X-Spin-ID", http.StatusBadRequest)
		return false
	}

	if id == we.SpinID {
		// Already seen this Spin ID
		return false
	}

	// Have not seen this ID, record it
	we.mu.Lock()
	we.SpinID = id
	we.mu.Unlock()

	return true
}

// Spin produces a random string from a list of words,
// currently it's a randomized selection from the slice.
func (we *Wheel) Spin(rot int) string {
	spaces := len(*we.Entries)
	spun := rand.Int32N(int32(spaces)) * int32(rot)
	selected := (*we.Entries)[spun%int32(spaces)]

	return selected
}

// SpinClient is a built-in CLI for the Wheel server
func SpinClient(url, token, json string) (string, error) {
	client := http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(json))
	if err != nil {
		slog.Error("Failed to create request", slog.Any("error", err))
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(token))
	mac.Write([]byte(json))
	hexmac := hex.EncodeToString(mac.Sum(nil))
	req.Header.Set("X-Hub-Signature", "sha256="+hexmac)
	req.Header.Set("X-Spin-ID", uuid.New().String())

	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Failed to send request", slog.Any("error", err))
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	// Some error code checking here will be nice

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", slog.Any("error", err))
		return "", err
	}

	slog.Info("Wheel spun!")
	return string(read), nil
}
