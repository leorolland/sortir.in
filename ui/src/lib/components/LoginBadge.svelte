<script lang="ts">
  import { onDestroy } from "svelte";
  import { authModel, client, webauthnRegister } from "../pocketbase";
  import Alerts, { alerts } from "./Alerts.svelte";
  import Dialog from "./Dialog.svelte";
  import LoginForm from "./LoginForm.svelte";
  const { signupAllowed = true } = $props();
  async function logout() {
    client.authStore.clear();
  }
  const unsubscribe = client.authStore.onChange((token, model) => {
    if (model) {
      const { name, username } = model;
      alerts.success(`Signed in as ${name || username || "Admin"}`, 5000);
    } else {
      alerts.success(`Signed out`, 5000);
    }
  }, false);
  onDestroy(() => {
    unsubscribe();
  });
</script>

{#if $authModel}
  <Dialog>
    {#snippet trigger(show)}
      <button class="badge" onclick={show}>
        {#if $authModel.avatar}
          <img
            src={client.getFileUrl($authModel, $authModel.avatar)}
            alt="profile pic"
          />
        {/if}
        <samp
          >{$authModel?.name || $authModel?.username || $authModel?.email}</samp
        >
      </button>
    {/snippet}
    <div class="wrapper">
      <div class="badge">
        {#if $authModel.avatar}
          <img
            src={client.getFileUrl($authModel, $authModel.avatar)}
            alt="profile pic"
          />
        {/if}
        <samp
          >{$authModel?.name ?? $authModel?.username ?? $authModel?.email}</samp
        >
      </div>
      <button onclick={() => webauthnRegister($authModel?.email)}
        >Register Passkey</button
      >
      <button onclick={logout}>Sign Out</button>
    </div>
  </Dialog>
{:else}
  <Dialog>
    {#snippet trigger(show)}
      <button class="login-button" onclick={show}>
        {signupAllowed ? "Connexion" : "Connexion"}
      </button>
    {/snippet}
    <Alerts />
    <LoginForm {signupAllowed} />
  </Dialog>
{/if}

<style lang="scss">
  .badge {
    padding: 0;
    background-color: transparent;
    cursor: pointer;
    display: flex;
    align-items: center;
    gap: 8px;

    > img {
      height: 2.2em;
      width: 2.2em;
      border-radius: 50%;
      object-fit: cover;
      border: 2px solid rgba(255, 255, 255, 0.8);
      box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    }

    > samp {
      display: inline-block;
      border-radius: 20px;
      padding: 0.5rem 0.8rem;
      text-align: center;
      line-height: 1.5rem;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
      font-weight: 500;
      color: #333;
      background-color: rgba(240, 240, 245, 0.7);
      backdrop-filter: blur(5px);
      -webkit-backdrop-filter: blur(5px);
      box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
    }
  }

  .wrapper {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding: 8px 0;

    button {
      border-radius: 8px;
      padding: 8px 16px;
      background-color: rgba(240, 240, 245, 0.7);
      color: #007AFF;
      font-weight: 500;
      border: none;
      transition: all 0.2s ease;

      &:hover {
        background-color: rgba(0, 122, 255, 0.1);
      }
    }
  }

  .login-button {
    border-radius: 20px;
    padding: 8px 16px;
    background-color: rgba(0, 122, 255, 0.1);
    color: #007AFF;
    font-weight: 500;
    border: none;
    transition: all 0.2s ease;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;

    &:hover {
      background-color: rgba(0, 122, 255, 0.2);
    }
  }
</style>
