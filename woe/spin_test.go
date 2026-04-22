package woe_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	wheel "github.com/maroda/popg/woe"
)

/*

1. Domain Events (Core)
   - what events exist
   - what they mean in business terms

2. Event Contract (Port)
   - Payload / Event schema (what fields, what types)
     - Generate a UUID per event, include in every delivery attempt, never reuse it
   - Event taxonomy (what event types exist, naming conventions)
     - Default events are available without auth for testing: spin, random entities
   - Delivery semantics (at-least-once, idempotency key)
     - UUID per event is used to dedupe retries
   - Versioning strategy

3. Delivery Mechanism (Adapter)
   - HTTP specifics
   - authentication
   - retry + idempotency
   - status code semantics

*/

func TestWheel_Spin(t *testing.T) {
	t.Run("Returns a default wheel spin", func(t *testing.T) {
		words := []string{"one", "two", "three", "four", "five", "six", "seven"}
		we, err := wheel.NewWheel(&words)
		assertError(t, err, nil)

		r := httptest.NewRequest("GET", "/spin", nil)
		w := httptest.NewRecorder()
		mux := we.SetupMux()

		mux.ServeHTTP(w, r)
		assertStatus(t, w.Code, http.StatusOK)

		t.Log(we.Token, w.Body.String())
	})

	t.Run("Returns a browser wheel spin", func(t *testing.T) {
		json := `{
  "id": "331c7a00-1e70-11f1-85c8-53e0cbee6e98",
  "version": "0.1.0",
  "event_type": "spin.browser.wheel",
  "timestamp": "2026-03-12T10:00:00Z",
  "data": {
    "entries": ["eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen"]
  }
}`
		reader := bytes.NewReader([]byte(json))
		sendbody := io.Reader(reader)

		// Create a new default wheel using default entries
		words := []string{"one", "two", "three", "four", "five", "six", "seven"}
		we, err := wheel.NewWheel(&words)
		assertError(t, err, nil)

		// Run the server
		ts := httptest.NewServer(we.SetupMux())
		defer ts.Close()

		resp, err := http.Post(ts.URL+"/spin", "application/json", sendbody)
		assertError(t, err, nil)
		assertStatus(t, resp.StatusCode, http.StatusOK)

		url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		assertError(t, err, nil)
		defer ws.Close()

		// Read from the websocket
		result := &wheel.SpinDataWS{}
		err = ws.ReadJSON(result)
		assertError(t, err, nil)
		t.Log(result)

		err = ws.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		assertError(t, err, nil)
	})
}

func TestWheel_SpinArgsHandler(t *testing.T) {
	// TODO: This should also include a test for idempotent retries and auth failures

	t.Run("Returns a configured wheel spin", func(t *testing.T) {
		// example_payload.json
		json := `{
  "id": "331c7a00-1e70-11f1-85c8-53e0cbee6e98",
  "version": "0.1.0",
  "event_type": "spin.custom.wheel",
  "timestamp": "2026-03-12T10:00:00Z",
  "data": {
    "entries": ["eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen"]
  }
}`
		reader := bytes.NewReader([]byte(json))
		sendbody := io.Reader(reader)

		// Create a new default wheel using default entries
		words := []string{"one", "two", "three", "four", "five", "six", "seven"}
		we, err := wheel.NewWheel(&words)
		assertError(t, err, nil)

		// When a wheel is created, it comes with a token.
		// This is used to authenticate.
		mac := hmac.New(sha256.New, []byte(we.Token))
		mac.Write([]byte(json))
		hexmac := hex.EncodeToString(mac.Sum(nil))

		// Our API Contract ('port') requires:
		// - message auth code (HMAC) provided by the client
		// - spin ID provided by the client
		r := httptest.NewRequest("POST", "/s/test", sendbody)
		r.Header.Set("X-Hub-Signature", "sha256="+hexmac)
		r.Header.Set("X-Spin-ID", uuid.New().String())
		w := httptest.NewRecorder()
		mux := we.SetupMux()

		mux.ServeHTTP(w, r)
		assertStatus(t, w.Code, http.StatusOK)
	})
}

func TestSpinClient(t *testing.T) {
	t.Run("Client can retrieve a configured wheel spin", func(t *testing.T) {
		json := `{
  "id": "331c7a00",
  "version": "0.1.0",
  "event_type": "spin.custom.wheel",
  "timestamp": "2026-03-12T10:00:00Z",
  "data": {
    "entries": ["eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen"]
  }
}`

		we, err := wheel.NewWheel(&[]string{"one, two"})
		assertError(t, err, nil)

		// Run the server
		ts := httptest.NewServer(we.SetupMux())
		defer ts.Close()

		// Run the client with a payload

		url := ts.URL + "/s/test"
		spun, err := wheel.SpinClient(url, we.Token, json)
		assertError(t, err, nil)
		assertStringInList(t, spun, []string{"eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen"})
	})
}

func TestWheel_RandTermHandler(t *testing.T) {
	t.Run("Returns a random entities wheel spin", func(t *testing.T) {
		// This list will be replaced below by the internal function
		words := []string{"one", "two", "three", "four", "five", "six", "seven"}
		we, err := wheel.NewWheel(&words)
		assertError(t, err, nil)

		body := []byte(strings.Join(words, ", "))
		sendbody := io.Reader(bytes.NewReader(body))

		r := httptest.NewRequest("POST", "/randomize", sendbody)
		w := httptest.NewRecorder()
		mux := we.SetupMux()

		mux.ServeHTTP(w, r)
		assertStatus(t, w.Code, http.StatusOK)
		assertStringContains(t, w.Body.String(), "not implemented")
	})
}

// Helpers //

func assertError(t testing.TB, got, want error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Errorf("got error %q want %q", got, want)
	}
}

func assertGotError(t testing.TB, got error) {
	t.Helper()
	if got == nil {
		t.Errorf("Expected an error but got %q", got)
	}
}

func assertStatus(t testing.TB, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct status, got %d, want %d", got, want)
	}
}

func assertInt(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("did not get correct value, got %d, want %d", got, want)
	}
}

func assertStringContains(t *testing.T, full, want string) {
	t.Helper()
	if !strings.Contains(full, want) {
		t.Errorf("Did not find %q, expected string contains %q", want, full)
	}
}

func assertStringInList(t *testing.T, full string, want []string) {
	t.Helper()
	for _, w := range want {
		if full == w {
			return
		}
	}
	t.Errorf("Did not find %q, expected string contains %q", want, full)
}
