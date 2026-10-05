<script lang="ts">
  import { metadata } from "$lib/metadata.js";
  import MapView from "$lib/components/MapView.svelte";

  const { data } = $props();
  $effect(() => {
    // the title is owned by MapView (events count and place name)
    $metadata.headline = ""; // Empty string instead of null
  });
</script>

<div class="fullscreen-map">
  <MapView />
</div>

<style>
  /* The height:100% chain doesn't resolve to the real screen in the iOS
     standalone app: pin the map to the layout viewport instead */
  .fullscreen-map {
    position: fixed;
    inset: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
</style>
