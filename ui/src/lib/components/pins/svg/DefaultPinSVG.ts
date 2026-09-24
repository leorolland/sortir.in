import { createPinSVG } from './PinShape';

export function createDefaultPinSVG(color: string): string {
  return createPinSVG(color, `<path d="M22 10.5 C22.9 15.9 27.1 20.1 32.5 21 C27.1 21.9 22.9 26.1 22 31.5 C21.1 26.1 16.9 21.9 11.5 21 C16.9 20.1 21.1 15.9 22 10.5 Z" fill="#fff"/>`, 1.1);
}
