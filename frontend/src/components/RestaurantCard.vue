<script setup lang="ts">
import Fa7SolidStar from '~icons/fa7-solid/star';
import { computed } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { useFavorites } from '../stores/useFavorites';
import type { RestaurantRecord } from '../models/restaurant';
import { useRestaurants } from '../stores/useRestaurants';
import { useNow } from '../composables/useNow';
import { getLatestMenu } from '../utils/menu';
import { formatRelativePastLabel } from '../utils/date';
import { getMenuFreshnessMeta } from '../utils/menuFreshness';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { getFileUrl, applySearch, getRestaurantDistanceKm, sortBy, coords } = useRestaurants();
const { isFavorite, toggleFavorite } = useFavorites();

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

const nowMs = useNow(30_000);
const currentWeekday = computed(() => WEEKDAYS[new Date(nowMs.value).getDay()]);
const isClosed = computed(() => props.restaurant.rest_days.includes(currentWeekday.value));
const isFavorited = computed(() => isFavorite(props.restaurant.id));
const thumbnailUrl = computed(() => getFileUrl(props.restaurant));
const latestMenuCreated = computed(() => {
  return getLatestMenu(props.restaurant.expand?.menus)?.created ?? null;
});
const lastCheck = computed(() => {
  return props.restaurant.last_check ?? null;
});
const menuFreshness = computed(() => {
  if (!latestMenuCreated.value) return null;
  return getMenuFreshnessMeta({
    menuDate: latestMenuCreated.value,
    cron: props.restaurant.cron,
    method: props.restaurant.method,
    lastCheck: lastCheck.value,
    now: nowMs.value,
  });
});
const lastCheckText = computed(() => {
  if (!lastCheck.value) return '';
  return `${formatRelativePastLabel(lastCheck.value.at, nowMs.value)} zuletzt versucht`;
});
const distanceKm = computed(() => getRestaurantDistanceKm(props.restaurant));
const showDistance = computed(() => sortBy.value === 'distance-asc' && coords.value && distanceKm.value !== null);
const distanceLabel = computed(() => {
  if (distanceKm.value === null) return '';

  if (distanceKm.value < 1) {
    return `~${Math.round(distanceKm.value * 1000)} m`;
  }

  if (distanceKm.value < 10) {
    return `~${distanceKm.value.toFixed(1)} km`;
  }

  return `~${Math.round(distanceKm.value)} km`;
});
const lastCheckTitle = computed(() => {
  if (!lastCheck.value) return '';
  if (lastCheck.value.status === 'success') {
    return 'Erfolgreich aktualisiert';
  }
  if (lastCheck.value.status === 'not_changed') {
    return 'Keine Änderung gefunden';
  }
  if (lastCheck.value.status === 'error') {
    return `Fehler${lastCheck.value.detail ? `: ${lastCheck.value.detail}` : ''}`;
  }
  return '';
});

function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('');
}
</script>

<template>
  <article
    class="group card card-border overflow-hidden rounded-xl bg-base-100 opacity-80 shadow-md transition-[shadow,opacity] duration-200 hover:opacity-100 hover:shadow-xl"
  >
    <figure class="relative h-30 overflow-hidden bg-base-300">
      <img
        v-if="thumbnailUrl"
        :src="thumbnailUrl"
        :alt="props.restaurant.name"
        :class="['h-full w-full object-cover transition-transform duration-500 group-hover:scale-105', isClosed ? 'opacity-40 grayscale' : '']"
        loading="lazy"
      />
      <div
        v-else
        :class="['flex h-full w-full items-center justify-center text-2xl font-semibold', isClosed ? 'opacity-40 grayscale' : 'text-base-content/70']"
      >
        {{ getInitials(props.restaurant.name) }}
      </div>

      <div class="absolute inset-x-0 top-0 flex items-start justify-between px-3 pt-3">
        <div class="flex flex-col items-start gap-1.5">
          <span v-if="isClosed" class="badge badge-sm badge-info backdrop-blur">Heute geschlossen</span>
          <span
            v-else-if="menuFreshness"
            :class="['badge badge-sm', menuFreshness.className]"
            :title="menuFreshness.title"
            >{{ menuFreshness.label }}</span
          >
          <span v-if="showDistance" class="badge badge-sm badge-neutral/85 backdrop-blur w-fit">{{ distanceLabel }}</span>
        </div>
        <button
          type="button"
          :class="[
            'cursor-pointer border-0 bg-transparent p-0 text-lg leading-none drop-shadow-sm transition-colors opacity-80 hover:opacity-100 focus:outline-none',
            isFavorited ? 'text-warning' : 'text-white hover:text-warning',
          ]"
          :aria-label="isFavorited ? 'Favorit entfernen' : 'Als Favorit markieren'"
          :aria-pressed="isFavorited"
          @click="toggleFavorite(props.restaurant.id)"
        >
          <Fa7SolidStar aria-hidden="true" />
        </button>
      </div>
      <!-- bottom row: tags -->
      <div class="absolute inset-x-0 bottom-0 bg-linear-to-t from-base-100/90 to-transparent px-3 pb-3 pt-8">
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="tag in props.restaurant.tags"
            :key="tag"
            type="button"
            class="badge badge-outline badge-xs cursor-pointer border-base-300/60 bg-base-100/70 px-2 py-2.5 text-xs font-medium backdrop-blur transition-colors hover:bg-base-100"
            :title="`Nach Tag ${tag} filtern`"
            @click="applySearch(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </figure>

    <div class="card-body gap-3 p-3">
      <div>
        <h3 class="text-base font-semibold leading-tight">
          {{ props.restaurant.name }}
        </h3>

        <p v-if="lastCheck" class="text-xs text-base-content/65" :title="lastCheckTitle" aria-label="Letzter Pruefstatus">
          {{ lastCheckText }}
        </p>
        <p class="text-xs text-base-content/65" v-else>Noch nicht geprüft</p>
      </div>

      <RestaurantActions :restaurant="props.restaurant" />
    </div>
  </article>
</template>
