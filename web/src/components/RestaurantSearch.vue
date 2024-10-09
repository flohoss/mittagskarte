<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { ref, watch } from 'vue';
import NavRestaurant from './NavRestaurant.vue';
import { services_CleanRestaurant } from 'src/openapi';

const store = useRestaurantStore();
const search = ref('');
const result = ref<services_CleanRestaurant[]>([]);
const show = ref(false);

watch(search, (val) => {
  if (val === '') {
    result.value = [];
    show.value = false;
  } else {
    const lowerCaseSearch = search.value.toLowerCase();
    result.value = Object.values(store.restaurants).filter((restaurant) => {
      return (
        restaurant.id.toLowerCase().includes(lowerCaseSearch) ||
        restaurant.name.toLowerCase().includes(lowerCaseSearch) ||
        restaurant.description.toLowerCase().includes(lowerCaseSearch) ||
        restaurant.address.toLowerCase().includes(lowerCaseSearch) ||
        restaurant.group.toLowerCase().includes(lowerCaseSearch)
      );
    });
    show.value = true;
  }
});
</script>

<template>
  <q-input
    v-model="search"
    style="max-width: 20rem"
    class="full-width"
    dense
    placeholder="Suchen"
    :autofocus="false"
    type="search"
    @keyup.esc="search = ''"
  >
    <template v-slot:append><q-icon name="fa-solid fa-search" /></template>

    <q-menu
      :no-parent-event="true"
      :no-focus="true"
      v-model="show"
      fit
      anchor="bottom left"
      self="top left"
      class="q-pl-md"
    >
      <q-list style="min-width: 15rem">
        <NavRestaurant
          v-for="restaurant in result"
          :key="restaurant.id"
          :restaurant="restaurant"
          :search="true"
        />
      </q-list>
    </q-menu>
  </q-input>
</template>
