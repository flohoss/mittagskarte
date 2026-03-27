import { createGlobalState } from '@vueuse/core';
import { ref } from 'vue';

import PocketBase, { type RecordModel } from 'pocketbase';
import { BackendURL } from '../main';

export const useRestaurants = createGlobalState(() => {
  const pb = new PocketBase(BackendURL);

  const restaurants = ref<RecordModel[]>([]);
  const isLoaded = ref(false);
  const isLoading = ref(false);

  async function fetchRestaurants() {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const records = await pb.collection('restaurants').getFullList({
        sort: 'name',
      });
      restaurants.value = records;
      isLoaded.value = true;
    } finally {
      isLoading.value = false;
    }
  }

  if (!isLoaded.value && !isLoading.value) {
    void fetchRestaurants();
  }

  return {
    restaurants,
    isLoaded,
    isLoading,
    fetchRestaurants,
  };
});
