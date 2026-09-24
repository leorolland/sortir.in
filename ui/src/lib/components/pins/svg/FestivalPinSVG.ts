import { createPinSVG } from './PinShape';

export function createFestivalPinSVG(color: string): string {
  return createPinSVG(
    color,
    `<rect x="15.6" y="10.5" width="2.3" height="21" rx="1.15" fill="#fff"/>
  <path d="M17.9 11.5 L32.8 15.2 L17.9 19.4 Z" fill="#fff"/>`,
    1.1
  );
}
