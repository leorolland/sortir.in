<script lang="ts">
  import MapLibre from 'svelte-maplibre/MapLibre.svelte';
  import GeoJSON from 'svelte-maplibre/GeoJSON.svelte';
  import CircleLayer from 'svelte-maplibre/CircleLayer.svelte';
  import SymbolLayer from 'svelte-maplibre/SymbolLayer.svelte';
  import Popup from 'svelte-maplibre/Popup.svelte';
  import { pinsStore, type Pin } from '$lib/stores/pins';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import { GeolocateControl } from 'maplibre-gl';
  import MapSidebar from '$lib/components/MapSidebar.svelte';
  import { pinsToGeoJSON } from '$lib/utils/geoJsonUtils';
  import EventPopup from '$lib/components/EventPopup.svelte';
  import DateRangeSelector from '$lib/components/DateRangeSelector.svelte';
  import AppVersion from '$lib/components/AppVersion.svelte';
  // @ts-ignore
  import type { Feature, Geometry } from 'geojson';
  import { pinSVGs } from '$lib/components/pins/svg';
  import { writable } from 'svelte/store';
  import { DateRange, getDateWindow } from '$lib/utils/dateUtils';
  import { eventsStore } from '$lib/stores/events';
  import { placeDisplayPhrase, reverseGeocode, type Place } from '$lib/utils/geocode';
  import { metadata } from '$lib/metadata.js';

  const pins = $derived($pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);
  let sidebarCollapsed = $state<boolean>(window.innerWidth < 768);
  const initialZoom = window.innerWidth < 768 ? 4.5 : 5.5;
  let geoJsonData = $state(pinsToGeoJSON([]));
  let initialized = $state(false);
  let place = $state<Place>({});
  let zoom = $state<number>(0);

  const eventsOnScreen = $derived(pins.reduce((sum, pin) => sum + pin.amount, 0));

  let geocodeTimer: ReturnType<typeof setTimeout> | undefined;

  function updatePlace() {
    if (!map) return;

    zoom = map.getZoom();

    clearTimeout(geocodeTimer);
    geocodeTimer = setTimeout(async () => {
      const center = map?.getCenter();
      if (!center) return;

      place = await reverseGeocode(center.lat, center.lng);
    }, 500);
  }

  // Create a store for the selected date range
  export const selectedDateRange = writable<DateRange>(DateRange.TODAY);

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

    const window = getDateWindow($selectedDateRange);
    const pins = await pinsStore.loadPins(map.getBounds(), window);
    eventsStore.getEventsInBounds(map.getBounds(), window);

    geoJsonData = pinsToGeoJSON(pins);

    updatePlace();
  }


  $effect(() => {
    if (!map) return;

    map.on('moveend', updatePins);
    map.on('load', loadPinImages);

    const handleMapClick = (e: any) => {
      if (!map) return;

      const features = map.queryRenderedFeatures(e.point, { layers: ['event_points'] });

      if (features.length > 0) {
        const feature = features[0];
        if (feature.geometry && feature.geometry.type === 'Point') {
          const coordinates = feature.geometry.coordinates.slice();
          const amount = feature.properties?.amount;

          const screenFactor = 0.0011*window.outerHeight; // factor to adjust the offsetY to the screen height
          const offsetY = (-100*screenFactor) - (Math.min(3, amount/5)*80*screenFactor);

          map.flyTo({
            center: coordinates as [number, number],
            duration: 300,
            freezeElevation: true,
            offset: [0, offsetY],
            padding: window.outerHeight,
          });
        }
      }
    };

    map.on('click', handleMapClick);

    if (!initialized) {
      loadPinImages();
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
      initialized = true;
    }

    updatePins();

    return () => {
      map?.off('moveend', updatePins);
      map?.off('load', loadPinImages);
      map?.off('click', handleMapClick);
      clearTimeout(geocodeTimer);
    };
  });

  $effect(() => {
    const location = placeDisplayPhrase(place, zoom) || 'autour de vous';
    const title = `${eventsOnScreen} ${eventsOnScreen === 1 ? 'sortie' : 'sorties'} ${location}`;
    metadata.update((m) => ({ ...m, title }));
  });

  let prevPinsLength = 0;
  let prevPinsString = '';

  $effect(() => {
    const currentPinsString = JSON.stringify(pins.map((p: any) => p.id));

    if (pins.length !== prevPinsLength || currentPinsString !== prevPinsString) {
      prevPinsLength = pins.length;
      prevPinsString = currentPinsString;

      geoJsonData = pinsToGeoJSON(pins);
    }
  });
</script>

<div class="map-container">
  <MapSidebar
    map={map}
    pins={pins}
    collapsed={sidebarCollapsed}
    events={eventsStore.subscribeEventsForBounds}
    on:collapsedChange={(e) => sidebarCollapsed = e.detail}
  />

  <DateRangeSelector selectedDateRange={selectedDateRange} />

  <AppVersion />

  <MapLibre
    center={[2.4, 46.6]}
    zoom={initialZoom}
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
        <Popup openOn="click" closeButton={true}>
          {#snippet children({ data }: { data: Feature<Geometry, Pin> | undefined })}
            <EventPopup feature={data ?? undefined} dateRange={$selectedDateRange} />
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

  /* MapLibre's built-in popup close button: only used on touch devices,
     where the full-screen popup has no "click outside" to close. */
  :global(.maplibregl-popup-close-button) {
    display: none;
  }

  /* On touch devices, the pin popup stops floating over the map and becomes
     a full-screen page: the popup element is stretched over the viewport and
     the frosted-glass panel fills it, the blurred map staying behind. */
  @media (hover: none) and (pointer: coarse) {
    :global(.maplibregl-popup) {
      position: fixed !important;
      inset: 0 !important;
      transform: none !important;
      max-width: none !important;
      z-index: 800 !important; /* above map controls, below global alerts */
      animation: popup-fullscreen-in 250ms ease-out;
    }

    :global(.maplibregl-popup-content) {
      width: 100% !important;
      height: 100% !important;
      max-width: none !important;
      padding: 0 !important;
      display: flex !important;
      flex-direction: column;
      overflow: hidden !important;
    }

    :global(.maplibregl-popup .sv-popup) {
      flex: 1;
      min-height: 0;
      display: flex;
      flex-direction: column;
    }

    :global(.maplibregl-popup-content .floating-panel) {
      flex: 1;
      min-height: 0;
      border-radius: 0;
    }

    :global(.maplibregl-popup-tip) {
      display: none !important;
    }

    :global(.maplibregl-popup-close-button) {
      display: flex !important;
      align-items: center;
      justify-content: center;
      top: calc(10px + env(safe-area-inset-top)) !important;
      right: calc(10px + env(safe-area-inset-right)) !important;
      width: 36px;
      height: 36px;
      padding: 0;
      border-radius: 50%;
      background-color: rgba(255, 255, 255, 0.95);
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.18);
      font-size: 22px;
      line-height: 1;
      color: #333;
      pointer-events: auto; /* container has pointer-events: none */
    }
  }

  @keyframes popup-fullscreen-in {
    from {
      translate: 0 32px;
      opacity: 0;
    }
    to {
      translate: 0 0;
      opacity: 1;
    }
  }

  @media (hover: none) and (pointer: coarse) and (prefers-reduced-motion: reduce) {
    :global(.maplibregl-popup) {
      animation: none;
    }
  }
</style>
