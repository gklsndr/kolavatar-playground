//go:build kolavatardev

// kolavatar-dev is a small HTTP harness for visually inspecting generated
// kolavatars. It is gated by the `kolavatardev` build tag and must NEVER be
// linked into the production monolith.
package main

import (
	"flag"
	"fmt"
	"html"
	"log"
	"net/http"
	"strings"

	"github.com/gklsndr/kolavatar"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	flag.Parse()

	gen, err := kolavatar.New()
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	gen.RegisterRoutes(mux, nil)
	mux.Handle("GET /", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seeds := []string{
			"asha@example.com", "ravi@example.com", "priya@example.com",
			"karan@example.com", "meena@example.com", "vivek@example.com",
			"lakshmi@example.com", "arun@example.com", "divya@example.com",
		}
		var b strings.Builder
		b.WriteString(`<!doctype html><html><head><meta charset="utf-8"><title>kolavatar-dev</title>`)
		b.WriteString(`<style>body{font-family:system-ui;margin:24px}img{margin:8px;border:1px solid #ddd}</style></head><body>`)
		b.WriteString(`<h1>kolavatar-dev preview</h1>`)
		for _, s := range seeds {
			for _, score := range []string{"0.1", "0.5", "0.9"} {
				url := fmt.Sprintf("/v1/avatars/%s.svg?score=%s", html.EscapeString(s), score)
				fmt.Fprintf(&b, `<figure style="display:inline-block">`+
					`<img src="%s" width="128" height="128" alt="%s @ %s"/>`+
					`<figcaption>%s @ %s</figcaption></figure>`,
					url, html.EscapeString(s), score, html.EscapeString(s), score)
			}
		}
		b.WriteString(`</body></html>`)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(b.String()))
	}))

	log.Printf("kolavatar-dev listening on %s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}
