<script lang="ts">
  let {
    size = 14,
    thickness = 2
  }: { size?: number; thickness?: number } = $props();
</script>

<!--
  @component
  Drawn cross (two rotated bars): geometrically centered, independent of
  font metrics — the × glyph sits slightly low in its line box.
  Color follows the parent's `color`.
-->
<span
  class="cross"
  style:width="{size}px"
  style:height="{size}px"
  style:--thickness="{thickness}px"
  aria-hidden="true"
></span>

<style>
  .cross {
    position: relative;
    display: block;
    flex-shrink: 0;
  }

  .cross::before,
  .cross::after {
    content: '';
    position: absolute;
    top: 50%;
    left: 50%;
    width: 100%;
    height: var(--thickness);
    border-radius: calc(var(--thickness) / 2);
    background-color: currentColor;
  }

  .cross::before {
    transform: translate(-50%, -50%) rotate(45deg);
  }

  .cross::after {
    transform: translate(-50%, -50%) rotate(-45deg);
  }
</style>
