<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed } from 'vue';
import NavRestaurant from './NavRestaurant.vue';

const store = useRestaurantStore();
const groups: ComputedRef<Record<string, handler_Restaurant[]>> = computed(
  () => store.grouped
);

const amountOfRestaurants = (restaurants: handler_Restaurant[]) => {
  const amount = restaurants.length;
  return amount === 1 ? amount + ' Restaurant' : amount + ' Restaurants';
};

const defaultOpened = ['Fasanenhof', 'Leinfelden-Echterdingen', 'Degerloch'];
</script>

<template>
  <q-list>
    <q-expansion-item
      v-for="(restaurants, key) in groups"
      :key="key"
      :label="key"
      :caption="amountOfRestaurants(restaurants)"
      :default-opened="defaultOpened.includes(key)"
    >
      <NavRestaurant
        v-for="(restaurant, index) in restaurants"
        :key="index"
        :restaurant="restaurant"
        :search="false"
      />
    </q-expansion-item>
  </q-list>
</template>
