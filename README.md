# kolavatar-playground

Development tools and visual fixtures for the [kolavatar](https://github.com/gklsndr/kolavatar) project. Nothing here ships to end users — this repo exists so contributors can iterate on the renderers and confirm the TS client matches the Go reference byte for byte.

## Layout

```
ts/        Vite-based Storybook-style preview of the TS renderer.
           Aliases @gklsndr/kolavatar-client-ts to the sibling repo's src/
           for HMR. Cross-checks against ../samples for parity with the
           Go renderer.
go/        Standalone Go module with the dev/sample/gallery/playground
           binaries. Imports github.com/gklsndr/kolavatar via a replace
           directive pointing at the sibling kolavatar-go checkout.
samples/   Reference SVGs produced by the Go renderer. The TS playground's
           "Go reference" cross-check column reads these directly. Regenerate
           via `cd go && go run -tags=kolavatardev ./cmd/kolavatar-gallery`.
```

This repo expects sibling checkouts of [kolavatar-client-ts](https://github.com/gklsndr/kolavatar-client-ts) and [kolavatar-go](https://github.com/gklsndr/kolavatar-go) at `../kolavatar-client-ts` and `../kolavatar-go`.

## TS playground

```sh
cd ts
npm install
npm run dev          # vite dev server on :5173, HMR over the sibling src/
npm run build        # static build into ts/dist/
```

The TS client is consumed by package name (`@gklsndr/kolavatar-client-ts`); a Vite + tsconfig alias resolves it to `../../kolavatar-client-ts/src/index.ts` so renderer edits hot-reload without an intermediate build. Once the package is published to npm, switch to a real version range and drop the alias.

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

## Cultural attribution

These tools render kolams from the South Indian sikku kolam tradition. Cultural and ethnographic context: [Sahapedia](https://www.sahapedia.org/kolam-floor-art-of-tamil-nadu). Algorithmic foundation: Siromoney, Siromoney, and Krithivasan, *Array Grammars and Kolam* (CGIP 3, 1974). Symmetry classification: Gopalan, *Symmetry Classification and Enumeration of Square-Tile Sikku Kolams* (J. Math. & the Arts, 2024); arXiv:2304.14134.
