//go:build kolavatardev

// kolavatar-gallery-json generates a curated 100-descriptor JSON catalogue
// for the TS playground's Gallery view.
//
// The composition is deliberately uniform so the wall reads as a "potential
// of the output" exhibit at a glance:
//
//   4 high-symmetry groups × 5 seeds per group × 5 score points = 100
//
//   symmetries: 4mm_d, 4, 2mm, 2m_dm_d   (the four most ornate Gopalan groups)
//   grid:       13 (the largest odd grid the SDK supports up to 15;
//                   13 leaves room for figure-ground breathing without
//                   flattening into noise)
//   scores:     0.20, 0.40, 0.60, 0.80, 0.95
//
// Each row in the output corresponds to one (symmetry, seed) pair, and the
// five entries within a row walk the score axis low → high — so the gallery
// UI can render rows as vertical strips that read as score progressions.
//
// Output: a JSON array on stdout. Pipe to ts/public/gallery.json:
//
//   go run -tags=kolavatardev ./cmd/kolavatar-gallery-json > ../ts/public/gallery.json
//
// Re-run after any SDK change that affects descriptors so the gallery stays
// in sync with the live generator.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gklsndr/kolavatar"
)

// galleryEntry is the per-design row in the emitted JSON. Descriptor is held
// as RawMessage so the Generator's hand-marshalled, byte-stable output is
// preserved verbatim (re-marshalling through the standard encoder would
// change field order and break parity with /v1/avatars/* responses).
type galleryEntry struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Seed        string          `json:"seed"`
	Score       float64         `json:"score"`
	Grid        int             `json:"grid"`
	Symmetry    string          `json:"symmetry"`
	Descriptor  json.RawMessage `json:"descriptor"`
}

func main() {
	const grid = 13
	symmetries := []string{"4mm_d", "4", "2mm", "2m_dm_d"}

	// Five hand-picked seeds per symmetry. The names are first-name-flavoured
	// so the gallery feels populated (vs. cryptographic-looking hashes), but
	// they're just strings — the seed-driven layout is the point. Picked by
	// eyeballing a few sweeps; replace freely if a particular row reads
	// poorly after a generator change.
	seedsByGroup := map[string][]string{
		"4mm_d":   {"asha", "ravi", "meera", "vikram", "tara"},
		"4":       {"karan", "divya", "arjun", "priya", "kavya"},
		"2mm":     {"aryan", "anjali", "raghav", "shreya", "aditya"},
		"2m_dm_d": {"deepika", "rohan", "neha", "vivek", "sanjay"},
	}

	scores := []float64{0.20, 0.40, 0.60, 0.80, 0.95}

	entries := make([]galleryEntry, 0, len(symmetries)*5*len(scores))

	for _, sym := range symmetries {
		seeds, ok := seedsByGroup[sym]
		if !ok {
			log.Fatalf("missing seed list for symmetry %q", sym)
		}
		for _, seed := range seeds {
			g, err := kolavatar.New(
				kolavatar.WithGrid(grid),
				kolavatar.WithSymmetryPolicy(kolavatar.SymmetryForced(sym)),
			)
			if err != nil {
				log.Fatalf("New(grid=%d, sym=%s): %v", grid, sym, err)
			}
			for _, score := range scores {
				d, err := g.Generate(context.Background(), seed, score)
				if err != nil {
					log.Fatalf("Generate(seed=%s, score=%g, sym=%s): %v",
						seed, score, sym, err)
				}
				descBytes, err := d.MarshalJSON()
				if err != nil {
					log.Fatalf("MarshalJSON: %v", err)
				}
				entries = append(entries, galleryEntry{
					ID:         fmt.Sprintf("%s-%s-%03d", sym, seed, int(score*100)),
					Title:      fmt.Sprintf("%s @ %d%%", seed, int(score*100)),
					Seed:       seed,
					Score:      score,
					Grid:       grid,
					Symmetry:   sym,
					Descriptor: descBytes,
				})
			}
		}
	}

	out, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Fatalf("MarshalIndent: %v", err)
	}
	if _, err := os.Stdout.Write(out); err != nil {
		log.Fatalf("write: %v", err)
	}
	if _, err := os.Stdout.Write([]byte("\n")); err != nil {
		log.Fatalf("write: %v", err)
	}

	log.Printf("wrote %d gallery entries (%d sym × %d seeds × %d scores) to stdout",
		len(entries), len(symmetries), 5, len(scores))
}
