<script setup lang="ts">
import RestaurantGroup from '../components/RestaurantGroup.vue';
import { useRestaurants } from '../stores/useRestaurants';

const { groupedRestaurants } = useRestaurants();
</script>

<template>
  <div class="grid gap-8">
    <RestaurantGroup v-for="(restaurants, group) in groupedRestaurants" :key="group" :restaurants="restaurants" :group="group" />
  </div>
</template>
