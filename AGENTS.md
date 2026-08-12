# Agent Guidance

Read before making changes. Rule-oriented and self-contained.

## Stack

- **Backend:** Go + PocketBase v0.39. Serves frontend dist, `/api/restaurants/scrape` endpoint, and `sitemap.xml` on `:8090`.
- **Frontend:** Vue 3 + Vite + Tailwind v4 + daisyUI. Built to `frontend/dist`, served by backend. No i18n — UI is hardcoded German.
- **Auth:** Email/password via PocketBase `users` collection. No OIDC.
- **Scraping:** Playwright (Go) for browser-based menu scraping, SnapOtter for image/PDF conversion to WebP. Cron-driven per-restaurant schedules.

## Architecture

- `main` — entry point only. `config` — env parsing (`caarlos0/env`). `api` — HTTP routes, rate limits, static serving (SSR-lite template injection).
- `internal/mittag` — core orchestrator (cron, scrape queue, menu pipeline). `internal/restaurant` — domain model, data access, slug generation.
- `internal/sitemap` — sitemap XML and robots.txt. `internal/snapotter` — SnapOtter client. `internal/web` — Playwright wrapper. `internal/placeholder` — placeholder image generation.
- `pkg/` — `checksum` (CRC32 for menu dedup), `curl` (HTTP downloads), `fsutil` (filesystem utils), `pdfinfo` (PDF metadata via poppler-utils), `snapotter` (ogen-generated OpenAPI client — do not edit, `go generate` regenerates).
- `migrations` — PocketBase migrations (auto-run on serve).
- Frontend: `backendClient.ts` (PocketBase SDK, auth, realtime), `config.ts`, `router.ts` (custom `scrollBehavior`), `stores/` (`useFavorites`, `useLogin`, `useRestaurants`), `composables/` (`useNow`, `usePrefersDark`), `models/`, `utils/` (pure functions with vitest tests), `views/` (`HomeView`, `RestaurantView`, `PrivacyView`).
- Icons: `@iconify/tailwind4` plugin, use as CSS classes `icon-[<set>--<name>]`.

## Code style

- **No comments.** Use descriptive function or service names instead.
- **No code markers** like `// ... existing code ...` in edits.
- Go imports: stdlib, then external, then internal (`github.com/flohoss/mittagskarte/...`), each block alphabetical.

## Critical conventions

### PocketBase collection rules

- **`nil` rule = superuser-only** (not public). **`""` = public access** — do not use unless intentional.
- Use `publicRule()` (returns `""`) for public read access and `authRule()` (returns `@request.auth.id != ""`) for authenticated access.
- Only expose the **minimum** rules needed by the frontend. If a collection is only mutated server-side via custom routes, leave Create/Update/Delete as `nil`.
- `@request.auth.isAdmin` is **not** a valid PocketBase field. Use `nil` rules for superuser-only.

### Dev vs production

Features toggled by `!Dev` (only enabled in production). `Dev` comes from the `DEV` env var (`false` default). In production: SMTP, gzip (non-admin routes), trusted proxy (`X-Forwarded-For`), rate limiting, superuser OTP/MFA. `SkipSuccessActivityLog` bound on `/health`, `/sitemap.xml`, `/robots.txt`, and frontend.

### Frontend serving (SSR-lite)

The backend parses `dist/index.html` as a Go `text/template` and injects `window.__RESTAURANTS__` (restaurant data as JSON) and `window.__EMAIL__` (base64-encoded imprint email) at serve time. The frontend hydrates from these globals on init, then subscribes to realtime updates.

### Scrape pipeline

1. Cron triggers per-restaurant cron group → `Scraper.Enqueue` (dedup + cooldown check)
2. Single-worker queue processes sequentially: `scrape` (Playwright) / `download` (HTTP) / `upload` (manual)
3. Menu file → `processToWebp` (PDF via SnapOtter → PNG → stitch → WebP; or direct image → WebP)
4. CRC32 checksum vs latest menu hash → reject unchanged (`ErrMenuUnchanged`)
5. Store menu record → retention cleanup (enforce `restaurants.menus` MaxSelect)
6. Realtime broadcast: `restaurants/status` topic with `{id, status, coolDownSeconds}`

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

## Verification

Before committing, always run:

- **Backend:** `docker compose run --rm go fmt ./...`
- **Format:** `docker compose run --rm format`

Only commit if all pass.
