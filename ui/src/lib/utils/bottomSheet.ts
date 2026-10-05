import { get } from 'svelte/store';
import { sheetState, type SheetPosition } from '$lib/stores/sheet';

/**
 * Gesture-driven bottom sheet (ADR 0006): the panel is moved with
 * translateY — a drag tracks the finger 1:1, release settles to the nearest
 * detent biased by velocity. Detents are offsets, not layout states:
 * - peek: only the title bar is visible (PEEK_HEIGHT)
 * - normal: the sheet hugs its content, capped at 60% of the viewport
 * - expanded: the full sheet (content height, capped by the panel max-height)
 *
 * The panel keeps its natural CSS height (hugging content, capped by its
 * max-height); this action only owns `transform` (and toggles the settle
 * transition). `sheetState` stays the source of truth for which detent is
 * active, so class-driven styles and external sets (map pan -> peek,
 * tap bar -> normal) keep working.
 */

const PEEK_HEIGHT = 76;
const NORMAL_CAP = 0.6;
const SETTLE_MS = 320;
const SETTLE_EASE = 'cubic-bezier(0.32, 0.72, 0, 1)';
const HEADER_ZONE = 56;
const RUBBER = 0.25;
const FLICK_PROJECT_MS = 140;
const DEAD_ZONE = 4;
const DETENT_MERGE = 8;

type Mode = 'undecided' | 'sheet' | 'scroll';

interface Detent {
  pos: SheetPosition;
  y: number;
}

