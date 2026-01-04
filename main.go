package main

import (
	"fmt"
	"log/slog"
	"os"

	pop "github.com/maroda/popg/cmd"
)

func main() {
	term := os.Args[1:]
	q := pop.NewMBQuestion(term[0], "artist")
	ok, name := q.FindArtist()
	if !ok {
		slog.Error("failed to find term!")
	}
	fmt.Println("Found it! ::: " + name)
}
