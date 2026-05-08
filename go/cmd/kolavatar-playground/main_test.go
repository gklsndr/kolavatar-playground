//go:build kolavatardev

package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestIndexHandler asserts the playground page is served at "/" and is
// well-formed enough for the script to bootstrap.
func TestIndexHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	indexHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d, want 200", w.Code)
	}
	body := w.Body.String()
	for _, marker := range []string{
		`<title>kolavatar — playground</title>`,
		`id="seed"`,
		`id="score"`,
		`id="coverage"`,
		`id="strand_length"`,
		`function render()`,
		`function randomizeSeed()`,
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("playground HTML missing marker %q", marker)
		}
	}
}

// TestRenderHandler exercises the /render endpoint with each control set,
// verifying SVG content is returned and the metadata headers are present.
func TestRenderHandler(t *testing.T) {
	cases := []struct {
		name  string
		query string
	}{
		{"defaults", "seed=alice@example.com"},
		{"forced-symmetry", "seed=alice&sym=2mm&grid=7"},
		{"explicit-coverage", "seed=alice&coverage=0.85&strand_length=2"},
		{"bw-palette", "seed=alice&palette=bw"},
		{"low-score", "seed=alice&score=0.0"},
		{"high-score", "seed=alice&score=1.0&grid=13"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/render?"+tc.query, nil)
			w := httptest.NewRecorder()
			renderHandler(w, r)
			if w.Code != http.StatusOK {
				t.Fatalf("status %d, want 200; body=%s", w.Code, w.Body.String())
			}
			ct := w.Header().Get("Content-Type")
			if !strings.HasPrefix(ct, "image/svg+xml") {
				t.Errorf("content-type %q, want image/svg+xml", ct)
			}
			body := w.Body.String()
			if !strings.HasPrefix(body, "<svg") {
				t.Errorf("body doesn't start with <svg: %q", body[:min(60, len(body))])
			}
			for _, h := range []string{
				"X-Kolavatar-Group",
				"X-Kolavatar-Palette",
				"X-Kolavatar-Tier",
				"X-Kolavatar-Cells-Covered",
			} {
				if w.Header().Get(h) == "" {
					t.Errorf("missing header %s", h)
				}
			}
		})
	}
}

// TestRenderHandler_BWPalette confirms that the bw palette is actually used
// (white strokes on black background).
func TestRenderHandler_BWPalette(t *testing.T) {
	r := httptest.NewRequest("GET", "/render?seed=alice&palette=bw", nil)
	w := httptest.NewRecorder()
	renderHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, `fill="#000000"`) {
		t.Error("expected black background in bw palette")
	}
	if !strings.Contains(body, `stroke="#ffffff"`) {
		t.Error("expected white strokes in bw palette")
	}
}

// TestRenderHandler_Errors verifies invalid input returns 4xx.
func TestRenderHandler_Errors(t *testing.T) {
	cases := []struct {
		name   string
		query  string
		status int
	}{
		{"missing-seed", "", http.StatusBadRequest},
		{"bad-score", "seed=alice&score=2", http.StatusBadRequest},
		{"bad-grid", "seed=alice&grid=4", http.StatusBadRequest},
		{"bad-coverage", "seed=alice&coverage=1.5", http.StatusBadRequest},
		{"bad-strand-length", "seed=alice&strand_length=99", http.StatusBadRequest},
		{"unknown-palette", "seed=alice&palette=does-not-exist", http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/render?"+tc.query, nil)
			w := httptest.NewRecorder()
			renderHandler(w, r)
			if w.Code != tc.status {
				t.Errorf("status %d, want %d", w.Code, tc.status)
			}
		})
	}
}

// TestRandomSeedHandler verifies a fresh 8-char hex string is returned.
func TestRandomSeedHandler(t *testing.T) {
	r := httptest.NewRequest("GET", "/random-seed", nil)
	w := httptest.NewRecorder()
	randomSeedHandler(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	seed := w.Body.String()
	if len(seed) != 8 {
		t.Errorf("seed length %d, want 8 (hex of 4 bytes)", len(seed))
	}
	for _, c := range seed {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex char %q in seed %q", c, seed)
		}
	}
}
