<script>
  import { MapLibre } from 'svelte-maplibre';
</script>

<MapLibre
  center={[-1.6794, 48.1147]}
  zoom={12}
  class="map"
  standardControls
  style="https://basemaps.cartocdn.com/gl/positron-gl-style/style.json" />

<style>
  :global(.map) {
    height: 100%;
    width: 100%;
  }
</style>
