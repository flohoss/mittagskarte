<script setup lang="ts">
import Fa7SolidMagnifyingGlass from '~icons/fa7-solid/magnifying-glass';
import Fa7SolidArrowLeft from '~icons/fa7-solid/arrow-left';
import SimpleIconsPocketbase from '~icons/simple-icons/pocketbase';
import Github from '~icons/simple-icons/github';
import { useMagicKeys, whenever } from '@vueuse/core';
import { computed, ref } from 'vue';
import { RouterLink, useRoute } from 'vue-router';
import LoginModal from './LoginModal.vue';

import { AppVersion, BackendURL, RepoURL } from '../config';
import { useRestaurants } from '../stores/useRestaurants';

defineProps<{
  title: string;
  description: string;
}>();

const route = useRoute();
const { searchQuery } = useRestaurants();
const searchInput = ref<HTMLInputElement | null>(null);
const keys = useMagicKeys({
  passive: false,
  onEventFired(event) {
    if (event.type !== 'keydown' || event.key.toLowerCase() !== 'k') {
      return;
    }

    if (isMenuHistoryRoute.value) {
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

const isMenuHistoryRoute = computed(() => route.name === 'restaurant-menu-history');

function focusSearch() {
  searchInput.value?.focus();
  searchInput.value?.select();
}

whenever(
  () => Boolean(keys['Meta+K']?.value || keys['Ctrl+K']?.value),
  () => {
    if (isMenuHistoryRoute.value) {
      return;
    }
    focusSearch();
  }
);
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <header class="navbar bg-base-200 shadow-sm">
      <div class="container py-2">
        <template v-if="isMenuHistoryRoute">
          <div class="flex w-full items-center justify-between gap-4">
            <RouterLink to="/" class="btn btn-soft shrink-0" aria-label="Zur Startseite zurück">
              <Fa7SolidArrowLeft class="size-4" aria-hidden="true" />
              Zurück
            </RouterLink>

            <div class="shrink-0">
              <LoginModal />
            </div>
          </div>
        </template>

        <template v-else>
          <div class="flex flex-col items-stretch gap-3 md:flex-row md:items-center md:justify-between md:gap-4">
            <RouterLink to="/" class="flex items-center gap-2 shrink-0">
              <img src="/static/schniddzl.webp" :alt="title" class="size-10" />
              <div class="font-light font-stretch-semi-expanded min-w-0">
                <div class="text-xl font-semibold truncate">{{ title }}</div>
                <div class="text-xs truncate">{{ description }}</div>
              </div>
            </RouterLink>

            <div class="flex w-full items-center gap-2 md:max-w-xl">
              <label
                class="input w-full rounded-lg relative flex items-center gap-3 transition-all duration-200 focus-within:ring-2 focus-within:ring-primary/20"
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

              <LoginModal />
            </div>
          </div>
        </template>
      </div>
    </header>
    <main class="my-5 lg:my-10 grow">
      <div class="container">
        <slot />
      </div>
    </main>
    <footer class="navbar bg-base-200">
      <div class="container flex flex-col items-center gap-3 py-2 text-center md:flex-row md:items-center md:justify-between md:text-left">
        <div class="text-sm leading-relaxed">{{ title }} - {{ description }}</div>
        <div class="flex gap-2 items-center">
          <a target="_blank" rel="noreferrer" class="btn btn-circle btn-ghost" :href="BackendURL + '/_/'" title="Admin Panel">
            <SimpleIconsPocketbase class="size-5" />
          </a>
          <a target="_blank" rel="noreferrer" class="btn btn-circle btn-ghost" :href="RepoURL" :title="AppVersion">
            <Github class="size-5" />
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>
