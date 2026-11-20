<script setup lang="ts">
import RestaurantGroup from '../components/RestaurantGroup.vue';
import RestaurantsEmpty from '../components/state/RestaurantsEmpty.vue';
import { useRestaurants } from '../stores/useRestaurants';

const { groupedRestaurants, groupBy, hasVisibleRestaurants } = useRestaurants();
</script>

<template>
  <div class="grid gap-8">
    <RestaurantGroup
      v-for="(restaurants, group) in groupedRestaurants"
      :key="group"
      :restaurants="restaurants"
      :group="group"
      :show-heading="groupBy !== 'none' || group === 'Favoriten'"
    />
    <RestaurantsEmpty v-if="!hasVisibleRestaurants" />
  </div>
</template>
