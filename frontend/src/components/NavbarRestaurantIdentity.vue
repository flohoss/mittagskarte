<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useRestaurants } from '../stores/useRestaurants';
import RestaurantLinkButtons from './buttons/RestaurantLinkButtons.vue';

const route = useRoute();
const { restaurants } = useRestaurants();

const restaurant = computed(() => {
  const slug = String(route.params.restaurantSlug ?? '').trim();
  return restaurants.value.find((entry) => entry.slug === slug) ?? null;
});
</script>

<template>
  <div v-if="restaurant" class="flex min-w-0 items-center gap-2">
    <h1 class="min-w-0 flex-1 truncate text-center text-base font-semibold sm:text-lg">{{ restaurant.name }}</h1>
    <RestaurantLinkButtons :restaurant="restaurant" square />
  </div>
</template>
