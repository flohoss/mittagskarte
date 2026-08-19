# Agent Guidance

Read before making changes. Rule-oriented and self-contained.

## Stack

- **Backend:** Go + PocketBase v0.39 on `:8090`. Serves frontend dist, `/api/restaurants/scrape`, `sitemap.xml`.
- **Frontend:** Vue 3 + Vite + Tailwind v4 + daisyUI. Built to `frontend/dist`, served by backend. UI is hardcoded German.
- **Auth:** Email/password via PocketBase `users` collection. No OIDC.
- **Scraping:** Playwright (Go) for browser-based scraping, SnapOtter for image/PDF → WebP. Cron-driven per-restaurant.

## Architecture

- `main` — entry only. `config` — env (`caarlos0/env`). `api` — HTTP routes, rate limits, SSR-lite template injection.
- `internal/mittag` — orchestrator (cron, scrape queue, menu pipeline). `internal/restaurant` — domain model, data access, slugs.
- `internal/{sitemap,snapotter,web,placeholder}` — sitemap/robots, SnapOtter client, Playwright wrapper, placeholder images.
- `pkg/` — `checksum` (CRC32 dedup), `curl` (HTTP), `fsutil` (fs utils), `pdfinfo` (poppler-utils), `snapotter` (ogen-generated, do not edit).
- `migrations/` — PocketBase migrations (auto-run on serve).
- Frontend: `backendClient.ts` (PB SDK, auth, realtime), `config.ts`, `router.ts`, `stores/` (`useFavorites`, `useLogin`, `useRestaurants`), `composables/` (`useNow`, `usePrefersDark`), `models/`, `utils/` (pure, vitest tests), `views/` (`HomeView`, `RestaurantView`, `PrivacyView`).
- Icons: `@iconify/tailwind4` plugin, CSS classes `icon-[<set>--<name>]`.

## Principles

- **CLEAN code.** Small functions, single responsibility, descriptive names, no dead code, no overengineering.
- **No comments.** Use descriptive function or service names instead.
- **No code markers** like `// ... existing code ...` in edits.
- Go imports: stdlib → external → internal (`github.com/flohoss/mittagskarte/...`), each block alphabetical.
- Never edit generated files (`backend/pkg/snapotter/`).

## Conventions

### PocketBase rules

- `nil` = superuser-only. `""` = public (avoid unless intentional).
- `publicRule()` → `""` (public read). `authRule()` → `@request.auth.id != ""` (authed).
- Expose minimum rules. Server-side-only collections: leave Create/Update/Delete as `nil`.
- `@request.auth.isAdmin` is invalid. Use `nil` for superuser-only.

### Dev vs prod

`!Dev` toggles prod-only features. `Dev` from `DEV` env (`false` default). Prod: SMTP, gzip (non-admin), trusted proxy, rate limiting, OTP/MFA. `SkipSuccessActivityLog` on `/health`, `/sitemap.xml`, `/robots.txt`, frontend.

### Frontend serving (SSR-lite)

Backend parses `dist/index.html` as Go `text/template`, injects `window.__RESTAURANTS__` (JSON) and `window.__EMAIL__` (base64). Frontend hydrates from globals, then subscribes to realtime.

### Scrape pipeline

1. Cron → `Scraper.Enqueue` (dedup + cooldown)
2. Single-worker queue: `scrape` (Playwright) / `download` (HTTP) / `upload` (manual)
3. `processToWebp`: PDF → SnapOtter → PNG → stitch → WebP; or image → WebP
4. CRC32 vs latest hash → reject unchanged (`ErrMenuUnchanged`)
5. Store menu → retention cleanup (`restaurants.menus` MaxSelect)
6. Realtime broadcast: `restaurants/status` → `{id, status, coolDownSeconds}`

## Tooling — always via Docker Compose, never on the host

- **Dev server:** `docker compose up --build --force-recreate` (separate backend and frontend containers).
  - URLs: frontend `:5173`, backend `:8090`, dashboard `:8090/_/`, SnapOtter `:1349`.
  - Services: `backend` (air hot reload), `frontend` (Vite), `snapotter`, `snapotter-schema`, `ogen`, `go`, `npm`, `release`.
- **Code generation:** `docker compose run --rm ogen` — generates the SnapOtter client via [ogen](https://github.com/ogen-go/ogen) from OpenAPI spec (filtered to `/api/v1/(tools|download|pipeline|admin|features)`), output in `backend/pkg/snapotter/api/` (gitignored).
- **Format (Go):** `docker compose run --rm go fmt ./...`
- **Format (all non-Go files):** `docker compose run --rm format`

Run formatting after every code change, even small edits. Only commit if all pass.

## Common Commands

```sh
docker compose run --rm npm install
docker compose run --rm --entrypoint npx npm --yes npm-check-updates -u && docker compose run --rm npm install
docker compose run --rm go get -u ./...
docker compose run --rm go mod tidy
docker compose run --rm go fmt ./...
docker compose run --rm npm run build
```

### Playwright driver version

`playwright-go` in `go.mod` and `V_PLAYWRIGHT` in `compose.yml` must match. After `go get -u`, if `playwright-go` was upgraded, update `V_PLAYWRIGHT` to the same version and rebuild:

```sh
docker compose up --build --force-recreate backend
```

A mismatch causes: `could not start playwright: please install the driver (vX.Y.Z) first`.

### TypeScript major

`vue-tsc` breaks on TS majors it doesn't support yet (`ERR_PACKAGE_PATH_NOT_EXPORTED` for `./lib/tsc`). After `npm-check-updates -u`, if `typescript` was bumped to a new major, check if `vue-tsc` supports it (`docker compose run --rm npm run build`). If not, revert `typescript` in `frontend/package.json` to the previous major before installing.

## Git

- Do not commit automatically — wait until explicitly asked.
- One commit per concern — never batch unrelated changes.
- Title only, no body. Capitalize first letter after the prefix:
  - `[fix]` bug fix
  - `[feature]` new functionality
  - `[improve]` improvement to existing functionality
  - `[refactor]` formatting, renaming, structural-only
  - `[meta]` deployment, CI
  - `[docs]` documentation
