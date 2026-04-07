import { createGlobalState } from '@vueuse/core';
import { computed, ref } from 'vue';

import type { RestaurantRecord, RestaurantStatusEvent } from '../models/restaurant';
import { backendClient } from '../services/backendClient';
import { useFavorites } from './useFavorites';

export enum RestaurantStatus {
  QUEUED = 'queued',
  UPDATING = 'updating',
  COOLDOWN = 'cooldown',
  IDLE = 'idle',
}

export enum RestaurantMethod {
  SCRAPE = 'scrape',
  DOWNLOAD = 'download',
  UPLOAD = 'upload',
}
export const useRestaurants = createGlobalState(() => {
  const { favorites } = useFavorites();

  const restaurants = ref<RestaurantRecord[]>([]);
  const isLoaded = ref(false);
  const isLoading = ref(false);
  const searchQuery = ref('');

  function upsertRestaurant(record: RestaurantRecord) {
    const index = restaurants.value.findIndex((r) => r.id === record.id);
    if (index !== -1) {
      const current = restaurants.value[index];
      const { status: _incomingStatus, ...nextFields } = record;

      restaurants.value[index] = {
        ...current,
        ...nextFields,
        expand: record.expand ?? current.expand,
      };
    }
  }

  async function subscribeRealtime() {
    await backendClient.subscribeRestaurantStatus((e: RestaurantStatusEvent) => {
      const index = restaurants.value.findIndex((r) => r.id === e.id);
      if (index !== -1) {
        restaurants.value[index].status = e.status as RestaurantStatus;
      }
    });
    await backendClient.subscribeRestaurants((action, record) => {
      if (action === 'update') {
        upsertRestaurant(record);
      }
    });
  }

  function preloadThumbnails(records: ReturnType<typeof restaurants.value.slice>) {
    return Promise.allSettled(
      records
        .filter((r) => r.thumbnail)
        .map(
          (r) =>
            new Promise<void>((resolve) => {
              const img = new Image();
              img.onload = () => resolve();
              img.onerror = () => resolve();
              img.src = backendClient.getFileUrl(r);
            })
        )
    );
  }

  async function fetchRestaurants() {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const records = await backendClient.fetchRestaurants();
      await preloadThumbnails(records);
      restaurants.value = records;
      isLoaded.value = true;
    } finally {
      isLoading.value = false;
    }
  }

  if (!isLoaded.value && !isLoading.value) {
    void fetchRestaurants();
    void subscribeRealtime();
  }

  function getFileUrl(restaurant: RestaurantRecord) {
    return backendClient.getFileUrl(restaurant);
  }

  function getMapUrl(restaurant: RestaurantRecord) {
    return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(restaurant.address)}`;
  }

  function getPhoneUrl(restaurant: RestaurantRecord) {
    return `tel:${restaurant.phone.replace(/[^+\d]/g, '')}`;
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

  const groupedRestaurants = computed<Record<string, RestaurantRecord[]>>(() => {
    const groups: Record<string, RestaurantRecord[]> = {};
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

  function applySearch(query: string) {
    searchQuery.value = query;
  }

  return {
    restaurants,
    searchQuery,
    groupedRestaurants,
    filteredRestaurants,
    isLoaded,
    isLoading,
    fetchRestaurants,
    getFileUrl,
    getMapUrl,
    getPhoneUrl,
    applySearch,
  };
});
