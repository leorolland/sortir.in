import type { EventWithCoordinates } from '$lib/stores/pins';
import { getRelativeTimeDisplay } from '$lib/utils/dateUtils';
import { Marker, Popup, Map as MaplibreMap } from 'maplibre-gl';
import './EventMarker.css';

/**
 * Extracts the domain name from a URL
 * @param url The full URL
 * @returns The domain name or null if invalid URL
 */
function extractDomain(url: string | null | undefined): string | null {
  if (!url) return null;
  try {
    const domain = new URL(url).hostname;
    return domain;
  } catch (e) {
    console.error('Invalid URL:', url);
    return null;
  }
}

/**
 * Creates a custom event marker with popup
 * @param event The event data
 * @param map Map instance for fly animation
 * @param onClick Click handler for the marker
 * @returns The created marker instance
 */
export function createEventMarker(
  event: EventWithCoordinates,
  map: MaplibreMap,
  onClick: (eventId: string) => void
): Marker {
  const timeStatus = getRelativeTimeDisplay(event.begin, event.end);
  const statusClass = timeStatus === 'En cours' ? 'status-ongoing' :
                     timeStatus === 'Terminé' ? 'status-ended' : 'status-upcoming';

  const popup = new Popup({
    offset: [0, -10],
    className: 'custom-popup',
    closeButton: false
  })
    .setHTML(`
      <div class="popup-content">
        <div class="popup-title">${event.name}</div>
        <div class="popup-date ${statusClass}">${timeStatus}</div>
        ${event.place ? `<div class="popup-place">${event.place}</div>` : ''}
        ${event.address ? `<div class="popup-address">${event.address}</div>` : ''}
        ${event.kind && event.kind.toLowerCase() !== "event" ? `<div class="popup-kind">${event.kind}</div>` : ''}
        ${Array.isArray(event.genres) && event.genres.length > 0 ? `<div class="popup-genres">
          ${event.genres.map((genre: string) => `<span class="genre-tag">${genre}</span>`).join('')}
        </div>` : ''}
        ${event.price !== undefined && event.price_currency ? `<div class="popup-price">${event.price} ${event.price_currency}</div>` : ''}
        ${event.img ? `<div class="popup-image"><img decoding="async" loading="lazy" src="${event.img}" alt="${event.name}" /></div>` : ''}
        <div class="popup-actions">
          <a href="${event.source}" target="_blank" rel="noopener noreferrer" class="popup-button">
            ${extractDomain(event.source) || 'Détails'}
            <svg class="external-link-icon" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"></path>
              <path d="M15 3h6v6"></path>
              <path d="M10 14L21 3"></path>
            </svg>
          </a>
        </div>
      </div>
    `);

  const el = document.createElement('div');
  el.className = 'custom-marker';

  if (timeStatus === 'En cours') {
    el.classList.add('marker-ongoing');
  } else if (timeStatus === 'Terminé') {
    el.classList.add('marker-ended');
  } else {
    el.classList.add('marker-upcoming');
  }

  el.addEventListener('click', () => {
    // Call the onClick handler
    onClick(event.id);

    // Fly to the marker location
    const coordinates = event.getCoordinates();
    map.flyTo({
      center: coordinates,
      speed: 1.2,
      curve: 1.4,
      essential: true
    });
  });

  return new Marker({ element: el, draggable: false })
    .setLngLat(event.getCoordinates())
    .setPopup(popup);
}
