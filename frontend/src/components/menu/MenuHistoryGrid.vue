<script setup lang="ts">
import { computed } from 'vue';
import type { MenuRecord } from '../../models/restaurant';
import { backendClient } from '../../services/backendClient';

const props = defineProps<{
  menus: MenuRecord[];
}>();

const dateFormatter = new Intl.DateTimeFormat('de-DE', {
  dateStyle: 'medium',
  timeStyle: 'short',
});

const menuUrlsById = computed(() => {
  const lookup = new Map<string, string>();
  props.menus.forEach((menu) => {
    lookup.set(menu.id, backendClient.getMenuFileUrl(menu));
  });
  return lookup;
});

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
  <div class="timeline-masonry columns-1 gap-3 sm:columns-2 lg:columns-3 2xl:columns-4">
    <a
      v-for="menu in props.menus"
      :key="menu.id"
      :href="menuUrlsById.get(menu.id)"
      target="_blank"
      rel="noreferrer"
      class="group relative mb-3 block break-inside-avoid overflow-hidden rounded-xl transition-opacity duration-200 hover:opacity-90"
    >
      <img :src="menuUrlsById.get(menu.id)" :alt="getMenuAlt(menu)" class="h-auto w-full" loading="lazy" />
      <div
        class="pointer-events-none absolute inset-x-0 bottom-0 bg-linear-to-t from-black/65 via-black/25 to-transparent px-3 py-2 text-xs text-white opacity-0 transition-opacity duration-150 group-hover:opacity-100"
      >
        {{ formatDate(menu.created) }}
      </div>
    </a>
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
