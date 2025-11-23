<script lang="ts">
  import { type Writable } from 'svelte/store';
  import FloatingPanel from './FloatingPanel.svelte';
  import { onMount } from 'svelte';

  // Date range options
  import { DateRange } from '$lib/utils/dateUtils';

  // Accept the selected date range as a prop
  export let selectedDateRange: Writable<DateRange>;

  // Get the appropriate label based on time of day
  function getTodayLabel(): string {
    const now = new Date();
    const hour = now.getHours();
    // "Ce soir" between 16h and 4h
    return (hour >= 16 || hour < 4) ? "Ce soir" : "Aujourd'hui";
  }

  let todayLabel = getTodayLabel();

  // Update the today label when mounted
  onMount(() => {
    todayLabel = getTodayLabel();
  });
</script>

<div class="date-range-selector-container">
  <FloatingPanel compact withAnimation className="date-range-panel">
    <div class="date-range-options">
      <button
        class="date-range-option"
        class:active={$selectedDateRange === DateRange.TODAY}
        on:click={() => selectedDateRange.set(DateRange.TODAY)}
      >
        {todayLabel}
      </button>
      <div class="separator"></div>
      <button
        class="date-range-option"
        class:active={$selectedDateRange === DateRange.TOMORROW}
        on:click={() => selectedDateRange.set(DateRange.TOMORROW)}
      >
        Demain
      </button>
      <div class="separator"></div>
      <button
        class="date-range-option"
        class:active={$selectedDateRange === DateRange.THIS_WEEK}
        on:click={() => selectedDateRange.set(DateRange.THIS_WEEK)}
      >
        Cette semaine
      </button>
    </div>
  </FloatingPanel>
</div>

<style>
  .date-range-selector-container {
    position: absolute;
    top: 16px;
    left: 50%;
    transform: translateX(-50%);
    z-index: 10;
    width: auto;
  }

  :global(.date-range-panel) {
    padding: 0;
    border-radius: 24px;
  }

  :global(.date-range-panel .floating-panel-content) {
    padding: 0;
  }

  .date-range-options {
    display: flex;
    align-items: center;
    height: 40px;
  }

  .date-range-option {
    background: transparent;
    border: none;
    padding: 0 16px;
    font-size: 14px;
    font-weight: 500;
    color: #4f4f4f;
    cursor: pointer;
    height: 100%;
    transition: color 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  .date-range-option:hover {
    color: #2196f3;
  }

  .date-range-option.active {
    color: #2196f3;
    font-weight: 600;
  }

  .separator {
    width: 1px;
    height: 20px;
    background-color: rgba(0, 0, 0, 0.1);
  }
</style>
