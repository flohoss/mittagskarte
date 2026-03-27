<script setup lang="ts">
import AppShell from './components/AppShell.vue';
import RestaurantGroup from './components/RestaurantGroup.vue';
import RestaurantsEmpty from './components/RestaurantsEmpty.vue';
import { useRestaurants } from './stores/useRestaurants';

const { groupedRestaurants } = useRestaurants();
</script>

<template>
  <AppShell title="Schniddzl.de" description="deine Mittagskarte für die Region Stuttgart">
    <div v-if="Object.keys(groupedRestaurants).length" class="grid gap-8">
      <RestaurantGroup v-for="(restaurants, group) in groupedRestaurants" :key="group" :restaurants="restaurants" :group="group" />
    </div>

    <RestaurantsEmpty v-else />
  </AppShell>
</template>
