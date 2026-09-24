import { createPinSVG } from './PinShape';

export function createPartyPinSVG(color: string): string {
  return createPinSVG(
    color,
    `<g transform="rotate(45 22 21.5)">
    <path d="M22 31 L17.5 17.5 L26.5 17.5 Z" fill="#fff"/>
    <ellipse cx="22" cy="17.5" rx="4.5" ry="1.6" fill="#fff"/>
    <path d="M19 17.5 L20.7 17.5 L22.1 26 Z" fill="${color}"/>
    <path d="M23.3 17.5 L25 17.5 L23.9 26 Z" fill="${color}"/>
  </g>
  <circle cx="27.5" cy="14" r="1.7" fill="#fff"/>
  <circle cx="31.5" cy="18.5" r="1.4" fill="#fff"/>
  <circle cx="23.5" cy="11" r="1.2" fill="#fff"/>`,
    1.2
  );
}
