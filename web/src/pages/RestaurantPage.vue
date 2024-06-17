<script setup lang="ts">
import { Loading } from 'quasar';
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed } from 'vue';

const store = useRestaurantStore();
const restaurant: ComputedRef<handler_Restaurant> = computed(
  () => store.restaurant
);

defineOptions({
  preFetch({ currentRoute }) {
    Loading.show();
    const store = useRestaurantStore();
    store
      .getRestaurant(currentRoute.params.name as string)
      .then(() => Loading.hide());
  },
});

const thumbnail = computed(
  () =>
    process.env.BASE_URL + '/public/thumbnails/' + restaurant.value.id + '.webp'
);
</script>

<template>
  <q-page class="row align-start justify-center" padding>
    <div class="container">
      <div class="q-pa-md example-row-equal-width">
        <div class="row">
          <div class="col-3">
            <q-img
              :src="thumbnail"
              fit="cover"
              style="max-height: 12rem; border-radius: 0.5rem"
            />
          </div>
          <div class="col">
            <q-chip
              :icon="restaurant.icon"
              :label="restaurant.name"
            />
            <q-btn
              selectable
              round
              tabindex="100"
              color="primary"
              icon="fa-solid fa-location-dot"
              :href="
                'https://www.google.com/maps/search/?api=1&query=' +
                restaurant.address
              "
            />
          </div>
        </div>
      </div>
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
