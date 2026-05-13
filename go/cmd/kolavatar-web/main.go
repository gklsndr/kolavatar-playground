//go:build kolavatardev

// kolavatar-web is the hosted-playground binary: one process that serves
//
//	/              → the TS Vite playground (embedded static bundle)
//	/go/           → the Go inline-HTML playground (same handlers as
//	                 cmd/kolavatar-playground; mounted under a prefix)
//	/v1/avatars/*  → the SDK's built-in HTTP surface (Generator.RegisterRoutes)
//
// Same-origin everywhere, so the TS app's existing relative-URL fetches
// to /v1/avatars/* and the Go playground's relative ./render fetches both
// work without any CORS plumbing or hardcoded backend URL.
//
// Listens on $PORT (Fly.io / Cloud Run / Heroku convention) or :8080.
//
// The TS bundle is embedded from ./dist; Dockerfile copies the Vite build
// output (kolavatar-playground/ts/dist) into that directory before the Go
// build runs. The committed dist/placeholder.html keeps //go:embed happy
// when no bundle has been staged yet (local `go build` without the npm
// build will produce a working binary that just serves the placeholder at
// /).
package main

import (
	"embed"
	"io"
	"io/fs"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gklsndr/kolavatar"
	"github.com/gklsndr/kolavatar-playground/go/internal/goplayground"
)

//go:embed all:dist
var distFS embed.FS

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}

	mux := http.NewServeMux()

	// Liveness probe — Fly.io's http_service.checks hits this.
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok"))
	})

	// SDK API surface — same routes the TS playground's Live API panel
	// already calls.
	apiGen, err := kolavatar.New()
	if err != nil {
		log.Fatal(err)
	}
	apiGen.RegisterRoutes(mux, nil)

	// Go inline-HTML playground at /go/.
	goplayground.Register(mux, "/go")

	// TS Vite bundle at /. Catch-all that falls through to index.html for
	// any non-asset path (the SPA does client-side routing for hash links).
	tsFS, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/", spaHandler(tsFS))

	log.Printf("kolavatar-web listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// spaHandler serves files out of fsys, falling back to /index.html for
// unknown paths so deep links to the SPA still resolve. Returns 404 only
// if index.html itself is missing.
func spaHandler(fsys fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		if clean == "." || clean == "/" {
			clean = "index.html"
		}
		serveFile := func(name string) bool {
			f, err := fsys.Open(name)
			if err != nil {
				return false
			}
			defer f.Close()
			stat, err := f.Stat()
			if err != nil || stat.IsDir() {
				return false
			}
			ctype := mime.TypeByExtension(path.Ext(name))
			if ctype != "" {
				w.Header().Set("Content-Type", ctype)
			}
			// Long-cache hashed assets; let HTML revalidate.
			if strings.HasPrefix(name, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			_, _ = io.Copy(w, f)
			return true
		}
		if serveFile(clean) {
			return
		}
		// Fall through: any unknown path under / serves the SPA shell.
		if serveFile("index.html") {
			return
		}
		http.NotFound(w, r)
	})
}
