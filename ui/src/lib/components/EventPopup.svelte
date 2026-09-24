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
  let mapsOpen = false;
  let linkCopied = false;
  let copiedTimer: ReturnType<typeof setTimeout> | undefined;

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
    mapsOpen = false;
    linkCopied = false;
    clearTimeout(copiedTimer);
    sheetState.set('normal');

    const window = getDateWindow(currentDateRange);

    // Load events for this location and kind
    eventsStore.loadEventsForLocation(currentPin.loc, window);
  }

  // When pin or dateRange changes, load events for this location
  $: if (pin && dateRange) {
    loadEventsForPin(pin, dateRange);
  }

  async function sharePage() {
    const url = window.location.href;
    if (navigator.share) {
      try {
        await navigator.share({ title: events[0]?.place || document.title, url });
      } catch {
        // user dismissed the share sheet
      }
      return;
    }
    try {
      await navigator.clipboard.writeText(url);
      linkCopied = true;
      clearTimeout(copiedTimer);
      copiedTimer = setTimeout(() => (linkCopied = false), 2000);
    } catch {
      // clipboard blocked (permissions) — nothing else to offer
    }
  }

  function mapsLinks(event: EventsResponse, loc: { lat: number; lon: number }) {
    const name = encodeURIComponent([event.place, event.address].filter(Boolean).join(', '));
    // OSM search (Nominatim) is weak on venue names: query by address only,
    // falling back to the pin coordinates when no address is set
    const address = encodeURIComponent(event.address || `${loc.lat},${loc.lon}`);
    return {
      google: `https://www.google.com/maps/search/?api=1&query=${name}`,
      osm: `https://www.openstreetmap.org/search?query=${address}`,
      waze: `https://www.waze.com/ul?ll=${loc.lat}%2C${loc.lon}&q=${name}`,
    };
  }
</script>

