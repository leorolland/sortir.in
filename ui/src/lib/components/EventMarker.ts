import type { EventWithCoordinates } from '$lib/stores/pins';
import { getRelativeTimeDisplay } from '$lib/utils/dateUtils';
import { Marker, Popup, Map as MaplibreMap } from 'maplibre-gl';
import './EventMarker.css';

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
