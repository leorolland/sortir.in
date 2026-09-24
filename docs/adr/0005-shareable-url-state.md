# ADR 0005 — Shareable URL state for map position and pin focus

- **Status:** Accepted
- **Date:** 2026-09-24

## Context

The map state (center, zoom, date range, focused pin) lives only in memory, so
a copied URL always reopens the default view. Also, the popup was bound to
svelte-maplibre's layer click handling and could not be opened
programmatically.

## Decision

- The URL query mirrors the view: `lat`, `lng`, `z` (rounded), `range`
  (omitted when default `today`) and `pin=<lat>,<lon>`. Pin coordinates keep
  full float precision: the events filter compares `loc.lat`/`loc.lon`
  exactly. Unknown parameters are preserved.
- Writes: `replaceState` on map moves (`moveend`) and range changes (no
  history spam); `pushState` on pin focus so Back closes the popup; closing
  the popup drops `pin` with `replaceState` (never `history.back()`, which
  could leave the site). Writes wait for router initialization
  (`afterNavigate`), and the readiness flag lives in module state to survive
  HMR.
- Reads: on load, params initialize position, range and popup; invalid values
  fall back to defaults. A `popstate` listener re-applies the encoded view so
  Back/Forward restore the full snapshot.
- Pin focus becomes app state (`selectedPin`); the popup is a manual
  svelte-maplibre `Popup` anchored to it. UX (bottom sheet, viewport fitting,
  outside-click close) is unchanged.

## Consequences

- Any copied URL reopens the same position, zoom, range and popup. Ranges are
  relative names, so shared links stay live (resolved against the recipient's
  "now").
- The events popup content comes from `loadEventsForLocation`, so a link works
  even when the pin itself is no longer rendered (expired, clustered).
