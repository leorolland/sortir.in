# ADR 0002 — Event dates are stored in PocketBase's canonical UTC layout

- **Status:** Accepted
- **Date:** 2026-09-19

## Context

Event dates are filtered in SQL with text comparisons (SQLite). Text comparison
is only correct when all values share one unambiguous format: mixed offsets
(`+02:00` vs `Z`) or free-form layouts do not compare chronologically.

## Decision

- `begin` and `end` are stored in PocketBase's canonical layout
  `2006-01-02 15:04:05.000Z` in UTC (`types.DefaultDateLayout`).
- On write (`PUT /api/events`), dates are normalized before insert
  (pkg/infrastructure/server/requests/events.go).
- SQL comparisons bind the bounds as strings formatted with the same layout
  (pkg/infrastructure/repository/event.go), making text comparison exact.
- The frontend mirrors the layout for PocketBase filters with
  `formatDateForFilter` (ui/src/lib/utils/dateUtils.ts).
- A one-off migration (migrations/1789796243_normalized_events_dates.go) rewrote
  existing rows to the canonical layout, deduplicating rows that would collide
  on the `(name, begin, end)` unique index.

## Alternatives considered

- Binding RFC3339 strings instead of `time.Time`: keeps the legacy format and
  its mixed-offset fragility forever.
- SQLite datetime functions on every query: slower and still wrong on mixed
  formats.

## Consequences

- All date values in `events` are UTC; display conversion happens client-side.
- Any new writer must use the canonical layout — date fields are only
  lexicographically comparable in this format.
