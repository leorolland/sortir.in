# ADR 0004 — Worldwide event collection: stateless clock-sharded scheduling

- **Status:** Discussed
- **Date:** 2026-09-19

## Context

Today `cmd/populate` is a one-shot batch, triggered every 12 hours: it walks a
hardcoded list of French cities in fixed order, runs every collector per city,
and upserts results through `PUT /api/events`. The goal is to extend coverage
to a very large number of cities worldwide.

## Constraints

1. **Scale** — from ~30 hardcoded cities to a very large set (order of 10³–10⁴)
   worldwide; the location list cannot stay in Go source, it must live in data
   and grow without a new release.
2. **Freshness ceiling** — priority cities must never be staler than 12h;
   smaller cities may accept looser targets (24h+), so freshness targets are
   per-tier rather than global.
3. **No bursts** — collection must be spread over time (steady trickle, not one
   big batch): external sources rate-limit per window, and SQLite (event
   upserts plus the every-minute expired-events cron) lives in the same
   process.
4. **Capacity arithmetic** — the refresh cycle must fit its budget:
   `cities × per-city latency / concurrency ≤ freshness target`; the concurrency
   model must be explicit (order of magnitude: at ~2s per city, one stream
   covers ~20k cities in 12h).
5. **Fairness** — when the set exceeds what one window can serve, rotation must
   be due-based / round-robin; today the same first N cities always win.
6. **Crash-resilience** — scheduling state (last fetch, next due) survives
   restarts and redeploys; a restart resumes instead of resetting all clocks.
7. **Fault isolation** — a failing source or city is skipped and retried with
   backoff, never blocking the others (currently a single failing collector
   fails the whole city — composite returns early on error).
8. **Boring infra** — single Go binary, PocketBase built-in cron, state in
   SQLite; no external queue, scheduler or cache service.
9. **Source heterogeneity** — collectors differ in per-city cost, latency, rate
   limits and geographic coverage (Paris opendata is Paris-only, bobine is
   France-focused, allevents.in is worldwide and costs ≥2 API calls per city);
   the design must handle sources that do not apply to every location.
10. **Write-path invariants** — however collection is triggered, writes keep
    `Event.IsValid`, the `(name, begin, end)` upsert key, canonical UTC date
    layout (ADR 0002) and the pins/window semantics (ADR 0001, ADR 0003).
11. **Observability** — it must be possible to answer "when was city X last
    fetched, how late is it?" in order to prove the freshness ceiling holds.
12. **Timezones (soft)** — worldwide cities have local evenings; a single
    global tick ignores local prime time. Aligning with local time is a soft
    preference, not a hard requirement.

## Decision

**Stateless clock-sharding, driven by the PocketBase builtin cron.**

- Locations live in SQLite (tier, name, coordinates, radius), provisioned from
  an embedded catalog (see Provisioning below) — no city list in Go source.
  The schedule itself holds no state: **the clock is the only scheduler
  state**.
- A 1-minute in-process tick (same cron as the expired-events job) computes the
  current slot `unix_minutes mod S` and collects the cities whose
  `hash(slug) mod S` equals it. With **S = 720** (12h at 1-minute resolution),
  every city is fetched exactly once per 12h period, spread evenly across 720
  ticks. The slot index is plain modulo of the clock — deliberately not
  `hash(now)`, which would only hit ~63% of slots per period and break the 12h
  ceiling.
- Ticks run their batch through a bounded worker pool (constraint 4) and skip
  if the previous tick is still running (run-guard).
- Tiers map to independent rings: the default ring is S = 720 (12h); a looser
  tier for small cities (e.g. 48h) is its own ring with S = 2880. The same
  tick evaluates all rings.
- A failed city is simply retried at its next slot; per-city `last_fetched_at`
  and `last_error` are recorded for observability (constraint 11).
- Collection runs in-process and writes through the same validation path as
  `PUT /api/events` (constraint 10). `cmd/populate` remains a manual / backfill
  tool; the server becomes the sole scheduler.

### Provisioning

