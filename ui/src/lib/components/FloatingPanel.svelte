<script lang="ts">
  /**
   * Reusable component for floating panels with blur effect and scrollable content
   * @prop {boolean} compact - Compact style (for popups)
   * @prop {boolean} withAnimation - Add an appearance animation
   * @prop {boolean} scrollable - Allow scrolling of content
   * @prop {string} className - Additional CSS classes
   */
  export let compact = false;
  export let withAnimation = false;
  export let scrollable = false;
  export let className = '';
</script>

<div class="floating-panel {compact ? 'compact' : ''} {withAnimation ? 'with-animation' : ''} {scrollable ? 'scrollable' : ''} {className}">
  <div class="floating-panel-content">
    <slot />
  </div>
</div>

<style>
  .floating-panel {
    background: rgba(240, 240, 245, 0.7);
    backdrop-filter: blur(15px);
    -webkit-backdrop-filter: blur(15px);
    box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
    border-radius: 24px;
    overflow: hidden;
    transition: transform 100ms ease, opacity 200ms ease;
    display: flex;
    flex-direction: column;
    color: #4f4f4f;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
  }

  .floating-panel.scrollable {
    height: 100%;
    max-height: 100%;
    overflow: hidden;
  }

  .floating-panel-content {
    padding: 24px;
    box-sizing: border-box;
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    scrollbar-width: none; /* Firefox */
    -ms-overflow-style: none; /* IE and Edge */
    height: 100%;
    max-width: none;
    width: auto;
  }

  .floating-panel-content::-webkit-scrollbar {
    display: none; /* Chrome, Safari and Opera */
  }

  /* Variante plus petite pour les popups */
  .floating-panel.compact {
    border-radius: 16px;
  }

  .floating-panel.compact .floating-panel-content {
    padding: 16px;
  }

  /* Animation d'apparition */
  @keyframes floatingPanelFadeIn {
    from {
      opacity: 0;
      transform: translateY(10px);
    }
    to {
      opacity: 1;
      transform: translateY(0);
    }
  }

  .floating-panel.with-animation {
    animation: floatingPanelFadeIn 0.3s ease-out;
  }
</style>
