//go:build kolavatardev

// kolavatar-playground serves an interactive HTML page with controls for
// every kolavatar generation parameter. Used to fine-tune the look of a
// kolam by hand: seed, score, grid, symmetry, coverage, strand length,
// motif pool size, and palette are all live-controllable.
//
// Run: go run -tags=kolavatardev ./cmd/kolavatar-playground
//      → open http://localhost:8080/
//
// The UI handlers live in internal/goplayground so cmd/kolavatar-web can
// mount the same playground at a nested path for the hosted deployment.
package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gklsndr/kolavatar"
	"github.com/gklsndr/kolavatar-playground/go/internal/goplayground"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	mux := http.NewServeMux()
	goplayground.Register(mux, "")

	// Mount the SDK's built-in HTTP surface (/v1/avatars/*) on the same
	// server so the TS playground's "Live API" panel can talk to this
	// binary directly — without needing kolavatar-dev to also be running.
	apiGen, err := kolavatar.New()
	if err != nil {
		log.Fatal(err)
	}
	apiGen.RegisterRoutes(mux, nil)

	log.Printf("kolavatar-playground listening on http://localhost%s/", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
