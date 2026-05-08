//go:build kolavatardev

// kolavatar-gallery generates a sample gallery: N deterministic 32-bit hex
// seeds × multiple scores × both states of WithCoverage (score-derived vs
// forced high). Used to produce the browseable samples/gallery/ output.
package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"html"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gklsndr/kolavatar"
)

func main() {
	count := flag.Int("count", 100, "number of seeds to generate")
	grid := flag.Int("grid", 11, "grid dimension")
	outDir := flag.String("out", "samples/gallery", "output directory")
	bw := flag.Bool("bw", false, "use black background + white strands (testing)")
	flag.Parse()

	var paletteOpt kolavatar.Option
	if *bw {
		// Force a single high-contrast palette so every kolam in the
		// gallery uses the same look. The first colour is the stroke,
		// the second is the background. The 4-colour minimum is padded
		// with neutral greys that the renderer doesn't currently use.
		bwPalette, err := kolavatar.NewPaletteSet("test-bw",
			kolavatar.Palette{
				Name:   "test-bw",
				Colors: []string{"#ffffff", "#000000", "#888888", "#444444"},
			},
		)
		if err != nil {
			log.Fatal(err)
		}
		paletteOpt = kolavatar.WithPaletteSet(bwPalette)
	}

	scores := []float64{0.1, 0.4, 0.7, 1.0}
	coverageModes := []struct {
		label string
		opt   kolavatar.Option // nil → score-derived
	}{
		{"natural", nil},
		{"forced", kolavatar.WithCoverage(0.9)},
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatal(err)
	}

	seeds := generateSeeds(*count)

	type entry struct {
		seed   string
		score  float64
		mode   string
		file   string
		group  string
		tier   uint8
	}
	var entries []entry

	for i, seed := range seeds {
		for _, score := range scores {
			for _, cm := range coverageModes {
				opts := []kolavatar.Option{kolavatar.WithGrid(*grid)}
				if cm.opt != nil {
					opts = append(opts, cm.opt)
				}
				if paletteOpt != nil {
					opts = append(opts, paletteOpt)
				}
				g, err := kolavatar.New(opts...)
				if err != nil {
					log.Fatal(err)
				}
				d := g.MustGenerate(seed, score)
				svg, err := d.SVG()
				if err != nil {
					log.Fatal(err)
				}
				name := fmt.Sprintf("s%03d-%s-%.2f-%s.svg", i, seed, score, cm.label)
				path := filepath.Join(*outDir, name)
				if err := os.WriteFile(path, []byte(svg), 0o644); err != nil {
					log.Fatal(err)
				}
				entries = append(entries, entry{
					seed: seed, score: score, mode: cm.label,
					file: name, group: d.SymmetryGroup, tier: d.Tier,
				})
			}
		}
		if (i+1)%10 == 0 {
			fmt.Fprintf(os.Stderr, "  generated %d/%d seeds\n", i+1, len(seeds))
		}
	}

	// Build the HTML index. Each row = one seed; columns alternate
	// natural/forced coverage at each score so the user can compare side
	// by side.
	idxPath := filepath.Join(*outDir, "index.html")
	f, err := os.Create(idxPath)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	fmt.Fprintln(f, `<!doctype html><html><head><meta charset="utf-8">`)
	fmt.Fprintln(f, `<title>kolavatar — gallery (100 seeds × 4 scores × 2 coverage modes)</title>`)
	pageBg := "#fafafa"
	pageFg := "#222"
	subtitleFg := "#666"
	thumbBg := "#fff"
	headerBg := "#eee"
	headerFg := "#333"
	if *bw {
		pageBg = "#1a1a1a"
		pageFg = "#eee"
		subtitleFg = "#aaa"
		thumbBg = "#000"
		headerBg = "#2a2a2a"
		headerFg = "#ccc"
	}
	fmt.Fprintf(f, `<style>
body { font-family: system-ui, sans-serif; margin: 16px; background: %s; color: %s; }
h1 { margin-bottom: 4px; }
.subtitle { color: %s; margin-top: 0; }
table { border-collapse: collapse; }
td { padding: 4px; vertical-align: top; }
.seed-cell { width: 110px; font-family: monospace; font-size: 12px; color: %s; padding-right: 8px; line-height: 1.3; }
img { width: 110px; height: 110px; display: block; }
.cap { font-size: 9px; color: #888; text-align: center; }
.scoreband { background: %s; padding: 2px; border-radius: 4px; }
.col-natural { border-left: 1px solid #444; }
.col-forced { border-right: 1px solid #444; }
.header td { font-size: 10px; font-weight: bold; color: %s; padding: 8px 4px 4px 4px; text-align: center; background: %s; }
a { color: %s; }
</style></head><body>`, pageBg, pageFg, subtitleFg, subtitleFg, thumbBg, headerFg, headerBg, pageFg)
	titleSuffix := ""
	if *bw {
		titleSuffix = " (test palette: white on black)"
	}
	fmt.Fprintf(f, `<h1>kolavatar — gallery%s</h1>`, html.EscapeString(titleSuffix))
	fmt.Fprintf(f, `<p class="subtitle">%d seeds (32-bit hex) × scores 0.1 / 0.4 / 0.7 / 1.0 × coverage modes (<b>natural</b> = score-derived, <b>forced</b> = WithCoverage(0.9)). Grid %d. Same seed → same group across all cells in its row; score increases complexity left→right within a coverage mode; forced coverage adds extra fill regardless of score. <a href="../index.html">← back to overview</a></p>`, *count, *grid)
	fmt.Fprintln(f, `<p class="subtitle"><b>How to read it:</b> some seeds (especially under high-symmetry groups like <code>4mm_d</code>) saturate the fundamental domain at low scores, so further score increase has nothing to add — those rows look static. Most rows visibly grow in density and motif diversity left→right.</p>`)

	fmt.Fprintln(f, `<table>`)
	// Column header
	fmt.Fprintln(f, `<tr class="header">`)
	fmt.Fprintln(f, `<td>seed · group</td>`)
	for _, score := range scores {
		fmt.Fprintf(f, `<td class="col-natural">%.1f natural</td>`, score)
		fmt.Fprintf(f, `<td class="col-forced">%.1f forced</td>`, score)
	}
	fmt.Fprintln(f, `</tr>`)

	idx := 0
	for i, seed := range seeds {
		fmt.Fprintln(f, `<tr>`)
		// Seed identity (group is invariant across scores for a given seed).
		first := entries[idx]
		fmt.Fprintf(f, `<td class="seed-cell">#%03d<br>%s<br>%s</td>`,
			i, html.EscapeString(seed), first.group)
		for range scores {
			natural := entries[idx]
			forced := entries[idx+1]
			fmt.Fprintf(f, `<td class="col-natural"><div class="scoreband"><img src="%s" loading="lazy"></div></td>`,
				html.EscapeString(natural.file))
			fmt.Fprintf(f, `<td class="col-forced"><div class="scoreband"><img src="%s" loading="lazy"></div></td>`,
				html.EscapeString(forced.file))
			idx += 2
		}
		fmt.Fprintln(f, `</tr>`)
	}
	fmt.Fprintln(f, `</table></body></html>`)

	fmt.Fprintf(os.Stderr, "wrote %d entries to %s\nindex: %s\n", len(entries), *outDir, idxPath)
}

// generateSeeds produces n deterministic 32-bit hex seeds. Using SHA-256 of
// "kolavatar-gallery-i" gives stable, varied seeds without depending on
// runtime randomness.
func generateSeeds(n int) []string {
	out := make([]string, n)
	for i := 0; i < n; i++ {
		h := sha256.Sum256([]byte(fmt.Sprintf("kolavatar-gallery-%d", i)))
		var b [4]byte
		copy(b[:], h[:4])
		out[i] = hex.EncodeToString(b[:])
	}
	// Ensure unique (collision odds in 100 random 32-bit values: ~1.16e-6,
	// but cheap to verify).
	seen := map[string]struct{}{}
	for i, s := range out {
		if _, dup := seen[s]; dup {
			// Resolve by appending the index — extremely unlikely to fire.
			out[i] = fmt.Sprintf("%s%02x", s, i&0xff)
		}
		seen[out[i]] = struct{}{}
	}
	return out
}

// avoid unused-import errors when compiling under unusual flags
var _ = strings.Builder{}
var _ = binary.LittleEndian