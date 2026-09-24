<script lang="ts">
  const version = import.meta.env.VITE_APP_VERSION ?? "dev";
  const releaseUrl = `https://github.com/leorolland/sortir.in/releases/tag/v${version}`;
</script>

<a class="app-version" href={releaseUrl} target="_blank" rel="noreferrer">
  v{version}
</a>

<style lang="scss">
  .app-version {
    position: absolute;
    bottom: calc(34px + env(safe-area-inset-bottom));
    right: calc(10px + env(safe-area-inset-right));
    z-index: 100;
    padding: 2px 10px;
    border-radius: 20px;
    font-size: 0.75rem;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    color: #555;
    text-decoration: none;
    background-color: rgba(240, 240, 245, 0.7);
    backdrop-filter: blur(5px);
    -webkit-backdrop-filter: blur(5px);
    box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    transition: all 0.2s ease;

    &:hover {
      color: #007aff;
      background-color: rgba(240, 240, 245, 0.9);
    }
  }
</style>
