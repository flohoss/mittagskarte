<script setup lang="ts">
import RestaurantCard from './RestaurantCard.vue';
import type { Restaurant } from '../types/restaurant';

defineProps<{
  restaurants: Restaurant[];
  group: string;
}>();
</script>

<template>
  <section id="restaurants-container" class="grid gap-2">
    <div class="text-sm opacity-75 tracking-wide font-medium text-base-content/80 flex items-center gap-2">
      <span>{{ group }}</span> <span class="text-xs opacity-60">({{ restaurants.length }})</span>
    </div>
    <div class="card-grid">
      <RestaurantCard v-for="restaurant in restaurants" :key="restaurant.id" :restaurant="restaurant" />
    </div>
  </section>
</template>
