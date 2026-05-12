import { defineConfig } from 'vite';
import { resolve } from 'node:path';

// The TS playground consumes the kolavatar-client-ts library by its published
// package name, but during development we alias the import to the sibling
// repo's src/ so renderer edits hot-reload. Switch to a real version range
// once the package is on npm.
const clientSrc = resolve(__dirname, '../../kolavatar-client-ts/src/index.ts');

// Go-rendered reference SVGs live in ../samples (this monorepo) and are served
// by a custom dev middleware so the cross-check column doesn't have to copy
// them in. To regenerate: `cd ../go && go run -tags=kolavatardev ./cmd/kolavatar-gallery -out ../samples/gallery`.
const goSamplesDir = resolve(__dirname, '../samples');

// Live-API proxy target. The "Live API" panel POSTs query params to
// /v1/avatars/{seed}.json via this proxy; either Go binary in ../go/cmd/
// (kolavatar-playground or kolavatar-dev) serves the SDK's built-in
// /v1/avatars/* surface from kolavatar.RegisterRoutes(). Both default to
// :8080. Override with KOLAVATAR_API_URL=http://localhost:NNNN if you
// pass a different -addr to the Go server.
const apiTarget = process.env.KOLAVATAR_API_URL ?? 'http://localhost:8080';

export default defineConfig({
  server: {
    port: 5173,
    fs: {
      // Allow Vite to read the aliased client-ts source and the sibling samples/.
      allow: [resolve(__dirname), resolve(__dirname, '../..'), goSamplesDir],
    },
    proxy: {
      // Forward the SDK's /v1/avatars/* surface to the Go server. CORS is
      // sidestepped because the browser sees a same-origin URL.
      '/v1': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      '@gklsndr/kolavatar-client-ts': clientSrc,
    },
  },
  build: {
    outDir: resolve(__dirname, 'dist'),
    emptyOutDir: true,
  },
  plugins: [
    {
      name: 'serve-go-samples',
      configureServer(server) {
        server.middlewares.use('/go-samples', (req, res, next) => {
          if (!req.url) {
            next();
            return;
          }
          const cleanedRel = decodeURIComponent(req.url.split('?')[0]!).replace(/^\/+/, '');
          if (cleanedRel.includes('..')) {
            res.statusCode = 400;
            res.end('bad path');
            return;
          }
          const fsPath = resolve(goSamplesDir, cleanedRel);
          if (!fsPath.startsWith(goSamplesDir)) {
            res.statusCode = 400;
            res.end('bad path');
            return;
          }
          import('node:fs').then(({ readFile, statSync }) => {
            try {
              const stat = statSync(fsPath);
              if (!stat.isFile()) {
                res.statusCode = 404;
                res.end('not a file');
                return;
              }
            } catch {
              res.statusCode = 404;
              res.end('not found');
              return;
            }
            readFile(fsPath, (err, data) => {
              if (err) {
                res.statusCode = 500;
                res.end(err.message);
                return;
              }
              res.setHeader(
                'Content-Type',
                fsPath.endsWith('.svg') ? 'image/svg+xml' : 'application/octet-stream',
              );
              res.setHeader('Cache-Control', 'no-cache');
              res.end(data);
            });
          });
        });
      },
    },
  ],
});
