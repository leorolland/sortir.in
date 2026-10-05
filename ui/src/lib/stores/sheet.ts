import { writable } from 'svelte/store';

/**
 * Detent of the mobile bottom sheet (pin popup on touch devices). Movement
 * is gesture-driven (utils/bottomSheet.ts, ADR 0006): dragging tracks the
 * finger and release settles to the nearest detent. The store only records
 * which detent is active:
 * - normal: sheet hugging its content, capped at 60dvh
 * - expanded: full sheet, capped by the panel max-height (95dvh)
 * - peek: collapsed to its title bar (after the user pans the map),
 *   so the map takes most of the screen and the pin stays reachable
 */
export type SheetPosition = 'normal' | 'expanded' | 'peek';

export const sheetState = writable<SheetPosition>('normal');
