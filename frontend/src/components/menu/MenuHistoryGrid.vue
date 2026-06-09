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

function isLandscape(menu: MenuRecord) {
  return menu.dimensions?.landscape === true;
}

const landscapeLayout = computed(() => props.menus.length > 0 && isLandscape(props.menus[0]));

function formatDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Unbekanntes Datum';
  return dateFormatter.format(date);
}
</script>

<template>
  <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4 lg:gap-4">
    <a
      v-for="menu in props.menus"
      :key="menu.id"
      :href="menuUrlsById.get(menu.id)"
      target="_blank"
      rel="noreferrer"
      :class="['card group block overflow-hidden', landscapeLayout ? 'lg:col-span-2' : 'lg:col-span-1']"
    >
      <div
        :class="['relative w-full overflow-hidden bg-base-200', landscapeLayout ? 'aspect-4/3' : 'aspect-3/4']"
      >
        <img
          :src="menuUrlsById.get(menu.id)"
          :alt="`Speisekarte vom ${formatDate(menu.created)}`"
          loading="lazy"
          decoding="async"
          class="h-full w-full object-cover"
        />
        <div
          class="pointer-events-none absolute inset-0 flex items-center justify-center bg-black/50 opacity-0 transition-opacity duration-200 group-hover:opacity-100"
        >
          <span class="badge badge-neutral badge-lg">
            {{ formatDate(menu.created) }}
          </span>
        </div>
      </div>
    </a>
  </div>
</template>
