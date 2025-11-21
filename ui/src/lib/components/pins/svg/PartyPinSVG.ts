/**
 * SVG pour le pin soirée avec des lumières/disco
 */
export function createPartyPinSVG(color: string): string {
  return `<svg width="40" height="50" viewBox="0 0 40 50" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <filter id="shadow-party" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="2" stdDeviation="2" flood-color="#000000" flood-opacity="0.3" />
      </filter>
    </defs>

    <!-- Main pin shape -->
    <path d="M20 0C9 0 0 9 0 20C0 31 20 50 20 50C20 50 40 31 40 20C40 9 31 0 20 0Z"
      fill="${color}" filter="url(#shadow-party)" />

    <!-- Party icon background -->
    <circle cx="20" cy="20" r="14" fill="white" opacity="0.9" />

    <!-- Disco ball -->
    <circle cx="20" cy="20" r="10" fill="${color}" />

    <!-- Light reflections -->
    <path d="M20,10 L20,14" stroke="white" stroke-width="1.5" />
    <path d="M20,26 L20,30" stroke="white" stroke-width="1.5" />
    <path d="M10,20 L14,20" stroke="white" stroke-width="1.5" />
    <path d="M26,20 L30,20" stroke="white" stroke-width="1.5" />
    <path d="M13,13 L16,16" stroke="white" stroke-width="1.5" />
    <path d="M27,13 L24,16" stroke="white" stroke-width="1.5" />
    <path d="M13,27 L16,24" stroke="white" stroke-width="1.5" />
    <path d="M27,27 L24,24" stroke="white" stroke-width="1.5" />
  </svg>`;
}
