<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import type { MenuRecord } from '../models/restaurant';
import { useRestaurants } from '../stores/useRestaurants';
import { backendClient } from '../services/backendClient';

const route = useRoute();
const { restaurants } = useRestaurants();

const restaurantId = computed(() => String(route.params.restaurantId ?? '').trim());
const restaurant = computed(() => restaurants.value.find((entry) => entry.id === restaurantId.value) ?? null);

const sortedMenus = computed(() => {
  const menus = restaurant.value?.expand?.menus ?? [];
  return [...menus].sort((a, b) => (a.created > b.created ? -1 : 1));
});

const dateFormatter = new Intl.DateTimeFormat('de-DE', {
  dateStyle: 'medium',
});

function getMenuUrl(menu: MenuRecord) {
  return backendClient.getMenuFileUrl(menu);
}

function getMenuAlt(menu: MenuRecord) {
  return `Menü ${menu.id}`;
}

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Unbekanntes Datum';
  return dateFormatter.format(date);
}
</script>

<template>
  <div class="grid gap-4 md:gap-6">
    <div v-if="!restaurant" class="rounded-xl border border-dashed border-base-300 bg-base-100 p-6 text-base-content/70">
      Restaurant wurde nicht gefunden.
    </div>

    <div v-else-if="!sortedMenus.length" class="rounded-xl border border-dashed border-base-300 bg-base-100 p-6 text-base-content/70">
      Für dieses Restaurant wurde noch kein Menü gespeichert.
    </div>

    <div v-else>
      <div class="timeline-masonry columns-1 gap-3 sm:columns-2 lg:columns-3 2xl:columns-4">
        <a
          v-for="menu in sortedMenus"
          :key="menu.id"
          :href="getMenuUrl(menu)"
          target="_blank"
          rel="noreferrer"
          class="group relative mb-3 block break-inside-avoid overflow-hidden rounded-xl transition-opacity duration-200 hover:opacity-90"
        >
          <img :src="getMenuUrl(menu)" :alt="getMenuAlt(menu)" class="h-auto w-full" loading="lazy" />
          <div
            class="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/65 via-black/25 to-transparent px-3 py-2 text-xs text-white opacity-0 transition-opacity duration-150 group-hover:opacity-100"
          >
            {{ formatDate(menu.created) }}
          </div>
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.timeline-masonry {
  column-fill: balance;
}

@media (min-width: 1024px) {
  .timeline-masonry {
    column-gap: 0.9rem;
  }
}
</style>
