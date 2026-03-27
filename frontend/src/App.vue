<script setup lang="ts">
import { useRestaurants } from './stores/useRestaurants';

const { isLoaded, restaurants } = useRestaurants();
</script>

<template>
  <div v-if="isLoaded" class="container pb-2 pt-3 md:py-4 lg:py-6">
    {{ restaurants }}
  </div>
</template>
