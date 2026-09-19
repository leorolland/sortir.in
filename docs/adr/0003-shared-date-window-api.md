# ADR 0003 — One shared date window across pins API, sidebar and popup

- **Status:** Accepted
- **Date:** 2026-09-19

## Context

Map pins, the "Suggestions" sidebar and the event popup all display events for
the selected time range. Their filtering must be identical so the surfaces can
never disagree (e.g. a pin without its sidebar entry).

## Decision

- The pins service and repository take an explicit window:
  `ByBoundsAndDateRange(bounds, minDate, maxDate)`; the repository filters
  `begin <= maxDate AND end >= minDate` (ADR 0001 semantics).
- `/api/pins` accepts an optional `min_time` (RFC3339); when absent there is no
  lower bound. `max_time` stays required.
- The sidebar (`eventsStore.getEventsInBounds`) and the event popup
  (`eventsStore.loadEventsForLocation`) apply the same window via the same
  `getDateWindow` helper.

## Consequences

- Changing the range refetches pins and events consistently; the title count
  and the sidebar always reflect the same window.
- Integration tests cover the window semantics: ongoing multi-day events, night
  events ending after the window, terminated events and events starting after
  the window (test/integration/pins_test.go).
