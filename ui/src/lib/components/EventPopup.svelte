<script lang="ts">
  import { getDateWindow, type DateRange } from "$lib/utils/dateUtils";
  import FloatingPanel from "./FloatingPanel.svelte";
  import type { FocusedPin } from "$lib/stores/pins";
  import { eventsStore } from "$lib/stores/events";
  import type { EventsResponse } from "$lib/pocketbase/generated-types";
  import { onMount, onDestroy } from "svelte";
  import { get } from "svelte/store";
  import EventDescription from "./EventDescription.svelte";
  import { sheetState } from "$lib/stores/sheet";

  export let pin: FocusedPin | undefined = undefined;
  export let dateRange: DateRange;

  // Local state
  let loading = false;
  let events: EventsResponse[] = [];
  let unsubscribe: () => void;

  /**
   * Bottom sheet behavior (touch devices): drives the sheet position store.
   * - scrolling inside the sheet expands it (peek -> normal -> expanded)
   * - scrolling back to the very top collapses an expanded sheet
   * - tapping the sheet restores it when peeked
   */
  function expandSheetOnScroll(node: HTMLElement) {
    const scroller = node.closest<HTMLElement>('.floating-panel-content');
    if (!scroller) return;

    const onScroll = () => {
      const top = scroller.scrollTop;
      const state = get(sheetState);
      if (top > 140) {
        sheetState.set('expanded');
      } else if (state === 'expanded' && top < 2) {
        sheetState.set('normal');
      } else if (state === 'peek' && top > 24) {
        sheetState.set('normal');
      }
    };

    const onTap = () => {
      if (get(sheetState) === 'peek') sheetState.set('normal');
    };

    scroller.addEventListener('scroll', onScroll, { passive: true });
    node.addEventListener('click', onTap);
    return {
      destroy: () => {
        scroller.removeEventListener('scroll', onScroll);
        node.removeEventListener('click', onTap);
      }
    };
  }

  // Subscribe to events store
  onMount(() => {
    unsubscribe = eventsStore.subscribeEventsForLocation((value) => {
      events = value;
      loading = false;
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  function loadEventsForPin(currentPin: FocusedPin, currentDateRange: DateRange) {
    loading = true;
    sheetState.set('normal');

    const window = getDateWindow(currentDateRange);

    // Load events for this location and kind
    eventsStore.loadEventsForLocation(currentPin.loc, window);
  }

  // When pin or dateRange changes, load events for this location
  $: if (pin && dateRange) {
    loadEventsForPin(pin, dateRange);
  }
</script>

{#if pin}
  <FloatingPanel
    compact={events.length <= 1}
    withAnimation
    className="dynamic-panel {$sheetState === 'expanded' ? 'sheet-expanded' : ''} {$sheetState === 'peek' ? 'sheet-peeked' : ''}"
  >
    <div class="popup-content" use:expandSheetOnScroll>
      {#if loading}
        <div class="loading">
          <div class="spinner"></div>
          <div>Chargement des événements...</div>
        </div>
      {:else if events.length === 0}
        <div class="popup-header">
          <div class="popup-title">Aucun événement trouvé</div>
          {#if pin.kind}
            <div class="popup-kind">{pin.kind}</div>
          {/if}
          <div class="popup-place">
            Lat: {pin.loc.lat.toFixed(4)}, Lon: {pin.loc.lon.toFixed(4)}
          </div>
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
        <div class="events-container" class:single={events.length === 1}>
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
  }

  .location-address {
    font-size: 14px;
    color: #666;
    margin-bottom: 4px;
    word-wrap: break-word;
    overflow-wrap: break-word;
    max-width: 100%;
  }

  .location-title, .location-address {
    contain: inline-size;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .events-container {
    display: grid;
    /* Auto-fill grid with minimum 220px columns */
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    grid-gap: 20px;
    /* Never let the popup outgrow the viewport: the grid takes the space
       left by the popup chrome (paddings + header) and scrolls inside */
    max-height: min(65vh, calc(100vh - 240px));
    overflow-y: auto;
    padding-right: 12px; /* Increased padding to accommodate scrollbar */
    width: min(calc(80vw - 82px), 544px);
  }

  .events-container.single {
    width: min(calc(80vw - 82px), 240px);
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

  /* Bottom sheet on touch devices: the panel hugs its content up to a cap
     and expands when the user scrolls inside it (positions are driven by
     the sheetState store). The map stays visible above the sheet. */
  @media (hover: none) and (pointer: coarse) {
    :global(.maplibregl-popup-content .dynamic-panel) {
      max-height: 60vh;
      max-height: 60dvh;
      transition: max-height 300ms cubic-bezier(0.32, 0.72, 0, 1);
    }

    .popup-content {
      padding: 20px 17px calc(20px + env(safe-area-inset-bottom));
    }

    .popup-header {
      padding-right: 56px;
    }

    .events-container {
      width: 100%;
      max-height: none;
      overflow: visible;
      padding-right: 0;
    }

    /* Expanded: stretched after scrolling inside the sheet */
    :global(.maplibregl-popup-content .dynamic-panel.sheet-expanded) {
      max-height: 90vh;
      max-height: 90dvh;
    }

    /* Peek position: collapsed to its title bar after a map move, so the
       map takes most of the screen; the content is clipped, not hidden, so
       the max-height transition stays smooth. Tapping the bar restores. */
    :global(.maplibregl-popup-content .dynamic-panel.sheet-peeked) {
      max-height: 76px;
      border-radius: 22px;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.18);
    }

    :global(.dynamic-panel.sheet-peeked .floating-panel-content) {
      padding: 0 !important;
    }

    :global(.dynamic-panel.sheet-peeked) .popup-content {
      padding: 16px 12px 16px 17px;
    }

    :global(.dynamic-panel.sheet-peeked) .popup-header {
      margin-bottom: 0;
      padding-bottom: 0;
      border-bottom: none;
    }

    :global(.dynamic-panel.sheet-peeked) .location-title {
      font-size: 17px;
      display: block;
      -webkit-line-clamp: unset;
      line-clamp: unset;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    :global(.dynamic-panel.sheet-peeked) .popup-title {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
</style>
