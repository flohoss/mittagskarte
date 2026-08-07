# Agent Guidance

> **Purpose.** This file is the primary onboarding and guardrail document for any LLM
> (Claude, GPT, Gemini, Cursor, Copilot, etc.) that reads, writes, or reviews code in
> this repository. Read it before making changes. It is intentionally rule-oriented and
> self-contained.

## Stack

- **Backend:** Go + PocketBase v0.39 (pocketbase/pocketbase). Serves the frontend dist, a custom `/api/restaurants/scrape` endpoint, and a `sitemap.xml` on `:8090`.
- **Frontend:** Vue 3 + Vite + Tailwind v4 + daisyUI. Built to `frontend/dist` and served by the backend. No i18n — UI is hardcoded German.
- **Auth:** Email/password via PocketBase `users` collection. No OIDC.
- **Scraping:** Playwright (Go) for browser-based menu scraping, SnapOtter for image/PDF conversion to WebP. Cron-driven per-restaurant schedules.

## Architecture

### Backend packages

| Package                | Responsibility                                                                            |
| ---------------------- | ----------------------------------------------------------------------------------------- |
| `main`                 | Entry point only — loads config, creates app, registers hooks, starts                     |
| `config`               | Env parsing (`caarlos0/env`), validation, `Config` struct                                 |
| `api`                  | HTTP route handlers, rate limit config, static file serving (SSR-lite template injection) |
| `internal/mittag`      | Core orchestrator — cron scheduling, scrape queue, menu processing pipeline               |
| `internal/restaurant`  | Domain model, data access functions, slug generation                                      |
| `internal/sitemap`     | Sitemap XML generation and robots.txt                                                     |
| `internal/snapotter`   | Client for the SnapOtter image processing service                                         |
| `internal/web`         | Playwright browser wrapper for scraping                                                   |
| `internal/placeholder` | Placeholder image generation                                                              |
| `pkg/checksum`         | CRC32 checksum utilities for menu deduplication                                           |
| `pkg/curl`             | HTTP download helpers                                                                     |
| `pkg/fsutil`           | Filesystem utilities (local path resolution, temp files)                                  |
| `pkg/pdfinfo`          | PDF metadata extraction via poppler-utils                                                 |
| `pkg/snapotter`        | ogen-generated OpenAPI client for SnapOtter (do not edit — `go generate` regenerates)     |
| `migrations`           | PocketBase app migrations (auto-run on serve)                                             |

### Frontend layout

- `src/services/backendClient.ts` — shared PocketBase client, auth, data fetching, realtime subscriptions, file URL helpers
- `src/config.ts` — `BackendURL`, `RepoURL`, `AppVersion`
- `src/router.ts` — vue-router with custom `scrollBehavior` (home scroll restoration)
- `src/stores/` — global state via `createGlobalState` + `useStorage` (VueUse): `useFavorites`, `useLogin`, `useRestaurants`
- `src/composables/` — reusable composition functions: `useNow`, `usePrefersDark`
- `src/models/` — TypeScript interfaces for PocketBase records (`restaurant.ts`)
- `src/utils/` — pure utility functions with vitest tests (`date`, `menu`, `menuFreshness`, `regionColor`)
- `src/views/` — `HomeView`, `RestaurantView`, `PrivacyView`
- Icons: `@iconify/tailwind4` plugin, use as CSS classes `icon-[<set>--<name>]`

## Code style

- **No comments.** Use descriptive function or service names instead.
- **No code markers** like `// ... existing code ...` in edits.
- Go imports: stdlib, then external, then internal (`github.com/flohoss/mittagskarte/...`), each block alphabetical.

## Critical conventions

### PocketBase collection rules

- **`nil` rule = superuser-only** (not public). This is PocketBase's default for `NewBaseCollection`.
- **`""` (empty string) = public access** for non-superusers. Do not use unless intentional.
- Use `publicRule()` (returns `""`) for public read access and `authRule()` (returns `@request.auth.id != ""`) for authenticated access.
- Only expose the **minimum** rules needed by the frontend. If a collection is only mutated server-side via custom routes, leave Create/Update/Delete as `nil`.
- `@request.auth.isAdmin` is **not** a valid PocketBase field. Use `nil` rules for superuser-only.

### Dev vs production behavior

Many features are toggled by `!Dev` (i.e. only enabled in production). The `Dev` flag comes from the `DEV` env var (`false` by default).

| Setting                  | Dev (`Dev=true`)      | Production (`Dev=false`)                                        |
| ------------------------ | --------------------- | --------------------------------------------------------------- |
| SMTP                     | disabled              | enabled                                                         |
| Gzip middleware          | not applied           | applied (non-admin routes)                                      |
| `SkipSuccessActivityLog` | not bound (logs kept) | bound on `/health`, `/sitemap.xml`, `/robots.txt`, and frontend |
| Trusted proxy            | disabled              | enabled (`X-Forwarded-For`)                                     |
| Rate limiting            | disabled              | enabled                                                         |
| Superuser OTP/MFA        | disabled              | enabled                                                         |

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

Commit message — title only, no body, capitalize first letter:

- `[fix]` bug fix
- `[feature]` new functionality
- `[improve]` improvement to existing functionality
- `[meta]` changes outside the codebase (deployment, CI)
- `[docs]` documentation
- `[refactor]` formatting, renaming, structural-only

## Verification

Before committing, always run:

- **Backend:** `docker compose run --rm go fmt ./...`
- **Format:** `docker compose run --rm format`

Only commit if all pass.
