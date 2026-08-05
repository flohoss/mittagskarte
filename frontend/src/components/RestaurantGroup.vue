<script setup lang="ts">
import type { RestaurantRecord } from '../models/restaurant';
import RestaurantCard from './RestaurantCard.vue';
import { useRestaurants } from '../stores/useRestaurants';

const { groupBy } = useRestaurants();

defineProps<{
  restaurants: RestaurantRecord[];
  group: string;
  showHeading?: boolean;
}>();
</script>

<template>
  <section id="restaurants-container" class="grid gap-2">
    <div v-if="showHeading !== false" class="text-sm opacity-75 tracking-wide font-medium text-base-content/80 flex items-center gap-2">
      <span>{{ group }}</span>
      <span class="text-xs opacity-60">({{ restaurants.length }})</span>
    </div>
    <div class="card-grid">
      <RestaurantCard
        v-for="restaurant in restaurants"
        :key="restaurant.id"
        :restaurant="restaurant"
        :show-region="groupBy !== 'group' || group === 'Favoriten'"
      />
    </div>
  </section>
</template>
