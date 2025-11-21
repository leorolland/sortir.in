/**
 * SVG pour le pin théâtre avec des masques de comédie/tragédie
 */
export function createTheaterPinSVG(color: string): string {
  return `<svg width="40" height="50" viewBox="0 0 40 50" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <filter id="shadow-theater" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="2" stdDeviation="2" flood-color="#000000" flood-opacity="0.3" />
      </filter>
    </defs>

    <!-- Main pin shape -->
    <path d="M20 0C9 0 0 9 0 20C0 31 20 50 20 50C20 50 40 31 40 20C40 9 31 0 20 0Z"
      fill="${color}" filter="url(#shadow-theater)" />

    <!-- Theater masks background -->
    <circle cx="20" cy="20" r="14" fill="white" opacity="0.9" />

    <!-- Comedy mask (left) -->
    <path d="M13,15 Q10,18 12,24 Q15,22 17,19 Z" fill="${color}" />
    <circle cx="13" cy="18" r="1.2" fill="white" />

    <!-- Tragedy mask (right) -->
    <path d="M27,15 Q30,18 28,24 Q25,22 23,19 Z" fill="${color}" />
    <circle cx="27" cy="18" r="1.2" fill="white" />

    <!-- Mask connecting line -->
    <path d="M17,19 Q20,25 23,19" fill="none" stroke="${color}" stroke-width="1.5" />
  </svg>`;
}
