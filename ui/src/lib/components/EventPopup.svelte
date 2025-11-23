<script lang="ts">
  // @ts-ignore
  import type { Feature, Geometry } from "geojson";
  import { getMaxDateForRange, type DateRange } from "$lib/utils/dateUtils";
  import FloatingPanel from "./FloatingPanel.svelte";
  import type { Pin } from "$lib/stores/pins";
  import { eventsStore } from "$lib/stores/events";
  import type { EventsResponse } from "$lib/pocketbase/generated-types";
  import { onMount, onDestroy } from "svelte";
  import EventDescription from "./EventDescription.svelte";

  export let feature: Feature<Geometry, Pin> | undefined = undefined;
  export let dateRange: DateRange;

  // Local state
  let loading = false;
  let events: EventsResponse[] = [];
  let unsubscribe: () => void;

  // Subscribe to events store
  onMount(() => {
    unsubscribe = eventsStore.subscribe((value) => {
      events = value;
      loading = false;
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  function loadEventsForFeature(feature: Feature<Geometry, Pin>, currentDateRange: DateRange) {
    const pin = feature.properties;
    loading = true;

    const maxDate = getMaxDateForRange(currentDateRange);

    // Parse location if it's a string
    let location;
    if (typeof pin.loc === 'string') {
      location = JSON.parse(pin.loc);
    } else {
      location = pin.loc;
    }

    // Load events for this location and kind
    eventsStore.loadEventsForLocation(location, maxDate);
  }

  // When feature changes or dateRange changes, load events for this location
  $: if (feature?.properties && dateRange) {
    loadEventsForFeature(feature, dateRange);
  }

  // Set CSS variable for event count to control grid width
  $: if (events) {
    setTimeout(() => {
      const container = document.querySelector('.events-container') as HTMLElement;
      if (container) {
        container.style.setProperty('--event-count', String(events.length));
      }
    }, 0);
  }
</script>

{#if feature?.properties}
  <FloatingPanel compact={events.length <= 1} withAnimation className="dynamic-panel">
    <div class="popup-content">
      {#if loading}
        <div class="loading">
          <div class="spinner"></div>
          <div>Chargement des événements...</div>
        </div>
      {:else if events.length === 0}
        <div class="popup-header">
          <div class="popup-title">Aucun événement trouvé</div>
          <div class="popup-kind">{feature.properties.kind}</div>
          {#if typeof feature.properties.loc === "string"}
            {@const locObj = JSON.parse(feature.properties.loc)}
            <div class="popup-place">
              Lat: {locObj.lat.toFixed(4)}, Lon: {locObj.lon.toFixed(4)}
            </div>
          {:else}
            <div class="popup-place">
              Lat: {feature.properties.loc.lat.toFixed(4)}, Lon: {feature.properties.loc.lon.toFixed(4)}
            </div>
          {/if}
        </div>
      {:else}
        <div class="popup-header">
          {#if events[0]?.place}
            <h2 class="location-title" title={events[0].place}>{events[0].place}</h2>
            {#if events[0].address}
              <div class="location-address">{events[0].address}</div>
            {/if}
          {/if}
        </div>

        <!-- Events grid -->
        <div class="events-container">
          {#each events.sort((a, b) => new Date(a.begin).getTime() - new Date(b.begin).getTime()) as event (event.id)}
            <EventDescription {event} />
          {/each}
        </div>
      {/if}
    </div>
  </FloatingPanel>
{/if}

<style>
  .popup-content {
    padding: 17px;
  }

  .loading {
    text-align: center;
    padding: 20px;
    color: #666;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .spinner {
    width: 40px;
    height: 40px;
    border: 3px solid rgba(0, 0, 0, 0.1);
    border-radius: 50%;
    border-top-color: #2196f3;
    animation: spin 1s ease-in-out infinite;
  }

  @keyframes spin {
    to { transform: rotate(360deg); }
  }

  .popup-header {
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #eee;
  }
  .popup-kind {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .popup-place {
    font-size: 14px;
    margin-bottom: 4px;
    color: #666;
  }

  .location-title {
    font-size: 24px;
    font-weight: 1200;
    color: #323232;
    margin: 0;
    line-height: 1.3;
    hyphens: auto;
    display: inline-block;
    width: 100%;
    max-width: calc(240px * min(5, var(--event-count, 1)) + (min(5, var(--event-count, 1)) - 1) * 20px);
  }

  .location-address {
    font-size: 14px;
    color: #666;
    margin-bottom: 4px;
    word-wrap: break-word;
    overflow-wrap: break-word;
    max-width: 100%;
  }

  .events-container {
    display: grid;
    /* Auto-fill grid with minimum 220px columns */
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    grid-gap: 20px;
    max-height: 65vh;
    overflow-y: auto;
    padding-right: 12px; /* Increased padding to accommodate scrollbar */
    /* Make width 100% to fit inside the popup */
    width: 100%;
    max-width: calc(240px * min(5, var(--event-count, 1)) + (min(5, var(--event-count, 1)) - 1) * 20px);
  }

  /* Modern scrollbar styling */
  .events-container {
    scrollbar-width: thin; /* Firefox */
    scrollbar-color: rgba(0, 0, 0, 0.2) transparent; /* Firefox */
  }

  .events-container::-webkit-scrollbar {
    width: 4px;
  }

  .events-container::-webkit-scrollbar-track {
    background: transparent;
    margin: 4px 0;
  }

  .events-container::-webkit-scrollbar-thumb {
    background: rgba(0, 0, 0, 0.2);
    border-radius: 20px;
    transition: background 0.2s ease;
  }

  .events-container::-webkit-scrollbar-thumb:hover {
    background: rgba(0, 0, 0, 0.3);
  }

  /* Override FloatingPanel styles for dynamic sizing */
  :global(.dynamic-panel) {
    width: auto !important;
    max-width: none !important;
  }

  :global(.dynamic-panel .floating-panel-content) {
    width: auto !important;
    display: block !important;
  }
</style>
