<script setup lang="ts">
import type { Restaurant } from '../types/restaurant';

defineProps<{
  restaurants: Restaurant[];
  group: string;
}>();
</script>

<template>
  <section id="restaurants-container" class="grid gap-2">
    <div class="text-sm opacity-75 tracking-wide pl-5 font-medium text-base-content/80 flex items-center gap-2">
      <span>{{ group }}</span> <span class="text-xs opacity-60">({{ restaurants.length }})</span>
    </div>
    <div class="px-4 card-grid">
      <div v-for="restaurant in restaurants" :key="restaurant.id" class="group my-card">
        {{ restaurant.name }}
      </div>
    </div>
  </section>
</template>
