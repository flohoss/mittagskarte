<script setup lang="ts">
import Fa7SolidMagnifyingGlass from '~icons/fa7-solid/magnifying-glass';
import { computed } from 'vue';

import { useRestaurants } from '../../stores/useRestaurants';

const { searchQuery } = useRestaurants();

const isFiltering = computed(() => searchQuery.value.trim().length > 0);
</script>

<template>
  <section class="grid min-h-[40vh] place-items-center rounded-4xl border border-dashed border-base-300 bg-base-200/40 p-8 text-center">
    <div class="max-w-md space-y-4">
      <div class="mx-auto flex size-16 items-center justify-center rounded-2xl border border-base-300 bg-base-100/80 text-base-content/55 shadow-sm">
        <Fa7SolidMagnifyingGlass class="size-7" aria-hidden="true" />
      </div>

      <div class="space-y-2">
        <div class="text-lg font-semibold">Keine Restaurants gefunden</div>
        <p v-if="isFiltering" class="text-sm text-base-content/65">
          Für Ihre Suche wurden keine passenden Restaurants gefunden. Versuchen Sie einen anderen Namen, Ort oder Tag.
        </p>
        <p v-else class="text-sm text-base-content/65">
          Zurzeit sind keine Restaurantdaten verfügbar. Prüfen Sie die Verbindung zum Backend oder laden Sie die Seite erneut.
        </p>
      </div>
    </div>
  </section>
</template>
