<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed } from 'vue';

const store = useRestaurantStore();
const restaurant: ComputedRef<handler_Restaurant> = computed(
  () => store.restaurant
);

defineOptions({
  preFetch({ currentRoute }) {
    const store = useRestaurantStore();
    store.getRestaurant(currentRoute.params.name as string);
  },
});
</script>

<template>
  <q-page class="row items-center justify-evenly">
    {{ restaurant.name }}
  </q-page>
</template>
