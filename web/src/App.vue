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
    store.getRestaurants().then(() => Loading.hide());
  },
});
</script>

<template>
  <router-view />
</template>
