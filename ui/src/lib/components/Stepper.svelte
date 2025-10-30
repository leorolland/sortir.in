<script lang="ts">
  interface Step {
    name: string;
    label: string;
  }
  interface StepperProps {
    data?: Record<string, any>;
    steps: Step[];
    currentStep?: number;
    interactive?: boolean;
    hideLabels?: boolean;
    children?: (props: {
      data: Record<string, any>;
      index: number;
      total: number;
      step: Step;
      next: () => void;
      previous: () => void;
    }) => any;
  }
  let {
    data = $bindable({}),
    steps = [],
    currentStep = 0,
    interactive = false,
    hideLabels = false,
    children,
  }: StepperProps = $props();
  let next = () => {
    if (currentStep < steps.length - 1) {
      currentStep += 1;
    }
  };
  let previous = () => {
    if (currentStep > 0) {
      currentStep -= 1;
    }
  };
</script>

<ol role="tablist">
  {#each steps as { label }, index}
    {@const active = index === currentStep}
    {#if interactive}
      <button class:active onclick={() => (currentStep = index)}>
        {#if !hideLabels}
          {label}
        {/if}
      </button>
    {:else}
      <li class:active>
        {#if !hideLabels}
          {label}
        {/if}
      </li>
    {/if}
  {/each}
</ol>

{#if children}
  {@render children({
    data,
    index: currentStep,
    total: steps.length,
    step: steps[currentStep],
    next,
    previous,
  })}
{/if}
