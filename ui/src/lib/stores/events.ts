import { writable } from 'svelte/store';
import { client } from '$lib/pocketbase';
import type { EventsResponse, GeoPoint } from '$lib/pocketbase/generated-types';

function createEventsStore() {
  const { subscribe, set } = writable<EventsResponse[]>([]);

  return {
    subscribe,

    loadEventsForLocation: async (location: GeoPoint, maxDate: Date) => {
      try {
        let filter = `(begin<='${maxDate.toISOString()}'&&loc.lat=${location.lat}&&loc.lon=${location.lon})`;

        const eventsResult = await client.collection('events').getList<EventsResponse>(
          1,
          100,
          { filter, skipTotal: true }
        );

        set(eventsResult.items);
        return eventsResult.items;
      } catch (error) {
        console.error('Error loading events for location:', error);
        set([]);
        return [];
      }
    },

    // Clear the events list
    clear: () => set([])
  };
}

export const eventsStore = createEventsStore();
