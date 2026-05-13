//go:build kolavatardev

package goplayground

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestServer mounts the playground at the given prefix and returns a
// running httptest.Server. Caller closes.
func newTestServer(t *testing.T, prefix string) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	Register(mux, prefix)
	return httptest.NewServer(mux)
}

// TestIndexHandler asserts the playground page is served at "/" and is
// well-formed enough for the script to bootstrap.
func TestIndexHandler(t *testing.T) {
	srv := newTestServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d, want 200", resp.StatusCode)
	}
	body := readBody(t, resp)
	for _, marker := range []string{
		`<title>kolavatar — playground</title>`,
		`id="seed"`,
		`id="score"`,
		`id="coverage"`,
		`id="strand_length"`,
		`function render()`,
		`function randomizeSeed()`,
		// Relative URLs so the page works at any mount prefix.
		`"./render?"`,
		`"./random-seed"`,
	} {
		if !strings.Contains(body, marker) {
			t.Errorf("playground HTML missing marker %q", marker)
		}
	}
}

// TestIndexHandler_Prefixed exercises mounting under a non-empty prefix —
// the index must be served at /go/ and a bare /go must redirect.
func TestIndexHandler_Prefixed(t *testing.T) {
	srv := newTestServer(t, "/go")
	defer srv.Close()

	// /go/ → 200
	resp, err := http.Get(srv.URL + "/go/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("/go/ status %d, want 200", resp.StatusCode)
	}

	// /go (no slash) → 301 to /go/
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp2, err := client.Get(srv.URL + "/go")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusMovedPermanently {
		t.Errorf("/go status %d, want 301", resp2.StatusCode)
	}
	if loc := resp2.Header.Get("Location"); loc != "/go/" {
		t.Errorf("redirect Location %q, want /go/", loc)
	}
}

// TestRenderHandler exercises the /render endpoint with each control set,
// verifying SVG content is returned and the metadata headers are present.
func TestRenderHandler(t *testing.T) {
	srv := newTestServer(t, "")
	defer srv.Close()

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
			resp, err := http.Get(srv.URL + "/render?" + tc.query)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("status %d, want 200; body=%s", resp.StatusCode, readBody(t, resp))
			}
			ct := resp.Header.Get("Content-Type")
			if !strings.HasPrefix(ct, "image/svg+xml") {
				t.Errorf("content-type %q, want image/svg+xml", ct)
			}
			body := readBody(t, resp)
			if !strings.HasPrefix(body, "<svg") {
				t.Errorf("body doesn't start with <svg: %q", body[:min(60, len(body))])
			}
			for _, h := range []string{
				"X-Kolavatar-Group",
				"X-Kolavatar-Palette",
				"X-Kolavatar-Tier",
				"X-Kolavatar-Cells-Covered",
			} {
				if resp.Header.Get(h) == "" {
					t.Errorf("missing header %s", h)
				}
			}
		})
	}
}

// TestRenderHandler_BWPalette confirms the bw palette is actually used
// (white strokes on black background).
func TestRenderHandler_BWPalette(t *testing.T) {
	srv := newTestServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/render?seed=alice&palette=bw")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	body := readBody(t, resp)
	if !strings.Contains(body, `fill="#000000"`) {
		t.Error("expected black background in bw palette")
	}
	if !strings.Contains(body, `stroke="#ffffff"`) {
		t.Error("expected white strokes in bw palette")
	}
}

// TestRenderHandler_Errors verifies invalid input returns 4xx.
func TestRenderHandler_Errors(t *testing.T) {
	srv := newTestServer(t, "")
	defer srv.Close()

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
			resp, err := http.Get(srv.URL + "/render?" + tc.query)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tc.status {
				t.Errorf("status %d, want %d", resp.StatusCode, tc.status)
			}
		})
	}
}

// TestRandomSeedHandler verifies a fresh 8-char hex string is returned.
func TestRandomSeedHandler(t *testing.T) {
	srv := newTestServer(t, "")
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/random-seed")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	seed := readBody(t, resp)
	if len(seed) != 8 {
		t.Errorf("seed length %d, want 8 (hex of 4 bytes)", len(seed))
	}
	for _, c := range seed {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("non-hex char %q in seed %q", c, seed)
		}
	}
}

func readBody(t *testing.T, resp *http.Response) string {
	t.Helper()
	var sb strings.Builder
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			sb.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
	return sb.String()
}
