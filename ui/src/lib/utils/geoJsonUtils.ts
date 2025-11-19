import type { EventWithCoordinates } from '$lib/stores/pins';
// @ts-ignore
import type { Feature, FeatureCollection } from 'geojson';

/**
 * Converts an array of EventWithCoordinates to a GeoJSON FeatureCollection
 */
export function eventsToGeoJSON(events: EventWithCoordinates[]): FeatureCollection {
  const features: Feature[] = events.map(event => {
    const coordinates = event.getCoordinates();

    return {
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: coordinates
      },
      properties: {
        ...event
      }
    };
  });

  return {
    type: 'FeatureCollection',
    features
  };
}
