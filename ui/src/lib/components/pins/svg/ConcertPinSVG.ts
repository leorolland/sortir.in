import { createPinSVG } from './PinShape';

export function createConcertPinSVG(color: string): string {
  return createPinSVG(
    color,
    `<ellipse cx="16.9" cy="27.9" rx="3.1" ry="2.4" fill="#fff"/>
  <ellipse cx="27.1" cy="25.9" rx="3.1" ry="2.4" fill="#fff"/>
  <rect x="17.9" y="13.8" width="2.1" height="14.4" fill="#fff"/>
  <rect x="28.1" y="11.8" width="2.1" height="14.4" fill="#fff"/>
  <path d="M17.9 13.8 L30.2 11.6 L30.2 15 L17.9 17.2 Z" fill="#fff"/>`,
    1.1,
    -1
  );
}
