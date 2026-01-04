package cmd_test

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
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

func assertInt64(t *testing.T, got, want int64) {
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
