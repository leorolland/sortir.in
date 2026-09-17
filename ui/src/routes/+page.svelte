<script lang="ts">
  import { base } from "$app/paths";
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
  .fullscreen-map {
    width: 100%;
    height: 100%;
    position: relative;
    overflow: hidden;
    display: flex;
    flex-direction: column;
  }
</style>
