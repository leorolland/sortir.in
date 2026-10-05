<script lang="ts">
  import { base } from "$app/paths";
  import { fly } from "svelte/transition";
  import { cubicOut } from "svelte/easing";
  import CrossIcon from "./CrossIcon.svelte";

  /*
    Self-contained iOS "Add to Home Screen" invitation: iOS Safari only,
    never in standalone; 7-day cooldown once hidden (dismiss or auto-hide).
  */
  const HIDDEN_KEY = "sortir:a2hs-hidden-at";
  const COOLDOWN_DAYS = 7;
  const SHOW_DELAY_MS = 6000;
  const AUTO_HIDE_MS = 30000;

  let visible = $state(false);

  function flyParams() {
    const reducedMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    return { y: 28, duration: reducedMotion ? 0 : 320, easing: cubicOut };
  }

  // iPadOS 13+ reports a Mac user agent: trust touch points over the UA string
  function isIosSafari(): boolean {
    const ua = navigator.userAgent;
    const ios =
      /iPad|iPhone|iPod/.test(ua) || (ua.includes("Macintosh") && navigator.maxTouchPoints > 1);
    const thirdPartyBrowser = /CriOS|FxiOS|EdgiOS|OpiOS/.test(ua);
    const standalone =
      (navigator as { standalone?: boolean }).standalone === true ||
      window.matchMedia("(display-mode: standalone)").matches;
    return ios && !thirdPartyBrowser && !standalone;
  }

  function hiddenByCooldown(): boolean {
    const hiddenAt = Number(localStorage.getItem(HIDDEN_KEY) ?? 0);
    return Date.now() - hiddenAt < COOLDOWN_DAYS * 86_400_000;
  }

  $effect(() => {
    // Dev builds ignore cooldowns: iterating on the prompt must not require
    // clearing storage on the device
    if (!import.meta.env.DEV && hiddenByCooldown()) return;
    if (!isIosSafari()) return;
    const timer = setTimeout(() => (visible = true), SHOW_DELAY_MS);
    return () => clearTimeout(timer);
  });

  $effect(() => {
    if (!visible) return;
    const timer = setTimeout(hide, AUTO_HIDE_MS);
    return () => clearTimeout(timer);
  });

  function hide() {
    visible = false;
    localStorage.setItem(HIDDEN_KEY, String(Date.now()));
  }
</script>

<!--
  @component
  Discreet invitation to install the app on the iOS home screen, teaching the
  Share → "Sur l'écran d'accueil" gesture (iOS offers no programmatic prompt).
  Anchored above the "…" button of Safari's bottom bar; never shows in
  standalone mode.
-->
{#if visible}
  <aside
    class="install-prompt"
    transition:fly={flyParams()}
    aria-label="Installer l'application web"
  >
    <div class="row">
      <img class="app-icon" src="{base}/apple-touch-icon.png" alt="" />
      <div class="titles">
        <strong>sortir.in</strong>
        <span>Installer la webapp</span>
      </div>
      <button
        class="close"
        type="button"
        onclick={hide}
        aria-label="Ne plus afficher"
      >
        <CrossIcon size={9} thickness={1.5} />
      </button>
    </div>
    <div class="steps">
      <span class="chip">
        <svg
          class="chip-icon"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          aria-hidden="true"
        >
          <path d="M4 12v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2v-8" />
          <polyline points="16 6 12 2 8 6" />
          <line x1="12" x2="12" y1="2" y2="15" />
        </svg>
        Partager
      </span>
      <svg
        class="steps-arrow"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2.5"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <path d="m9 18 6-6-6-6" />
      </svg>
      <span class="chip">Sur l'écran d'accueil</span>
    </div>
  </aside>
{/if}

<style>
  /* Anchored above the "…" button of Safari's bottom bar (right edge), with a
     caret pointing at it — a coach-mark, not a floating card */
  .install-prompt {
    position: fixed;
    left: auto;
    right: 10px;
    bottom: calc(12px + env(safe-area-inset-bottom));
    z-index: 700; /* above map controls, below the bottom sheet and alerts */
    width: min(266px, calc(100vw - 20px));
    padding: 12px 12px 11px;
    border-radius: 18px;
    border: 1px solid rgba(0, 0, 0, 0.06);
    background: rgba(255, 255, 255, 0.94);
    backdrop-filter: blur(20px) saturate(1.5);
    -webkit-backdrop-filter: blur(20px) saturate(1.5);
    box-shadow:
      0 12px 40px rgba(0, 0, 0, 0.16),
      0 2px 8px rgba(0, 0, 0, 0.06);
    color: #1c1c1e;
  }

  .install-prompt::after {
    content: '';
    position: absolute;
    bottom: -11px;
    right: 36px;
    border-left: 9px solid transparent;
    border-right: 9px solid transparent;
    border-top: 11px solid rgba(255, 255, 255, 0.94);
  }

  .row {
    display: flex;
    align-items: center;
    gap: 11px;
  }

  .app-icon {
    width: 35px;
    height: 35px;
    border-radius: 8.5px;
    border: 0.5px solid rgba(0, 0, 0, 0.08);
    flex-shrink: 0;
  }

  .titles {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
  }

  .titles strong {
    font-size: 15px;
    font-weight: 650;
    letter-spacing: -0.1px;
    line-height: 1.05;
  }

  .titles span {
    font-size: 16px;
    color: #6d6d72;
    line-height: 1.1;
    white-space: nowrap;
    /* glyphs render ~2px low in their line box: optical correction */
    transform: translateY(-6px);
  }

  .close {
    margin-left: auto;
    width: 24px;
    height: 24px;
    padding: 0;
    display: grid;
    place-items: center;
    border: none;
    border-radius: 50%;
    background: rgba(0, 0, 0, 0.05);
    color: #8e8e93;
    cursor: pointer;
    flex-shrink: 0;
  }

  .close:active {
    background: rgba(0, 0, 0, 0.1);
  }

  .steps {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 6px;
    margin-top: 11px;
    padding-top: 10px;
    border-top: 1px solid rgba(0, 0, 0, 0.06);
  }

  .chip {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    padding: 4px 8px;
    border-radius: 8px;
    background: rgba(0, 0, 0, 0.055);
    font-size: 11.5px;
    font-weight: 500;
    color: #3a3a3c;
    white-space: nowrap;
  }

  .chip-icon {
    width: 12px;
    height: 12px;
    color: #007aff;
  }

  .steps-arrow {
    width: 11px;
    height: 11px;
    color: #8e8e93;
    flex-shrink: 0;
  }
</style>
