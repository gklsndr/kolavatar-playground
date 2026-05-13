//go:build kolavatardev

// Package goplayground hosts the inline-HTML Go playground UI. Extracted
// out of cmd/kolavatar-playground so both that standalone binary and the
// hosted (cmd/kolavatar-web) binary can mount the same handlers — the
// only difference between them is the URL prefix.
package goplayground

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gklsndr/kolavatar"
)

// Register wires the playground's three endpoints onto mux under prefix.
// prefix should be "" (root) or "/foo" (no trailing slash); index is served
// at prefix+"/" and the auxiliary fetch endpoints at prefix+"/render" and
// prefix+"/random-seed". The inline JS uses relative URLs so the same HTML
// works at any prefix; if prefix is non-empty Register also installs a
// redirect from prefix → prefix+"/" so the relative URLs resolve cleanly.
func Register(mux *http.ServeMux, prefix string) {
	prefix = strings.TrimRight(prefix, "/")

	indexPath := prefix + "/"
	mux.HandleFunc(indexPath, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != indexPath {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, playgroundHTML)
	})

	if prefix != "" {
		mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, indexPath, http.StatusMovedPermanently)
		})
	}

	mux.HandleFunc(prefix+"/render", renderHandler)
	mux.HandleFunc(prefix+"/random-seed", randomSeedHandler)
}

// renderHandler returns an SVG for the requested parameters. Query
// parameters: seed, score, grid, sym, coverage, strand_length, palette,
// bw, min_score, score_bands.
func renderHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	seed := q.Get("seed")
	if seed == "" {
		http.Error(w, "seed required", http.StatusBadRequest)
		return
	}

	score, err := strconv.ParseFloat(orDefault(q.Get("score"), "0.5"), 64)
	if err != nil || score < 0 || score > 1 {
		http.Error(w, "score must be in [0,1]", http.StatusBadRequest)
		return
	}

	grid, err := strconv.Atoi(orDefault(q.Get("grid"), "11"))
	if err != nil || grid < 3 || grid > 15 || grid%2 == 0 {
		http.Error(w, "grid must be odd in [3,15]", http.StatusBadRequest)
		return
	}

	opts := []kolavatar.Option{kolavatar.WithGrid(grid)}

	if sym := q.Get("sym"); sym != "" && sym != "auto" {
		opts = append(opts, kolavatar.WithSymmetryPolicy(kolavatar.SymmetryForced(sym)))
	}

	if cov := q.Get("coverage"); cov != "" && cov != "auto" {
		v, err := strconv.ParseFloat(cov, 64)
		if err != nil || v < 0 || v > 1 {
			http.Error(w, "coverage must be in [0,1] or 'auto'", http.StatusBadRequest)
			return
		}
		opts = append(opts, kolavatar.WithCoverage(v))
	}

	if sl := q.Get("strand_length"); sl != "" && sl != "auto" {
		v, err := strconv.Atoi(sl)
		if err != nil || v < 0 || v > 3 {
			http.Error(w, "strand_length must be 0..3 or 'auto'", http.StatusBadRequest)
			return
		}
		opts = append(opts, kolavatar.WithStrandLength(v))
	}

	if ms := q.Get("min_score"); ms != "" {
		v, err := strconv.ParseFloat(ms, 64)
		if err != nil || v < 0 || v > 1 {
			http.Error(w, "min_score must be in [0,1]", http.StatusBadRequest)
			return
		}
		if v > 0 {
			opts = append(opts, kolavatar.WithMinScore(v))
		}
	}

	if sb := q.Get("score_bands"); sb != "" {
		v, err := strconv.Atoi(sb)
		if err != nil || v < 1 || v > 16 {
			http.Error(w, "score_bands must be in [1,16]", http.StatusBadRequest)
			return
		}
		opts = append(opts, kolavatar.WithScoreBands(v))
	}

	switch p := q.Get("palette"); p {
	case "":
		// default
	case "bw":
		bwPalette, err := kolavatar.NewPaletteSet("test-bw",
			kolavatar.Palette{
				Name:   "test-bw",
				Colors: []string{"#ffffff", "#000000", "#888888", "#444444"},
			},
		)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		opts = append(opts, kolavatar.WithPaletteSet(bwPalette))
	default:
		full := kolavatar.DefaultPaletteSet()
		palette, err := full.ByName(p)
		if err != nil {
			http.Error(w, "unknown palette: "+p, http.StatusBadRequest)
			return
		}
		single, err := kolavatar.NewPaletteSet(p, palette)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		opts = append(opts, kolavatar.WithPaletteSet(single))
	}

	g, err := kolavatar.New(opts...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	d, err := g.Generate(r.Context(), seed, score)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	svg, err := d.SVG()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	w.Header().Set("X-Kolavatar-Group", d.SymmetryGroup)
	w.Header().Set("X-Kolavatar-Palette", d.Palette.Name)
	w.Header().Set("X-Kolavatar-Tier", strconv.Itoa(int(d.Tier)))
	w.Header().Set("X-Kolavatar-Seed-Hash", d.SeedHashHex)
	covered, strands := tilesStats(int(d.Grid), d.Tiles)
	w.Header().Set("X-Kolavatar-Cells-Covered", strconv.Itoa(covered))
	w.Header().Set("X-Kolavatar-Strands", strconv.Itoa(strands))
	_, _ = io.WriteString(w, svg)
}

// tilesStats counts non-zero tiles and a rough strand-count approximation.
func tilesStats(N int, tiles []byte) (covered, strands int) {
	for _, p := range tiles {
		if p != 0 {
			covered++
		}
	}
	for _, p := range tiles {
		switch p {
		case 0:
		case 15:
			strands += 2
		default:
			strands++
		}
	}
	strands /= 2
	if strands < 1 && covered > 0 {
		strands = 1
	}
	return covered, strands
}

func randomSeedHandler(w http.ResponseWriter, r *http.Request) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, hex.EncodeToString(b[:]))
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func paletteList() string {
	ps := kolavatar.DefaultPaletteSet()
	names := ps.Names()
	return `"` + strings.Join(names, `","`) + `"`
}
