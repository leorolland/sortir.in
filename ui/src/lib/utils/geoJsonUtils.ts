import type { Pin } from '$lib/stores/pins';
// @ts-ignore
import type { Feature, FeatureCollection, Geometry } from 'geojson';

/**
 * Converts an array of Pin to a GeoJSON FeatureCollection
 */
export function pinsToGeoJSON(pins: Pin[]): FeatureCollection {
  const features: Feature[] = pins.map(pin => {
    return {
      type: 'Feature',
      geometry: {
        type: 'Point',
        coordinates: [pin.loc.lon, pin.loc.lat]
      },
      properties: pin
    } as Feature<Geometry, Pin>;
  });

  return {
    type: 'FeatureCollection',
    features
  };
}
