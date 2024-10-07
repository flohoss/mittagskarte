<script setup lang="ts">
import { services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed } from 'vue';

const props = defineProps<{
  restaurant: services_CleanRestaurant;
}>();
const store = useRestaurantStore();

const starIcon = computed(() => {
  if (store.favorites.includes(props.restaurant.id)) {
    return 'fa-solid fa-star';
  } else {
    return 'fa-regular fa-star';
  }
});
</script>

<template>
  <q-icon
    @click.prevent="store.toggleFavorite(restaurant)"
    :name="starIcon"
    class="star cursor-pointer"
    color="secondary"
    size="xs"
  />
</template>

<style>
.star {
  opacity: 0.6;
}

.star:hover {
  opacity: 1;
}
</style>
