<script lang="ts">
  import { getRelativeTimeDisplay } from '$lib/utils/dateUtils';
  import type { Map as MaplibreMap } from 'maplibre-gl';
  import FloatingPanel from './FloatingPanel.svelte';
  import type { EventsResponse } from '$lib/pocketbase/generated-types';
  import { onDestroy, onMount } from 'svelte';
  import { eventsStore } from '$lib/stores/events';

  export let map: MaplibreMap | undefined;
  export let collapsed = true;
  let events: EventsResponse[] = [];
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

  // Group events by Kinds of events not terminated with kind not unknown or not movie
  $: groupedEvents = events.
  filter((event) => getRelativeTimeDisplay(event.begin, event.end).status !== 'Terminé').
  filter((event) => event.kind !== 'unknown' && event.kind !== 'movie').
  reduce((acc, event) => {
    acc[event.kind] = [...(acc[event.kind] || []), event];
    return acc;
  }, {} as Record<string, EventsResponse[]>);

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
        <div class="events-list">
          {#if events.length > 0}
            {#each Object.keys(groupedEvents) as kind}
              <h3 class="kind-title">{kind.charAt(0).toUpperCase() + kind.slice(1)}</h3>
              {#each groupedEvents[kind] as event}
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
                  <img src={event.img} alt={event.name} class="event-img" />
                {/if}
                <div class="event-name">{event.name}</div>
                <div class="event-time">{getRelativeTimeDisplay(event.begin, event.end).status}</div>
              </button>
            {/each}
            {/each}
          {:else}
            <div class="no-events-message">
              Aucun évènement dans cette zone
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
    min-height: 200px;
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
    min-height: 200px;
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
    margin-top: 20px;
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
    padding: 18px;
    border-radius: 16px;
    background-color: rgba(255, 255, 255, 0.5);
    cursor: pointer;
    transition: all 0.2s ease, opacity 0.3s ease, transform 0.3s ease;
    width: 100%;
    text-align: left;
    border: none;
    display: block;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.02);
    margin-bottom: 2px;
    position: relative;
    overflow: hidden;
    flex-shrink: 0;
    min-height: 90px;
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
    display: inline-block;
    width: 20%;
  }

  .event-name {
    display: inline-block;
    width: 60%;
    font-weight: 600;
    font-size: 17px;
    margin-bottom: 10px;
    color: #000;
    letter-spacing: -0.2px;
    line-height: 1.3;
  }

  .event-time {
    font-size: 14px;
    font-weight: 500;
    padding: 5px 12px;
    border-radius: 20px;
    display: inline-block;
    background-color: rgba(0, 122, 255, 0.1);
    color: #007AFF; /* iOS blue */
    letter-spacing: -0.1px;
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
    color: #8E8E93; /* iOS gray color */
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
