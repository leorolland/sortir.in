import { createPinSVG } from './PinShape';

export function createTheaterPinSVG(color: string): string {
  return createPinSVG(
    color,
    `<path d="M13.7 14.2 C13.7 13.2 14.5 12.4 15.5 12.4 L28.5 12.4 C29.5 12.4 30.3 13.2 30.3 14.2 L30.3 20.9 C30.3 25.8 26.7 29.6 22 29.6 C17.3 29.6 13.7 25.8 13.7 20.9 Z" fill="#fff"/>
  <circle cx="17.8" cy="19.3" r="1.5" fill="${color}"/>
  <circle cx="26.2" cy="19.3" r="1.5" fill="${color}"/>
  <path d="M17.4 22.9 C19.2 25.3 24.8 25.3 26.6 22.9" fill="none" stroke="${color}" stroke-width="2" stroke-linecap="round"/>`,
    1.1
  );
}
