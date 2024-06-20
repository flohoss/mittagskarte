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
const cardUrl = computed(
  () => process.env.BASE_URL + '/' + restaurant.value.menu.card
);

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    restaurant.value.address
);

const menu = ref(false);
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
              @click="menu = true"
            />
          </div>
          <div class="text-subtitle">
            {{ restaurant.description }}
            <span v-for="i in restaurant.price" :key="i">€</span>
          </div>
        </div>
      </div>
      <WeeklyFood
        v-if="restaurant.menu.food.length > 0"
        :restaurant="restaurant"
      />
      <q-img
        width="50rem"
        style="border-radius: 1em; margin: 2rem 1rem"
        round
        v-else-if="cardUrl"
        :src="cardUrl"
      />
    </div>
  </q-page>

  <q-dialog v-model="menu">
    <div class="q-pa-md">
      <q-img
        v-if="cardUrl"
        width="50rem"
        style="border-radius: 1em; margin: 2rem 1rem"
        round
        :src="cardUrl"
      />
    </div>
  </q-dialog>
</template>

<style lang="scss">
.container {
  width: 100%;
  max-width: $breakpoint-sm-max;
}
</style>
