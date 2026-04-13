<script setup lang="ts">
import Fa7SolidStar from '~icons/fa7-solid/star';
import { computed } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { useFavorites } from '../stores/useFavorites';
import type { RestaurantRecord } from '../models/restaurant';
import { useRestaurants } from '../stores/useRestaurants';
import { useNow } from '../composables/useNow';
import { getLatestMenu } from '../utils/menu';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { getFileUrl, applySearch } = useRestaurants();
const { isFavorite, toggleFavorite } = useFavorites();

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const RELATIVE_TIME_UNITS: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 60 * 60 * 24 * 365],
  ['month', 60 * 60 * 24 * 30],
  ['week', 60 * 60 * 24 * 7],
  ['day', 60 * 60 * 24],
  ['hour', 60 * 60],
  ['minute', 60],
];

const relativeTimeFormatter = new Intl.RelativeTimeFormat('de', {
  numeric: 'auto',
  style: 'long',
});

const nowMs = useNow(30_000);
const currentWeekday = computed(() => WEEKDAYS[new Date(nowMs.value).getDay()]);
const isClosed = computed(() => props.restaurant.rest_days.includes(currentWeekday.value));
const isFavorited = computed(() => isFavorite(props.restaurant.id));
const thumbnailUrl = computed(() => getFileUrl(props.restaurant));
const latestMenuCreated = computed(() => {
  return getLatestMenu(props.restaurant.expand?.menus)?.created ?? null;
});
const lastCheck = computed(() => {
  const lc = props.restaurant.last_check;
  return lc ?? null;
});
const lastCheckText = computed(() => {
  if (!lastCheck.value) return '';
  return `${formatRelativeDate(lastCheck.value.at)} zuletzt versucht`;
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

function formatRelativeDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Unbekannt';

  const rawDiffSeconds = Math.round((date.getTime() - nowMs.value) / 1000);
  // Backend and client clocks can drift slightly; avoid showing future times in the UI.
  const diffSeconds = Math.min(0, rawDiffSeconds);

  if (diffSeconds > -60) return 'gerade eben';

  for (const [unit, seconds] of RELATIVE_TIME_UNITS) {
    if (Math.abs(diffSeconds) >= seconds) {
      return relativeTimeFormatter.format(Math.round(diffSeconds / seconds), unit);
    }
  }

  return 'gerade eben';
}

function getRelativeDateBadgeClass(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'badge-neutral';

  const ageMs = nowMs.value - date.getTime();
  const dayMs = 24 * 60 * 60 * 1000;
  const weekMs = 7 * dayMs;
  const monthMs = 30 * dayMs;

  if (ageMs <= dayMs) return 'badge-success';
  if (ageMs <= 2 * dayMs) return 'badge-info';
  if (ageMs <= weekMs) return 'badge-warning';
  if (ageMs <= monthMs) return 'badge-error';
  return 'badge-neutral';
}

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
        <span v-if="isClosed" class="badge badge-sm badge-error backdrop-blur">Heute geschlossen</span>
        <span v-else-if="latestMenuCreated" :class="['badge badge-sm backdrop-blur', getRelativeDateBadgeClass(latestMenuCreated)]">{{
          formatRelativeDate(latestMenuCreated)
        }}</span>
        <span v-else />
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
        <p v-else>
          <span class="text-xs text-base-content/65">Noch nicht geprüft</span>
        </p>
      </div>

      <RestaurantActions :restaurant="props.restaurant" />
    </div>
  </article>
</template>
