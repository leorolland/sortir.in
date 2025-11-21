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
  import { pinSVGs } from '$lib/components/pins/svg';

  const pins = $derived($pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);
  let sidebarCollapsed = $state<boolean>(window.innerWidth < 768);
  let geoJsonData = $state(eventsToGeoJSON([]));

  function loadPinImages() {
    if (!map) return;

    Object.entries(pinSVGs).forEach(([name, svg]) => {
      const img = new Image();
      img.onload = () => {
        if (map && !map.hasImage(`pin-${name}`)) {
          map.addImage(`pin-${name}`, img);
          console.log(`Image pin-${name} chargée`);
        }
      };

      const blob = new Blob([svg as string], { type: 'image/svg+xml' });
      const url = URL.createObjectURL(blob);
      img.src = url;
    });
  }

  async function updatePins() {
    if (!map) return;
    const events = await pinsStore.loadPins(map.getBounds());
    geoJsonData = eventsToGeoJSON(events);
  }

  $effect(() => {
    if (!map) return;

    map.on('moveend', updatePins);
    map.on('load', loadPinImages);

    updatePins();
    loadPinImages();

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
      map?.off('load', loadPinImages);
    };
  });

  let prevPinsLength = 0;
  let prevPinsString = '';

  $effect(() => {
    const currentPinsString = JSON.stringify(pins.map((p: any) => p.id));

    if (pins.length !== prevPinsLength || currentPinsString !== prevPinsString) {
      prevPinsLength = pins.length;
      prevPinsString = currentPinsString;

      geoJsonData = eventsToGeoJSON(pins);
    }
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
        radius: 100,
        maxZoom: 13
      }}
    >
      <CircleLayer
        id="cluster_circles"
        applyToClusters
        cursor="pointer"
        paint={{
          'circle-color': [
            'interpolate',
            ['linear'],
            ['get', 'point_count'],
            1, 'rgba(255, 240, 50, 0.95)',
            5, 'rgba(255, 150, 0, 0.95)',
            15, 'rgba(255, 0, 50, 0.95)',
            30, 'rgba(200, 0, 100, 0.95)',
            50, 'rgba(100, 0, 150, 0.95)'
          ],
          'circle-radius': [
            'interpolate',
            ['linear'],
            ['get', 'point_count'],
            1, 35,
            5, 45,
            15, 55,
            30, 65,
            50, 75
          ],
          'circle-blur': 1.5,
          'circle-opacity': 0.8,
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
          'text-color': '#222222'
        }}
      />

      <SymbolLayer
        id="event_points"
        applyToClusters={false}
        hoverCursor="pointer"
        layout={{
          'icon-image': [
            'match',
            ['get', 'kind'],
            'movie', 'pin-movie',
            'concert', 'pin-concert',
            'festival', 'pin-festival',
            'theater', 'pin-theater',
            'party', 'pin-party',
            'pin-default'
          ],
          'icon-size': 1.0,
          'icon-allow-overlap': true,
          'icon-anchor': 'bottom'
        }}
      >
        <Popup openOn="click">
          {#snippet children({ data }: { data: Feature<Geometry, EventWithCoordinates> | undefined })}
            <EventPopup feature={data ?? undefined} />
          {/snippet}
        </Popup>
      </SymbolLayer>
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

  :global(.maplibregl-marker) {
    background: none !important;
  }
</style>
