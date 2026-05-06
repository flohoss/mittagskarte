<script setup lang="ts">
import Fa7SolidMagnifyingGlass from '~icons/fa7-solid/magnifying-glass';
import Fa7SolidArrowDownAZ from '~icons/fa7-solid/arrow-down-a-z';
import Fa7SolidArrowDownZA from '~icons/fa7-solid/arrow-down-z-a';
import Fa7SolidLocationDot from '~icons/fa7-solid/location-dot';
import MdiSortCalendarAscending from '~icons/mdi/sort-calendar-ascending';
import MdiSortCalendarDescending from '~icons/mdi/sort-calendar-descending';
import Fa7SolidLayerGroup from '~icons/fa7-solid/layer-group';
import Fa7SolidBan from '~icons/fa7-solid/ban';
import Fa7SolidUser from '~icons/fa7-solid/user';
import Fa7SolidUserShield from '~icons/fa7-solid/user-shield';
import Fa7SolidRightFromBracket from '~icons/fa7-solid/right-from-bracket';
import Fa7SolidXmark from '~icons/fa7-solid/xmark';
import { computed, ref } from 'vue';
import { useLogin } from '../stores/useLogin';
import { restaurantSortOptions, useRestaurants } from '../stores/useRestaurants';

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

const getSortIcon = (sortValue: string) => {
  switch (sortValue) {
    case 'name-asc':
      return Fa7SolidArrowDownAZ;
    case 'name-desc':
      return Fa7SolidArrowDownZA;
    case 'distance-asc':
      return Fa7SolidLocationDot;
    case 'menu-newest':
      return MdiSortCalendarDescending;
    case 'menu-oldest':
      return MdiSortCalendarAscending;
    default:
      return Fa7SolidArrowDownAZ;
  }
};

const sortIconComponent = computed(() => {
  return getSortIcon(sortBy.value);
});



const authStatus = computed(() => {
  return isAuthenticated.value ? '✓ Angemeldet' : '○ Gast';
});

const authIconComponent = computed(() => {
  return isAuthenticated.value ? Fa7SolidUserShield : Fa7SolidUser;
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
      <Fa7SolidMagnifyingGlass class="size-5 text-base-content/60 shrink-0" aria-hidden="true" />
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
        <Fa7SolidXmark class="size-4" aria-hidden="true" />
      </button>
      <kbd v-else class="hidden lg:inline-flex kbd kbd-sm font-mono opacity-50">
        <span class="me-1 text-sm">{{ searchShortcut.primary }}</span
        >{{ searchShortcut.key }}
      </kbd>
    </label>

    <!-- Sort Button -->
    <div class="dropdown dropdown-end">
      <button 
        type="button" 
        class="btn btn-soft btn-square shrink-0"
        :aria-label="`Sortierung: ${sortLabel}`" 
        tabindex="0"
      >
        <Transition name="sort-indicator" mode="out-in">
          <span
            v-if="geolocationLoading"
            key="spinner"
            class="sort-indicator sort-indicator-spinner loading loading-spinner size-5"
            aria-hidden="true"
          />
          <component v-else :is="sortIconComponent" key="icon" class="sort-indicator size-5" aria-hidden="true" />
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
            <component :is="getSortIcon(option.value)" class="size-4" aria-hidden="true" />
            <span>{{ option.label }}</span>
          </a>
        </li>
      </ul>
    </div>

    <!-- Group Button -->
    <label
      class="btn btn-soft btn-square shrink-0 swap swap-rotate"
      :class="{ 'swap-active': groupBy === 'group' }"
      :aria-label="groupBy === 'group' ? 'Gruppierung aufheben' : 'Nach Gruppe gruppieren'"
      @click="groupBy = groupBy === 'group' ? 'none' : 'group'"
    >
      <Fa7SolidLayerGroup class="swap-on size-5" aria-hidden="true" />
      <Fa7SolidBan class="swap-off size-5" aria-hidden="true" />
    </label>

    <!-- Auth Button -->
    <template v-if="isAuthenticated">
      <div class="dropdown dropdown-end">
        <button type="button" class="btn btn-soft btn-square btn-success shrink-0" :aria-label="authStatus" tabindex="0">
          <component :is="authIconComponent" class="size-5" aria-hidden="true" />
        </button>
        <ul tabindex="0" class="dropdown-content menu bg-base-100 rounded-box z-50 w-48 p-2 shadow border border-base-300/50">
          <li>
            <a @click="clearAuthentication" class="flex items-center gap-2 text-error">
              <Fa7SolidRightFromBracket class="size-5" aria-hidden="true" />
              <span>Abmelden</span>
            </a>
          </li>
        </ul>
      </div>
    </template>
    <template v-else>
      <button type="button" class="btn btn-soft btn-square btn-ghost shrink-0" :aria-label="authStatus" @click="onOpenAuthModal?.()">
        <component :is="authIconComponent" class="size-5" aria-hidden="true" />
      </button>
    </template>
  </div>
</template>

<style scoped>
.sort-indicator-enter-active,
.sort-indicator-leave-active {
  transition: opacity 140ms ease, transform 140ms ease;
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
