import { writable } from 'svelte/store';
import { client } from '$lib/pocketbase';
import type { EventsResponse } from '$lib/pocketbase/generated-types';

// Helper type for map display that extends EventsResponse with a convenience function
export interface EventWithCoordinates extends EventsResponse {
  getCoordinates(): [number, number];
}

export interface MapBounds {
  getNorth(): number;
  getSouth(): number;
  getEast(): number;
  getWest(): number;
}

// Date range enum for filtering events
export enum DateRange {
  TODAY = 'today',
  TOMORROW = 'tomorrow',
  THIS_WEEK = 'this_week'
}

function createPinsStore() {
  const { subscribe, set } = writable<EventWithCoordinates[]>([]);

  // Store the current date range and bounds
  let currentDateRange = DateRange.TODAY;
  let currentBounds: MapBounds | null = null;

  // Helper function to convert EventsResponse to EventWithCoordinates
  const enhanceEvent = (event: EventsResponse): EventWithCoordinates => {
    return {
      ...event,
      getCoordinates: () => [event.loc?.lon || 0, event.loc?.lat || 0] as [number, number]
    };
  };

  const getMaxDateForRange = (range: DateRange): Date => {
    const now = new Date();
    const maxDate = new Date();
    const currentHour = now.getHours();
    console.log('currentHour', currentHour);

    switch (range) {
      case DateRange.TODAY:
        // If it's between 16h and 4h, it's "Ce soir" and we set max to 6am
        if (currentHour >= 16 || currentHour < 4) {
          if (currentHour >= 16) {
            // Between 16h and midnight - set to 6am next day
            maxDate.setDate(maxDate.getDate() + 1);
            maxDate.setHours(6, 0, 0, 0);
          } else {
            // Between midnight and 4h - set to 6am same day
            maxDate.setHours(6, 0, 0, 0);
          }
        } else {
          // Regular "Aujourd'hui" - end of today
          maxDate.setHours(23, 59, 59, 999);
        }
        break;
      case DateRange.TOMORROW:
        // End of tomorrow
        maxDate.setDate(maxDate.getDate() + 1);
        maxDate.setHours(23, 59, 59, 999);
        break;
      case DateRange.THIS_WEEK:
        // End of the week (next Sunday)
        const daysToSunday = 7 - now.getDay();
        maxDate.setDate(maxDate.getDate() + daysToSunday);
        maxDate.setHours(23, 59, 59, 999);
        break;
    }

    return maxDate;
  };

  return {
    subscribe,
    loadPins: async (bounds: MapBounds) => {
      currentBounds = bounds;
      const page = 1; // Fixed page value
      const perPage = 800; // Fixed perPage value
      try {
        // Get the max date based on the current date range
        const maxBeginDate = getMaxDateForRange(currentDateRange);

        const filter = `(begin<='${maxBeginDate.toISOString()}'&&loc.lat>${bounds.getSouth()}&&loc.lat<${bounds.getNorth()}&&loc.lon<${bounds.getEast()}&&loc.lon>${bounds.getWest()})`;

        const eventsResult = await client.collection('events').getList<EventsResponse>(
          page,
          perPage,
          { filter, skipTotal: true }
        );

        // Enhance events with coordinates helper
        const enhancedEvents = eventsResult.items.map(enhanceEvent);

        set(enhancedEvents);
        return enhancedEvents;
      } catch (error) {
        console.error('Error loading pins:', error);
        return [];
      }
    },
    updateDateRange: async (range: DateRange) => {
      currentDateRange = range;

      // If we have bounds (map is loaded), reload pins
      if (currentBounds) {
        return await pinsStore.loadPins(currentBounds);
      }
    },
    reset: () => set([])
  };
}

export const pinsStore = createPinsStore();
