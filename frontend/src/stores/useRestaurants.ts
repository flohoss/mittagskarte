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

declare global {
  interface Window {
    __RESTAURANTS__?: RestaurantRecord[] | undefined;
  }
}

export const useRestaurants = createGlobalState(() => {
  const { favorites } = useFavorites();

  const restaurants = ref<RestaurantRecord[]>([]);
  const isLoading = ref(true);
  const searchQuery = ref('');
  const prefetchedMenuUrls = new Set<string>();

  function findRestaurantIndexById(id: string) {
    return restaurants.value.findIndex((restaurant) => restaurant.id === id);
  }

  function setRestaurantStatus(id: string, status: RestaurantStatus) {
    const index = findRestaurantIndexById(id);
    if (index === -1) return;
    restaurants.value[index].status = status;
  }

  function mergeRestaurant(record: RestaurantRecord) {
    const next = backendClient.normalizeRestaurant(record);
    const index = findRestaurantIndexById(next.id);

    if (index === -1) {
      restaurants.value = backendClient.prepareRestaurants([...restaurants.value, next]);
      return;
    }

    const current = restaurants.value[index];
    restaurants.value[index] = {
      ...current,
      ...next,
      status: next.status || current.status,
      expand: next.expand ?? current.expand,
    };
  }

  function removeRestaurant(id: string) {
    const index = findRestaurantIndexById(id);
    if (index === -1) return;
    restaurants.value.splice(index, 1);
  }

  async function subscribeRealtime() {
    await backendClient.subscribeRestaurantStatus(({ id, status }: RestaurantStatusEvent) => {
      setRestaurantStatus(id, status as RestaurantStatus);
    });

    await backendClient.subscribeRestaurants((action, record) => {
      if (action === 'delete') {
        removeRestaurant(record.id);
        return;
      }

      if (action === 'create' || action === 'update') {
        mergeRestaurant(record);
      }
    });
  }

  function preloadImageUrls(urls: string[]) {
    if (typeof Image === 'undefined' || !urls.length) {
      return Promise.resolve([]);
    }

    return Promise.allSettled(
      urls.filter(Boolean).map(
        (url) =>
          new Promise<void>((resolve) => {
            const img = new Image();
            img.onload = () => resolve();
            img.onerror = () => resolve();
            img.src = url;
          })
      )
    );
  }

  function preloadThumbnails(records: RestaurantRecord[]) {
    return preloadImageUrls(records.filter((restaurant) => restaurant.thumbnail).map((restaurant) => backendClient.getFileUrl(restaurant)));
  }

  function scheduleMenuPrefetch(records: RestaurantRecord[]) {
    const browserWindow = typeof window === 'undefined' ? null : window;

    if (!browserWindow) {
      return;
    }

    const scheduleWhenIdle = (callback: () => void, timeout: number) => {
      if ('requestIdleCallback' in browserWindow) {
        browserWindow.requestIdleCallback(callback, { timeout });
        return;
      }

      globalThis.setTimeout(callback, Math.min(timeout, 500));
    };

    const prefetchUrls = (urls: string[], onDone?: () => void) => {
      const nextUrls = urls.filter((url) => Boolean(url) && !prefetchedMenuUrls.has(url));

      if (!nextUrls.length) {
        onDone?.();
        return;
      }

      nextUrls.forEach((url) => prefetchedMenuUrls.add(url));
      void preloadImageUrls(nextUrls).finally(() => onDone?.());
    };

    scheduleWhenIdle(() => {
      const currentMenuUrls = records
        .flatMap((restaurant) => {
          const currentMenu = restaurant.expand?.menus?.[0];
          return currentMenu ? [backendClient.getMenuFileUrl(currentMenu)] : [];
        })
        .slice(0, 24);

      prefetchUrls(currentMenuUrls, () => {
        scheduleWhenIdle(() => {
          const historyMenuUrls = records
            .flatMap((restaurant) => (restaurant.expand?.menus ?? []).slice(1).map((menu) => backendClient.getMenuFileUrl(menu)))
            .slice(0, 64);

          prefetchUrls(historyMenuUrls);
        }, 2500);
      });
    }, 1200);
  }

  function applyRestaurants(records: RestaurantRecord[]) {
    const nextRestaurants = backendClient.prepareRestaurants(records);
    restaurants.value = nextRestaurants;
    return nextRestaurants;
  }

  async function fetchRestaurants() {
    isLoading.value = true;

    try {
      const nextRestaurants = applyRestaurants(await backendClient.fetchRestaurants());
      await preloadThumbnails(nextRestaurants);
      scheduleMenuPrefetch(nextRestaurants);
    } finally {
      isLoading.value = false;
    }
  }

  function takeInitialRestaurants() {
    if (typeof window === 'undefined') {
      return null;
    }

    const initialRestaurants = window.__RESTAURANTS__;
    window.__RESTAURANTS__ = undefined;
    return Array.isArray(initialRestaurants) ? initialRestaurants : null;
  }

  async function hydrateInitialRestaurants() {
    const initialRestaurants = takeInitialRestaurants();
    if (!initialRestaurants) return;

    const nextRestaurants = applyRestaurants(initialRestaurants);
    await preloadThumbnails(nextRestaurants);
    scheduleMenuPrefetch(nextRestaurants);
  }

  function initialize() {
    void (async () => {
      await hydrateInitialRestaurants();
      await fetchRestaurants();
    })();

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
    initialize,
    isLoading,
    fetchRestaurants,
    getFileUrl,
    getMapUrl,
    getPhoneUrl,
    applySearch,
  };
});
