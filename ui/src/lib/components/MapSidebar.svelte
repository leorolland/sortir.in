<script lang="ts">
  import { getRelativeTimeDisplay } from '$lib/utils/dateUtils';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import FloatingPanel from './FloatingPanel.svelte';
  import type { EventsResponse } from '$lib/pocketbase/generated-types';
  import { onDestroy, onMount } from 'svelte';
  import { eventsStore } from '$lib/stores/events';

  type SidebarEvent = EventsResponse & { timeInfo: { status: string; display: string } };

  const KIND_LABELS: Record<string, string> = {
    concert: 'Concerts',
    theater: 'Théâtre',
    festival: 'Festivals',
    party: 'Soirées',
    karaoke: 'Karaoké',
    business: 'Professionnel',
    'food-drinks': 'Food & boissons',
    sports: 'Sports',
    exhibitions: 'Expositions',
    'health-wellness': 'Bien-être',
    circus: 'Cirque',
    workshop: 'Ateliers',
    'flea-market': 'Brocantes',
    solidarity: 'Solidarité'
  };

  function kindLabel(kind: string): string {
    return KIND_LABELS[kind] ?? kind.charAt(0).toUpperCase() + kind.slice(1);
  }

  const FILTER_STORAGE_KEY = 'sidebar-kind-filter';

  function loadSelectedKind(): string | null {
    try {
      return localStorage.getItem(FILTER_STORAGE_KEY);
    } catch {
      return null;
    }
  }

  function setSelectedKind(kind: string | null): void {
    selectedKind = kind;
    try {
      if (kind === null) localStorage.removeItem(FILTER_STORAGE_KEY);
      else localStorage.setItem(FILTER_STORAGE_KEY, kind);
    } catch {
      // storage unavailable (private mode) — selection just won't persist
    }
  }

  export let map: MaplibreMap | undefined;
  export let collapsed = true;
  let events: EventsResponse[] = [];
  let selectedKind: string | null = loadSelectedKind();
  let unsubscribe: () => void;

  // Subscribe to events store
  onMount(() => {
    unsubscribe = eventsStore.subscribeEventsForBounds((value) => {
      events = value;
    });
  });

  onDestroy(() => {
    if (unsubscribe) unsubscribe();
  });

  // Precompute time info once per event, then drop terminated/unknown/movie
  // and sort by soonest begin (PocketBase dates sort lexicographically)
  $: sidebarEvents = events
    .map((event) => ({ ...event, timeInfo: getRelativeTimeDisplay(event.begin, event.end) }))
    .filter((event) => event.timeInfo.status !== 'Terminé')
    .filter((event) => event.kind !== 'unknown' && event.kind !== 'movie')
    .sort((a, b) => a.begin.localeCompare(b.begin));

  // Group events by kind
  $: groupedEvents = sidebarEvents.reduce((acc, event) => {
    acc[event.kind] = [...(acc[event.kind] || []), event];
    return acc;
  }, {} as Record<string, SidebarEvent[]>);

  $: kinds = Object.keys(groupedEvents);

  // Ignore a stale selection (kind no longer present after a map move)
  $: visibleGroups =
    selectedKind && groupedEvents[selectedKind]
      ? { [selectedKind]: groupedEvents[selectedKind] }
      : groupedEvents;

  function timeLabel(event: SidebarEvent): string {
    return event.timeInfo.status === 'En cours' ? 'En cours' : event.timeInfo.display;
  }

  // Function to toggle sidebar and update map padding
  function toggleSidebar() {
    collapsed = !collapsed;

    if (map) {
      const padding = { left: collapsed ? 0 : 380 };
      map.easeTo({
        padding,
        duration: 100
      });
    }
  }
</script>

