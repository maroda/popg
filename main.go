package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	pop "github.com/maroda/popg/cmd"
	wheel "github.com/maroda/popg/woe"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
	// Init OTel for Grafana OTLP endpoint
	tp, err := pop.InitOTelGRF()
	if err != nil {
		slog.Error("Failed to init TraceProvider, continuing...", slog.Any("error", err))
	}
	defer tp.Shutdown(context.Background())

	// Runtime flags
	// -artist <term>
	// Without flags, the server starts.
	artist := flag.String("artist", "", "Search for artist in Musicbrainz")
	client := flag.String("client", "", "Client mode with server Token")
	url := flag.String("url", "", "URL for Spinning the Wheel")
	json := flag.String("json", "", "JSON Payload")
	flag.Parse()
	find := *artist

	// Client Mode //

	if *client != "" {
		if *json == "" {
			slog.Error("Usage: -url <wheel.url> -client <token> -json <payload>")
			os.Exit(1)
		}
		if *url == "" {
			slog.Error("Usage: -url <wheel.url> -client <token> -json <payload>")
			os.Exit(1)
		}

		spun, err := wheel.SpinClient(*url, *client, *json)
		if err != nil {
			slog.Error("Failed to spin!", slog.Any("error", err))
			os.Exit(1)
		}

		fmt.Println(spun)
		os.Exit(0)
	}

	// Single Lookup //

	if find != "" {
		fmt.Printf("Looking for %s in MusicBrainz database...\n", find)
		p := pop.NewMBQuestion(find, "artist")
		ok, name, err := p.ArtistSearch(context.Background())
		if !ok || err != nil {
			slog.Error("failed to find term!",
				slog.Bool("ok", ok), slog.Any("error", err))
		}
		fmt.Println("Found ::: " + name)
		os.Exit(0)
	}

	// Server Startup //

	// Create a default wheel
	words := []string{"one", "two", "three", "four", "five", "six", "seven"}
	we, err := wheel.NewWheel(&words)
	if err != nil {
		fmt.Printf("Failed to init Wheel: %q", err)
		slog.Error("Failed to init Wheel", slog.Any("error", err))
		os.Exit(1)
	}

	we.Server = &http.Server{
		Addr: ":1234",
		Handler: otelhttp.NewHandler(we.SetupMux(), "Wheel of Expertise",
			otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
				return r.Method + " " + operation + " " + r.URL.Path
			})),
	}

	slog.Info("Starting server on port 1234", slog.String("token", we.Token))
	if err = we.Server.ListenAndServe(); err != nil {
		slog.Error("Failed to start server!", slog.Any("error", err))
		os.Exit(1)
	}
}
