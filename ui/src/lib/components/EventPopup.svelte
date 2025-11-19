<script lang="ts">
  // @ts-ignore
  import type { Feature, Geometry } from "geojson";
  import { getRelativeTimeDisplay } from "$lib/utils/dateUtils";
  import FloatingPanel from "./FloatingPanel.svelte";
  import type { EventWithCoordinates } from "$lib/stores/pins";

  export let feature: Feature<Geometry, EventWithCoordinates> | undefined =
    undefined;

  $: event = feature?.properties;
  $: timeStatus = event ? getRelativeTimeDisplay(event.begin, event.end) : "";
  $: statusClass =
    timeStatus === "En cours"
      ? "status-ongoing"
      : timeStatus === "Terminé"
        ? "status-ended"
        : "status-upcoming";

  /**
   * Extracts the domain name from a URL
   */
  function extractDomain(url: string | null | undefined): string | null {
    if (!url) return null;
    try {
      const domain = new URL(url).hostname;
      return domain;
    } catch (e) {
      console.error("Invalid URL:", url);
      return null;
    }
  }
</script>

{#if event}
  <FloatingPanel compact withAnimation>
    <div class="popup-content">
      <div class="popup-title">{event.name}</div>
      <div class="popup-date {statusClass}">{timeStatus}</div>
      {#if event.place}
        <div class="popup-place">{event.place}</div>
      {/if}
      {#if event.address}
        <div class="popup-address">{event.address}</div>
      {/if}
      {#if event.kind && event.kind.toLowerCase() !== "event"}
        <div class="popup-kind">{event.kind}</div>
      {/if}
      {#if Array.isArray(event.genres) && event.genres.length > 0}
        <div class="popup-genres">
          {#each event.genres as genre}
            <span class="genre-tag">{genre}</span>
          {/each}
        </div>
      {/if}
      {#if event.price !== undefined && event.price_currency}
        <div class="popup-price">{event.price} {event.price_currency}</div>
      {/if}
      {#if event.img}
        <div class="popup-image">
          <img
            decoding="async"
            loading="lazy"
            src={event.img}
            alt={event.name}
          />
        </div>
      {/if}
      <div class="popup-actions">
        <a
          href={event.source}
          target="_blank"
          rel="noopener noreferrer"
          class="popup-button"
        >
          {extractDomain(event.source) || "Détails"}
          <svg
            class="external-link-icon"
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
          >
            <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"
            ></path>
            <path d="M15 3h6v6"></path>
            <path d="M10 14L21 3"></path>
          </svg>
        </a>
      </div>
    </div>
  </FloatingPanel>
{/if}

<style>
  .popup-content {
    padding: 12px;
    max-width: 280px;
  }

  .popup-title {
    font-weight: bold;
    font-size: 16px;
    margin-bottom: 8px;
    color: #000; /* Noir pur pour le titre */
  }

  .popup-date {
    font-size: 14px;
    margin-bottom: 8px;
    padding: 2px 6px;
    border-radius: 4px;
    display: inline-block;
  }

  .status-ongoing {
    background-color: #4caf50;
    color: white;
  }

  .status-ended {
    background-color: #9e9e9e;
    color: white;
  }

  .status-upcoming {
    background-color: #2196f3;
    color: white;
  }

  .popup-place,
  .popup-address,
  .popup-kind {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .popup-genres {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    margin-bottom: 8px;
  }

  .genre-tag {
    background-color: #f0f0f0;
    padding: 2px 6px;
    border-radius: 12px;
    font-size: 12px;
  }

  .popup-price {
    font-weight: bold;
    margin-bottom: 8px;
  }

  .popup-image {
    margin-bottom: 8px;
  }

  .popup-image img {
    max-width: 100%;
    border-radius: 4px;
  }

  .popup-actions {
    display: flex;
    justify-content: flex-end;
  }

  .popup-button {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    background-color: #2196f3;
    color: white;
    text-decoration: none;
    border-radius: 4px;
    font-size: 14px;
  }

  .external-link-icon {
    width: 14px;
    height: 14px;
  }
</style>
