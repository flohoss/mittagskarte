<script setup lang="ts">
import { computed, ref } from 'vue';
import { useLogin } from '../stores/useLogin';
import { restaurantGroupingOptions, restaurantSortOptions, useRestaurants } from '../stores/useRestaurants';

const { searchQuery, sortBy, groupBy, requestGeolocation, geolocationLoading } = useRestaurants();
const { isAuthenticated, clearAuthentication } = useLogin();
const input = ref<HTMLInputElement | null>(null);

interface Props {
  onOpenAuthModal?: () => void;
}

defineProps<Props>();

const searchShortcut = computed(() => {
  if (typeof navigator === 'undefined') {
    return { primary: 'Ctrl', key: 'K' };
  }

  const platform = navigator.userAgent || '';
  const isMac = /mac/i.test(platform);

  return isMac ? { primary: '⌘', key: 'K' } : { primary: 'Ctrl', key: 'K' };
});

const hasSearchQuery = computed(() => searchQuery.value.trim().length > 0);

const sortLabel = computed(() => {
  return restaurantSortOptions.find((opt) => opt.value === sortBy.value)?.label || '';
});

const getSortIconClass = (sortValue: string) => {
  switch (sortValue) {
    case 'name-asc':
      return 'icon-[fa7-solid--arrow-down-a-z]';
    case 'name-desc':
      return 'icon-[fa7-solid--arrow-down-z-a]';
    case 'distance-asc':
      return 'icon-[fa7-solid--location-dot]';
    case 'menu-newest':
      return 'icon-[mdi--sort-calendar-descending]';
    case 'menu-oldest':
      return 'icon-[mdi--sort-calendar-ascending]';
    default:
      return 'icon-[fa7-solid--arrow-down-a-z]';
  }
};

const sortIconClass = computed(() => {
  return getSortIconClass(sortBy.value);
});

const groupLabel = computed(() => {
  return restaurantGroupingOptions.find((opt) => opt.value === groupBy.value)?.label || '';
});

const getGroupIconClass = (groupValue: string) => {
  switch (groupValue) {
    case 'group':
      return 'icon-[fa7-solid--layer-group]';
    case 'none-favorites':
      return 'icon-[fa7-solid--star]';
    case 'favorites':
      return 'icon-[fa7-solid--star]';
    case 'none':
    default:
      return 'icon-[fa7-solid--ban]';
  }
};

const groupIconClass = computed(() => {
  return getGroupIconClass(groupBy.value);
});

const authStatus = computed(() => {
  return isAuthenticated.value ? '✓ Angemeldet' : '○ Gast';
});

const authIconClass = computed(() => {
  return isAuthenticated.value ? 'icon-[fa7-solid--user-shield]' : 'icon-[fa7-solid--user]';
});

function focusSearch() {
  input.value?.focus();
  input.value?.select();
}

function clearSearch() {
  searchQuery.value = '';
  input.value?.focus();
}

function closeDropdown() {
  (document.activeElement as HTMLElement)?.blur();
}

defineExpose({
  focusSearch,
});
</script>

