# Map pins (SVG)

Flat, Apple-Maps-style location pins. All pins share one balloon shape and
differ only by a white glyph and a fill color. Rendered on the map by
`MapView.svelte` as MapLibre images (`pin-<kind>`).

## Files

| File | Role |
|---|---|
| `PinShape.ts` | Shared balloon path, `PIN_VIEWBOX`, `PIN_PIXEL_RATIO`, `createPinSVG(color, glyph)` |
| `DefaultPinSVG.ts` … `PartyPinSVG.ts` | One file per kind; each returns `createPinSVG(color, glyphMarkup)` |
| `index.ts` | `PIN_COLORS` palette, `pinSVGs` registry, re-exports `PIN_PIXEL_RATIO` |
| `MapView.svelte` | Loads blobs as `Image`s → `map.addImage('pin-' + name, img, { pixelRatio: 2 })`; maps `kind` → image in the `icon-image` `match` expression |

## Shape geometry (PinShape.ts)

ViewBox `0 0 44 52`, tip at the bottom: `icon-anchor: bottom` pins the tip to
the event coordinate, so **the tip must stay at the bottom of the viewBox**
(≤ 0.5 px slack) or every pin will float above its real location.

The balloon is a single continuous teardrop — no visible joint between the
round body and the tail:

- Body: circle, center `(22, 21)`, radius `19.5`.
- Tail exits the circle tangentially at ±48° from the bottom vertical
  (points `(35.05, 35.49)` / `(8.95, 35.49)`); the first Bézier control point
  of each tail curve lies on the circle tangent at the exit point
  (`c1 = exit + 6 · tangent`), which is what makes the connection smooth.
- Tip: rounded by a quadratic cap `Q22 53` between `(23.4, 50.9)` and
  `(20.6, 50.9)` — keep it if you want the blunt tip; remove the `Q` and join
  the curves at a single point for a sharp tip.

## Glyph rules

- White (`#fff`) on the solid color; details inside the glyph use the pin
  color (stripes, eyes, mouth).
- Optical centering around `(22, 21)`, kept within ~13 px of that point.
  Asymmetric glyphs (flag, popper) are shifted so their visual mass — not
  their bounding box — sits at the center.
- Per-kind tuning via `createPinSVG(color, glyph, glyphScale, glyphDx)`:
  `glyphScale` zooms about the body center, `glyphDx` shifts horizontally
  (e.g. concert `1.1, -1`, party `1.2`).
- One or two simple filled shapes max; SF-Symbols-like, flat, no strokes
  except for deliberate line details (theater smile).

## Palette

Apple system colors in `index.ts` (`PIN_COLORS`), all high-contrast with
white glyphs. Adding a kind: new `XxxPinSVG.ts` + entry in `pinSVGs` +
`match` arm in `MapView.svelte` (`'xxx', 'pin-xxx'`).

## Retina crispness

SVGs carry `width`/`height` at 2× the viewBox (`PIN_PIXEL_RATIO = 2`) and are
registered with `map.addImage(..., { pixelRatio: PIN_PIXEL_RATIO })`. If you
change `PIN_VIEWBOX` or the ratio, both sides must stay in sync or pins render
blurry or mis-sized.

## Previewing the pins locally

Render a contact sheet from the real modules (Node ≥ 22.6 type stripping) and
open it:

```sh
SRC=ui/src/lib/components/pins/svg
mkdir -p /tmp/opencode/pinsrc
for f in "$SRC"/*.ts; do sed -E "s|from './([A-Za-z]+)'|from './\1.ts'|g" "$f" > "/tmp/opencode/pinsrc/$(basename "$f")"; done
```

Then run a small script that imports `pinSVGs` from `/tmp/opencode/pinsrc/index.ts`,
writes each SVG plus a `<g>`-nested sheet, and thumbnail it:

```sh
qlmanage -t -s 1024 sheet.svg -o /tmp/opencode/pins   # → sheet.svg.png
```

(Copy of the script used during the 2026 redesign: `/tmp/opencode/render-pins.mjs`
— recreate it if /tmp was cleared.)
