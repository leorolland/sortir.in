import { writable } from 'svelte/store';

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

function createPinsStore() {
  const { subscribe, set } = writable<Pin[]>([]);

  let currentBounds: MapBounds | null = null;

  return {
    subscribe,
    loadPins: async (bounds: MapBounds, maxBeginDate: Date) => {
      currentBounds = bounds;
      try {
        const url = new URL('/api/pins', window.location.origin);
        url.searchParams.append('north', bounds.getNorth().toString());
        url.searchParams.append('south', bounds.getSouth().toString());
        url.searchParams.append('east', bounds.getEast().toString());
        url.searchParams.append('west', bounds.getWest().toString());
        url.searchParams.append('max_time', maxBeginDate.toISOString());

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
