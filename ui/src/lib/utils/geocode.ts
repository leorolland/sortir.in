const PHOTON_REVERSE_URL = 'https://photon.komoot.io/reverse';

const CACHE_TTL_MS = 24 * 60 * 60 * 1000;
// coordinates are rounded to 2 decimals to share cache entries, ~1km grid
const CACHE_PRECISION = 2;

export type Place = {
  city?: string;
  town?: string;
  village?: string;
  municipality?: string;
  locality?: string;
  hamlet?: string;
  suburb?: string;
  borough?: string;
  district?: string;
  county?: string;
  state?: string;
  country?: string;
};

// city/zone level keys, checked in order, the first non empty one is used
const cityTierProperties = [
  'city', 'town', 'village', 'municipality', 'locality', 'hamlet',
  'suburb', 'borough', 'district'
] as const;

const placeProperties = [...cityTierProperties, 'county', 'state', 'country'] as const;

type PhotonResponse = {
  features: {
    properties?: Record<string, unknown>;
  }[];
};

const cache = new Map<string, CacheEntry>();

type CacheEntry = {
  place: Place;
  expiresAt: number;
};

function cacheKey(lat: number, lon: number): string {
  return `${lat.toFixed(CACHE_PRECISION)}:${lon.toFixed(CACHE_PRECISION)}`;
}

/**
 * Resolves coordinates into the place hierarchy (city/zone, department, country)
 * using the public Photon reverse geocoding API (OpenStreetMap data, fair-use policy).
 * Returns an empty object when no place could be found or on failure.
 */
export async function reverseGeocode(lat: number, lon: number): Promise<Place> {
  const key = cacheKey(lat, lon);

  const cached = cache.get(key);
  if (cached && cached.expiresAt > Date.now()) {
    return cached.place;
  }

  try {
    const url = new URL(PHOTON_REVERSE_URL);
    url.searchParams.append('lat', lat.toString());
    url.searchParams.append('lon', lon.toString());
    url.searchParams.append('lang', 'fr');
    url.searchParams.append('radius', '50');

    const response = await fetch(url.toString());
    if (!response.ok) {
      throw new Error(`Reverse geocoding failed: ${response.statusText}`);
    }

    const result: PhotonResponse = await response.json();
    const place = extractPlace(result);

    cache.set(key, { place, expiresAt: Date.now() + CACHE_TTL_MS });
    return place;
  } catch (error) {
    console.error('Error during reverse geocoding:', error);
    return {};
  }
}

function extractPlace(result: PhotonResponse): Place {
  const properties = result.features?.[0]?.properties;
  if (!properties) {
    return {};
  }

  const place: Place = {};
  for (const key of placeProperties) {
    const value = properties[key];
    if (typeof value === 'string' && value !== '') {
      place[key] = value;
    }
  }

  return place;
}

/**
 * Returns the display phrase for the place matching the given zoom level:
 * zoomed in on a city or zone ("autour de Rennes"), department scale
 * ("en Ille-et-Vilaine"), or country when zoomed way out ("en France").
 * Returns an empty string when no place is available at that level.
 */
export function placeDisplayPhrase(place: Place, zoom: number): string {
  // country scale
  if (zoom < 8) {
    return place.country ? `en ${place.country}` : '';
  }

  // department scale, falls back to the region
  if (zoom < 11) {
    const department = place.county ?? place.state;
    return department ? `en ${department}` : '';
  }

  // city/zone scale
  for (const key of cityTierProperties) {
    const city = place[key];
    if (city) {
      return /^[aeiouyéèêàâîôû]/i.test(city) ? `autour d'${city}` : `autour de ${city}`;
    }
  }

  return '';
}
