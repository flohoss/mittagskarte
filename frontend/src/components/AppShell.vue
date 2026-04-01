<script setup lang="ts">
import Fa7SolidMagnifyingGlass from '~icons/fa7-solid/magnifying-glass';
import Github from '~icons/simple-icons/github';
import { useMagicKeys, whenever } from '@vueuse/core';
import { computed, ref } from 'vue';

import { useRestaurants } from '../stores/useRestaurants';

defineProps<{
  title: string;
  description: string;
}>();

const { searchQuery } = useRestaurants();
const searchInput = ref<HTMLInputElement | null>(null);
const keys = useMagicKeys({
  passive: false,
  onEventFired(event) {
    if (event.type !== 'keydown' || event.key.toLowerCase() !== 'k') {
      return;
    }

    if (event.metaKey || event.ctrlKey) {
      event.preventDefault();
    }
  },
});

const searchShortcut = computed(() => {
  if (typeof navigator === 'undefined') {
    return { primary: 'Ctrl', key: 'K' };
  }

  const platform = navigator.userAgent || '';
  const isMac = /mac/i.test(platform);

  return isMac ? { primary: '⌘', key: 'K' } : { primary: 'Ctrl', key: 'K' };
});

function focusSearch() {
  searchInput.value?.focus();
  searchInput.value?.select();
}

whenever(
  () => Boolean(keys['Meta+K']?.value || keys['Ctrl+K']?.value),
  () => {
    focusSearch();
  }
);
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <header class="navbar bg-base-200 shadow-sm">
      <div class="container flex flex-col items-stretch gap-3 py-2 md:flex-row md:items-center md:justify-between md:gap-4">
        <a href="/" class="flex items-center gap-2 shrink-0">
          <img src="/static/schniddzl.webp" :alt="title" class="size-10" />
          <div class="font-light font-stretch-semi-expanded min-w-0">
            <div class="text-xl font-semibold truncate">{{ title }}</div>
            <div class="text-xs truncate">{{ description }}</div>
          </div>
        </a>

        <label
          class="input w-full rounded-lg relative flex items-center gap-3 transition-all duration-200 focus-within:ring-2 focus-within:ring-primary/20 md:max-w-md"
        >
          <Fa7SolidMagnifyingGlass class="size-4 text-base-content/60 shrink-0" aria-hidden="true" />
          <input
            id="search-input"
            ref="searchInput"
            v-model="searchQuery"
            type="search"
            name="q"
            class="grow bg-transparent border-0 outline-0 focus:ring-0 placeholder:text-base-content/50"
            placeholder="Restaurant suchen..."
            autocomplete="off"
            spellcheck="false"
            aria-label="Restaurants suchen"
          />
          <kbd class="hidden lg:inline-flex kbd kbd-sm font-mono opacity-50">
            <span class="me-1 text-sm">{{ searchShortcut.primary }}</span
            >{{ searchShortcut.key }}
          </kbd>
        </label>
      </div>
    </header>
    <main class="my-5 lg:my-10 grow">
      <div class="container">
        <slot />
      </div>
    </main>
    <footer class="navbar bg-base-200">
      <div class="container flex flex-col items-center gap-3 py-2 text-center md:flex-row md:items-center md:justify-between md:text-left">
        <div class="text-sm leading-relaxed">
          {{ title }} - {{ description }} -
          <a class="link" title="Source Code auf GitHub" target="_blank" data-lg-blank="" href="https://github.com/flohoss/mittagskarte">Source Code</a>
        </div>
        <a target="_blank" data-lg-blank="" class="btn btn-circle btn-ghost" title="GitHub" href="https://github.com/flohoss/">
          <Github class="size-5" />
        </a>
      </div>
    </footer>
  </div>
</template>
