package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/time/rate"
)

const (
	mzEndpoint        = "https://musicbrainz.org/ws/2/"
	mzFmtString       = "&fmt=json"
	httpClientTimeout = 5 * time.Second
)

// MBQuestion is the heart of operations, where data and client configs are kept
type MBQuestion struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
	userAgent   string
	QString     string
	QType       string
	QFullURL    string
	RespBody    string
}

// MBAnswerArtist is the primary target for unmarshalling the results JSON.
type MBAnswerArtist struct {
	Count   int      `json:"count"`
	Offset  int      `json:"offset"`
	Artists []Artist `json:"artists"`
}

// Artist is a sub-target for unmarshalling the specific Artist keys we care about.
type Artist struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Type           string `json:"type"`
	Country        string `json:"country,omitempty"`
	Disambiguation string `json:"disambiguation,omitempty"`
	Score          int    `json:"score"`
}

// NewMBQuestion creates the object for asking and storing the answer.
func NewMBQuestion(qSearch, qType string) *MBQuestion {
	var q string
	switch qType {
	case "artist":
		q = "/?query=artist:"
	}
	full := CatURL(mzEndpoint, qType, q, qSearch, mzFmtString)
	slog.Debug("Question URL", slog.String("url", full))

	return &MBQuestion{
		httpClient:  &http.Client{Timeout: httpClientTimeout},
		rateLimiter: rate.NewLimiter(rate.Limit(1*time.Second), 1),
		userAgent:   "popg/v0.1.0 ( https://github.com/maroda/popg )",
		QString:     qSearch,
		QType:       qType,
		QFullURL:    full,
	}
}

// ArtistSearch uses context for the request and includes retry logic for backoff
func (mbq *MBQuestion) ArtistSearch(ctx context.Context) (bool, string, error) {
	ctx, span := otel.Tracer("musicbrainz/artistsearch").Start(ctx, "ArtistSearch")
	defer span.End()

	span.SetAttributes(
		attribute.String("find.string", mbq.QString),
		attribute.String("http.url", mbq.QFullURL))

	// Initiate rate limiting
	if err := mbq.rateLimiter.Wait(ctx); err != nil {
		span.RecordError(err)
		slog.Error("rate limit wait", slog.String("search", mbq.QString))
		return false, "", fmt.Errorf("rate limit wait: %w", err)
	}

	// Wrap request in retry logic
	retries := 3
	base := 1 * time.Second
	for a := 0; a <= retries; a++ {
		span.SetAttributes(attribute.Int("retry.attempt", a))

		// Try to Fetch the body contents
		code, err := mbq.FetchBody(ctx)
		isRetry := err != nil || code == 503

		if !isRetry {
			if err != nil {
				span.RecordError(err)
				slog.Error("Unrecoverable error, no retry", slog.String("url", mbq.QFullURL))
				return false, "", fmt.Errorf("unrecoverable error: %w", err)
			}

			// Success!
			// No more retries needed
			break
		}

		// Total retries reached, log and error out
		if a == retries {
			if err != nil {
				span.RecordError(err)
				slog.Error("Max retries reached with error", slog.Any("error", err))
				return false, "", fmt.Errorf("max retries reached with error: %w", err)
			}
			span.RecordError(fmt.Errorf("max retries reached"))
			slog.Error("Max retries reached, last status: 503", slog.String("url", mbq.QFullURL))
			return false, "", fmt.Errorf("max retries exceeded, last status: 503")
		}

		// Retries can continue
		// bitwise /a/ for exponential backoff on /delay/ time: 1s, 2s, 4s
		delay := base * time.Duration(1<<a)
		select {
		case <-time.After(delay):
			// wait /delay/ seconds before continuing
			continue
		case <-ctx.Done():
			span.RecordError(ctx.Err())
			return false, "", fmt.Errorf("context done: %w", ctx.Err())
		}
	}

	newartist := &MBAnswerArtist{}
	err := json.Unmarshal([]byte(mbq.RespBody), newartist)
	if err != nil {
		span.RecordError(err)
		slog.Error("Failed to unmarshal artist info", slog.String("url", mbq.QFullURL))
		return false, "", fmt.Errorf("failed to unmarshal artist info: %w", err)
	}

	if newartist.Count == 0 {
		return false, "Not Found: " + mbq.QString, nil
	}

	name := newartist.Artists[0].Name
	return true, name, nil
}

// FetchBody reads a url and returns the status code and any errors,
// the actual contents of the fetch are put in the struct.
func (mbq *MBQuestion) FetchBody(ctx context.Context) (int, error) {
	ctx, span := otel.Tracer("musicbrainz/artistsearch").Start(ctx, "FetchBody")
	defer span.End()

	span.SetAttributes(
		attribute.String("http.url", mbq.QFullURL),
		attribute.String("http.method", "GET"))

	// Create an http request and initiate new context,
	// passed around for tracing and rate limiting
	req, err := http.NewRequestWithContext(ctx, "GET", mbq.QFullURL, nil)
	if err != nil {
		span.RecordError(err)
		slog.Error("Failed to make request", slog.String("url", mbq.QFullURL))
		return 0, err
	}
	req.Header.Set("User-Agent", mbq.userAgent)

	resp, err := mbq.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		slog.Error("Failed to fetch body", slog.String("url", mbq.QFullURL))
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	code := resp.StatusCode
	span.SetAttributes(attribute.Int("http.status_code", code))

	if code == 503 {
		span.RecordError(fmt.Errorf("rate limit exceeded"))
		slog.Warn("Request hit rate limit", slog.String("url", mbq.QFullURL))
		return code, nil
	} else if code != 200 {
		span.RecordError(fmt.Errorf("unexpected status code: %d", code))
		slog.Error("Failed to fetch",
			slog.String("url", mbq.QFullURL),
			slog.String("status", resp.Status))
		return code, errors.New(resp.Status)
	}

	read, err := io.ReadAll(resp.Body)
	if err != nil {
		span.RecordError(err)
		slog.Error("Failed to read response body", slog.String("url", mbq.QFullURL))
		return code, err
	}

	mbq.RespBody = string(read)

	slog.Info("Data fetched", slog.Int("status", code), slog.String("url", mbq.QFullURL))
	return code, nil
}

// CatURL takes arbitrary number of strings and concatenates them together
func CatURL(u ...string) string {
	var fullURL string
	for _, p := range u {
		fullURL = fullURL + p
	}
	slog.Debug("URL created", slog.String("url", fullURL))
	return fullURL
}
