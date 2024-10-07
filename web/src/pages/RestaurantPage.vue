<script setup lang="ts">
import { Loading } from 'quasar';
import RestaurantHeader from 'src/components/RestaurantHeader.vue';
import { services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const store = useRestaurantStore();
const restaurant: ComputedRef<services_CleanRestaurant> = computed(
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
  let url = process.env.BASE_URL + restaurant.value.image_url;
  if (route.query.cache !== undefined) {
    url += '?rnd=' + route.query.cache;
  }
  return url;
});

const menu = ref(false);
</script>

<template>
  <q-page class="row align-start justify-center q-pt-md">
    <div class="container" v-if="restaurant.name != ''">
      <RestaurantHeader :restaurant="restaurant" />
    </div>
  </q-page>

  <q-dialog v-model="menu" full-height>
    <q-img v-if="cardUrl" :src="cardUrl" />
  </q-dialog>
</template>
