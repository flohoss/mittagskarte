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

const euroFormatter = new Intl.NumberFormat('de-DE', {
  style: 'currency',
  currency: 'EUR',
});
</script>

<template>
  <q-page class="row align-start justify-center" padding>
    <div class="container">
      <q-card flat v-if="restaurant.name != ''">
        <q-img
          :src="thumbnail"
          fit="cover"
          style="max-height: 10rem; border-radius: 0.5rem"
        />

        <q-card-section>
          <q-btn
            fab
            color="primary"
            icon="fa-solid fa-location-dot"
            class="absolute"
            style="top: 0; right: 12px; transform: translateY(-50%)"
          />

          <div class="row no-wrap items-center">
            <div class="col text-h5 ellipsis">{{ restaurant.name }}</div>
          </div>
          <div class="text-subtitle1">
            <span v-for="i in restaurant.price" :key="i">€</span>
            ・{{ restaurant.description }}
          </div>
        </q-card-section>

        <q-separator />

        <q-card-actions>
          <q-list style="width: 100%">
            <q-item v-for="(entry, id) in restaurant.menu.food" :key="id">
              <q-item-section>
                <q-item-label>{{ entry.name }}</q-item-label>
                <q-item-label caption>{{ entry.description }}</q-item-label>
              </q-item-section>

              <q-item-section side top>
                <q-item-label caption>
                  {{ euroFormatter.format(restaurant.price) }}
                </q-item-label>
              </q-item-section>
            </q-item>
          </q-list>
        </q-card-actions>
      </q-card>
    </div>
  </q-page>
</template>

<style lang="scss">
.container {
  width: 100%;
  max-width: $breakpoint-sm-max;
}
</style>
