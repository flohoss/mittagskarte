import { createGlobalState } from '@vueuse/core';
import { computed, ref } from 'vue';

import PocketBase, { type RecordModel } from 'pocketbase';
import { BackendURL } from '../main';
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
interface StatusChangeEvent {
  id: string;
  status: RestaurantStatus;
}

export const useRestaurants = createGlobalState(() => {
  const pb = new PocketBase(BackendURL);
  const { favorites } = useFavorites();

  const restaurants = ref<RecordModel[]>([]);
  const isLoaded = ref(false);
  const isLoading = ref(false);
  const searchQuery = ref('');

  function upsertRestaurant(record: RecordModel) {
    const index = restaurants.value.findIndex((r) => r.id === record.id);
    if (index !== -1) {
      restaurants.value[index] = record;
    }
  }

  async function subscribeRealtime() {
    await pb.realtime.subscribe('restaurants/status', (e: StatusChangeEvent) => {
      const index = restaurants.value.findIndex((r) => r.id === e.id);
      if (index !== -1) {
        restaurants.value[index].status = e.status;
      }
    });
    await pb.collection('restaurants').subscribe('*', (e) => {
      if (e.action === 'update') {
        upsertRestaurant(e.record);
      }
    });
  }

  async function fetchRestaurants() {
    if (isLoading.value) return;

    isLoading.value = true;
    try {
      const records = await pb.collection('restaurants').getFullList({
        sort: 'group,name',
      });
      restaurants.value = records as RecordModel[];
      isLoaded.value = true;
    } finally {
      isLoading.value = false;
    }
  }

  if (!isLoaded.value && !isLoading.value) {
    void fetchRestaurants();
    void subscribeRealtime();
  }

  function getFileUrl(restaurant: RecordModel) {
    const url = pb.files.getURL(restaurant, restaurant.thumbnail);
    return url;
  }

  function getMapUrl(restaurant: RecordModel) {
    return `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(restaurant.address)}`;
  }

  function getPhoneUrl(restaurant: RecordModel) {
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

  const groupedRestaurants = computed<Record<string, RecordModel[]>>(() => {
    const groups: Record<string, RecordModel[]> = {};
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
