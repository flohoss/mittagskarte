<script setup lang="ts">
import { Loading } from 'quasar';
import { OpenAPI } from './openapi';
import { useRestaurantStore } from './stores/restaurants';

const BASE_URL = process.env.BASE_URL || '';
OpenAPI.BASE = BASE_URL + 'api/v1';

defineOptions({
  name: 'App',
  preFetch() {
    Loading.show();
    const store = useRestaurantStore();
    store.getRestaurants().finally(() => Loading.hide());
  },
});
</script>

<template>
  <router-view />
</template>

<style lang="scss">
.container {
  width: 100%;
  max-width: $breakpoint-sm-max;
  padding-top: 1rem;
}
</style>
