# kolavatar-playground

Development tools and visual fixtures for the [kolavatar](https://github.com/gklsndr/kolavatar) project. Nothing here ships to end users — this repo exists so contributors can iterate on the renderers and confirm the TS client matches the Go reference byte for byte.

## Layout

```
ts/         Vite-based Storybook-style preview of the TS renderer.
            Aliases @gklsndr/kolavatar-ts to the sibling repo's src/
            for HMR. Cross-checks against ../samples for parity with the
            Go renderer.
go/         Standalone Go module with the dev/sample/gallery/playground
            binaries (cmd/) plus the deployable cmd/kolavatar-web that
            embeds the TS bundle and serves both playgrounds + the SDK
            HTTP API from a single process. Imports github.com/gklsndr/
            kolavatar via a replace directive pointing at the sibling
            kolavatar-go checkout.
samples/    Reference SVGs produced by the Go renderer. The TS playground's
            "Go reference" cross-check column reads these directly.
            Regenerate via
            `cd go && go run -tags=kolavatardev ./cmd/kolavatar-gallery`.
Dockerfile, fly.toml, scripts/prepare-deploy.sh
            Single-container Fly.io deploy of cmd/kolavatar-web. See the
            "Hosted deploy (Fly.io)" section below.
```

This repo expects sibling checkouts of [kolavatar-ts](https://github.com/gklsndr/kolavatar-ts) and [kolavatar-go](https://github.com/gklsndr/kolavatar-go) at `../kolavatar-ts` and `../kolavatar-go`.

## TS playground

```sh
cd ts
npm install
npm run dev          # vite dev server on :5173, HMR over the sibling src/
npm run build        # static build into ts/dist/
```

The TS client is consumed by package name (`@gklsndr/kolavatar-ts`); a Vite + tsconfig alias resolves it to `../../kolavatar-ts/src/index.ts` so renderer edits hot-reload without an intermediate build. Once the package is published to npm, switch to a real version range and drop the alias.

## Go dev tools

All four binaries are gated behind the `kolavatardev` build tag and require the sibling kolavatar-go checkout for the `replace` directive in [go/go.mod](go/go.mod) to resolve.

```sh
cd go
go mod tidy

# Single SVG to stdout / file
go run -tags=kolavatardev ./cmd/kolavatar-sample \
    -seed='asha@example.com' -score=0.5 -grid=7 -out=sample.svg

# Live preview server with controls for every generation parameter
go run -tags=kolavatardev ./cmd/kolavatar-playground
# → http://localhost:8080/

# Browseable preview grid of N seeds × scores
go run -tags=kolavatardev ./cmd/kolavatar-dev -addr=:8081
# → http://localhost:8081/

# Batch gallery generator (writes into ../samples/gallery)
go run -tags=kolavatardev ./cmd/kolavatar-gallery -out ../samples/gallery
```

## Hosted deploy (Fly.io)

A single Fly.io machine serves the whole demo from one process via
`cmd/kolavatar-web`:

| Path                          | Served by                                          |
| ----------------------------- | -------------------------------------------------- |
| `/`                           | TS Vite playground (embedded via `embed.FS`)       |
| `/go/`                        | Go inline-HTML playground (`internal/goplayground`) |
| `/v1/avatars/{seed}.{format}` | SDK HTTP surface (`Generator.RegisterRoutes`)      |
| `/healthz`                    | Liveness probe                                     |

Same-origin everywhere — no CORS, no proxy, no separately-hosted backend.

### One-off setup

1. Install [`flyctl`](https://fly.io/docs/hands-on/install-flyctl/) and
   `fly auth login`.
2. From this repo's root, set the app name + region (the default
   `kolavatar-web` in `fly.toml` may already be taken — pick something
   unique):
   ```sh
   fly launch --no-deploy --copy-config
   ```

### Each deploy

The Docker build runs entirely inside Fly's builder, so it can't reach
your sibling `kolavatar-ts` and `kolavatar-go` checkouts directly. The
prepare script copies them into the build context first:

```sh
bash scripts/prepare-deploy.sh   # vendors kolavatar-ts/src + kolavatar-go
fly deploy
```

Both vendored paths are `.gitignore`d. Override their source locations
via `KOLAVATAR_TS=/abs/path` and `KOLAVATAR_GO=/abs/path` if your
sibling checkouts live elsewhere.

### Local dry run

```sh
bash scripts/prepare-deploy.sh
cd ts && npm ci && npm run build
cp -r dist/* ../go/cmd/kolavatar-web/dist/
cd ../go && go build -tags=kolavatardev ./cmd/kolavatar-web
PORT=8080 ./kolavatar-web
# → open http://localhost:8080/
```

To skip the npm step, the committed `dist/index.html` is a placeholder
that explains what's missing — `/go/`, `/v1/...`, and `/healthz` work
without it, only `/` shows the placeholder instead of the SPA.

## Cultural attribution

These tools render kolams from the South Indian sikku kolam tradition. Cultural and ethnographic context: [Sahapedia](https://www.sahapedia.org/kolam-floor-art-of-tamil-nadu). Algorithmic foundation: Siromoney, Siromoney, and Krithivasan, *Array Grammars and Kolam* (CGIP 3, 1974). Symmetry classification: Gopalan, *Symmetry Classification and Enumeration of Square-Tile Sikku Kolams* (J. Math. & the Arts, 2024); arXiv:2304.14134.
