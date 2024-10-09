<script setup lang="ts">
import { emptyRestaurant, useRestaurantStore } from 'src/stores/restaurants';
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import FavStar from './FavStar.vue';

const route = useRoute();
const store = useRestaurantStore();
const restaurant = computed(
  () => store.restaurants[route.params.name as string] ?? emptyRestaurant
);

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    restaurant.value.address
);
</script>

<template>
  <div class="row q-gutter-x-sm" v-if="restaurant.id !== ''">
    <FavStar :restaurant="restaurant" />
    <q-btn
      size="sm"
      flat
      round
      color="secondary"
      icon="fa-solid fa-map-marker-alt"
      :href="googleSearch"
    >
      <q-tooltip class="bg-accent">Karte öffnen</q-tooltip>
    </q-btn>
    <q-btn
      size="sm"
      flat
      round
      color="secondary"
      icon="fa-solid fa-phone"
      :href="'tel:' + restaurant.phone"
    >
      <q-tooltip class="bg-accent">Anrufen</q-tooltip>
    </q-btn>
    <q-btn
      v-if="restaurant.page_url"
      size="sm"
      flat
      round
      color="secondary"
      icon="fa-solid fa-globe"
      :href="restaurant.page_url"
    >
      <q-tooltip class="bg-accent">Restaurant öffnen</q-tooltip>
    </q-btn>
  </div>
</template>
