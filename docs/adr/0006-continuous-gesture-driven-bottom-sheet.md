# ADR 0006 — Continuous gesture-driven bottom sheet

- **Status:** Accepted
- **Date:** 2026-09-24

## Context

The mobile pin popup is a bottom sheet with three discrete positions —
`peek` (76px title bar), `normal` (content, capped 60dvh) and `expanded`
(90dvh) — driven by the `sheetState` store. Transitions between positions are
scripted CSS animations (`max-height 300ms`), triggered by scroll-position
thresholds inside the sheet (expand at `scrollTop > 140`, collapse at
`scrollTop < 2`, pull-down 60px, map pan). The sheet never follows the finger:
gestures produce delayed, fixed-duration jumps that feel non-native.

## Decision

- The sheet keeps the same three detents and the same store, but positions are
  no longer layout states. The panel is content-hugging up to 95dvh
  (max-height set in CSS, read by the sheet geometry code) and moved with
  `translateY`; each detent is an offset. Changing detent never animates
  `max-height`.
- Movement is direct manipulation: a touch drag sets `translateY` to the
  finger position 1:1 (no transition while dragging), with rubber-banding
  beyond the detent range. On release the sheet settles to the nearest detent,
  biased by release velocity (a flick moves one detent even from a short
  drag). The settle transition is the only scripted animation.
- Gesture arbitration: the header/grabber area always drags the sheet. On the
  content, the gesture drags the sheet only when the inner scroller cannot
  consume it (scroller at top, or no overflow at the current detent);
  otherwise the content scrolls natively. Scrolling inside the sheet no longer
  changes the detent. At `expanded` with overflowing content, upward drags
  scroll content; downward drags at `scrollTop = 0` drag the sheet.
- Width keeps morphing only during settle: the peek pill (6% side insets) is
  class-driven with the existing left/right transition; drags always show the
  current width.
- The close button lives inside the sheet (FloatingPanel `chrome` slot, drawn
  with the shared CrossIcon): MapLibre's built-in button is anchored to the
  untranslated popup container, so it would float over the map once the sheet
  moves with a transform.
- The popup content box does not follow the transform either; it gets
  `pointer-events: none` on touch (the panel re-enables them), so the empty
  region above a translated sheet pans the map instead of swallowing touches.
- The entrance waits for content: the sheet mounts only once its events are
  loaded and the entrance animation starts when it appears, with the detent
  offset already applied — the slide-up lands directly at its detent.
  FloatingPanel's own fadeIn is disabled on touch: it animates transform,
  which overrides the sheet's inline translateY (CSS animations beat inline
  styles) and teleports the sheet to its detent when it ends.
- Preserved behaviors: map pan settles the sheet to `peek`; tapping the peeked
  bar restores it; `prefers-reduced-motion` disables transitions. Desktop
  (hover) popups are untouched.

## Consequences

- The sheet feels native: position always matches the finger during drag;
  velocity carries intent; no threshold jumps.
- Sheet geometry (height, offset) is JS-owned (`utils/bottomSheet.ts`); the
  store remains the source of truth for which detent is active, so CSS class
  hooks (`sheet-peeked`, `sheet-expanded`) and the map-pan/tap behaviors keep
  working unchanged.
- Gesture takeover relies on `touch-action` and `overscroll-behavior` rather
  than `preventDefault`, keeping iOS Safari scrolling predictable; it must be
  validated on a real device (no iOS test automation in this repo —
  verification is on-device).
