# Multi-stage build for the kolavatar-web binary deployed to Fly.io.
#
# The deployable artifact is a single Go binary that embeds the Vite-built
# TS playground and serves it alongside the Go inline-HTML playground and
# the SDK's HTTP API surface.
#
# Three stages:
#   1. node:alpine   — builds the TS bundle via Vite. Reads vendored
#                      kolavatar-ts source from ts/vendor-kolavatar-ts/.
#   2. golang:alpine — builds the Go binary. Reads vendored kolavatar-go
#                      from vendor-kolavatar-go/ (matches the go.mod
#                      replace directive `../../kolavatar-go` once the
#                      playground module is placed at /src/playground/go).
#                      Stages ts/dist into the embed target and compiles
#                      with `-tags=kolavatardev` so the SDK's SVG renderer
#                      is included.
#   3. alpine:latest — minimal runtime image carrying just the binary.
#                      Listens on $PORT (Fly.io sets this to 8080 by
#                      default; fly.toml's internal_port matches).
#
# Run `bash scripts/prepare-deploy.sh` before `fly deploy` to populate
# the vendored directories above. Both are .gitignored.

# ── Stage 1: TS bundle ────────────────────────────────────────────────────
FROM node:20-alpine AS ts-builder
WORKDIR /src
COPY ts/package.json ts/package-lock.json* ./
RUN npm ci
COPY ts/ ./
# Sanity check: the vendored kolavatar-ts source must be present, else
# vite.config.ts's alias would fall back to a sibling path that doesn't
# exist in the container.
RUN test -d vendor-kolavatar-ts/src || (echo "ts/vendor-kolavatar-ts/src not found — run scripts/prepare-deploy.sh before docker build" >&2 && exit 1)
RUN npm run build

# ── Stage 2: Go binary ────────────────────────────────────────────────────
FROM golang:1.25-alpine AS go-builder
WORKDIR /src

# Lay out playground at /src/playground/ and kolavatar-go at
# /src/kolavatar-go/ so the go.mod replace path `../../kolavatar-go`
# (relative to /src/playground/go) resolves to /src/kolavatar-go.
COPY vendor-kolavatar-go /src/kolavatar-go
COPY go /src/playground/go

WORKDIR /src/playground/go
RUN go mod download

# Stage the TS dist into the embed target.
COPY --from=ts-builder /src/dist ./cmd/kolavatar-web/dist

# Build a small static binary.
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -tags=kolavatardev -ldflags="-s -w" -o /out/kolavatar-web ./cmd/kolavatar-web

# ── Stage 3: runtime ──────────────────────────────────────────────────────
FROM alpine:3.20
RUN apk add --no-cache ca-certificates && adduser -D -u 10001 app
COPY --from=go-builder /out/kolavatar-web /usr/local/bin/kolavatar-web
USER app
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/usr/local/bin/kolavatar-web"]
