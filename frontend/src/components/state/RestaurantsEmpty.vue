<script setup lang="ts">
import { computed } from 'vue';

import { RestaurantEmptyReason, useRestaurants, type RestaurantEmptyReason as RestaurantEmptyReasonType } from '../../stores/useRestaurants';

const { emptyViewReason } = useRestaurants();

const iconClass = computed(() => {
  if (emptyViewReason.value === RestaurantEmptyReason.NO_FAVORITES) {
    return 'icon-[fa7-solid--star]';
  }
  return 'icon-[fa7-solid--magnifying-glass]';
});

const heading = computed(() => {
  if (emptyViewReason.value === RestaurantEmptyReason.NO_FAVORITES) {
    return 'Keine Favoriten ausgewählt';
  }
  return 'Keine Restaurants gefunden';
});

const description: Record<RestaurantEmptyReasonType, string> = {
  [RestaurantEmptyReason.NO_FAVORITES]: 'Markiere ein Restaurant mit dem Stern als Favorit, um es hier ohne weitere Filter zu sehen.',
  [RestaurantEmptyReason.SEARCH]: 'Für deine Suche wurden keine passenden Restaurants gefunden. Versuche einen anderen Namen, Ort oder Tag.',
  [RestaurantEmptyReason.NO_DATA]: 'Zurzeit sind keine Restaurantdaten verfügbar. Prüfe die Verbindung zum Backend oder lade die Seite erneut.',
};

const descriptionText = computed(() => description[emptyViewReason.value ?? RestaurantEmptyReason.NO_DATA]);
</script>

<template>
  <section class="grid min-h-[40vh] place-items-center rounded-4xl border border-dashed border-base-300 bg-base-200/40 p-8 text-center">
    <div class="max-w-md space-y-4">
      <div class="mx-auto flex size-16 items-center justify-center rounded-2xl border border-base-300 bg-base-100/80 text-base-content/55 shadow-sm">
        <span :class="[iconClass, 'size-7']" aria-hidden="true"></span>
      </div>

      <div class="space-y-2">
        <div class="text-lg font-semibold">{{ heading }}</div>
        <p class="text-sm text-base-content/65">{{ descriptionText }}</p>
      </div>
    </div>
  </section>
</template>
