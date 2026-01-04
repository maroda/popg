package cmd_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	pop "github.com/maroda/popg/cmd"
)

func TestBasicFetch(t *testing.T) {
	t.Run("Fetches a basic URL", func(t *testing.T) {
		newServ := MakeTestWebServer("a body string")
		defer newServ.Close()

		url := newServ.URL
		want := "a body string"
		code, got, err := pop.FetchBody(url)

		if code != 200 {
			t.Errorf("want 200, got %d", code)
		}
		if want != got {
			t.Errorf("want %q, got %q", want, got)
		}
		if err != nil {
			t.Errorf("want no error, got %v", err)
		}
	})
}

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

func TestQueryArtist_Integration(t *testing.T) {
	t.Run("INTEGRATION: Retrieve artist from MusicBrainz", func(t *testing.T) {
		want := "Craque"
		question := "Craque"
		qtype := "artist"

		ask := pop.NewMBQuestion(question, qtype)
		ok, got := ask.FindArtist()
		if !ok {
			t.Errorf("want true, got false")
		}
		if got != want {
			t.Errorf("want %q, got %q", want, got)
		}
		t.Log(got)
	})
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
