import { writable } from 'svelte/store';

/**
 * Position of the mobile bottom sheet (pin popup on touch devices):
 * - normal: hugging its content, capped at 60dvh
 * - expanded: stretched to 90dvh (after scrolling inside the sheet)
 * - peek: collapsed to its title bar (after the user pans the map),
 *   so the map takes most of the screen and the pin stays reachable
 */
export type SheetPosition = 'normal' | 'expanded' | 'peek';

export const sheetState = writable<SheetPosition>('normal');
