import { PIN_PIXEL_RATIO } from './PinShape';

// White rounded rectangle stretched around the pin label text by MapLibre
// (icon-text-fit + nine-patch metadata, see loadLabelBackgroundImage in
// MapView.svelte). One shared image for every pin: only the text changes.
//
// With icon-text-fit, the drawn size is text + padding + the image border
// around its content box, so the content box must span nearly the whole
// image — only ~2css px of border on each side. The image is 96x40 image
// pixels (48x20 css at PIN_PIXEL_RATIO), the pill's natural height; only
// its width really flexes with the text.
export const LABEL_BACKGROUND_IMAGE = 'label-bg';
const LABEL_BG_WIDTH = 96;
const LABEL_BG_HEIGHT = 40;
const LABEL_BG_CORNER = 16;

export function createLabelBackgroundSVG(): string {
	return `<svg width="${LABEL_BG_WIDTH}" height="${LABEL_BG_HEIGHT}" viewBox="0 0 ${LABEL_BG_WIDTH} ${LABEL_BG_HEIGHT}" xmlns="http://www.w3.org/2000/svg">
  <rect width="${LABEL_BG_WIDTH}" height="${LABEL_BG_HEIGHT}" rx="${LABEL_BG_CORNER}" fill="#fff"/>
</svg>`;
}

// map.addImage nine-patch metadata, in image pixels: stretch bands live in
// the middle, away from the rounded corners; the content box is the image
// minus a 4px border.
export const LABEL_BACKGROUND_STRETCH: [number, number][] = [
	[LABEL_BG_CORNER * 2, LABEL_BG_WIDTH - LABEL_BG_CORNER * 2]
];
export const LABEL_BACKGROUND_STRETCH_Y: [number, number][] = [
	[LABEL_BG_HEIGHT / 2 - 4, LABEL_BG_HEIGHT / 2 + 4]
];
export const LABEL_BACKGROUND_CONTENT: [number, number, number, number] = [
	4,
	4,
	LABEL_BG_WIDTH - 4,
	LABEL_BG_HEIGHT - 4
];