<template>
  <div class="flex w-full sm:w-auto items-center gap-2">
    <label
      class="input flex-1 min-w-0 sm:flex-none sm:w-64 md:w-72 rounded-lg relative flex items-center gap-3 focus-within:ring-2 focus-within:ring-primary/20"
    >
      <span class="icon-[fa7-solid--magnifying-glass] size-5 text-base-content/60 shrink-0" aria-hidden="true"></span>
      <input
        id="search-input"
        ref="input"
        v-model="searchQuery"
        type="text"
        name="q"
        class="grow bg-transparent border-0 outline-0 focus:ring-0 placeholder:text-base-content/50"
        placeholder="Restaurant suchen..."
        autocomplete="off"
        spellcheck="false"
        aria-label="Restaurants suchen"
        @keydown.esc.stop.prevent="clearSearch"
      />
      <button
        v-if="hasSearchQuery"
        type="button"
        class="btn btn-ghost btn-xs btn-circle"
        aria-label="Suche leeren"
        title="Suche leeren"
        @click.stop.prevent="clearSearch"
      >
        <span class="icon-[fa7-solid--xmark] size-4" aria-hidden="true"></span>
      </button>
      <kbd v-else class="hidden lg:inline-flex kbd kbd-sm font-mono opacity-50">
        <span class="me-1 text-sm">{{ searchShortcut.primary }}</span
        >{{ searchShortcut.key }}
      </kbd>
    </label>

    <!-- Sort Button -->
    <div class="dropdown dropdown-end">
      <button type="button" class="btn btn-soft btn-square shrink-0" :aria-label="`Sortierung: ${sortLabel}`" tabindex="0">
        <Transition name="sort-indicator" mode="out-in">
          <span v-if="geolocationLoading" key="spinner" class="sort-indicator sort-indicator-spinner loading loading-spinner size-5" aria-hidden="true" />
          <span v-else :class="['sort-indicator', sortIconClass, 'size-5']" key="icon" aria-hidden="true"></span>
        </Transition>
      </button>
      <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-50 w-52 p-2 shadow border border-base-300/50">
        <li v-for="option in restaurantSortOptions" :key="option.value">
          <a
            :class="{ active: sortBy === option.value }"
            @click="
              sortBy = option.value;
              if (option.value === 'distance-asc') requestGeolocation();
              closeDropdown();
            "
            class="flex items-center gap-2"
          >
            <span :class="[getSortIconClass(option.value), 'size-4']" aria-hidden="true"></span>
            <span>{{ option.label }}</span>
          </a>
        </li>
      </ul>
    </div>

    <!-- Group Button -->
    <div class="dropdown dropdown-end">
      <button type="button" class="btn btn-soft btn-square shrink-0" :aria-label="`Gruppierung: ${groupLabel}`" tabindex="0">
        <span :class="[groupIconClass, 'size-5']" aria-hidden="true"></span>
      </button>
      <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-50 w-52 p-2 shadow border border-base-300/50">
        <li v-for="option in restaurantGroupingOptions" :key="option.value">
          <a
            :class="{ active: groupBy === option.value }"
            @click="
              groupBy = option.value;
              closeDropdown();
            "
            class="flex items-center gap-2"
          >
            <span :class="[getGroupIconClass(option.value), 'size-4']" aria-hidden="true"></span>
            <span>{{ option.label }}</span>
          </a>
        </li>
      </ul>
    </div>

    <!-- Auth Button -->
    <template v-if="isAuthenticated">
      <div class="dropdown dropdown-end">
        <button type="button" class="btn btn-square btn-success shrink-0" :aria-label="authStatus" tabindex="0">
          <span :class="[authIconClass, 'size-5']" aria-hidden="true"></span>
        </button>
        <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-50 w-48 p-2 shadow border border-base-300/50">
          <li>
            <a @click="clearAuthentication" class="flex items-center gap-2 text-error">
              <span class="icon-[fa7-solid--right-from-bracket] size-5" aria-hidden="true"></span>
              <span>Abmelden</span>
            </a>
          </li>
        </ul>
      </div>
    </template>
    <template v-else>
      <button type="button" class="btn btn-soft btn-square shrink-0" :aria-label="authStatus" @click="onOpenAuthModal?.()">
        <span :class="[authIconClass, 'size-5']" aria-hidden="true"></span>
      </button>
    </template>
  </div>
</template>

<style scoped>
.sort-indicator-enter-active,
.sort-indicator-leave-active {
  transition:
    opacity 140ms ease,
    transform 140ms ease;
}

.sort-indicator-enter-from,
.sort-indicator-leave-to {
  opacity: 0;
  transform: scale(0.92);
}

.sort-indicator-spinner.sort-indicator-enter-active {
  transition-delay: 110ms;
}
</style>
