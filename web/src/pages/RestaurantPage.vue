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
  <q-page class="row align-start justify-center" padding>
    <div class="container">
      <q-img
        src="https://cdn.quasar.dev/img/chicken-salad.jpg"
        fit="cover"
        style="max-height: 10rem; border-radius: 1rem"
      >
        <q-btn
          selectable
          round
          tabindex="100"
          color="primary"
          icon="fa-solid fa-location-dot"
          class="absolute"
          style="bottom: 1rem; right: 1rem"
          :href="
            'https://www.google.com/maps/search/?api=1&query=' +
            restaurant.address
          "
        />
        <q-chip
          :icon="restaurant.icon"
          :label="restaurant.name"
          class="absolute"
          style="top: 1rem; left: 1rem"
        />
      </q-img>

      <q-card-actions>
        {{ restaurant.menu }}
      </q-card-actions>
    </div>
  </q-page>
</template>

<style lang="scss">
.container {
  width: 100%;
  max-width: $breakpoint-sm-max;
}
</style>
