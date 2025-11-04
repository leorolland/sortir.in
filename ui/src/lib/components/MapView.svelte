<script lang="ts">
  import { MapLibre } from 'svelte-maplibre';
  import { pinsStore } from '$lib/stores/pins';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import { GeolocateControl, Marker } from 'maplibre-gl';
  import { createEventMarker } from '$lib/components/EventMarker';
  import MapSidebar from '$lib/components/MapSidebar.svelte';

  // Subscribe to pins store
  const pins = $derived($pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);
  let sidebarCollapsed = $state(false);

  // Function to update pins based on current map bounds
  async function updatePins() {
    if (!map) return;
    await pinsStore.loadPins(map.getBounds());
  }

  // Effect to add map event listeners when map is available
  $effect(() => {
    if (!map) return;

    map.on('moveend', updatePins);

    updatePins();

    // Add geolocate control to the map
    map.addControl(
      new GeolocateControl({
        positionOptions: {
          enableHighAccuracy: true
        },
        trackUserLocation: true,
        showAccuracyCircle: true,
        showUserLocation: true
      }),
      'top-right'
    );

    return () => {
      map?.off('moveend', updatePins);
    };
  });

  interface MarkerWithId {
    marker: Marker;
    eventId: string;
  }
  let markerMap: MarkerWithId[] = [];

  // Effect to efficiently update markers when map is available or pins change
  $effect(() => {
    if (!map) return;

    const refreshedPinIds = new Set(pins.map(pin => pin.id));
    const markersToKeep: MarkerWithId[] = [];

    // Remove markers for pins that are no longer in the list
    markerMap.forEach(({ marker, eventId }) => {
      if (!refreshedPinIds.has(eventId)) {
        marker.remove();
      } else {
        markersToKeep.push({ marker, eventId });
      }
    });

    // Add markers for pins that are not already in the list
    const existingMarkerIds = new Set(markersToKeep.map(m => m.eventId));
    pins.forEach(event => {
      if (!existingMarkerIds.has(event.id)) {
        const marker = createEventMarker(event, map!, () => {}).addTo(map!);
        markersToKeep.push({ marker, eventId: event.id });
      }
    });

    markerMap = markersToKeep;
  });
</script>

<div class="map-container">
  <MapSidebar
    {map}
    {pins}
    bind:collapsed={sidebarCollapsed}
  />

  <MapLibre
    center={[-1.6794, 48.1147]}
    zoom={12}
    class="map"
    style="https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json"
    bind:map={map}
  />
</div>

<style>
  .map-container {
    width: 100%;
    height: 100%;
    position: relative;
  }

  :global(.map) {
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
  }

  :global(.maplibregl-ctrl-top-right) {
    top: 10px;
    right: 10px;
  }
</style>
