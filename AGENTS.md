# AGENTS.md

Instructions for AI agents working on this repository. Read this before making
changes; follow the verification checklist before claiming any task is done.

## What this is

sortir.in — a French "events around you" web app. A Go backend built on
PocketBase (SQLite) collects events from external sources and serves aggregated
map pins; a SvelteKit (Svelte 5) UI with a MapLibre map is embedded in the
single binary. Events are date-ranged and geolocated; the UI filters them by
time range (Aujourd'hui / Demain / Cette semaine).

## Repo map

```
cmd/
  main.go            server entrypoint (PocketBase + migrations + routes + crons)
  populate/          one-shot population CLI: runs collectors over French cities,
                     pushes events to PUT /api/events (server must be running)
pkg/
  application/       domain layer: interfaces (EventRepository, PinsService,
                     Collector), services (pins, populator), Event model and
                     validation, mocks/ (mockgen — regenerate with make generate)
  infrastructure/
    server/          HTTP wiring: routes, request handlers, expired-events cron
    repository/      dbx/SQLite queries against the events table
    collector/       event sources (bobine, allevents, opendata Paris, random)
    client/pb/       HTTP client used by cmd/populate
migrations/          PocketBase Go migrations (run on server start)
ui/                  SvelteKit app (Svelte 5 runes), embedded via ui/embed.go
  src/lib/components/MapView.svelte   map + layers + stores wiring (core file)
  src/lib/components/pins/svg/        map pin SVGs — see README.md there before
                                      regenerating or adding pins
  src/lib/stores/    pins (map aggregates) and events (lists)
  src/lib/utils/dateUtils.ts          range windows + date formatting
docs/adr/            architecture decision records — read before changing the
                     areas they cover; new significant decisions get a new ADR
pb_data/             local dev database (gitignored, disposable)
```

## Commands

| Task | Command |
|---|---|
| Dev (server + embedded UI) | `make dev` (modd) — serves on http://localhost:8090 |
| Dev (UI with HMR) | `make dev-ui` (vite) |
| Full build (UI must build first: it is embedded) | `make build` |
| Regenerate mocks | `make generate` (after changing `pkg/application` interfaces) |
| Go checks | `go build ./... && go vet ./... && go test ./...` |
| UI checks | `cd ui && npx svelte-check` |
| UI typegen (after schema changes) | `cd ui && pnpm run typegen` |
| Populate local DB | `go run cmd/populate/populate.go <city_limit>` (needs :8090 up) |

Integration tests (`test/integration/`) spin a real PocketBase on
127.0.0.1:8035 — they are part of `go test ./...`.

## Architecture invariants

These are decided and documented in `docs/adr/` — do not silently deviate:

- **Date-range filtering is an interval overlap** (ADR 0001): an event is shown
  for a range when `begin <= window.max && end >= window.min`. All window
  boundaries are 3am local cuts (night activities belong to the day that
  started them); "Cette semaine" ends Monday 3am. Windows live in
  `ui/src/lib/utils/dateUtils.ts` (`getDateWindow`).
- **Dates are stored in PocketBase's canonical UTC layout**
  (`2006-01-02 15:04:05.000Z`, `types.DefaultDateLayout`) — ADR 0002. Any SQL
  comparison against `begin`/`end` must bind a string in that exact layout
  (never a raw `time.Time`): SQLite compares dates as text.
- **One shared window** (ADR 0003): `/api/pins` (optional `min_time`), the
  sidebar (`eventsStore.getEventsInBounds`) and the event popup all use the
  same `getDateWindow` output. Keep them in sync.
- **Pins are aggregates**: `/api/pins` groups events by (location, kind) into
  `Pin{loc, kind, amount}` (pkg/application/pins.go); the UI resolves full
  events per location only when a pin is clicked.
- **`Event.IsValid`** (pkg/application/event.go) rejects events that already
  ended or span more than 15 days. Upsert key: unique index on
  `(name, begin, end)`.
- **Expired events** are deleted every minute by the cron in
  pkg/infrastructure/server/pocketbase.go.

## Conventions

- Conventional commits (`feat:`, `fix:`), lowercase, imperative; semantic-release
  cuts versions from these.
- Significant decisions (schema, filtering semantics, storage formats, new
  surfaces) get a numbered ADR in `docs/adr/`, matching the existing format.
- Go: respect the layering — domain interfaces in `pkg/application`, HTTP/DB/
  scraping details in `pkg/infrastructure`; mock external dependencies with
  gomock (regenerate mocks, never hand-edit `mocks/`).
- UI: Svelte 5 runes (`$state`, `$derived`, `$effect`) — not legacy stores
  syntax; user-facing strings in French, identifiers and comments in English.
- `end`/`begin` are SQL-adjacent keywords: backtick them in raw queries
  (`` `begin` ``, `` `end` ``).

## Gotchas

- **modd watches only `cmd/**/*.go`** — changes under `pkg/` do not restart the
  dev server. Kill the running `main` process; modd relaunches it.
- **svelte-check has a baseline of 21 pre-existing errors** (svelte-maplibre
  module resolution, `MapSidebar` props). Compare before/after your change;
  only new errors matter.
- `pb_data/` is disposable (`make clean` deletes it) and is also mutated live by
  the collectors and the expired-events cron — don't treat its contents as
  precious, and don't commit it.
- `ui/build` is gitignored but **required to build the Go binary** (embed).
- Dates in `pb_data` are UTC strings; when eyeballing them, remember France is
  UTC+1/+2 depending on DST.

## Definition of done

1. `go build ./... && go vet ./... && go test ./...` passes.
2. `cd ui && npx svelte-check` reports no new errors vs. the baseline.
3. New behavior is covered by a test where the structure allows it
   (application layer: mocks; HTTP/DB: `test/integration`).
4. Decisions with lasting impact are recorded in `docs/adr/`.
