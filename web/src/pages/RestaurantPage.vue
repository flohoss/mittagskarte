<script setup lang="ts">
import { emptyRestaurant, useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const store = useRestaurantStore();
const restaurant = computed(
  () => store.restaurants[route.params.name as string] ?? emptyRestaurant
);

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
  <q-page class="row align-start justify-center">
    <div
      :class="[$q.screen.gt.sm ? 'q-py-md' : 'q-pa-sm']"
      @click="menu = true"
      :style="{
        'border-radius': '0.5rem',
        width: '100%',
        'max-width': $q.screen.sizes.sm + 'px',
      }"
    >
      <q-img
        :src="cardUrl"
        :style="{
          'border-radius': '0.5rem',
          width: '100%',
        }"
      />
    </div>
  </q-page>

  <q-dialog v-model="menu" full-height>
    <q-img v-if="cardUrl" :src="cardUrl" />
  </q-dialog>
</template>
