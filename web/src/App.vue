<script setup lang="ts">
import { OpenAPI } from './openapi';
import { useRestaurantStore } from './stores/restaurants';

OpenAPI.BASE = 'http://localhost:4000/api/v1';

defineOptions({
  name: 'App',
  preFetch() {
    const store = useRestaurantStore();
    store.getRestaurants();
  },
});
</script>

<template>
  <router-view />
</template>
