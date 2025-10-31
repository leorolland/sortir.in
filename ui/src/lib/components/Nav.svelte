<script lang="ts">
  import { base } from "$app/paths";
  import { page } from "$app/stores";
  const links = [
    ["/", "Carte"],
  ];
</script>

<nav>
  {#each links as [path, label]}
    {@const active = $page.url.pathname == path}
    <a href={`${base}${path}`} class:active>{label}</a>
  {/each}
</nav>

<style lang="scss">
  nav {
    a {
      padding: 0.6em 1.2em;
      margin: 0;
      border-radius: 8px;
      color: #333;
      font-weight: 500;
      text-decoration: none;
      transition: all 0.2s ease;

      &:hover {
        background-color: rgba(0, 0, 0, 0.05);
      }

      &.active {
        background-color: rgba(0, 122, 255, 0.1);
        color: #007AFF;
        font-weight: 600;
      }
    }
  }
</style>
