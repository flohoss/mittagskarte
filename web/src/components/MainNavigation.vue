<script setup lang="ts">
import { services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import NavRestaurant from './NavRestaurant.vue';

const store = useRestaurantStore();

const amountOfRestaurants = (restaurants: services_CleanRestaurant[]) => {
  const amount = restaurants.length;
  return amount === 1 ? amount + ' Restaurant' : amount + ' Restaurants';
};
</script>

<template>
  <q-list>
    <q-expansion-item
      class="q-pb-md"
      label="Favoriten"
      :caption="amountOfRestaurants(Object.values(store.favorites))"
      v-if="store.favorites.length > 0"
      default-opened
      hide-expand-icon
      expand-icon-toggle
    >
      <transition-group
        appear
        enter-active-class="animated fadeIn"
        leave-active-class="animated fadeOutLeft"
      >
        <NavRestaurant
          v-for="restaurant in store.favoriteRestaurants"
          :key="'fav-' + restaurant.id"
          :restaurant="restaurant"
          :search="false"
          show-star
        />
      </transition-group>
    </q-expansion-item>
    <q-expansion-item
      v-for="(restaurants, key) in store.grouped"
      :key="key"
      :label="key"
      :caption="amountOfRestaurants(restaurants)"
    >
      <NavRestaurant
        v-for="restaurant in restaurants"
        :key="restaurant.id"
        :restaurant="restaurant"
        :search="false"
        show-star
      />
    </q-expansion-item>
  </q-list>
</template>
