---
name: live-data-probe
description: Inspect the live sortir.in data and API — query pb_data SQLite, call /api/pins with correct date windows, understand pin counts. Use when debugging event counts, verifying what the map displays, checking collector output, or investigating date/window behavior against real data.
---

# Probing live data

The dev stack runs via `make dev` (modd): server on http://localhost:8090 with
the local DB in `pb_data/data.db` (gitignored, disposable, but mutated live by
collectors and the expired-events cron — never edit its schema by hand).

## API

`GET /api/pins` — bounds + window, returns `[{loc: {lat, lon}, kind, amount}]`:

```bash
curl -s "http://localhost:8090/api/pins?north=51.5&south=41.0&east=10&west=-6\
&min_time=$(date -u +%Y-%m-%dT%H:%M:%SZ)&max_time=$(date -u -v+1d -v3H -v0M -v0S +%Y-%m-%dT%H:%M:%SZ)"
```

- `max_time` (required) and `min_time` (optional) are RFC3339.
- Semantics (ADR 0001): an event matches when `begin <= max_time && end >= min_time`.
- Pin `amount` sums give event counts; per-kind and per-day breakdowns come
  cheaper from SQL (below).

`PUT /api/events` accepts a JSON array of events (used by cmd/populate); it
silently skips events failing `Event.IsValid` (already ended, or > 15 days).

## SQLite

Dates are stored in PocketBase's canonical UTC layout
`2006-01-02 15:04:05.000Z` (ADR 0002). Always compare against strings in that
exact layout — a `time.Time`-bound param or an RFC3339 `T`-separated string
compares wrongly (`'T' > ' '`).

```bash
# events overlapping a window (same SQL the repository runs)
sqlite3 pb_data/data.db "SELECT COUNT(*) FROM events WHERE \`begin\` <= '<max>' AND \`end\` >= '<min>';"

# events per begin day
sqlite3 pb_data/data.db "SELECT date(\`begin\`), COUNT(*) FROM events GROUP BY 1 ORDER BY 1;"

# duration distribution (0 = single-instant, e.g. many showtimes)
sqlite3 pb_data/data.db "SELECT CAST((julianday(\`end\`)-julianday(\`begin\`)) AS INT), COUNT(*) FROM events GROUP BY 1 ORDER BY 2 DESC;"
```

`begin`/`end` are keywords: backtick them. France is UTC+1/+2 — a local 3am
window cut is `01:00`/`02:00` UTC depending on DST; compute UTC strings with
`date -u` rather than by hand.

## Read-only discipline

Treat `pb_data` as read-only unless the task is explicitly about writing data:
the dev server may be running against it and the user is often mid-testing.
