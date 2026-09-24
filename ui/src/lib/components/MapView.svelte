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
  import { pinSVGs, PIN_PIXEL_RATIO } from '$lib/components/pins/svg';
  import { writable } from 'svelte/store';
  import { DateRange, getDateWindow } from '$lib/utils/dateUtils';
  import { eventsStore } from '$lib/stores/events';
  import { sheetState } from '$lib/stores/sheet';
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

  // While a pin is open, any user move of the map (pan, pinch, wheel)
  // collapses the bottom sheet to its title bar ("peek"), keeping the pin
  // reachable: tapping the bar restores the sheet. MapLibre sets
  // originalEvent only on user-initiated moves, so programmatic moves
  // (flyTo on pin click, geolocation recentering) are ignored.
  function handleMoveStart(e: { originalEvent?: Event }) {
    if (!e.originalEvent) return;
    sheetState.set('peek');
  }

  // Create a store for the selected date range
  export const selectedDateRange = writable<DateRange>(DateRange.TODAY);

  function loadPinImages() {
    if (!map) return;

    Object.entries(pinSVGs).forEach(([name, svg]) => {
      const img = new Image();
      img.onload = () => {
        if (map && !map.hasImage(`pin-${name}`)) {
          map.addImage(`pin-${name}`, img, { pixelRatio: PIN_PIXEL_RATIO });
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
    map.on('movestart', handleMoveStart);

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
      map?.off('movestart', handleMoveStart);
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

  /* On touch devices, the pin popup becomes a bottom sheet: it
     slides up from the bottom edge with rounded top corners and the map
     stays visible above it. Its height is driven by EventPopup (max-height
     cap + .sheet-expanded when the user scrolls inside it). */
  @media (hover: none) and (pointer: coarse) {
    :global(.maplibregl-popup) {
      position: fixed !important;
      top: auto !important;
      right: 2.5% !important;
      bottom: 0 !important;
      left: 2.5% !important;
      transform: none !important;
      max-width: none !important;
      z-index: 800 !important; /* above map controls, below global alerts */
      animation: popup-sheet-in 300ms cubic-bezier(0.32, 0.72, 0, 1);
      transition:
        left 300ms cubic-bezier(0.32, 0.72, 0, 1),
        right 300ms cubic-bezier(0.32, 0.72, 0, 1);
    }

    /* Expanded sheet: edge to edge, no side gaps */
    :global(.maplibregl-popup:has(.sheet-expanded)) {
      left: 0 !important;
      right: 0 !important;
    }

    /* Peeking sheet: narrower pill, more map visible on both sides */
    :global(.maplibregl-popup:has(.sheet-peeked)) {
      left: 6% !important;
      right: 6% !important;
    }

    :global(.maplibregl-popup-content) {
      width: 100% !important;
      max-width: none !important;
      padding: 0 !important;
      display: flex !important;
      flex-direction: column;
    }

    :global(.maplibregl-popup .sv-popup) {
      display: flex;
      flex-direction: column;
    }

    :global(.maplibregl-popup-content .floating-panel) {
      position: relative;
      border-radius: 22px 22px 0 0;
      box-shadow: 0 -8px 32px rgba(0, 0, 0, 0.15);
    }

    /* Grabber: visual cue that the sheet can be expanded/collapsed */
    :global(.maplibregl-popup-content .floating-panel)::before {
      content: '';
      position: absolute;
      top: 6px;
      left: 50%;
      transform: translateX(-50%);
      width: 15%;
      min-width: 54px;
      max-width: 108px;
      height: 4px;
      border-radius: 2px;
      background-color: rgba(0, 0, 0, 0.2);
    }

    /* Let the panel's scroller shrink when the sheet is capped by its
       max-height, so it scrolls instead of clipping. */
    :global(.maplibregl-popup-content .floating-panel-content) {
      height: auto !important;
      min-height: 0 !important;
    }

    :global(.maplibregl-popup-tip) {
      display: none !important;
    }

    :global(.maplibregl-popup-close-button) {
      display: flex !important;
      align-items: center;
      justify-content: center;
      top: 12px !important;
      right: 12px !important;
      width: 36px;
      height: 36px;
      padding: 0;
      border-radius: 50%;
      background-color: rgba(255, 255, 255, 0.95);
      box-shadow: 0 2px 10px rgba(0, 0, 0, 0.18);
      color: transparent; /* MapLibre's × glyph sits off-center: hidden, the cross is drawn below */
      pointer-events: auto; /* container has pointer-events: none */
    }

    /* Drawn cross: geometrically centered, independent of font metrics */
    :global(.maplibregl-popup-close-button)::before,
    :global(.maplibregl-popup-close-button)::after {
      content: '';
      position: absolute;
      top: 50%;
      left: 50%;
      width: 16px;
      height: 2px;
      border-radius: 1px;
      background-color: #333;
    }

    :global(.maplibregl-popup-close-button)::before {
      transform: translate(-50%, -50%) rotate(45deg);
    }

    :global(.maplibregl-popup-close-button)::after {
      transform: translate(-50%, -50%) rotate(-45deg);
    }
  }

  @keyframes popup-sheet-in {
    from {
      translate: 0 100%;
      opacity: 0.4;
    }
    to {
      translate: 0 0;
      opacity: 1;
    }
  }

  @media (hover: none) and (pointer: coarse) and (prefers-reduced-motion: reduce) {
    :global(.maplibregl-popup) {
      animation: none;
      transition: none;
    }
  }
</style>
