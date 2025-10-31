<script lang="ts">
  import { getRelativeTimeDisplay } from '$lib/utils/dateUtils';
  import type { EventWithCoordinates } from '$lib/stores/pins';
  import type { Map as MaplibreMap } from 'maplibre-gl';

  export let map: MaplibreMap | undefined;
  export let pins: EventWithCoordinates[] = [];
  export let collapsed = true;

  // Function to toggle sidebar and update map padding
  function toggleSidebar() {
    collapsed = !collapsed;

    if (map) {
      const padding = { left: collapsed ? 0 : 350 };
      map.easeTo({
        padding,
        duration: 100
      });
    }
  }
</script>

<div class="sidebar-container">
  <div class="sidebar {collapsed ? 'collapsed' : ''}">
    <div class="sidebar-content">
      <h2 class="sidebar-title">Événements</h2>
      <div class="events-list">
        {#if pins.length > 0}
          {#each pins as event}
            <button
              class="event-item"
              onclick={() => {
                if (map) {
                  const coordinates = event.getCoordinates();
                  map.flyTo({
                    center: coordinates,
                    speed: 1.2,
                    curve: 1.4,
                    essential: true
                  });
                }
              }}
            >
              <div class="event-name">{event.name}</div>
              <div class="event-time">{getRelativeTimeDisplay(event.begin, event.end)}</div>
            </button>
          {/each}
        {:else}
          <div class="no-events-message">
            Aucun évènement dans cette zone
          </div>
        {/if}
      </div>
      <button
        type="button"
        class="sidebar-toggle"
        onclick={toggleSidebar}
        aria-label={collapsed ? 'Ouvrir le panneau latéral' : 'Fermer le panneau latéral'}
      >
        {collapsed ? '→' : '←'}
      </button>
    </div>
  </div>
</div>

<style>
  /* Sidebar container */
  .sidebar-container {
    position: absolute;
    top: 0;
    left: 0;
    height: 100%;
    z-index: 1;
    pointer-events: none;
  }

  /* Sidebar */
  .sidebar {
    width: 350px;
    height: 100%;
    transition: transform 100ms;
    pointer-events: auto;
    background: rgba(240, 240, 245, 0.7);
    backdrop-filter: blur(15px);
    -webkit-backdrop-filter: blur(15px);
    box-shadow: 0 0 20px rgba(0, 0, 0, 0.15);
  }

  .sidebar.collapsed {
    transform: translateX(-345px);
  }

  /* Content area */
  .sidebar-content {
    width: 100%;
    height: 100%;
    padding: 20px;
    box-sizing: border-box;
    overflow-y: auto;
    display: flex;
    flex-direction: column;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE and Edge */
  }

  .sidebar-content::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  /* Toggle button */
  .sidebar-toggle {
    position: absolute;
    width: 36px;
    height: 36px;
    top: 15px;
    right: -36px;
    background: rgba(240, 240, 245, 0.7);
    backdrop-filter: blur(15px);
    -webkit-backdrop-filter: blur(15px);
    border: none;
    border-radius: 0 18px 18px 0;
    box-shadow: 3px 0 8px -2px rgba(0, 0, 0, 0.15);
    cursor: pointer;
    display: flex;
    justify-content: center;
    align-items: center;
    font-size: 18px;
    padding: 0;
    color: #007AFF;
  }

  .sidebar-toggle:hover {
    background: rgba(255, 255, 255, 0.95);
  }

  /* Event list */
  .events-list {
    margin-top: 20px;
    display: flex;
    flex-direction: column;
    gap: 12px;
    flex: 1;
    overflow-y: auto;
    padding: 5px 0;
  }

  /* Hide scrollbar completely */
  .events-list {
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE and Edge */
  }

  .events-list::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  /* Event item */
  .event-item {
    padding: 15px;
    border-radius: 12px;
    background-color: rgba(255, 255, 255, 0.6);
    cursor: pointer;
    transition: all 0.2s ease;
    width: 100%;
    text-align: left;
    border: none;
    display: block;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
    margin-bottom: 2px;
  }

  .event-item:hover {
    background-color: rgba(255, 255, 255, 0.8);
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .event-name {
    font-weight: 600;
    font-size: 16px;
    margin-bottom: 8px;
    color: #000;
    letter-spacing: -0.2px;
  }

  .event-time {
    font-size: 13px;
    font-weight: 500;
    padding: 4px 10px;
    border-radius: 20px;
    display: inline-block;
    background-color: rgba(0, 122, 255, 0.1);
    color: #007AFF; /* iOS blue */
    letter-spacing: -0.1px;
  }

  .sidebar-title {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    font-size: 24px;
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
    padding: 30px 20px;
    border-radius: 12px;
    background-color: rgba(240, 240, 245, 0.4);
    margin: 20px 0;
    letter-spacing: -0.2px;
  }
</style>
