<script lang="ts">
  import { MapLibre } from 'svelte-maplibre';
  import { pinsStore } from '$lib/stores/pins';
  import 'maplibre-gl/dist/maplibre-gl.css';
  import { Marker, Popup, Map as MaplibreMap } from 'maplibre-gl';

  // Subscribe to pins store
  const pins = $derived($pinsStore);
  let map = $state<MaplibreMap | undefined>(undefined);

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

    return () => {
      map?.off('moveend', updatePins);
    };
  });

  // Effect to add markers when map is available or pins change
  $effect(() => {
    if (!map) return;

    // Clear existing markers first
    const markers = document.querySelectorAll('.maplibregl-marker');
    markers.forEach(marker => marker.remove());

    // Add markers for each event
    pins.forEach(event => {
      const popup = new Popup({ offset: [0, -10] })
        .setHTML(`
          <div style="font-weight: bold;">${event.name}</div>
        `);

      new Marker({ draggable: false })
        .setLngLat(event.getCoordinates())
        .setPopup(popup)
        .addTo(map as MaplibreMap);
    });
  });
</script>

<div class="map-container">
  <MapLibre
    center={[-1.6794, 48.1147]}
    zoom={12}
    class="map"
    standardControls
    style="https://basemaps.cartocdn.com/gl/voyager-gl-style/style.json"
    bind:map={map} />
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
</style>
