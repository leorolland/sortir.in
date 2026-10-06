# ADR 0008 — Pin labels (short place titles)

- **Status:** Accepted
- **Date:** 2026-10-06

## Context

Pins are aggregates `(location, kind)` (ADR 0003): they carry only `loc`,
`kind` and `amount`, so a pin is a colored balloon with no hint of *which*
place it marks. The user must open the popup to identify a venue. A small
white rounded pill on the pin head, with the 2-3 most identifying words of
the place in small bold black text, makes each pin evoke its venue
("Cinéma Saint-André des Prés" → "Saint-André", "Librairie Le Delta" →
"Le Delta").

Two constraints shape the design: `place` is often missing or a street
address (markets: "Avenue Jean Jaurès", "13 Rue Henri"), and a pin groups
several events that may disagree on the place name.

## Decision

- **`/api/pins` gains an optional `label`** (string, omitted when empty).
  The repository additionally selects `place`; `Pin.Place` stays internal
  (`json:"-"`) and never reaches the payload.
- **The label source is the most frequent non-empty, non-address-like place**
  among the pin's events, tie-broken lexicographically for determinism.
- **`application.ShortPlaceTitle`** (pure, table-tested) reduces the name:
  strip the leading venue-type word ("Cinéma", "Théâtre", …); drop the
  trailing prepositional phrase ("… des Prés") unless that leaves a generic
  word; cap at 3 words / 24 runes (ellipsis); title-case, keeping stop words
  lowercase; street addresses (leading number or street word) and single
  generic words ("Place") yield no label.
- **Rendering is one extra MapLibre symbol layer** (`event_labels`): a
  shared nine-patch white pill image (`label-bg`, corners outside the
  stretch bands) stretched around the text with `icon-text-fit: both`,
  `text-offset` centering it on the balloon head (15.5px above the anchor).
  MapLibre collision declutters overlapping labels; labeled pins stay
  clickable through the same click handler (both layers queried). Pins
  without label render exactly as before.

## Consequences

- Payload grows only by the labels (~30 bytes per labeled pin).
- The heuristic is French-oriented; English names only get the generic
  casing/cap rules. Bad source data (addresses in `place`) renders as
  plain pins instead of noise labels.
- The pill covers the pin's glyph: the kind stays readable through the
  balloon color, not the glyph.
- Adding a source that emits clean venue names in `place` instantly gets
  labels for free; no schema migration (label is derived, never stored).
