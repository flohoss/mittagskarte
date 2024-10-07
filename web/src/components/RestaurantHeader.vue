<script setup lang="ts">
import { services_CleanRestaurant } from 'src/openapi';
import { computed } from 'vue';
import FavStar from './FavStar.vue';

const props = defineProps<{ restaurant: services_CleanRestaurant }>();

const thumbnail = computed(
  () =>
    process.env.BASE_URL + 'config/thumbnails/' + props.restaurant.id + '.webp'
);

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    props.restaurant.address
);
</script>

<template>
  <div class="row justify-between items-center q-px-md q-gutter-y-md">
    <div class="row items-center q-gutter-x-md">
      <q-img
        :src="thumbnail"
        fit="cover"
        style="height: 5rem; width: 5rem; border-radius: 0.5rem"
      />
      <div class="column justify-start">
        <div class="text-h4 ellipsis">{{ restaurant.name }}</div>
        <div class="text-caption row items-baseline q-gutter-x-sm">
          <FavStar :restaurant="restaurant" />
          <div>{{ restaurant.description }}</div>
          <div><span v-for="i in restaurant.price" :key="i">€</span></div>
        </div>
      </div>
    </div>
    <div class="row q-gutter-x-sm">
      <q-btn
        outline
        round
        color="secondary"
        icon="fa-solid fa-map-marker-alt"
        :href="googleSearch"
      >
        <q-tooltip class="bg-accent">Karte öffnen</q-tooltip>
      </q-btn>
      <q-btn
        outline
        round
        color="secondary"
        icon="fa-solid fa-phone"
        :href="'tel:' + restaurant.phone"
      >
        <q-tooltip class="bg-accent">Anrufen</q-tooltip>
      </q-btn>
      <q-btn
        v-if="restaurant.page_url"
        outline
        round
        color="secondary"
        icon="fa-solid fa-globe"
        :href="restaurant.page_url"
      >
        <q-tooltip class="bg-accent">Restaurant öffnen</q-tooltip>
      </q-btn>
    </div>
  </div>
</template>
