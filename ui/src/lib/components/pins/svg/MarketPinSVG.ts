import { createPinSVG } from './PinShape';

export function createMarketPinSVG(color: string): string {
  return createPinSVG(
    color,
    `<path d="M14.8 19.8 h14.4 l-1.7 8.1 a2.6 2.6 0 0 1 -2.55 2.1 h-5.9 a2.6 2.6 0 0 1 -2.55 -2.1 Z" fill="#fff"/>
  <path d="M17.9 19.3 a4.1 4.1 0 0 1 8.2 0" stroke="#fff" stroke-width="2.1" fill="none" stroke-linecap="round"/>`
  );
}
