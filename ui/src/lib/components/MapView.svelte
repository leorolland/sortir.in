<script lang="ts">
  import MapLibre from 'svelte-maplibre/MapLibre.svelte';
  import GeoJSON from 'svelte-maplibre/GeoJSON.svelte';
  import CircleLayer from 'svelte-maplibre/CircleLayer.svelte';
  import SymbolLayer from 'svelte-maplibre/SymbolLayer.svelte';
  import Popup from 'svelte-maplibre/Popup.svelte';
  import { pinsStore, type EventWithCoordinates } from '$lib/stores/pins';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import { GeolocateControl } from 'maplibre-gl';
  import MapSidebar from '$lib/components/MapSidebar.svelte';
  import { eventsToGeoJSON } from '$lib/utils/geoJsonUtils';
  import EventPopup from '$lib/components/EventPopup.svelte';
  import DateRangeSelector from '$lib/components/DateRangeSelector.svelte';
  // @ts-ignore
  import type { Feature, Geometry } from 'geojson';

  // Subscribe to pins store
  const pins = $derived($pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);
  let sidebarCollapsed = $state(false);
  let geoJsonData = $state(eventsToGeoJSON([]));

  // Function to update pins based on current map bounds
  async function updatePins() {
    if (!map) return;
    const events = await pinsStore.loadPins(map.getBounds());
    geoJsonData = eventsToGeoJSON(events);
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

  // Update GeoJSON data when pins change
  $effect(() => {
    geoJsonData = eventsToGeoJSON(pins);
  });
</script>

<div class="map-container">
  <MapSidebar
    {map}
    {pins}
    bind:collapsed={sidebarCollapsed}
  />

  <DateRangeSelector />


  <MapLibre
    center={[-1.6794, 48.1147]}
    zoom={12}
    class="map"
    style="https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json"
    bind:map={map}
    zoomOnDoubleClick={true}
  >
    <GeoJSON
      id="events"
      data={geoJsonData}
      cluster={{
        radius: 40,
        maxZoom: 14
      }}
    >
      <CircleLayer
        id="cluster_circles"
        applyToClusters
        cursor="pointer"
        paint={{
          'circle-color': '#2196f3',
          'circle-radius': [
            'step',
            ['get', 'point_count'],
            20,  // Size for small clusters
            20,  // Threshold
            30,  // Size for medium clusters
            50,  // Threshold
            40   // Size for large clusters
          ] as any,
        }}
      >
      </CircleLayer>

      <SymbolLayer
        id="cluster_labels"
        interactive={false}
        applyToClusters
        layout={{
          'text-field': ['get', 'point_count_abbreviated'],
          'text-size': 12,
          'text-offset': [0, 0.1],
          'text-font': ['Open Sans Bold']
        }}
        paint={{
          'text-color': '#ffffff'
        }}
      />

      <CircleLayer
        id="events_circle"
        applyToClusters={false}
        hoverCursor="pointer"
        paint={{
          'circle-color': [
            'match',
            ['get', 'kind'],
            'movie', 'rgb(94, 37, 207)',
            '#2196f3'  // default color
          ] as any,
          'circle-radius': 8,
        }}
      >
        <Popup openOn="click">
          {#snippet children({ data }: { data: Feature<Geometry, EventWithCoordinates> | undefined })}
            <EventPopup feature={data ?? undefined} />
          {/snippet}
        </Popup>
      </CircleLayer>
    </GeoJSON>
  </MapLibre>

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

  /* Style des popups */
  :global(.maplibregl-popup-content) {
    background: transparent !important;
    padding: 0 !important;
    box-shadow: none !important;
    border-radius: 0 !important;
  }

  :global(.maplibregl-popup-tip) {
    border-bottom-color: rgb(240, 240, 245) !important;
    border-top-color: transparent !important;
  }

  :global(.maplibregl-popup-anchor-bottom .maplibregl-popup-tip) {
    border-top-color: rgb(240, 240, 245) !important;
    border-bottom-color: transparent !important;
  }

  :global(.maplibregl-popup-anchor-left .maplibregl-popup-tip) {
    border-right-color: rgb(240, 240, 245) !important;
    border-left-color: transparent !important;
  }

  :global(.maplibregl-popup-anchor-right .maplibregl-popup-tip) {
    border-left-color: rgb(240, 240, 245) !important;
    border-right-color: transparent !important;
  }
</style>
