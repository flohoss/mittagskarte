<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';

const store = useRestaurantStore();
const thumbnail = (restaurant: handler_Restaurant) =>
  process.env.BASE_URL + 'config/thumbnails/' + restaurant.id + '.webp';
</script>

<template>
  <q-page class="row justify-center items-center" style="padding-top: 1rem">
    <q-item
      :to="'/restaurants/' + restaurant.id"
      v-for="restaurant in store.restaurants"
      :key="restaurant.id"
    >
      <q-img :src="thumbnail(restaurant)" width="13rem" height="13rem">
        <div class="absolute-bottom text-subtitle2 text-center">
          {{ restaurant.name }}
        </div>
      </q-img>
    </q-item>
  </q-page>
</template>
