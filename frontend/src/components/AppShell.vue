<script setup lang="ts">
import Fa7SolidArrowLeft from '~icons/fa7-solid/arrow-left';
import SimpleIconsPocketbase from '~icons/simple-icons/pocketbase';
import Github from '~icons/simple-icons/github';
import { useMagicKeys, whenever } from '@vueuse/core';
import { computed, ref } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import LoginModal from './LoginModal.vue';
import RestaurantSearchControls from './RestaurantSearchControls.vue';

import { AppVersion, BackendURL, RepoURL } from '../config';

defineProps<{
  title: string;
  description: string;
}>();

const route = useRoute();
const router = useRouter();
const searchControls = ref<InstanceType<typeof RestaurantSearchControls> | null>(null);
const loginModal = ref<InstanceType<typeof LoginModal> | null>(null);
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

const isMenuHistoryRoute = computed(() => route.name === 'restaurant-view');
const adminPanelUrl = computed(() => `${BackendURL}_/`);

function focusSearch() {
  searchControls.value?.focusSearch();
}

function openAuthModal() {
  loginModal.value?.open();
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
            <button class="btn btn-soft shrink-0" aria-label="Zur Startseite zurück" @click="router.back()">
              <Fa7SolidArrowLeft class="size-4" aria-hidden="true" />
              Zurück
            </button>

            <div class="shrink-0">
              <LoginModal />
            </div>
          </div>
        </template>

        <template v-else>
          <div class="flex items-center justify-between gap-3 flex-wrap sm:flex-nowrap">
            <RouterLink to="/" class="flex w-full sm:w-auto items-center gap-2 min-w-0">
              <img src="/static/schniddzl.webp" :alt="title" class="size-10" />
              <div class="font-light font-stretch-semi-expanded min-w-0 w-full sm:w-auto">
                <div class="text-xl font-semibold truncate">{{ title }}</div>
                <div class="text-xs truncate">{{ description }}</div>
              </div>
            </RouterLink>

            <div class="w-full sm:w-auto max-w-full">
              <RestaurantSearchControls ref="searchControls" :on-open-auth-modal="openAuthModal" />
              <LoginModal ref="loginModal" :show-trigger="false" />
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
          <RouterLink to="/datenschutz" class="btn btn-ghost">Datenschutz</RouterLink>
          <a target="_blank" rel="noreferrer" class="btn btn-circle btn-ghost" :href="adminPanelUrl" title="Admin Panel">
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
