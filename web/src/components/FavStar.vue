<script setup lang="ts">
import { services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed } from 'vue';

const props = defineProps<{
  restaurant: services_CleanRestaurant;
}>();
const store = useRestaurantStore();

const favorite = computed(() => store.favorites.includes(props.restaurant.id));

const starIcon = computed(() => {
  if (favorite.value) {
    return 'fa-solid fa-star';
  } else {
    return 'fa-regular fa-star';
  }
});

const tooltip = computed(() => {
  if (favorite.value) {
    return 'Favorit entfernen';
  } else {
    return 'Favorisieren';
  }
});

const color = computed(() => {
  if (favorite.value) {
    return 'secondary';
  } else {
    return 'primary';
  }
});
</script>

<template>
  <q-btn
    round
    size="sm"
    flat
    :color="color"
    :icon="starIcon"
    @click.prevent="store.toggleFavorite(restaurant)"
  >
    <q-tooltip class="bg-accent">{{ tooltip }}</q-tooltip>
  </q-btn>
</template>