export function sheetDrag(node: HTMLElement) {
  if (!window.matchMedia('(hover: none) and (pointer: coarse)').matches) {
    return { destroy() {} };
  }

  const panelEl = node.closest<HTMLElement>('.floating-panel');
  const scrollerEl = node.closest<HTMLElement>('.floating-panel-content');
  if (!panelEl || !scrollerEl) return { destroy() {} };
  const panel: HTMLElement = panelEl;
  const scroller: HTMLElement = scrollerEl;

  const reducedMotion = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

  let contentH = 0;
  let y = 0;
  let dragging = false;
  let mode: Mode = 'undecided';
  let startInHeader = false;
  let suppressClick = false;
  let startY = 0;
  let yAtDragStart = 0;
  let prevSample = { t: 0, y: 0 };
  let lastSample = { t: 0, y: 0 };
  let resizeObserver: ResizeObserver | undefined;

  // scrollHeight is the natural content height; it is measured with the
  // un-peeked styles (peek hides the link chips and changes paddings)
  function measure() {
    if (get(sheetState) === 'peek') return;
    contentH = scroller.scrollHeight;
  }

  function panelH() {
    // The size cap is owned by CSS (.dynamic-panel max-height, EventPopup):
    // read it instead of duplicating the value here
    return Math.min(contentH || panel.offsetHeight, parseFloat(getComputedStyle(panel).maxHeight) || Infinity);
  }

  function detents(): Detent[] {
    const ph = panelH();
    const list: Detent[] = [{ pos: 'expanded', y: 0 }];
    const yNormal = ph - Math.min(contentH, window.innerHeight * NORMAL_CAP);
    if (yNormal > DETENT_MERGE) list.push({ pos: 'normal', y: yNormal });
    // Peek reads the live panel height: the .sheet-peeked styles (chips
    // hidden, reduced paddings) shrink small sheets, so an offset computed
    // from the un-peeked height would translate them past their own top
    // edge and hide the bar completely.
    list.push({ pos: 'peek', y: Math.max(0, panel.offsetHeight - PEEK_HEIGHT) });
    return list;
  }

  function yFor(pos: SheetPosition): number {
    const found = detents().find((d) => d.pos === pos);
    return found ? found.y : 0;
  }

  function currentY(): number {
    const t = getComputedStyle(panel).transform;
    if (t && t !== 'none') {
      const m = t.match(/matrix\(([^)]+)\)/);
      if (m) return parseFloat(m[1].split(',')[5]) || 0;
    }
    return y;
  }

  function applyY(next: number, animate: boolean) {
    panel.style.transition = animate && !reducedMotion ? `transform ${SETTLE_MS}ms ${SETTLE_EASE}` : 'none';
    panel.style.transform = `translateY(${next}px)`;
    y = next;
  }

  function sync(animate: boolean) {
    measure();
    if (!dragging) applyY(yFor(get(sheetState)), animate);
  }

  function rubber(v: number, min: number, max: number): number {
    if (v < min) return min - (min - v) * RUBBER;
    if (v > max) return max + (v - max) * RUBBER;
    return v;
  }

  // The sheet only takes over a gesture when the scroller cannot consume
  // it: header/grabber always drags; downward drags at scrollTop 0 drag the
  // sheet (overscroll-behavior: none means the browser scrolls nothing);
  // upward drags drag the sheet unless it is already fully expanded, where
  // the content scrolls instead.
  function decideMode(dy: number): Mode {
    if (startInHeader) return 'sheet';
    if (dy > 0) return scroller.scrollTop <= 0 ? 'sheet' : 'scroll';
    if (get(sheetState) === 'expanded') return 'scroll';
    return 'sheet';
  }

  const onTouchStart = (e: TouchEvent) => {
    const touch = e.touches[0];
    // freeze an in-flight settle where it currently is
    y = currentY();
    panel.style.transition = 'none';
    startY = touch.clientY;
    mode = 'undecided';
    suppressClick = false;
    prevSample = lastSample = { t: e.timeStamp, y: touch.clientY };
    startInHeader = touch.clientY - panel.getBoundingClientRect().top <= HEADER_ZONE;
  };

  const onTouchMove = (e: TouchEvent) => {
    if (mode === 'scroll') return;
    const touch = e.touches[0];
    const dy = touch.clientY - startY;
    if (mode === 'undecided') {
      if (Math.abs(dy) < DEAD_ZONE) return;
      mode = decideMode(dy);
      if (mode !== 'sheet') {
        suppressClick = false;
        return;
      }
      dragging = true;
      suppressClick = true;
      panel.style.transition = 'none';
      yAtDragStart = y;
    }
    // the sheet may own a gesture the browser already started scrolling
    // (content overflowing at a non-top detent, pulled up from scrollTop 0)
    if (dy < 0 && scroller.scrollTop > 0) scroller.scrollTop = 0;
    e.preventDefault();
    y = rubber(yAtDragStart + dy, 0, yFor('peek'));
    applyY(y, false);
    prevSample = lastSample;
    lastSample = { t: e.timeStamp, y: touch.clientY };
  };

  const onTouchEnd = () => {
    if (!dragging) {
      mode = 'undecided';
      return;
    }
    dragging = false;
    mode = 'undecided';
    const dt = lastSample.t - prevSample.t;
    const dy = lastSample.y - prevSample.y;
    const velocity = dt > 0 && dt < 100 ? dy / dt : 0;
    const projected = y + velocity * FLICK_PROJECT_MS;
    let target = detents()[0];
    for (const d of detents()) {
      if (Math.abs(d.y - projected) < Math.abs(target.y - projected)) target = d;
    }
    // the subscriber animates to the chosen detent
    sheetState.set(target.pos);
  };

  const onClick = (e: MouseEvent) => {
    if (suppressClick) {
      suppressClick = false;
      return;
    }
    if (get(sheetState) !== 'peek') return;
    const target = e.target as HTMLElement;
    if (target.closest('button, a')) return;
    sheetState.set('normal');
  };

  const onResize = () => sync(true);

  panel.addEventListener('touchstart', onTouchStart, { passive: true });
  panel.addEventListener('touchmove', onTouchMove, { passive: false });
  panel.addEventListener('touchend', onTouchEnd);
  panel.addEventListener('touchcancel', onTouchEnd);
  panel.addEventListener('click', onClick);
  window.addEventListener('resize', onResize);

  resizeObserver = new ResizeObserver(() => sync(true));
  resizeObserver.observe(panel);
  resizeObserver.observe(scroller);

  const unsubscribe = sheetState.subscribe(() => {
    requestAnimationFrame(() => sync(true));
  });
  // initial position, before the first rAF
  sync(false);

  return {
    destroy() {
      unsubscribe();
      resizeObserver?.disconnect();
      panel.removeEventListener('touchstart', onTouchStart);
      panel.removeEventListener('touchmove', onTouchMove);
      panel.removeEventListener('touchend', onTouchEnd);
      panel.removeEventListener('touchcancel', onTouchEnd);
      panel.removeEventListener('click', onClick);
      window.removeEventListener('resize', onResize);
    }
  };
}
