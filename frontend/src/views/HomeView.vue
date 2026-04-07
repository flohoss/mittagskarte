<script setup lang="ts">
import { computed } from 'vue';
import RestaurantGroup from '../components/RestaurantGroup.vue';
import RestaurantsEmpty from '../components/RestaurantsEmpty.vue';
import RestaurantsLoading from '../components/RestaurantsLoading.vue';
import { useRestaurants } from '../stores/useRestaurants';

const { groupedRestaurants, isLoading, isLoaded } = useRestaurants();

const hasRestaurants = computed(() => Object.keys(groupedRestaurants.value).length > 0);
const pageState = computed(() => {
  if (hasRestaurants.value) return 'content';
  if (!isLoaded.value) return 'loading';
  return 'empty';
});
</script>

<template>
  <Transition name="page-crossfade" mode="out-in">
    <div v-if="pageState === 'content'" key="content" class="grid gap-8">
      <RestaurantGroup v-for="(restaurants, group) in groupedRestaurants" :key="group" :restaurants="restaurants" :group="group" />
    </div>

    <RestaurantsLoading v-else-if="pageState === 'loading'" key="loading" :is-loading="isLoading" />

    <RestaurantsEmpty v-else key="empty" />
  </Transition>
</template>

<style scoped>
.page-crossfade-enter-active,
.page-crossfade-leave-active {
  transition: opacity 260ms ease;
}

.page-crossfade-enter-from,
.page-crossfade-leave-to {
  opacity: 0;
}
</style>
