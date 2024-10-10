<script setup lang="ts">
import { computed } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { services_CleanRestaurant } from 'src/openapi';
import { useQuasar } from 'quasar';
import RestaurantInfo from './RestaurantInfo.vue';

const $q = useQuasar();
defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
});

const size = computed(() => {
  if ($q.screen.lt.sm) {
    return 'md';
  }
  return 'sm';
});
</script>

<template>
  <div class="row q-gutter-x-sm">
    <RestaurantInfo
      v-if="$q.screen.gt.sm && restaurant.image_url !== ''"
      :restaurant="restaurant"
      :icon-size="size"
    />
    <RestaurantActions :restaurant="restaurant" :icon-size="size" />
  </div>
</template>
