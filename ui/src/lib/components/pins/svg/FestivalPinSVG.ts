/**
 * SVG pour le pin festival avec des drapeaux/fanions
 */
export function createFestivalPinSVG(color: string): string {
  return `<svg width="40" height="50" viewBox="0 0 40 50" xmlns="http://www.w3.org/2000/svg">
    <defs>
      <filter id="shadow-festival" x="-20%" y="-20%" width="140%" height="140%">
        <feDropShadow dx="0" dy="2" stdDeviation="2" flood-color="#000000" flood-opacity="0.3" />
      </filter>
    </defs>

    <!-- Main pin shape -->
    <path d="M20 0C9 0 0 9 0 20C0 31 20 50 20 50C20 50 40 31 40 20C40 9 31 0 20 0Z"
      fill="${color}" filter="url(#shadow-festival)" />

    <!-- Festival flags background -->
    <circle cx="20" cy="20" r="14" fill="white" opacity="0.9" />

    <!-- Festival flags -->
    <path d="M12,10 L12,30 L14,30 L14,10 Z" fill="${color}" />
    <path d="M14,12 L24,15 L14,18 Z" fill="${color}" />
    <path d="M14,21 L24,24 L14,27 Z" fill="${color}" />

    <path d="M28,10 L28,30 L26,30 L26,10 Z" fill="${color}" />
    <path d="M26,12 L16,15 L26,18 Z" fill="${color}" />
    <path d="M26,21 L16,24 L26,27 Z" fill="${color}" />
  </svg>`;
}
