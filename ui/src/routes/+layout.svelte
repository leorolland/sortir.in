<script lang="ts">
  import "../app.scss";
  import { base } from "$app/paths";
  import { page } from "$app/stores";
  import Alerts, { hasAlerts } from "$lib/components/Alerts.svelte";
  import { metadata } from "$lib/metadata";
  const { data, children } = $props();
  const config = $derived(data.config ?? {});

  $effect(() => {
    if ($page.error) {
      $metadata.title = $page.error.message;
    }
  });
</script>

<svelte:head>
  <title>{$metadata.title}{config.site?.name ? ` | ${config.site.name}` : ''}</title>
</svelte:head>

<!-- Rendered only when alerts exist: a transparent fixed overlay at the top
     edge makes iOS 26 Safari paint its chrome opaque, hiding the map. -->
{#if hasAlerts()}
  <div class="alerts-container">
    <Alerts />
  </div>
{/if}
<main class="fullscreen">
  {@render children()}
</main>

<style lang="scss">

  .alerts-container {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 900;
    padding: env(safe-area-inset-top) 1rem 0;
  }

  main.fullscreen {
    flex-grow: 1;
    height: 100%;
    width: 100%;
  }
</style>
