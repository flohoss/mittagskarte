<script setup lang="ts">
import { Loading } from 'quasar';
import { OpenAPI } from './openapi';
import { useRestaurantStore } from './stores/restaurants';

OpenAPI.BASE = process.env.API_URL || '';

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