<div class="sidebar-container">
  {#if collapsed}
    <button
      type="button"
      class="open-sidebar-button"
      onclick={toggleSidebar}
      aria-label="Ouvrir le panneau latéral"
    >
      ≡
    </button>
  {/if}
  <div class="sidebar {collapsed ? 'collapsed' : ''}">
    <FloatingPanel withAnimation scrollable className="sidebar-floating-panel">
      <div class="sidebar-inner-content">
        <h2 class="sidebar-title">Suggestions</h2>
        {#if kinds.length > 0}
          <div class="kind-filters">
            <button
              type="button"
              class="kind-chip"
              class:active={selectedKind === null}
              onclick={() => setSelectedKind(null)}
            >
              Tout
            </button>
            {#each kinds as kind}
              <button
                type="button"
                class="kind-chip"
                class:active={selectedKind === kind}
                onclick={() => setSelectedKind(selectedKind === kind ? null : kind)}
              >
                {kindLabel(kind)}
              </button>
            {/each}
          </div>
        {/if}
        <div class="events-list">
          {#if kinds.length > 0}
            {#each Object.keys(visibleGroups) as kind}
              <h3 class="kind-title">{kindLabel(kind)}</h3>
              {#each visibleGroups[kind] as event}
              <button
                class="event-item"
                onclick={() => {
                  if (map) {
                    const coordinates = { lat: event.loc.lat, lon: event.loc.lon };
                    map.flyTo({
                      center: coordinates,
                      speed: 1.2,
                      curve: 1.4,
                      zoom: 15,
                      essential: true
                    });
                  }
                }}
              >
                {#if event.img}
                  <img src={event.img} alt={event.name} class="event-img" loading="lazy" />
                {:else}
                  <div class="event-img event-img-placeholder" aria-hidden="true"></div>
                {/if}
                <div class="event-content">
                  <div class="event-name">{event.name}</div>
                  {#if event.place}
                    <div class="event-place">{event.place}</div>
                  {/if}
                  <div class="event-time" class:ongoing={event.timeInfo.status === 'En cours'}>{timeLabel(event)}</div>
                </div>
              </button>
            {/each}
            {/each}
          {:else}
            <div class="no-events-message">
              Aucune suggestion dans cette zone
            </div>
          {/if}
        </div>
        {#if !collapsed}
          <button
            type="button"
            class="sidebar-toggle"
            onclick={toggleSidebar}
            aria-label="Fermer le panneau latéral"
          >
            <span class="close-icon">×</span>
          </button>
        {/if}
      </div>
    </FloatingPanel>
  </div>
</div>

<style>
  /* Sidebar container */
  .sidebar-container {
    position: absolute;
    top: 0;
    left: 0;
    height: 100%;
    z-index: 11;
    pointer-events: none;
    padding: 20px;
    box-sizing: border-box;
    display: flex;
    align-items: center;
    justify-content: flex-start;
  }

  /* Sidebar */
  .sidebar {
    width: 380px;
    height: auto;
    max-height: calc(100% - 40px);
    transition: transform 100ms ease;
    pointer-events: auto;
    display: flex;
    flex-direction: column;
  }

  .sidebar.collapsed {
    transform: translateX(-450px);
  }

  :global(.sidebar-floating-panel) {
    width: 100%;
    height: 100%;
    max-height: calc(100% - 40px);
  }

  /* Inner content */
  .sidebar-inner-content {
    width: 100%;
    height: 100%;
    position: relative;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE and Edge */
  }

  .sidebar-inner-content::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  /* Toggle button */
  .sidebar-toggle {
    position: absolute;
    width: 32px;
    height: 32px;
    top: 16px;
    right: 16px;
    background: rgba(240, 240, 245, 0.7);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: none;
    border-radius: 50%;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    cursor: pointer;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 16px;
    padding: 0;
    color: #007AFF;
    z-index: 10;
    transition: all 0.2s ease;
  }

  .sidebar-toggle:hover {
    background: rgba(255, 255, 255, 0.9);
    transform: scale(1.05);
  }

  .sidebar-toggle:active {
    transform: scale(0.95);
  }

  .close-icon {
    font-size: 22px;
    line-height: 0;
    position: relative;
    top: 1px;
    font-weight: 300;
  }

  /* Open sidebar button */
  .open-sidebar-button {
    width: 40px;
    height: 40px;
    background: rgba(240, 240, 245, 0.8);
    backdrop-filter: blur(10px);
    -webkit-backdrop-filter: blur(10px);
    border: none;
    border-radius: 50%;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
    cursor: pointer;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 22px;
    padding: 0;
    color: #007AFF;
    transition: all 0.2s ease;
    pointer-events: auto;
  }

  .open-sidebar-button:hover {
    background: rgba(255, 255, 255, 0.9);
    transform: scale(1.05);
  }

  .open-sidebar-button:active {
    transform: scale(0.95);
  }

  /* Event list */
  .events-list {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    padding: 5px 0;
    flex: 1;
    transition: all 0.4s cubic-bezier(0.25, 1, 0.5, 1);
    overflow-y: auto;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE and Edge */
  }

  .events-list::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  /* Event item */
  .event-item {
    padding: 12px;
    border-radius: 16px;
    background-color: rgba(255, 255, 255, 0.5);
    cursor: pointer;
    transition: all 0.2s ease, opacity 0.3s ease, transform 0.3s ease;
    width: 100%;
    text-align: left;
    border: none;
    display: flex;
    align-items: center;
    gap: 12px;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
    position: relative;
    overflow: hidden;
    flex-shrink: 0;
    animation: fadeIn 0.3s ease-out;
  }

  @keyframes fadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .kind-title {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    font-size: 20px;
    font-weight: 700;
    color: #000;
  }

  .event-item:hover {
    background-color: rgba(255, 255, 255, 0.7);
    transform: translateY(-2px) scale(1.01);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.08);
  }

  .event-item:active {
    transform: translateY(0) scale(0.99);
    background-color: rgba(255, 255, 255, 0.8);
  }

  .event-img {
    width: 64px;
    height: 64px;
    border-radius: 12px;
    object-fit: cover;
    flex-shrink: 0;
    background-color: rgba(0, 122, 255, 0.08);
  }

  .event-img-placeholder {
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .event-content {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
  }

  .event-name {
    font-weight: 600;
    font-size: 15px;
    color: #000;
    letter-spacing: -0.2px;
    line-height: 1.3;
    display: -webkit-box;
    line-clamp: 2;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
    overflow: hidden;
  }

  .event-place {
    font-size: 13px;
    font-weight: 500;
    color: #8E8E93;
    max-width: 100%;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }

  .event-time {
    font-size: 14px;
    font-weight: 500;
    padding: 5px 12px;
    border-radius: 20px;
    display: inline-block;
    background-color: rgba(0, 122, 255, 0.1);
    color: #007AFF;
    letter-spacing: -0.1px;
  }

  .event-time.ongoing {
    background-color: #007AFF;
    color: #fff;
  }

  /* Kind filter chips */
  .kind-filters {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-top: 10px;
  }

  .kind-chip {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    font-size: 13px;
    font-weight: 600;
    padding: 6px 12px;
    border-radius: 16px;
    border: none;
    cursor: pointer;
    background-color: rgba(255, 255, 255, 0.6);
    color: #000;
    letter-spacing: -0.1px;
    transition: all 0.2s ease;
  }

  .kind-chip:hover {
    background-color: rgba(255, 255, 255, 0.9);
    transform: scale(1.03);
  }

  .kind-chip.active {
    background-color: #007AFF;
    color: #fff;
  }

  .sidebar-title {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    font-size: 28px;
    font-weight: 700;
    color: #000;
    margin: 0 0 5px 0;
    letter-spacing: -0.5px;
  }

  .no-events-message {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    color: #8E8E93;
    font-size: 16px;
    text-align: center;
    padding: 20px;
    border-radius: 16px;
    background-color: rgba(240, 240, 245, 0.3);
    margin: 10px 0;
    letter-spacing: -0.2px;
    backdrop-filter: blur(5px);
    -webkit-backdrop-filter: blur(5px);
    min-height: 60px;
    display: flex;
    align-items: center;
    justify-content: center;
    animation: fadeIn 0.4s ease-out;
    transition: all 0.3s ease;
  }

</style>