A fresh instance must not boot with an empty `locations` table — the ring
would have nothing to collect. The catalog is embedded in the binary:

- Source: the official **GeoNames dump `cities15000`** (~25k cities with
  population > 15k or capitals, CC BY 4.0). A small script fetches the dump
  and checks in a derived JSON catalog (~2–3 MB), `go:embed`ded in the
  binary. The dump is never fetched at boot: startup stays hermetic.
- On every boot the catalog is **idempotently upserted** into `locations`:
  the catalog is authoritative for its rows; runtime-created locations are
  left untouched. A fresh instance is instantly fully populated.
- **Slugs are `<name>-<geonameid>`**: the GeoNames `geonameid` guarantees
  uniqueness and stability, so city renames never move a hash slot.
- **Tiers derive from population**: pop ≥ 100k → 12h ring (S = 720, ~4.3k
  cities); smaller cities → 48h ring (S = 2880, ~21k cities). Both fit the
  tick budget with margin.
- License: CC BY 4.0 requires a credit line for GeoNames (README / UI).
- Growth: regenerate the catalog from a newer dump; `cities5000` (~50k) is
  the drop-in upgrade when more small cities are wanted.

## Options considered

- **B. SQLite due-queue (`next_fetch_at` per location)** — retry, backoff and
  tiers are natural, but the mutable schedule state must be kept correct across
  restarts by hand. Rejected: more moving parts for no hard-requirement gain.
- **C. Persisted-cursor round-robin under the external cron** — smallest
  change, but bursty and tied to an external cron's reliability. Rejected:
  only partially meets the no-bursts constraint and keeps scheduling outside
  the binary.

## Consequences

- Staleness ceiling = one period + downtime (12h while the process is up).
- Capacity check: 20k cities → ~28 cities per tick on average; at ~2s per city
  and 8 workers, a tick takes ~7s of its 60s budget (constraint 4 holds).
- The city→slot mapping is keyed on the geonameid-based slug, which never
  changes; adding cities never rebalances existing ones.
- No intra-period retry: a failed city waits for its next slot. If that ever
  hurts, the escape hatch is a small failed-cities retry list on top of the
  ring (hybrid), not a redesign.
- At very small city counts most ticks are empty; that is expected and free.
- Multiple server instances would double-collect; single-instance is assumed.

## Open points (to be decided later)

- **Rollout lever** — day-one scale (~25k cities, ~50k allevents.in calls per
  period, bobine queried for every city) risks throttling or an IP ban that
  would also kill current French coverage. Candidate: config cap (country
  allowlist / max enabled locations), ramping France → Europe → world.
- **Per-source coverage** (constraint 9) — collectors need a coverage
  predicate; today the Paris filter is a hardcoded city-name check
  (pkg/infrastructure/collector/opendata_paris_quefaire.go) and bobine queries
  every city blindly. Candidates: `country_code` column on locations plus a
  per-collector allowlist / `Supports(location)`.
- **Composite aborts the whole city on one source error**
  (pkg/infrastructure/collector/composite_collector.go returns early) —
  per-source isolation, partial save and per-source error recording to be
  decided.
- **Provisioning leftovers** — `radius` is not in GeoNames (derive from
  population?); cities removed from a newer dump are never deleted (disable vs
  keep); runtime-created locations have no decided creation mechanism (v1
  catalog-only vs PocketBase collection / admin endpoint); store the GeoNames
  `timezone` column now to serve constraint 12 later.
- **Shared write path** — the upsert + `Event.IsValid` logic lives in the HTTP
  handler (pkg/infrastructure/server/requests/events.go); extract to a
  repository/service shared with the in-process scheduler.
- **Stale-event semantics** — upsert-only means cancelled events linger until
  `end` passes (bounded by the 15-day `Event.IsValid` span + the
  expired-events cron); to be confirmed and recorded as an accepted
  consequence.
- **Test & measurement hooks** — pure ring functions (`slotFor`, `dueAt`) for
  table tests; integration test for catalog upsert idempotency; measure real
  per-city latency early (worst case 3 sources × 10s timeouts vs the ~2s
  assumption).
