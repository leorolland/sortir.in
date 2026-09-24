<script lang="ts">
  import MapLibre from 'svelte-maplibre/MapLibre.svelte';
  import GeoJSON from 'svelte-maplibre/GeoJSON.svelte';
  import CircleLayer from 'svelte-maplibre/CircleLayer.svelte';
  import SymbolLayer from 'svelte-maplibre/SymbolLayer.svelte';
  import Popup from 'svelte-maplibre/Popup.svelte';
  import { pinsStore, type FocusedPin } from '$lib/stores/pins';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import type { Map as MaplibreMap, Popup as MaplibrePopup } from 'maplibre-gl';
  import { GeolocateControl } from 'maplibre-gl';
  import MapSidebar from '$lib/components/MapSidebar.svelte';
  import { pinsToGeoJSON } from '$lib/utils/geoJsonUtils';
  import EventPopup from '$lib/components/EventPopup.svelte';
  import DateRangeSelector from '$lib/components/DateRangeSelector.svelte';
  import AppVersion from '$lib/components/AppVersion.svelte';
  import { pinSVGs, PIN_PIXEL_RATIO } from '$lib/components/pins/svg';
  import { writable, get } from 'svelte/store';
  import { untrack } from 'svelte';
  import { DateRange, getDateWindow } from '$lib/utils/dateUtils';
  import { eventsStore } from '$lib/stores/events';
  import { kindFilter } from '$lib/stores/filters';
  import { sheetState } from '$lib/stores/sheet';
  import { placeDisplayPhrase, reverseGeocode, type Place } from '$lib/utils/geocode';
  import { metadata } from '$lib/metadata.js';
  import { parseViewState, routerState, writeViewState } from '$lib/utils/viewState';
  import { afterNavigate, pushState, replaceState } from '$app/navigation';

  const initialView = parseViewState(window.location.search);

  // Pins currently displayed: narrowed by the kind filter shared with the
  // Suggestions pane (client-side, so toggling it needs no refetch)
  const pins = $derived($kindFilter ? $pinsStore.filter((p) => p.kind === $kindFilter) : $pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);
  let sidebarCollapsed = $state<boolean>(window.innerWidth < 768);
  const initialCenter: [number, number] = initialView.center
    ? [initialView.center.lon, initialView.center.lat]
    : [2.4, 46.6];
  const initialZoom = initialView.zoom ?? (window.innerWidth < 768 ? 4.5 : 5.5);
  let geoJsonData = $state(pinsToGeoJSON([]));
  let initialized = $state(false);
  let place = $state<Place>({});
  let zoom = $state<number>(0);

  // Pin focused in the popup. Also drives the `pin` URL parameter, so a
  // copied URL reopens the same view (ADR 0005).
  let selectedPin = $state<FocusedPin | null>(initialView.pin ? { loc: initialView.pin } : null);
  let popupOpen = $state(false);

  const eventsOnScreen = $derived(pins.reduce((sum, pin) => sum + pin.amount, 0));

  let geocodeTimer: ReturnType<typeof setTimeout> | undefined;

  let popupResizeObserver: ResizeObserver | undefined;

  // Same media query as the bottom-sheet CSS in the style block
  function isTouchDevice(): boolean {
    return window.matchMedia('(hover: none) and (pointer: coarse)').matches;
  }

  // Desktop only: pans the map by exactly the amount needed to bring the
  // popup fully inside the viewport. Measured from the rendered element
  // (not derived from screen heuristics), so it works for any popup size
  // and any viewport. panBy([dx, dy]) shifts map content by (-dx, -dy).
  function fitPopupToViewport(el: HTMLElement) {
    if (!map) return;

    const rect = el.getBoundingClientRect();
    if (rect.width === 0 || rect.height === 0) return;

    const margin = 16;
    let dx = 0;
    let dy = 0;

    if (rect.left < margin) dx = rect.left - margin;
    else if (rect.right > window.innerWidth - margin) {
      dx = rect.right - (window.innerWidth - margin);
    }

    if (rect.top < margin) dy = rect.top - margin;
    else if (rect.bottom > window.innerHeight - margin) {
      dy = rect.bottom - (window.innerHeight - margin);
    }

    if (dx || dy) map.panBy([dx, dy], { duration: 300 });
  }

  function handlePopupOpen(popup: MaplibrePopup) {
    if (isTouchDevice()) return; // mobile uses the bottom sheet, no fitting

    const el = popup.getElement();
    fitPopupToViewport(el);

    // The popup grows when its events finish loading: re-fit on each resize
    popupResizeObserver?.disconnect();
    popupResizeObserver = new ResizeObserver(() => fitPopupToViewport(el));
    popupResizeObserver.observe(el);
  }

  function handlePopupClose() {
    popupResizeObserver?.disconnect();
    popupResizeObserver = undefined;
  }

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
  export const selectedDateRange = writable<DateRange>(initialView.range ?? DateRange.TODAY);

  // Mirrors the current view (position, zoom, range, focused pin) into the
  // URL so it can be copied and reopened as-is (ADR 0005). replaceState keeps
  // the address bar in sync without polluting history; pushState is reserved
  // for pin focus so the Back button closes the popup.
  // URL writes wait for the SvelteKit router: its replaceState/pushState
  // throw before the initial navigation completes, and mount-time effects
  // (updatePins) can run before that. afterNavigate fires once the router is
  // initialized — and provides the first write.
  afterNavigate(() => {
    routerState.ready = true;
    syncUrl('replace');
  });

  function syncUrl(mode: 'replace' | 'push') {
    if (!map || !routerState.ready) return;

    const center = map.getCenter();
    const params = new URLSearchParams(window.location.search);
    writeViewState(params, {
      center: { lat: center.lat, lon: center.lng },
      zoom: map.getZoom(),
      range: get(selectedDateRange),
      // untrack: a pin focus must not re-run the caller effect (and refetch pins)
      pin: untrack(() => selectedPin)?.loc ?? null
    });

    const query = params.toString();
    const url = `${window.location.pathname}${query ? `?${query}` : ''}`;
    if (mode === 'push') pushState(url, {});
    else replaceState(url, {});
  }

  function focusPin(pin: FocusedPin) {
    selectedPin = pin;
    syncUrl('push');
  }

  function clearPinSelection() {
    if (!selectedPin) return;
    selectedPin = null;
    syncUrl('replace');
  }

  // The popup close button and outside clicks close the popup from inside
  // svelte-maplibre: reflect that in the selection (and URL).
  $effect(() => {
    popupOpen = selectedPin !== null;
  });

  $effect(() => {
    if (!popupOpen && selectedPin) clearPinSelection();
  });

  // Back/Forward restore the full view encoded in the target history entry
  // (position, range, pin focus) without rewriting the URL.
  $effect(() => {
    const handlePopState = () => {
      const view = parseViewState(window.location.search);

      if (map && (view.center || view.zoom !== undefined)) {
        map.jumpTo({
          ...(view.center ? { center: [view.center.lon, view.center.lat] as [number, number] } : {}),
          ...(view.zoom !== undefined ? { zoom: view.zoom } : {})
        });
      }

      selectedDateRange.set(view.range ?? DateRange.TODAY);
      selectedPin = view.pin ? { loc: view.pin } : null;
    };

    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  });

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

  // The pin properties carry the authoritative location (the events filter
  // matches it exactly); the geometry mirrors it as [lon, lat]. loc may
  // arrive as a JSON string depending on how the feature was serialized.
  function pinLocFromFeature(feature: any): FocusedPin['loc'] | undefined {
    const loc = feature?.properties?.loc;
    if (!loc) return undefined;
    if (typeof loc !== 'string') return loc;
    try {
      return JSON.parse(loc);
    } catch {
      return undefined;
    }
  }

  async function updatePins() {
    if (!map) return;

    syncUrl('replace');

    const window = getDateWindow($selectedDateRange);
    await pinsStore.loadPins(map.getBounds(), window);
    eventsStore.getEventsInBounds(map.getBounds(), window);

    // geoJsonData is derived from the pins store (and the kind filter) by
    // the effect below — no direct assignment here

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
          const focusedLoc = pinLocFromFeature(feature);

          if (focusedLoc) {
            focusPin({ loc: focusedLoc, kind: feature.properties?.kind });
          }

          if (isTouchDevice()) {
            // Mobile: bring the pin to the middle of the map strip above the
            // bottom sheet (which covers up to ~60% of the screen)
            map.flyTo({
              center: coordinates as [number, number],
              duration: 300,
              freezeElevation: true,
              offset: [0, -window.innerHeight * 0.3]
            });
          }
          // Desktop: no forced move; the popup opens next to the pin and
          // handlePopupOpen pans only if it overflows the viewport
        }
      } else {
        clearPinSelection();
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
      popupResizeObserver?.disconnect();
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
    const currentPinsString = JSON.stringify(pins);

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
    collapsed={sidebarCollapsed}
  />

  <DateRangeSelector selectedDateRange={selectedDateRange} />

  <AppVersion />

  <MapLibre
    center={initialCenter}
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
        radius: 60,
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
      </SymbolLayer>
    </GeoJSON>

    <!-- Manual popup driven by selectedPin: opens on pin clicks (handled in
         handleMapClick) and from the `pin` URL parameter, so a shared link
         restores the exact same popup (ADR 0005). -->
    <Popup
      openOn="manual"
      closeButton={true}
      maxWidth="none"
      bind:open={popupOpen}
      lngLat={selectedPin ? [selectedPin.loc.lon, selectedPin.loc.lat] : undefined}
      onopen={handlePopupOpen}
      onclose={handlePopupClose}
    >
      {#snippet children()}
        <EventPopup pin={selectedPin ?? undefined} dateRange={$selectedDateRange} />
      {/snippet}
    </Popup>
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
    top: calc(10px + env(safe-area-inset-top));
    right: calc(10px + env(safe-area-inset-right));
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
       max-height, so it scrolls instead of clipping. overscroll-behavior
       keeps drags on the sheet from chaining into map pans (which would
       peek the sheet mid-gesture). */
    :global(.maplibregl-popup-content .floating-panel-content) {
      height: auto !important;
      min-height: 0 !important;
      overscroll-behavior: contain;
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
