<script setup lang="ts">
import { useMagicKeys, whenever } from '@vueuse/core';
import { computed, ref } from 'vue';
import { RouterLink, useRoute, useRouter } from 'vue-router';
import LoginModal from './LoginModal.vue';
import RestaurantSearchControls from './RestaurantSearchControls.vue';
import NavbarRestaurantIdentity from './NavbarRestaurantIdentity.vue';

import { AppVersion, BackendURL, BuyMeACoffeeURL, RepoURL } from '../config';

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

function navigateBack() {
  if (router.options.history.state.back) {
    router.back();
    return;
  }

  void router.replace({ name: 'home', query: route.query });
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
          <div class="flex w-full min-w-0 items-center gap-2">
            <button class="btn btn-soft btn-square shrink-0" aria-label="Zur Startseite zurück" @click="navigateBack">
              <span class="icon-[fa7-solid--arrow-left] size-4" aria-hidden="true"></span>
            </button>
            <NavbarRestaurantIdentity class="min-w-0 flex-1" />
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
        <div class="flex flex-col items-center gap-2 md:flex-row">
          <RouterLink to="/datenschutz" class="btn btn-ghost">Datenschutz</RouterLink>
          <div class="flex items-center gap-2">
            <a target="_blank" rel="noreferrer" class="btn btn-circle btn-ghost" :href="adminPanelUrl" title="Admin Panel">
              <span class="icon-[simple-icons--pocketbase] size-5" aria-hidden="true"></span>
            </a>
            <a target="_blank" rel="noreferrer" class="btn btn-circle btn-ghost" :href="RepoURL" :title="AppVersion">
              <span class="icon-[simple-icons--github] size-5" aria-hidden="true"></span>
            </a>
          </div>
          <a
            v-if="BuyMeACoffeeURL"
            target="_blank"
            rel="noreferrer"
            class="btn w-56 max-w-full border-[#ffdd00] bg-[#ffdd00] text-black hover:border-[#ffe74d] hover:bg-[#ffe74d] md:w-auto"
            :href="BuyMeACoffeeURL"
            title="Kaffee spendieren"
          >
            <span class="icon-[simple-icons--buymeacoffee] size-5" aria-hidden="true"></span>
            Kaffee spendieren
          </a>
        </div>
      </div>
    </footer>
  </div>
</template>
