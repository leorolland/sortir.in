# ADR 0001 — Time-range filters are interval windows with a 3am night cutoff

- **Status:** Accepted
- **Date:** 2026-09-19

## Context

Time-range filters must answer a simple question: "what is going on today /
tomorrow / this week?" Two requirements shape the answer: events that are
actually running (including multi-day festivals and expos) must stay visible,
and night activities (a party starting at 23h and ending at 2am) must belong to
the day that started them.

## Decision

Each range defines a **window `[min, max]`**, and an event is displayed when its
`[begin, end]` interval overlaps the window:

```
begin <= max  AND  end >= min
```

Implemented in `getDateWindow` (ui/src/lib/utils/dateUtils.ts). All boundaries
are **3am cuts**, applied homogeneously:

| Range | Window |
|---|---|
| Aujourd'hui | `[now → tomorrow 3am]` |
| Demain | `[tomorrow 3am → day-after 3am]` |
| Cette semaine | `[now → Monday 3am]` (the next Sunday's night belongs to the week) |

"Cette semaine" ends on Monday 3am — not Sunday 23:59 — so that its window is
nested with "Demain" and no night is ever cut in half by a range boundary.

## Consequences

- Long events (up to 15 days, cf. `Event.IsValid`) appear in every overlapping
  range; this is intended ("actually running").
- Windows with nested scopes must remain nested; any new range must define both
  a `min` and a `max`.
- The "Ce soir" label in the range selector is purely cosmetic; filtering never
  depends on the time of day.
