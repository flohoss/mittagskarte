<script setup lang="ts">
import { Loading } from 'quasar';
import RestaurantHeader from 'src/components/RestaurantHeader.vue';
import WeeklyFood from 'src/components/WeeklyFood.vue';
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
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
      .finally(() => Loading.hide());
  },
});

const cardUrl = computed(() => {
  let url = process.env.BASE_URL + restaurant.value.menu.card;
  if (route.query.cache !== undefined) {
    url += '?rnd=' + route.query.cache;
  }
  return url;
});

const menu = ref(false);
</script>

<template>
  <q-page class="row align-start justify-center" style="padding-top: 1rem">
    <div class="container" v-if="restaurant.name != ''">
      <RestaurantHeader :restaurant="restaurant" @openMenu="menu = true" />
      <WeeklyFood
        v-if="restaurant.menu.food.length > 0"
        :restaurant="restaurant"
      />
      <q-img
        class="q-ma-md"
        style="border-radius: 1em"
        v-else-if="cardUrl"
        :src="cardUrl"
      />
    </div>
  </q-page>

  <q-dialog v-model="menu" full-height>
    <q-img v-if="cardUrl" :src="cardUrl" />
  </q-dialog>
</template>
