// Flat Apple-Maps-style balloon shared by every pin.
// viewBox is 44x52 with the tip at the bottom (icon-anchor: bottom).
export const PIN_VIEWBOX = { width: 44, height: 52 };

// Pins are rendered at 2x and registered with map.addImage(..., { pixelRatio: 2 })
// so they stay crisp on retina displays.
export const PIN_PIXEL_RATIO = 2;

// Teardrop: circle (center 22,21, r 19.5) with a tail that leaves the circle
// tangentially on both sides, softly rounded at the tip (see README.md).
const BALLOON_PATH =
  'M22 1.5 ' +
  'A19.5 19.5 0 0 1 35.05 35.49 ' +
  'C30.59 39.51 24.73 48.9 23.4 50.9 ' +
  'Q22 53 20.6 50.9 ' +
  'C19.27 48.9 13.41 39.51 8.95 35.49 ' +
  'A19.5 19.5 0 0 1 22 1.5 Z';

// glyphScale/glyphDx: glyph zoom (about body center 22,21) and horizontal
// optical shift, for per-kind size/balance tuning.
export function createPinSVG(color: string, glyph: string, glyphScale = 1, glyphDx = 0): string {
  const glyphTransform =
    glyphScale === 1 && glyphDx === 0
      ? ''
      : ` transform="translate(${22 + glyphDx} 21) scale(${glyphScale}) translate(-22 -21)"`;
  return `<svg width="${PIN_VIEWBOX.width * PIN_PIXEL_RATIO}" height="${PIN_VIEWBOX.height * PIN_PIXEL_RATIO}" viewBox="0 0 ${PIN_VIEWBOX.width} ${PIN_VIEWBOX.height}" xmlns="http://www.w3.org/2000/svg">
  <path d="${BALLOON_PATH}" fill="${color}"/>
  <g${glyphTransform}>${glyph}</g>
</svg>`;
}
