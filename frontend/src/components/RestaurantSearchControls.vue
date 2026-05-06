<script setup lang="ts">
import Fa7SolidMagnifyingGlass from '~icons/fa7-solid/magnifying-glass';
import Fa7SolidSliders from '~icons/fa7-solid/sliders';
import Fa7SolidCircleExclamation from '~icons/fa7-solid/circle-exclamation';
import Fa7SolidXmark from '~icons/fa7-solid/xmark';
import { computed, ref } from 'vue';
import { useLogin } from '../stores/useLogin';
import { restaurantGroupingOptions, restaurantSortOptions, useRestaurants } from '../stores/useRestaurants';

const { searchQuery, sortBy, groupBy } = useRestaurants();
const { isAuthenticated, authIdentity } = useLogin();
const input = ref<HTMLInputElement | null>(null);
const settingsDrawerToggle = ref<HTMLInputElement | null>(null);

const searchShortcut = computed(() => {
  if (typeof navigator === 'undefined') {
    return { primary: 'Ctrl', key: 'K' };
  }

  const platform = navigator.userAgent || '';
  const isMac = /mac/i.test(platform);

  return isMac ? { primary: '⌘', key: 'K' } : { primary: 'Ctrl', key: 'K' };
});

const activeSettingsLabel = computed(() => {
  const sortLabel = restaurantSortOptions.find((option) => option.value === sortBy.value)?.label;
  const groupLabel = restaurantGroupingOptions.find((option) => option.value === groupBy.value)?.label;
  return [sortLabel, groupLabel].filter(Boolean).join(' • ');
});

const hasNonDefaultDisplaySettings = computed(() => {
  return sortBy.value !== 'name-asc' || groupBy.value !== 'group';
});

const accountIdentityLabel = computed(() => {
  return authIdentity.value || 'Kein Benutzer aktiv';
});

const accountInitials = computed(() => {
  const identity = accountIdentityLabel.value;
  if (!identity || identity === 'Kein Benutzer aktiv') {
    return '?';
  }

  const base = identity.includes('@') ? identity.split('@')[0] : identity;
  const parts = base.split(/[.\-_\s]+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return `${parts[0][0]}${parts[1][0]}`.toUpperCase();
});

function focusSearch() {
  input.value?.focus();
  input.value?.select();
}

function openSettings() {
  if (settingsDrawerToggle.value) {
    settingsDrawerToggle.value.checked = true;
  }
}

function closeSettings() {
  if (settingsDrawerToggle.value) {
    settingsDrawerToggle.value.checked = false;
  }
}

defineExpose({
  focusSearch,
  openSettings,
});
</script>

<template>
  <div class="drawer drawer-end w-full sm:w-auto overflow-visible">
    <input id="header-settings-drawer" ref="settingsDrawerToggle" type="checkbox" class="drawer-toggle" />

    <div class="drawer-content flex w-full items-center gap-2">
      <label class="input flex-1 min-w-0 sm:flex-none sm:w-64 md:w-72 rounded-lg relative flex items-center gap-3 focus-within:ring-2 focus-within:ring-primary/20">
        <Fa7SolidMagnifyingGlass class="size-4 text-base-content/60 shrink-0" aria-hidden="true" />
        <input
          id="search-input"
          ref="input"
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

      <button type="button" class="btn btn-soft btn-square shrink-0 relative" aria-label="Einstellungen öffnen" @click="openSettings">
        <Fa7SolidSliders class="size-4" aria-hidden="true" />
        <Fa7SolidCircleExclamation
          v-if="hasNonDefaultDisplaySettings"
          class="size-3.5 text-warning absolute -top-1.5 -right-1.5 drop-shadow-sm"
          aria-hidden="true"
        />
      </button>
    </div>

    <div class="drawer-side z-40">
      <label for="header-settings-drawer" aria-label="close sidebar" class="drawer-overlay" />
      <aside class="bg-base-100 min-h-full w-80 border-s border-base-300/60 sm:w-96 flex flex-col">
        <div class="p-4 grow overflow-y-auto">
          <div class="flex items-center justify-between gap-2">
            <h3 class="font-semibold text-lg">Anzeigeeinstellungen</h3>
            <button type="button" class="btn btn-sm btn-ghost btn-square" aria-label="Einstellungen schließen" @click="closeSettings">
              <Fa7SolidXmark class="size-4" aria-hidden="true" />
            </button>
          </div>
          <p class="text-sm text-base-content/70 mt-1">{{ activeSettingsLabel }}</p>

          <div class="mt-5 grid gap-4">
            <label class="form-control w-full">
              <span class="label-text text-sm mb-1">Sortierung</span>
              <select v-model="sortBy" class="select select-bordered w-full">
                <option v-for="option in restaurantSortOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>

            <label class="form-control w-full">
              <span class="label-text text-sm mb-1">Gruppierung</span>
              <select v-model="groupBy" class="select select-bordered w-full">
                <option v-for="option in restaurantGroupingOptions" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </label>
          </div>
        </div>

        <div class="border-t border-base-300/70 bg-base-200/60 px-4 py-4">
          <div v-if="isAuthenticated" class="grid gap-3">
            <div class="flex items-center gap-3">
              <div class="avatar avatar-placeholder">
                <div class="size-10 rounded-full text-sm font-semibold bg-success text-success-content">
                  <span>{{ accountInitials }}</span>
                </div>
              </div>
              <div class="min-w-0 grow">
                <p class="text-sm font-medium leading-tight">Konto</p>
                <p class="text-xs text-base-content/75 truncate">{{ accountIdentityLabel }}</p>
              </div>
            </div>

            <div class="w-full [&_.btn]:w-full">
              <slot name="auth" />
            </div>
          </div>

          <div v-else class="w-full [&_.btn]:w-full">
            <slot name="auth" />
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>
