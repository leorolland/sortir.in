import { writable } from 'svelte/store';
import { client } from '$lib/pocketbase';
import type { EventsResponse, GeoPoint } from '$lib/pocketbase/generated-types';
import { formatDateForFilter, type DateWindow } from '$lib/utils/dateUtils';

function createEventsStore() {
  const { subscribe: subscribeEventsForLocation, set: setEventsForLocation } = writable<EventsResponse[]>([]);
  const { subscribe: subscribeEventsForBounds, set: setEventsForBounds } = writable<EventsResponse[]>([]);

  return {
    subscribeEventsForLocation: subscribeEventsForLocation,
    subscribeEventsForBounds: subscribeEventsForBounds,

    loadEventsForLocation: async (location: GeoPoint, { min, max }: DateWindow) => {
      try {
        let filter = `(begin<='${formatDateForFilter(max)}'&&end>='${formatDateForFilter(min)}'&&loc.lat=${location.lat}&&loc.lon=${location.lon})`;

        const eventsResult = await client.collection('events').getList<EventsResponse>(
          1,
          100,
          { filter, skipTotal: true }
        );

        setEventsForLocation(eventsResult.items);
        return eventsResult.items;
      } catch (error) {
        console.error('Error loading events for location:', error);
        setEventsForLocation([]);
        return [];
      }
    },

    getEventsInBounds: async (bounds: {
      getNorth: () => number;
      getSouth: () => number;
      getEast: () => number;
      getWest: () => number;
    }, { min, max }: DateWindow, page = 1, perPage = 100) => {
      try {
        const north = bounds.getNorth();
        const south = bounds.getSouth();
        const east = bounds.getEast();
        const west = bounds.getWest();
        const filter = `(begin<='${formatDateForFilter(max)}'&&end>='${formatDateForFilter(min)}'&&loc.lat>${south}&&loc.lat<${north}&&loc.lon<${east}&&loc.lon>${west})`;

        const eventsResult = await client.collection('events').getList<EventsResponse>(
          page,
          perPage,
          { filter, skipTotal: true }
        );

        setEventsForBounds(eventsResult.items);
        return eventsResult.items;
      } catch (error) {
        console.error('Error in getEventsInBounds:', error);
        setEventsForBounds([]);
        return [];
      }
    },
  };
}

export const eventsStore = createEventsStore();
