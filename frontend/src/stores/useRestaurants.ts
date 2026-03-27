import { createGlobalState } from '@vueuse/core';
import { computed, ref } from 'vue';

import PocketBase from 'pocketbase';
import { BackendURL } from '../main';
import { useFavorites } from './useFavorites';
import type { Restaurant } from '../types/restaurant';

export const useRestaurants = createGlobalState(() => {
  const pb = new PocketBase(BackendURL);
  const { favorites } = useFavorites();

  const restaurants = ref<Restaurant[]>([]);
  const isLoaded = ref(false);
  const isLoading = ref(false);
  const searchQuery = ref('');

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

  const filteredRestaurants = computed(() => {
    const query = searchQuery.value.trim().toLocaleLowerCase('de-DE');

    if (!query) {
      return restaurants.value;
    }

    return restaurants.value.filter((restaurant) => {
      const haystack = [restaurant.name, restaurant.group, restaurant.address, restaurant.website, ...restaurant.tags]
        .filter(Boolean)
        .join(' ')
        .toLocaleLowerCase('de-DE');

      return haystack.includes(query);
    });
  });

  const groupedRestaurants = computed<Record<string, Restaurant[]>>(() => {
    const groups: Record<string, Restaurant[]> = {};
    const favoriteRestaurants = filteredRestaurants.value.filter((restaurant) => favorites.value[restaurant.id]);

    if (favoriteRestaurants.length) {
      groups.Favoriten = favoriteRestaurants;
    }

    for (const restaurant of filteredRestaurants.value) {
      if (favorites.value[restaurant.id]) {
        continue;
      }

      if (!groups[restaurant.group]) {
        groups[restaurant.group] = [];
      }

      groups[restaurant.group].push(restaurant);
    }

    return groups;
  });

  return {
    restaurants,
    searchQuery,
    groupedRestaurants,
    filteredRestaurants,
    isLoaded,
    isLoading,
    fetchRestaurants,
  };
});
