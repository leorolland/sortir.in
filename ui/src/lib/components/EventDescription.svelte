<script lang="ts">
  import { getRelativeTimeDisplay } from "$lib/utils/dateUtils";
  import type { EventsResponse } from "$lib/pocketbase/generated-types";

  // Event to display
  export let event: EventsResponse;

  /**
   * Convert event status to CSS class name
   */
  function getStatusClass(status: string): string {
    return status === "En cours"
      ? "status-ongoing"
      : status === "Terminé"
        ? "status-ended"
        : "status-upcoming";
  }

  // Define variables for time status and class
  let timeStatus: string;
  let statusClass: string;

  // Calculate time status and class using the refactored function
  $: {
    // Get the time display and status from the utility function
    const timeInfo = getRelativeTimeDisplay(event.begin, event.end);

    // Set the time display and status class
    timeStatus = timeInfo.display;
    statusClass = getStatusClass(timeInfo.status);
  }

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

  /**
   * Maps currency code to its symbol
   */
  function getCurrencySymbol(currencyCode: string): string {
    const currencyMap: Record<string, string> = {
      'EUR': '€',
      'USD': '$',
      'GBP': '£',
      'JPY': '¥',
      'CAD': 'C$',
      'AUD': 'A$',
      'CHF': 'CHF',
      'CNY': '¥',
      'RUB': '₽',
      'INR': '₹',
      'BRL': 'R$',
      'KRW': '₩',
      'MXN': 'Mex$'
    };

    return currencyMap[currencyCode] || currencyCode;
  }
</script>

<div class="event-card">
  <div class="event-title">{event.name}</div>

  {#if Array.isArray(event.genres) && event.genres.length > 0}
    <div class="event-genres">
      {#each event.genres as genre}
        <span class="genre-tag">{genre}</span>
      {/each}
    </div>
  {/if}



  {#if event.img}
    <div class="event-image">
      <img decoding="async" loading="lazy" src={event.img} alt={event.name} />
    </div>
  {/if}
  <div class="event-info">
    <div class="status {statusClass}">{#if statusClass === "status-ongoing"}🔴{/if}{timeStatus}</div>
    {#if event.price !== undefined && event.price_currency}
      <a
        href={event.source}
        target="_blank"
        rel="noopener noreferrer"
        class="event-price"
      >
        {event.price} {getCurrencySymbol(event.price_currency)}
      </a>
    {/if}
  </div>




  <div class="event-actions">
    <a
      href={event.source}
      target="_blank"
      rel="noopener noreferrer"
      class="event-link"
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
        <path d="M18 13v6a2 2 0 01-2 2H5a2 2 0 01-2-2V8a2 2 0 012-2h6"></path>
        <path d="M15 3h6v6"></path>
        <path d="M10 14L21 3"></path>
      </svg>
    </a>
  </div>
</div>

<style>
  .event-card {
    padding: 16px;
    border-radius: 10px;
    max-width: none;
    width: 100%;
    background-color: #fff;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    height: 100%;
    display: flex;
    flex-direction: column;
    transition:
      transform 0.2s,
      box-shadow 0.2s;
  }

  .event-card:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  .event-title {
    font-weight: bold;
    font-size: 16px;
    margin-bottom: 8px;
    color: #000;
    /* Ensure long titles don't break the layout */
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }

  .event-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
  }

  .status {
    font-size: 14px;
    font-weight: 600;
    padding: 2px 6px;
    border-radius: 4px;
    display: inline-block;
    white-space: nowrap;
  }

  .status-ongoing {
    background-color: #2a5492;
    color: white;
  }

  .status-ended {
    background-color: #979797;
    color: white;
  }

  .status-upcoming {
    background-color: #4282e3;
    color: white;
  }

  .event-place,
  .event-address,
  .event-kind {
    font-size: 14px;
    margin-bottom: 4px;
  }

  .event-genres {
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

  .event-price {
    font-weight: bold;
    font-size: 14px;
    color: #333;
    border: 1px solid black;
    padding: 2px 6px;
    border-radius: 4px;
    text-decoration: none;
    cursor: pointer;
    transition: background-color 0.2s, transform 0.1s;
    display: inline-block;
  }

  .event-price:hover {
    background-color: #f0f0f0;
    transform: translateY(-1px);
  }

  .event-image {
    margin-bottom: 12px;
    height: 140px;
    overflow: hidden;
    border-radius: 6px;
  }

  .event-image img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    border-radius: 4px;
    transition: transform 0.3s;
  }

  .event-image:hover img {
    transform: scale(1.05);
  }

  .event-actions {
    display: flex;
    justify-content: flex-end;
    margin-top: auto;
    padding-top: 12px;
  }

  .event-link {
    color: #0066cc;
    text-decoration: none;
    font-size: 14px;
  }

  .event-link:hover {
    text-decoration: underline;
  }
</style>
