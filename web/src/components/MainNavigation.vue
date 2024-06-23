<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed } from 'vue';
import NavRestaurant from './NavRestaurant.vue';
import NavTitle from './NavTitle.vue';
import NavExtra from './NavExtra.vue';

const store = useRestaurantStore();
const groups: ComputedRef<Record<string, handler_Restaurant[]>> = computed(
  () => store.grouped
);

const amountOfRestaurants = (restaurants: handler_Restaurant[]) => {
  const amount = restaurants.length;
  return amount === 1 ? amount + ' Restaurant' : amount + ' Restaurants';
};

const defaultOpened = ['Fasanenhof', 'Leinfelden-Echterdingen'];
</script>

<template>
  <q-list>
    <NavTitle />
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
      />
    </q-expansion-item>
  </q-list>
  <NavExtra />
</template>
