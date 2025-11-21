/**
 * SVG pour le pin concert avec un symbole musical simple
 */
export function createConcertPinSVG(color: string): string {
  return `<svg width="40" height="50" viewBox="0 0 40 50" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <filter id="shadow-concert" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="2" stdDeviation="2" flood-color="#000000" flood-opacity="0.3" />
      </filter>
    </defs>

    <!-- Main pin shape -->
    <path d="M20 0C9 0 0 9 0 20C0 31 20 50 20 50C20 50 40 31 40 20C40 9 31 0 20 0Z"
      fill="${color}" filter="url(#shadow-concert)" />

    <!-- Background -->
    <circle cx="20" cy="20" r="14" fill="white" opacity="0.9" />

    <!-- Symbole musical simple et grand -->
    <text x="20" y="26" font-family="Arial" font-size="24" font-weight="bold" text-anchor="middle" fill="${color}">♫</text>
  </svg>`;
}
