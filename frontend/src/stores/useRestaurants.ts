import { createGlobalState } from '@vueuse/core';
import { computed, ref } from 'vue';

import PocketBase from 'pocketbase';
import { BackendURL } from '../main';
import type { Restaurant } from '../types/restaurant';

export const useRestaurants = createGlobalState(() => {
  const pb = new PocketBase(BackendURL);

  const restaurants = ref<Restaurant[]>([]);
  const isLoaded = ref(false);
  const isLoading = ref(false);

  async function fetchRestaurants() {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const records = await pb.collection('restaurants').getFullList({
        sort: 'name',
      });
      restaurants.value = records as unknown as Restaurant[];
      isLoaded.value = true;
    } finally {
      isLoading.value = false;
    }
  }

  if (!isLoaded.value && !isLoading.value) {
    void fetchRestaurants();
  }

  const groupedRestaurants = computed<Record<string, Restaurant[]>>(() => {
    const groups: Record<string, Restaurant[]> = {};
    for (const restaurant of restaurants.value) {
      if (!groups[restaurant.group]) {
        groups[restaurant.group] = [];
      }
      groups[restaurant.group].push(restaurant);
    }
    return groups;
  });

  return {
    restaurants,
    groupedRestaurants,
    isLoaded,
    isLoading,
    fetchRestaurants,
  };
});
