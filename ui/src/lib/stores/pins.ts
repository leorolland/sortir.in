import { writable } from 'svelte/store';
import { client } from '$lib/pocketbase';
import type { EventsResponse } from '$lib/pocketbase/generated-types';

// Helper type for map display that extends EventsResponse with a convenience function
export interface EventWithCoordinates extends EventsResponse {
  getCoordinates(): [number, number];
}

function createPinsStore() {
  const { subscribe, set } = writable<EventWithCoordinates[]>([]);

  // Helper function to convert EventsResponse to EventWithCoordinates
  const enhanceEvent = (event: EventsResponse): EventWithCoordinates => {
    return {
      ...event,
      getCoordinates: () => [event.loc?.lon || 0, event.loc?.lat || 0] as [number, number]
    };
  };

  return {
    subscribe,
    loadPins: async (page = 1, perPage = 50) => {
      try {
        // Query events from PocketBase
        const eventsResult = await client.collection('events').getList<EventsResponse>(page, perPage);

        // Enhance events with coordinates helper
        const enhancedEvents = eventsResult.items.map(enhanceEvent);

        set(enhancedEvents);
        return enhancedEvents;
      } catch (error) {
        console.error('Error loading pins:', error);
        return [];
      }
    },
    reset: () => set([])
  };
}

export const pinsStore = createPinsStore();
