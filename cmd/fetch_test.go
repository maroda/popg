package cmd_test

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	pop "github.com/maroda/popg/cmd"
)

func TestCatURL(t *testing.T) {
	t.Run("Creates a basic URL from pieces", func(t *testing.T) {
		want := "https://musicbrainz.org/ws/2/artist/?query=artist:Craque&fmt=json"
		root := "https://musicbrainz.org"
		api := "/ws/2/artist"
		query := "/?query=artist:"
		term := "Craque"
		fmt := "&fmt=json"

		got := pop.CatURL(root, api, query, term, fmt)
		assertStringContains(t, got, want)
	})
}

func TestMBQuestion_ArtistSearch(t *testing.T) {
	tests := []struct {
		question string
		qtype    string
		wantBool bool
		expect   string
	}{
		{
			question: "Craque", // Known artist
			qtype:    "artist",
			wantBool: true,
			expect:   "Craque",
		},
		{
			question: "Nslgzb", // nonsense (hopefully)
			qtype:    "artist",
			wantBool: false,
			expect:   "Not Found",
		},
	}

	for _, tt := range tests {
		t.Run("INTEGRATION - Artist search "+tt.question, func(t *testing.T) {
			ask := pop.NewMBQuestion(tt.question, tt.qtype)

			ok, got, err := ask.ArtistSearch(context.Background())
			assertError(t, err, nil)
			if ok != tt.wantBool {
				t.Errorf("Expected %v, got %v", tt.wantBool, ok)
			}
			assertStringContains(t, got, tt.expect)
		})
	}
}

// No testing for stdlib FetchBody functions: http.NewRequestWithContext(), io.ReadAll()
func TestMBQuestion_FetchBody_Errors(t *testing.T) {
	tests := []struct {
		name     string
		question string
		qType    string
		wantCode int
	}{
		{name: "503 rate limited", question: "Craque", qType: "artist", wantCode: 503},
		{name: "Non-200", question: "Craque", qType: "artist", wantCode: 404},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			webServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "", tt.wantCode)
			}))
			defer webServ.Close()

			q := pop.NewMBQuestion(tt.question, tt.qType)

			// set URL to the test server and then fetch
			q.QFullURL = webServ.URL
			code, _ := q.FetchBody(context.Background())
			assertStatus(t, code, tt.wantCode)
		})
	}

	t.Run("Errors on Host Unreachable", func(t *testing.T) {
		q := pop.NewMBQuestion("Craque", "artist")
		q.QFullURL = "https://badhost:4220"
		wantCode := 0
		code, err := q.FetchBody(context.Background())
		assertGotError(t, err)
		assertStatus(t, code, wantCode)
	})
}

func TestMBQuestion_ArtistSearch_Backoff(t *testing.T) {
	callCount := 0
	webServ := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer webServ.Close()

	q := pop.NewMBQuestion("Craque", "artist")
	q.QFullURL = webServ.URL

	start := time.Now()
	_, _, err := q.ArtistSearch(context.Background())
	elapsed := time.Since(start)

	// Should have retried 3 times (4 total attempts)
	if callCount != 4 {
		t.Errorf("Expected 4 calls, got %d", callCount)
	}

	// Backoff delays
	if elapsed < 7*time.Second {
		t.Errorf("Expected at least 7 seconds of backoff, got %s", elapsed)
	}

	// An error should be returned when reaching the backoff delay limit (3)
	assertGotError(t, err)
}

// Helpers //

func MakeTestWebServer(body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(body))
		if err != nil {
			log.Fatal(err)
		}
	}))
}

// Assertions //

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
