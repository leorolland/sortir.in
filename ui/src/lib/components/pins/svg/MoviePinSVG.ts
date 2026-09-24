import { createPinSVG } from './PinShape';

export function createMoviePinSVG(color: string): string {
  return createPinSVG(
    color,
    `<rect x="13" y="20" width="18" height="11" rx="2.2" fill="#fff"/>
  <g transform="rotate(-12 22 16.5)">
    <rect x="13.5" y="13.8" width="17" height="4.4" rx="1.4" fill="#fff"/>
    <rect x="16.8" y="13.8" width="1.9" height="4.4" fill="${color}"/>
    <rect x="21.05" y="13.8" width="1.9" height="4.4" fill="${color}"/>
    <rect x="25.3" y="13.8" width="1.9" height="4.4" fill="${color}"/>
  </g>`,
    1.1
  );
}
