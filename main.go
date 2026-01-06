package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	pop "github.com/maroda/popg/cmd"
)

func main() {
	term := os.Args[1:]
	find := term[0]

	// Init OTel for Grafana OTLP endpoint
	tp, err := pop.InitOTelGRF()
	if err != nil {
		slog.Error("Failed to init TraceProvider, continuing...", slog.Any("error", err))
	}
	defer tp.Shutdown(context.Background())

	// Run program
	fmt.Printf("Looking for %s in MusicBrainz database...\n", find)
	p := pop.NewMBQuestion(find, "artist")
	ok, name, err := p.ArtistSearch(context.Background())
	if !ok || err != nil {
		slog.Error("failed to find term!",
			slog.Bool("ok", ok), slog.Any("error", err))
	}
	fmt.Println("Found it! ::: " + name)
}
