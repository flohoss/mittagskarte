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
const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    restaurant.value.address
);
</script>

<template>
  <q-page class="row align-start justify-center" padding>
    <div class="container" v-if="restaurant.name != ''">
      <div class="full-width row no-wrap items-center q-gutter-md">
        <q-img
          :src="thumbnail"
          fit="cover"
          style="max-height: 10rem; max-width: 10rem; border-radius: 0.5rem"
        />
        <div class="column q-gutter-y-sm">
          <div class="text-h4 ellipsis">{{ restaurant.name }}</div>
          <div class="row wrap q-gutter-x-sm">
            <q-btn
              outline
              color="secondary"
              icon="fa-solid fa-map-marker-alt"
              :href="googleSearch"
            />
            <q-btn
              outline
              color="secondary"
              icon="fa-solid fa-phone"
              :href="'tel:' + restaurant.phone"
            />
            <q-btn
              outline
              color="secondary"
              icon="fa-solid fa-globe"
              :href="restaurant.page_url"
            />
          </div>
          <div class="text-subtitle">
            <span v-for="i in restaurant.price" :key="i">€</span>
            ・{{ restaurant.description }}
          </div>
        </div>
      </div>
      <q-card flat>
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
