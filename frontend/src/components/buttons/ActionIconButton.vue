<script setup lang="ts">
const props = withDefaults(
  defineProps<{
    href?: string;
    className: string;
    title: string;
    ariaLabel: string;
    target?: string;
    rel?: string;
    disabled?: boolean;
  }>(),
  {
    href: '',
    target: '_blank',
    rel: 'noreferrer',
    disabled: false,
  }
);
</script>

<template>
  <a
    v-if="props.href && !props.disabled"
    :href="props.href"
    :target="props.target"
    :rel="props.rel"
    :class="props.className"
    :title="props.title"
    :aria-label="props.ariaLabel"
  >
    <slot />
  </a>

  <button v-else type="button" :class="props.className" :title="props.title" :aria-label="props.ariaLabel" disabled>
    <slot />
  </button>
</template>
