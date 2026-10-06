---
name: new-event-collector
description: Add a new event source to sortir.in — implement application.Collector, parse dates to the Event model, register the collector, test it, populate. Use when the user asks for a new event source, scraper, API integration, or collector.
---

# Adding an event collector

A collector turns one external source (website/API) into `application.Event`s.
Read `pkg/application/event.go`, `pkg/application/collector.go`, and one
existing source first — `pkg/infrastructure/collector/bobine.go` (JSON API) and
`pkg/infrastructure/collector/opendata_paris_quefaire.go` (paged JSON) cover the
two common shapes.

## Steps

1. **Create** `pkg/infrastructure/collector/<source>.go` with a
   `New<Source>Collector() application.Collector`. `Collect(location
   application.CollectLocation)` returns `[]application.Event` and errors as
   error, never silently-empty.
2. **Map fields** into `application.Event`: `Name`, `Kind` (must be a known
   `Kind` — see `pkg/application/kinds.go`; use `KindUnknown` rather than
   inventing one unless the UI gets a pin SVG too), `Begin`, `End`, `Loc`
   (lat/lon), `Source` (stable identifier), optional `Place`, `Address`,
   `Price`, `Genres`, `Img`.
3. **Dates**: parse to `time.Time` with the source's real timezone when known
   (`time.Parse(time.RFC3339, ...)` or `time.ParseInLocation`). Storage
   normalization to UTC happens server-side (ADR 0002) — do not pre-format.
4. **Respect validation** (pkg/application/event.go): events that already ended
   or span > 15 days are dropped by `Event.IsValid` server-side; don't send
   them. `End` is required (the expired-events cron and all windows rely on it
   — if the source lacks an end, use the begin time, or the venue closing time,
   never a far-future sentinel).
5. **Register** in the composite: `cmd/populate/populate.go`
   (`collector.NewCompositeCollector(...)`).
6. **Test**: mirror `pkg/infrastructure/collector/random_test.go` — unit-test
   parsing with captured payload fixtures, including edge cases (no end time,
   multi-day events, timezone offsets).
7. **Verify live**: with `make dev` running on :8090, run
   `go run cmd/populate/populate.go 1` and inspect results (see the
   live-data-probe skill).

## Nationwide sources (GlobalCollector)

Sources covering a whole country with no per-location query (national CSV or
API — e.g. `marchesdefrance.go`, ADR 0007) implement
`application.GlobalCollector` instead:

1. **Create** the collector with `CollectAll() ([]application.Event, error)`,
   `New<Source>Collector() application.GlobalCollector`.
2. **Register** it in `cmd/populate/populate.go` via
   `application.NewGlobalPopulator(...)` — it runs once per populate run,
   before the city loop, and its save path batches (~1000 events per PUT).
3. **Recurring schedules** (markets, weekly slots) must be materialized into
   concrete `Begin`/`End` instances over the next 14 days (Europe/Paris) —
   see `expandMarket` in `marchesdefrance.go` for the pattern, including name
   disambiguation for shared names (the upsert key is `(name, begin, end)`).

## Constraints

- The events table upserts on `(name, begin, end)` — a source emitting unstable
  names or shifting begin times will create duplicates.
- Events are geolocated; if the source only gives an address, it needs a
  geocoding step before it is useful on the map.
- Follow the layering: HTTP/scraping details stay in
  `pkg/infrastructure/collector/`; no domain logic there.
