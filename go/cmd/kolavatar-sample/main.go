//go:build kolavatardev

// kolavatar-sample renders a single descriptor to stdout. Used to produce
// reference SVG samples; gated by the kolavatardev tag so it's not in prod.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gklsndr/kolavatar"
)

func main() {
	seed := flag.String("seed", "asha@example.com", "avatar seed")
	score := flag.Float64("score", 0.5, "score in [0,1]")
	grid := flag.Int("grid", 7, "grid dimension (odd, 3..15)")
	out := flag.String("out", "", "output SVG file (empty = stdout)")
	dumpJSON := flag.Bool("json", false, "also print descriptor JSON to stderr")
	sym := flag.String("symmetry", "", "force symmetry group (empty = score-driven)")
	coverage := flag.Float64("coverage", 0, "target dot coverage in [0,1]; 0 = natural density")
	strandLen := flag.Int("strand-length", 1, "strand-length level: 0=compact, 1=balanced, 2=long, 3=helix")
	flag.Parse()

	opts := []kolavatar.Option{kolavatar.WithGrid(*grid)}
	if *sym != "" {
		opts = append(opts, kolavatar.WithSymmetryPolicy(kolavatar.SymmetryForced(*sym)))
	}
	if *coverage > 0 {
		opts = append(opts, kolavatar.WithCoverage(*coverage))
	}
	opts = append(opts, kolavatar.WithStrandLength(*strandLen))
	g, err := kolavatar.New(opts...)
	if err != nil {
		log.Fatal(err)
	}
	d := g.MustGenerate(*seed, *score)
	svg, err := d.SVG()
	if err != nil {
		log.Fatal(err)
	}

	if *out != "" {
		if err := os.WriteFile(*out, []byte(svg), 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s (id=%s, symmetry=%s, tier=%d)\n",
			*out, d.ID(), d.SymmetryGroup, d.Tier)
	} else {
		fmt.Println(svg)
	}

	if *dumpJSON {
		j, _ := json.MarshalIndent(map[string]any{
			"id":        d.ID(),
			"tier":      d.Tier,
			"symmetry":  d.SymmetryGroup,
			"grid":      d.Grid,
			"template":  d.Template,
			"palette":   d.Palette.Name,
			"seed_hash": d.SeedHashHex,
		}, "", "  ")
		fmt.Fprintln(os.Stderr, string(j))
	}
}
