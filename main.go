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

	fmt.Printf("Looking for %s in MusicBrainz database...\n", find)
	p := pop.NewMBQuestion(find, "artist")

	// Init OTel for Grafana OTLP endpoint
	tp, err := pop.InitOTelGRF()
	if err != nil {
		slog.Error("Failed to init TraceProvider, continuing...", slog.Any("error", err))
	}
	defer tp.Shutdown(context.Background())

	okp, namep, err := p.ArtistSearch(context.Background())
	if !okp || err != nil {
		slog.Error("failed to find term!", slog.Bool("ok", okp), slog.Any("error", err))
	}
	fmt.Println("Found it! ::: " + namep)
}
