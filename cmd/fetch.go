package cmd

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type MBQuestion struct {
	QString  string
	QType    string
	QFullURL string
	RespBody string
}

type MBAnswerArtist struct {
	Count   int      `json:"count"`
	Offset  int      `json:"offset"`
	Artists []Artist `json:"artists"`
}

type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Country        string `json:"country,omitempty"`
	Disambiguation string `json:"disambiguation,omitempty"`
	Score          int    `json:"score"`
}

const (
	mzEndpoint  = "https://musicbrainz.org/ws/2/"
	mzFmtString = "&fmt=json"
)

func NewMBQuestion(qSearch, qType string) *MBQuestion {
	var q string
	switch qType {
	case "artist":
		q = "/?query=artist:"
	}
	full := CatURL(mzEndpoint, qType, q, qSearch, mzFmtString)
	slog.Info("Question URL", slog.String("url", full))

	return &MBQuestion{
		QString:  qSearch,
		QType:    qType,
		QFullURL: full,
	}
}

func (mbq *MBQuestion) FindArtist() (bool, string) {
	_, info, err := FetchBody(mbq.QFullURL)
	if err != nil {
		slog.Error("Failed to fetch artist", slog.String("url", mbq.QFullURL))
		return false, ""
	}

	newartist := &MBAnswerArtist{}
	err = json.Unmarshal([]byte(info), newartist)
	if err != nil {
		slog.Error("Failed to unmarshal artist info", slog.String("url", mbq.QFullURL))
	}

	name := newartist.Artists[0].Name
	return true, name
}

// FetchBody reads a url and returns the status code with body as a string
func FetchBody(url string) (int, string, error) {
	slog.Info("Fetching body", slog.String("url", url))
	client := &http.Client{}
	r, err := client.Get(url)
	if err != nil {
		slog.Error("fetch body error")
		return 0, "", err
	}
	defer r.Body.Close()
	read, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("read body error")
		return 0, "", err
	}

	return r.StatusCode, string(read), nil
}

// CatURL takes arbitrary number of strings and concatenates them together
func CatURL(u ...string) string {
	var fullURL string
	for _, p := range u {
		fullURL = fullURL + p
	}
	slog.Debug("Catting URL", slog.String("url", fullURL))
	return fullURL
}