{#snippet mapChips(links: ReturnType<typeof mapsLinks>)}
  <a
    class="map-button map-button--osm"
    href={links.osm}
    target="_blank"
    rel="noopener noreferrer"
    aria-label="Ouvrir dans OpenStreetMap"
    title="Ouvrir dans OpenStreetMap"
  >
    <svg class="osm-icon" viewBox="0 0 24 24" aria-hidden="true">
      <path fill="#fff" d="M2.672 23.969c-.352-.089-.534-.234-1.471-1.168C.085 21.688.014 21.579.018 20.999c0-.645-.196-.414 3.368-3.986 3.6-3.608 3.415-3.451 4.064-3.449.302 0 .378.016.62.14l.277.14 1.744-1.744-.218-.343c-.425-.662-.825-1.629-1.006-2.429a7.657 7.657 0 0 1 1.479-6.44c2.49-3.12 6.959-3.812 10.26-1.588 1.812 1.218 2.99 3.099 3.328 5.314.07.467.07 1.579 0 2.074a7.554 7.554 0 0 1-2.205 4.402 6.712 6.712 0 0 1-1.943 1.401c-.959.483-1.775.71-2.881.803-1.573.131-3.32-.305-4.656-1.163l-.343-.218-1.744 1.744.14.28c.125.241.14.316.14.617.003.651.156.467-3.426 4.049-2.761 2.756-3.186 3.164-3.398 3.261-.271.125-.69.171-.945.106zM17.485 13.95a6.425 6.425 0 0 0 4.603-3.51c1.391-2.899.455-6.306-2.227-8.108-.638-.43-1.529-.794-2.367-.962-.581-.117-1.809-.104-2.414.025a6.593 6.593 0 0 0-2.452 1.064c-.444.315-1.177 1.048-1.487 1.487a6.384 6.384 0 0 0 .38 7.907 6.406 6.406 0 0 0 3.901 2.136c.509.078 1.542.058 2.065-.037zm-3.738 7.376a80.97 80.97 0 0 1-2.196-.651c-.025-.028 1.207-4.396 1.257-4.449.023-.026 4.242 1.152 4.414 1.236.062.026-.003.288-.525 2.102a398.513 398.513 0 0 0-.635 2.236c-.025.087-.069.156-.097.156-.028-.003-1.028-.287-2.219-.631zm2.912.524c0-.053 1.227-4.333 1.246-4.347.047-.034 4.324-1.23 4.341-1.211.019.019-1.199 4.337-1.23 4.36-.02.019-4.126 1.191-4.259 1.218-.054.011-.098 0-.098-.019zm-7.105-1.911c.846-.852 1.599-1.627 1.674-1.728.171-.218.405-.732.472-1.015.026-.118.053-.352.058-.522l.011-.307.182-.051c.103-.028.193-.044.202-.034.023.025-1.207 4.321-1.246 4.36-.02.016-.677.213-1.464.436l-1.425.405 1.537-1.542zm8.289-3.06a1.371 1.371 0 0 1-.059-.187l-.044-.156.156-.028c1.339-.227 2.776-.856 3.908-1.713.16-.125.252-.171.265-.134.054.165.272.95.265.959-.034.034-4.48 1.282-4.492 1.261zm-15.083-1.3c-.05-.039-1.179-3.866-1.264-4.29-.016-.084.146-.044 2.174.536 2.121.604 2.192.629 2.222.74.028.098.011.129-.125.223-.084.059-.769.724-1.523 1.479a63.877 63.877 0 0 1-1.39 1.367c-.016 0-.056-.025-.093-.054zm.821-4.378c-1.188-.343-2.164-.623-2.167-.626-.016-.012 1.261-4.433 1.285-4.46.022-.022 4.422 1.211 4.469 1.252.009.009-.269 1.017-.618 2.239-.576 2.02-.643 2.224-.723 2.22-.05-.003-1.059-.285-2.247-.626zm2.959.538c.012-.031.212-.723.444-1.534l.42-1.476.056.321c.093.556.265 1.188.464 1.741.106.296.187.539.181.545-.008.006-.332.101-.719.212-.389.109-.741.21-.786.224-.058.016-.075.006-.059-.034zM4.905 6.112c-1.187-.339-2.167-.635-2.18-.654-.04-.062-1.246-4.321-1.23-4.338.026-.025 4.31 1.204 4.351 1.246.047.051 1.28 4.379 1.246 4.376L4.91 6.113zm2.148-1.713l-.519-1.806-.078-.28 1.693-.483c.934-.265 1.724-.495 1.76-.508.034-.016-.083.14-.26.336A8.729 8.729 0 0 0 7.69 5.23a4.348 4.348 0 0 0-.132.561c0 .293-.115-.025-.505-1.39z"/>
    </svg>
  </a>
  <a
    class="map-button map-button--waze"
    href={links.waze}
    target="_blank"
    rel="noopener noreferrer"
    aria-label="Ouvrir dans Waze"
    title="Ouvrir dans Waze"
  >
    <svg class="waze-icon" viewBox="0 0 50 50" aria-hidden="true">
      <path fill="#fff" d="M27.5,2.674C21.319,2.674 15.556,5.451 11.667,10.382C8.958,13.854 7.5,18.16 7.5,22.535L7.5,26.215C7.5,27.812 6.875,29.34 5.833,30.451C5,31.285 3.958,31.91 2.847,32.188C3.264,33.229 4.236,34.826 5.972,36.563C7.431,38.09 9.167,39.34 11.042,40.243L11.042,40.174C12.222,38.368 14.167,37.326 16.319,37.326C16.736,37.326 17.083,37.396 17.5,37.465C20,37.951 21.944,39.896 22.431,42.326L27.569,42.326C32.917,42.326 37.986,40.104 41.667,36.493C47.361,30.799 49.097,22.257 45.972,14.896C42.847,7.465 35.625,2.674 27.5,2.674Z"/>
      <path d="M27.5,0.174C20.625,0.174 14.167,3.229 9.792,8.715C6.667,12.674 5,17.535 5,22.604L5,26.215C5,28.09 3.681,29.826 1.111,29.965C0.486,29.965 0,30.451 -0.069,31.076C-0.139,32.743 1.667,35.868 4.167,38.368C5.903,40.104 7.917,41.493 10.069,42.604C9.375,46.424 12.361,49.896 16.25,49.896L16.319,49.896C19.306,49.896 21.806,47.813 22.431,44.965L27.639,44.965C28.194,47.813 30.694,49.896 33.75,49.896C34.444,49.896 35.208,49.757 35.903,49.549C37.639,48.993 38.958,47.674 39.583,45.937C40.139,44.34 40.069,42.743 39.583,41.424C40.972,40.521 42.222,39.549 43.403,38.368C47.639,34.201 50,28.507 50,22.604C50,16.632 47.639,11.076 43.403,6.84C39.167,2.465 33.472,0.174 27.5,0.174ZM27.5,2.674C35.556,2.674 42.847,7.535 45.972,14.965C49.097,22.396 47.361,30.937 41.667,36.563C37.986,40.243 32.917,42.396 27.569,42.396L22.431,42.396C21.944,39.896 20,38.021 17.5,37.535C17.083,37.465 16.736,37.396 16.319,37.396C14.236,37.396 12.222,38.438 11.042,40.243L11.042,40.313C9.167,39.34 7.5,38.09 5.972,36.632C4.236,34.896 3.264,33.229 2.847,32.257C4.028,31.979 5,31.354 5.833,30.521C6.875,29.34 7.5,27.882 7.5,26.285L7.5,22.604C7.5,18.229 8.958,13.924 11.667,10.451C15.556,5.382 21.319,2.674 27.5,2.674Z"/>
      <path d="M37.5,15.035C36.111,15.035 35,16.146 35,17.535C35,18.924 36.111,20.035 37.5,20.035C38.889,20.035 40,18.924 40,17.535C40,16.146 38.889,15.035 37.5,15.035Z"/>
      <path d="M22.5,15.035C21.111,15.035 20,16.146 20,17.535C20,18.924 21.111,20.035 22.5,20.035C23.889,20.035 25,18.924 25,17.535C25,16.146 23.889,15.035 22.5,15.035Z"/>
      <path d="M22.083,24.965C21.181,24.965 20.556,25.868 20.972,26.701C22.639,30.174 26.111,32.396 30,32.396C33.889,32.396 37.361,30.174 39.028,26.701C39.375,25.868 38.819,24.965 37.917,24.965C37.431,24.965 37.014,25.243 36.806,25.66C35.556,28.229 32.917,29.896 30.069,29.896C27.153,29.896 24.514,28.229 23.333,25.66C23.056,25.243 22.639,24.965 22.083,24.965Z"/>
    </svg>
  </a>
  <a
    class="map-button"
    href={links.google}
    target="_blank"
    rel="noopener noreferrer"
    aria-label="Ouvrir dans Google Maps"
    title="Ouvrir dans Google Maps"
  >
    <svg class="gmaps-icon" viewBox="0 0 92.3 132.3" aria-hidden="true">
      <path fill="#1a73e8" d="M60.2 2.2C55.8.8 51 0 46.1 0 32 0 19.3 6.4 10.8 16.5l21.8 18.3L60.2 2.2z"/>
      <path fill="#ea4335" d="M10.8 16.5C4.1 24.5 0 34.9 0 46.1c0 8.7 1.7 15.7 4.6 22l28-33.3-21.8-18.3z"/>
      <path fill="#4285f4" d="M46.2 28.5c9.8 0 17.7 7.9 17.7 17.7 0 4.3-1.6 8.3-4.2 11.4 0 0 13.9-16.6 27.5-32.7-5.6-10.8-15.3-19-27-22.7L32.6 34.8c3.3-3.8 8.1-6.3 13.6-6.3"/>
      <path fill="#fbbc04" d="M46.2 63.8c-9.8 0-17.7-7.9-17.7-17.7 0-4.3 1.5-8.3 4.1-11.3l-28 33.3c4.8 10.6 12.8 19.2 21 29.9l34.1-40.5c-3.3 3.9-8.1 6.3-13.5 6.3"/>
      <path fill="#34a853" d="M59.1 109.2c15.4-24.1 33.3-35 33.3-63 0-7.7-1.9-14.9-5.2-21.3L25.6 98c2.6 3.4 5.3 7.3 7.9 11.3 9.4 14.5 6.8 23.1 12.8 23.1s3.4-8.7 12.8-23.2"/>
    </svg>
  </a>
  <button
    class="map-button maps-share"
    onclick={sharePage}
    aria-label={linkCopied ? 'Lien copié !' : 'Partager'}
    title={linkCopied ? 'Lien copié !' : 'Partager'}
  >
    <svg
      class="share-icon"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      stroke-width="2"
      stroke-linecap="round"
      stroke-linejoin="round"
      aria-hidden="true"
    >
      {#if linkCopied}
        <path d="M20 6 9 17l-5-5" />
      {:else}
        <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
        <path d="m16 6-4-4-4 4" />
        <path d="M12 2v13" />
      {/if}
    </svg>
  </button>
{/snippet}

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
            {@const links = mapsLinks(events[0], pin.loc)}
            <div class="location-row">
              <div class="location-text">
                <h2 class="location-title" title={events[0].place}>{events[0].place}</h2>
                {#if events[0].address}
                  <div class="location-address">{events[0].address}</div>
                {/if}
              </div>
              {#if events.length > 1}
                <div class="map-links">
                  {@render mapChips(links)}
                </div>
              {:else}
                <div class="map-links">
                  <button
                    class="map-button maps-toggle"
                    class:open={mapsOpen}
                    onclick={() => (mapsOpen = !mapsOpen)}
                    aria-label="Voir sur une carte"
                    aria-expanded={mapsOpen}
                    title="Voir sur une carte"
                  >
                    <svg
                      class="chevron-icon"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      stroke-width="2"
                      stroke-linecap="round"
                      stroke-linejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M9 18l6-6-6-6" />
                    </svg>
                  </button>
                </div>
              {/if}
            </div>
            {#if events.length === 1}
              <div class="map-links map-links--below" class:open={mapsOpen}>
                {@render mapChips(links)}
              </div>
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

  .location-row {
    display: flex;
    align-items: flex-start;
    gap: 10px;
  }

  .location-text {
    flex: 1;
    min-width: 0;
  }

  .map-links {
    display: flex;
    gap: 8px;
    flex-shrink: 0;
  }

  .map-button {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    margin-top: 2px;
    border-radius: 50%;
    background: #fff;
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.06),
      0 1px 4px rgba(0, 0, 0, 0.12);
    transition: transform 0.15s ease, box-shadow 0.15s ease;
  }

  .map-links--below {
    display: none;
    margin-top: 8px;
  }

  .map-links--below.open {
    display: flex;
  }

  .maps-toggle {
    border: none;
    padding: 0;
    cursor: pointer;
    color: #666;
  }

  .maps-share {
    border: none;
    padding: 0;
    cursor: pointer;
    color: #007aff;
  }

  .share-icon {
    width: 14px;
    height: 14px;
  }

  .chevron-icon {
    width: 14px;
    height: 14px;
    transition: transform 0.15s ease;
  }

  .maps-toggle.open .chevron-icon {
    transform: rotate(90deg);
  }

  .map-button--osm {
    background: #2d3335;
  }

  .map-button--waze {
    background: #33ccff;
  }

  .map-button:hover {
    transform: scale(1.1);
    box-shadow:
      0 0 0 1px rgba(0, 0, 0, 0.06),
      0 3px 10px rgba(0, 0, 0, 0.18);
  }

  .map-button:active {
    transform: scale(0.95);
  }

  .map-button svg {
    display: block;
  }

  .gmaps-icon {
    width: 11px;
    height: 16px;
  }

  .osm-icon,
  .waze-icon {
    width: 16px;
    height: 16px;
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

    /* The map links don't fit next to the title on a phone: stack them
       as a full-width row below the place address instead */
    .location-row {
      flex-wrap: wrap;
    }

    .location-text {
      flex-basis: 100%;
    }

    .map-links {
      flex-basis: 100%;
    }

    .maps-toggle {
      display: none;
    }

    .map-links--below {
      display: flex;
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

    /* The peeked bar only fits the place name: hide the link chips so
       they don't get half-clipped by the 76px max-height */
    :global(.dynamic-panel.sheet-peeked) .map-links {
      display: none;
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
