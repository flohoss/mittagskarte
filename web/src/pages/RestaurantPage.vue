<script setup lang="ts">
import { Loading } from 'quasar';
import WeeklyFood from 'src/components/WeeklyFood.vue';
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed, ref } from 'vue';

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

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    restaurant.value.address
);

const menu = ref(true);
</script>

<template>
  <q-page class="row align-start justify-center" style="padding-top: 1rem">
    <div class="container" v-if="restaurant.name != ''">
      <div class="full-width row no-wrap items-center q-gutter-md">
        <q-img
          :src="thumbnail"
          fit="cover"
          style="max-height: 8rem; max-width: 8rem; border-radius: 0.5rem"
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
            <q-btn
              v-if="restaurant.menu.card"
              outline
              color="primary"
              icon="fa-solid fa-rectangle-list"
              label="Menu"
              :href="restaurant.menu.card"
            />
          </div>
          <div class="text-subtitle">
            {{ restaurant.description }}
            <span v-for="i in restaurant.price" :key="i">€</span>
          </div>
        </div>
      </div>
      <WeeklyFood :restaurant="restaurant" />
    </div>
  </q-page>

  <q-dialog v-model="menu">
    <q-card>
      <q-card-section>
        <div class="text-h6">Alert</div>
      </q-card-section>

      <q-card-section class="q-pt-none">
        Lorem ipsum dolor sit amet consectetur adipisicing elit. Rerum
        repellendus sit voluptate voluptas eveniet porro. Rerum blanditiis
        perferendis totam, ea at omnis vel numquam exercitationem aut, natus
        minima, porro labore.
      </q-card-section>

      <q-card-actions align="right">
        <q-btn flat label="OK" color="primary" v-close-popup />
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>

<style lang="scss">
.container {
  width: 100%;
  max-width: $breakpoint-sm-max;
}
</style>
