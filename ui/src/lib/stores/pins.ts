import { writable } from 'svelte/store';
import type { DateWindow } from '$lib/utils/dateUtils';

export type Pin = {
  loc: {
    lat: number;
    lon: number;
  };
  kind: string;
  amount: number;
}

export interface MapBounds {
  getNorth(): number;
  getSouth(): number;
  getEast(): number;
  getWest(): number;
}

// Eager loading: fetch pins for a larger area than displayed so that pins
// are already on screen when a pan finishes, instead of spawning. A factor
// of 3 spans one extra screen on each side (9x the visible area)
const EAGER_LOAD_FACTOR = 3;

function expandBounds(bounds: MapBounds, factor: number): MapBounds {
  const centerLat = (bounds.getNorth() + bounds.getSouth()) / 2;
  const centerLng = (bounds.getEast() + bounds.getWest()) / 2;
  const latSpan = (bounds.getNorth() - bounds.getSouth()) * factor;
  const lngSpan = (bounds.getEast() - bounds.getWest()) * factor;

  // Clamp to valid coordinate ranges: an expanded view can overshoot the
  // poles or the antimeridian, and the API expects bounded coordinates
  return {
    getNorth: () => Math.min(90, centerLat + latSpan / 2),
    getSouth: () => Math.max(-90, centerLat - latSpan / 2),
    getEast: () => Math.min(180, centerLng + lngSpan / 2),
    getWest: () => Math.max(-180, centerLng - lngSpan / 2)
  };
}

function createPinsStore() {
  const { subscribe, set } = writable<Pin[]>([]);

  return {
    subscribe,
    loadPins: async (bounds: MapBounds, { min, max }: DateWindow) => {
      const eagerBounds = expandBounds(bounds, EAGER_LOAD_FACTOR);
      try {
        const url = new URL('/api/pins', window.location.origin);
        url.searchParams.append('north', eagerBounds.getNorth().toString());
        url.searchParams.append('south', eagerBounds.getSouth().toString());
        url.searchParams.append('east', eagerBounds.getEast().toString());
        url.searchParams.append('west', eagerBounds.getWest().toString());
        url.searchParams.append('min_time', min.toISOString());
        url.searchParams.append('max_time', max.toISOString());

        const response = await fetch(url.toString());
        if (!response.ok) {
          throw new Error(`Failed to fetch pins: ${response.statusText}`);
        }

        const pins = await response.json();

        set(pins);
        return pins;
      } catch (error) {
        console.error('Error loading pins:', error);
        return [];
      }
    },
    reset: () => set([])
  };
}

export const pinsStore = createPinsStore();
