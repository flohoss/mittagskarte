<script setup lang="ts">
import Fa7SolidGlobe from '~icons/fa7-solid/globe';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import MenuPopover from './MenuPopover.vue';
import Fa7SolidPhone from '~icons/fa7-solid/phone';
import Fa7SolidMap from '~icons/fa7-solid/map';
import Fa7SolidStar from '~icons/fa7-solid/star';
import { computed } from 'vue';
import { useFavorites } from '../stores/useFavorites';
import type { RecordModel } from 'pocketbase';
import { useRestaurants } from '../stores/useRestaurants';

const props = defineProps<{
  restaurant: RecordModel;
}>();

const { getFileUrl, getMapUrl, getPhoneUrl } = useRestaurants();
const { isFavorite, toggleFavorite } = useFavorites();

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

const isClosed = computed(() => props.restaurant.rest_days.includes(WEEKDAYS[new Date().getDay()]));
const isFavorited = computed(() => isFavorite(props.restaurant.id));

const relativeTimeFormatter = new Intl.RelativeTimeFormat('de', {
  numeric: 'auto',
  style: 'long',
});

function formatRelativeDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Unbekannt';

  const diffSeconds = Math.round((date.getTime() - Date.now()) / 1000);
  const units: Array<[Intl.RelativeTimeFormatUnit, number]> = [
    ['year', 60 * 60 * 24 * 365],
    ['month', 60 * 60 * 24 * 30],
    ['week', 60 * 60 * 24 * 7],
    ['day', 60 * 60 * 24],
    ['hour', 60 * 60],
    ['minute', 60],
  ];

  for (const [unit, seconds] of units) {
    if (Math.abs(diffSeconds) >= seconds) {
      return relativeTimeFormatter.format(Math.round(diffSeconds / seconds), unit);
    }
  }

  return relativeTimeFormatter.format(diffSeconds, 'second');
}

function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('');
}

const menuDimensions = computed(() => {
  const raw = props.restaurant.menu_dimensions;

  if (!raw || typeof raw !== 'object') {
    return {
      width: null,
      height: null,
    };
  }

  const parsed = raw as Record<string, unknown>;
  const width = typeof parsed.width === 'number' && Number.isFinite(parsed.width) && parsed.width > 0 ? parsed.width : null;
  const height = typeof parsed.height === 'number' && Number.isFinite(parsed.height) && parsed.height > 0 ? parsed.height : null;

  return {
    width,
    height,
  };
});
</script>

<template>
  <article
    class="group card card-border overflow-hidden rounded-2xl bg-base-100 opacity-80 shadow-md transition-[shadow,opacity] duration-200 hover:opacity-100 hover:shadow-xl"
  >
    <figure class="relative h-30 overflow-hidden bg-base-300">
      <img
        v-if="getFileUrl(props.restaurant)"
        :src="getFileUrl(props.restaurant)"
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
        <span :class="['badge badge-sm backdrop-blur', isClosed ? 'badge-error' : 'badge-ghost']">
          {{ isClosed ? 'Heute geschlossen' : formatRelativeDate(props.restaurant.updated) }}
        </span>
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
          <span
            v-for="tag in props.restaurant.tags"
            :key="tag"
            class="badge badge-outline badge-xs border-base-300/60 bg-base-100/70 px-2 py-2.5 text-xs font-medium backdrop-blur"
          >
            {{ tag }}
          </span>
        </div>
      </div>
    </figure>

    <div class="card-body gap-3 p-3">
      <h3 class="text-base font-semibold leading-tight">{{ props.restaurant.name }}</h3>

      <div class="grid grid-cols-4 gap-1.5">
        <MenuPopover
          v-if="props.restaurant.menu"
          :menu-url="props.restaurant.menu"
          :menu-width="menuDimensions.width"
          :menu-height="menuDimensions.height"
        />
        <button v-else type="button" class="btn btn-primary" title="Keine Speisekarte verfügbar" aria-label="Keine Speisekarte verfügbar" disabled>
          <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
        </button>

        <a
          v-if="getMapUrl(props.restaurant)"
          :href="getMapUrl(props.restaurant)"
          target="_blank"
          rel="noreferrer"
          class="btn btn-soft hover:btn-warning"
          title="In Google Maps öffnen"
          aria-label="Karte öffnen"
        >
          <Fa7SolidMap class="btn-icon" aria-hidden="true" />
        </a>
        <button v-else type="button" class="btn btn-soft hover:btn-warning" title="Keine Karte verfügbar" aria-label="Keine Karte verfügbar" disabled>
          <Fa7SolidMap class="btn-icon" aria-hidden="true" />
        </button>

        <a
          v-if="props.restaurant.phone"
          :href="getPhoneUrl(props.restaurant)"
          class="btn btn-soft hover:btn-success"
          title="Anrufen"
          aria-label="Restaurant anrufen"
        >
          <Fa7SolidPhone class="btn-icon" aria-hidden="true" />
        </a>
        <button
          v-else
          type="button"
          class="btn btn-soft hover:btn-success"
          title="Keine Telefonnummer verfügbar"
          aria-label="Keine Telefonnummer verfügbar"
          disabled
        >
          <Fa7SolidPhone class="btn-icon" aria-hidden="true" />
        </button>

        <a
          v-if="props.restaurant.website"
          :href="props.restaurant.website"
          target="_blank"
          rel="noreferrer"
          class="btn btn-soft hover:btn-info"
          title="Website öffnen"
          aria-label="Website öffnen"
        >
          <Fa7SolidGlobe class="btn-icon" aria-hidden="true" />
        </a>
        <button v-else type="button" class="btn btn-soft hover:btn-info" title="Keine Website verfügbar" aria-label="Keine Website verfügbar" disabled>
          <Fa7SolidGlobe class="btn-icon" aria-hidden="true" />
        </button>
      </div>
    </div>
  </article>
</template>
