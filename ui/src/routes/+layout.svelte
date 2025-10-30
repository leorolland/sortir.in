<script lang="ts">
  import "../app.scss";
  import { base } from "$app/paths";
  import { page } from "$app/stores";
  import Alerts from "$lib/components/Alerts.svelte";
  import LoginBadge from "$lib/components/LoginBadge.svelte";
  import Nav from "$lib/components/Nav.svelte";
  import { metadata } from "$lib/metadata";
  const { data, children } = $props();
  const config = $derived(data.config ?? {});

  $effect(() => {
    if ($page.error) {
      $metadata.title = $page.error.message;
    }
  });
</script>

<svelte:head>
  <title>{$metadata.title} | {config.site?.name}</title>
</svelte:head>

<header>
  <div class="container header-content">
    <a href={`${base}/`} class="logo">
      <img src={`${base}/favicon.svg`} alt="application logo" />
    </a>
    <Nav />
    <LoginBadge signupAllowed={config.signupAllowed} />
  </div>
</header>
<div class="alerts-container">
  <Alerts />
</div>
<main class="fullscreen">
  {@render children()}
</main>

<style lang="scss">
  header {
    background-color: #fff;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1000;
    height: var(--header-height);

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: center;
      height: 100%;
    }

    .logo {
      width: 2rem;
      height: 2rem;
    }
  }

  .alerts-container {
    position: fixed;
    top: var(--header-height);
    left: 0;
    right: 0;
    z-index: 900;
    padding: 0 1rem;
  }

  main.fullscreen {
    flex-grow: 1;
    padding-top: var(--header-height);
    height: 100%;
    width: 100%;
  }
</style>
