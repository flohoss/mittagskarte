<script setup lang="ts">
import Fa7SolidGlobe from '~icons/fa7-solid/globe';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import Fa7SolidPhone from '~icons/fa7-solid/phone';
import Fa7SolidMap from '~icons/fa7-solid/map';
import Fa7SolidStar from '~icons/fa7-solid/star';

import { computed } from 'vue';
import { BackendURL } from '../main';
import type { Restaurant } from '../types/restaurant';

const props = defineProps<{
  restaurant: Restaurant;
}>();

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

const isClosed = computed(() => props.restaurant.rest_days.includes(WEEKDAYS[new Date().getDay()]));

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

function getThumbnailUrl(restaurant: Restaurant) {
  if (!restaurant.thumbnail) return '';
  return `${BackendURL}/api/files/${restaurant.collectionId}/${restaurant.id}/${restaurant.thumbnail}`;
}

function getMapUrl(restaurant: Restaurant) {
  const query = [restaurant.name, restaurant.address].filter(Boolean).join(' ');
  return query ? `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(query)}` : '';
}

function getPhoneUrl(phone: string) {
  if (!phone) return '';
  return `tel:${phone.replace(/[^+\d]/g, '')}`;
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
    class="group my-card card card-border overflow-hidden rounded-2xl border border-base-300 bg-base-100 opacity-80 shadow-md transition-[shadow,opacity] duration-200 hover:opacity-100 hover:shadow-xl"
  >
    <figure class="relative h-30 overflow-hidden bg-base-300">
      <img
        v-if="getThumbnailUrl(props.restaurant)"
        :src="getThumbnailUrl(props.restaurant)"
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
      <!-- top row: status badge + favourite -->
      <div class="absolute inset-x-0 top-0 flex items-start justify-between px-3 pt-3">
        <span :class="['badge badge-sm backdrop-blur', isClosed ? 'badge-error' : 'badge-ghost border-base-300/80 bg-base-100/85 text-base-content/70']">
          {{ isClosed ? 'Heute geschlossen' : formatRelativeDate(props.restaurant.updated) }}
        </span>
        <button type="button" class="btn btn-circle btn-ghost btn-xs text-warning backdrop-blur" aria-label="Favorit" @click="">
          <Fa7SolidStar class="size-4" aria-hidden="true" />
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

    <div class="card-body gap-3 p-4">
      <h3 class="text-base font-semibold leading-tight">{{ props.restaurant.name }}</h3>

      <div class="grid grid-cols-4 gap-1.5">
        <a
          v-if="props.restaurant.menu"
          :href="BackendURL + props.restaurant.menu"
          target="_blank"
          rel="noreferrer"
          class="btn btn-primary"
          title="Speisekarte"
          aria-label="Speisekarte öffnen"
        >
          <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
        </a>
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
          :href="getPhoneUrl(props.restaurant.phone)"
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
