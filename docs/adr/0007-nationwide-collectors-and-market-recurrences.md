# ADR 0007 — Nationwide one-shot collectors and market recurrences

- **Status:** Accepted
- **Date:** 2026-10-06

## Context

The `Collector` interface is per-location: populate runs every collector for
every city. Some national datasets have no per-location query and cover the
whole territory at once — the first one is **Les marchés de France**
(https://www.marches-de-france.fr), a merged base of ~5.7k French markets
built from OpenStreetMap (ODbL) and DATAtourisme (Licence Ouverte), served as
a single ~500 KB CSV. Re-querying it once per city is meaningless.

Markets are weekly recurrences ("Sa", "Tu-Su" 08:00-13:00), not dated events,
while the events table only stores concrete `begin`/`end` rows. Two more
dataset quirks shape the design: ~37% of rows carry no structured days
(DATAtourisme-only entries) and generic names are shared across communes
("Marché hebdomadaire" ×230) while events upsert on `(name, begin, end)`.

## Decision

- **New `application.GlobalCollector` interface** (`CollectAll() ([]Event,
  error)`) for nationwide sources, collected once per populate run, before
  the city loop. `application.NewGlobalPopulator` saves the result in batches
  of 1000: the write path upserts one row per event, so tens of thousands of
  events must not arrive in a single request.
- **Recurrences are materialized** into concrete instances over the next 14
  days (Europe/Paris), comfortably beyond the "Cette semaine" window that
  ends the following Monday 3am (ADR 0001). Rows without structured days are
  skipped. Re-runs regenerate the same instances, so the `(name, begin, end)`
  upsert stays idempotent; the expired-events cron removes past instances and
  the 12h populate cadence keeps the window filled — no new cron.
- **Duplicated names are disambiguated deterministically**: commune appended
  ("Marché hebdomadaire (Ville-A)"), address too when still colliding. Free
  hours ("le matin", empty) default to the 8:00-13:00 morning window; "/"
  separated hour variants are unioned; seasons ("Jul-Aug", "Sep-Jun" wrapping
  the new year) restrict the expansion, unparsable ones fail open.
- **New `market` kind** with its own pin SVG and UI label; DATAtourisme's
  generic `market` type and the `"marché"` kind-string move from
  `flea-market` to `market` (brocantes keep `flea-market`).
- **CSV fetched at populate time** (no embedding): a stale schedule is more
  harmful than a network hiccup, and the 12h cadence puts negligible load on
  the source.

## Consequences

- ~3.5k markets with structured days yield ~8-12k event instances per run,
  saved in 12 batches.
- A failing GlobalCollector never blocks the city loop (populate logs and
  continues).
- ODbL attribution is required in the project documentation.
- If a market's name disambiguation set changes (market added/removed), the
  affected instances change name; old rows linger at most 15 days (IsValid
  span + expired-events cron).

## Sources

- [Les marchés de France — données & licences](https://www.marches-de-france.fr/donnees-licences.html)
