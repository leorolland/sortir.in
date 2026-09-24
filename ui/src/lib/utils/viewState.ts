import { DateRange } from './dateUtils';

/**
 * Map view state mirrored into the URL query string so a copied URL reopens
 * the same visual location and popup (ADR 0005):
 *
 * - `lat`, `lng`: map center (6 decimals, ~0.1 m)
 * - `z`: zoom (2 decimals)
 * - `range`: DateRange value, omitted when it is the default (`today`)
 * - `pin=<lat>,<lon>`: focused pin location, full float precision — the
 *   events filter compares `loc.lat`/`loc.lon` exactly, so no rounding
 *
 * Unknown query parameters (e.g. OAuth `code=`) are preserved.
 */
export type MapViewState = {
  center: { lat: number; lon: number };
  zoom: number;
  range: DateRange | null;
  pin: { lat: number; lon: number } | null;
};

const ZOOM_MIN = 1;
const ZOOM_MAX = 22;

/**
 * Set once the SvelteKit router has completed its initial navigation:
 * `pushState`/`replaceState` from `$app/navigation` throw before that
 * (dev-mode guard), and map mount effects can run earlier. Module-level
 * so a vite hot swap of the map component doesn't lose it.
 */
export const routerState = { ready: false };

function parseCoordinate(value: string | null, min: number, max: number): number | undefined {
  if (value === null || value.trim() === '') return undefined;
  const n = Number(value);
  if (!Number.isFinite(n) || n < min || n > max) return undefined;
  return n;
}

function parsePin(value: string | null): { lat: number; lon: number } | null {
  if (!value) return null;
  const [lat, lon] = value.split(',').map((part) => Number(part));
  if (!Number.isFinite(lat) || !Number.isFinite(lon)) return null;
  if (lat < -90 || lat > 90 || lon < -180 || lon > 180) return null;
  return { lat, lon };
}

/**
 * Reads the view state from a query string. Invalid or missing values are
 * omitted, leaving the caller's defaults in place.
 */
export function parseViewState(search: string): Partial<MapViewState> {
  const params = new URLSearchParams(search);
  const view: Partial<MapViewState> = {};

  const lat = parseCoordinate(params.get('lat'), -90, 90);
  const lng = parseCoordinate(params.get('lng'), -180, 180);
  if (lat !== undefined && lng !== undefined) view.center = { lat, lon: lng };

  const zoom = parseCoordinate(params.get('z'), ZOOM_MIN, ZOOM_MAX);
  if (zoom !== undefined) view.zoom = zoom;

  const range = params.get('range');
  if (range && Object.values(DateRange).includes(range as DateRange)) {
    view.range = range as DateRange;
  }

  const pin = parsePin(params.get('pin'));
  if (pin) view.pin = pin;

  return view;
}

function roundTo(value: number, decimals: number): number {
  const factor = 10 ** decimals;
  return Math.round(value * factor) / factor;
}

/**
 * Updates `params` in place with the given view state, deleting keys that
 * carry no information (default range, no focused pin). Foreign parameters
 * are left untouched.
 */
export function writeViewState(params: URLSearchParams, view: MapViewState): void {
  params.set('lat', String(roundTo(view.center.lat, 6)));
  params.set('lng', String(roundTo(view.center.lon, 6)));
  params.set('z', String(roundTo(view.zoom, 2)));

  if (view.range && view.range !== DateRange.TODAY) {
    params.set('range', view.range);
  } else {
    params.delete('range');
  }

  if (view.pin) {
    params.set('pin', `${view.pin.lat},${view.pin.lon}`);
  } else {
    params.delete('pin');
  }
}
