import { createGlobalState, useStorage } from '@vueuse/core';
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

export type RestaurantSort = 'name-asc' | 'name-desc' | 'menu-newest' | 'menu-oldest' | 'distance-asc';
export type RestaurantGrouping = 'group' | 'none';

export const restaurantSortOptions: Array<{ value: RestaurantSort; label: string }> = [
  { value: 'name-asc', label: 'Name A-Z' },
  { value: 'name-desc', label: 'Name Z-A' },
  { value: 'distance-asc', label: 'Nächste zuerst' },
  { value: 'menu-newest', label: 'Neueste zuerst' },
  { value: 'menu-oldest', label: 'Älteste zuerst' },
];

export const restaurantGroupingOptions: Array<{ value: RestaurantGrouping; label: string }> = [
  { value: 'group', label: 'Nach Gruppe' },
  { value: 'none', label: 'Keine Gruppierung' },
];

declare global {
  interface Window {
    __RESTAURANTS__?: RestaurantRecord[] | undefined;
  }
}

export const useRestaurants = createGlobalState(() => {
  const { favorites } = useFavorites();

  const restaurants = ref<RestaurantRecord[]>([]);
  const isLoading = ref(true);
  const searchQuery = useStorage<string>('mittagskarte:restaurants:search', '');
  const sortBy = useStorage<RestaurantSort>('mittagskarte:restaurants:sort', 'name-asc');
  const groupBy = useStorage<RestaurantGrouping>('mittagskarte:restaurants:group', 'group');
  const coords = ref<{ latitude: number; longitude: number } | null>(null);
  const geolocationLoading = ref(false);

  function requestGeolocation() {
    if (!navigator.geolocation) {
      console.error('[Geolocation] API not available');
      sortBy.value = 'name-asc';
      return;
    }
    console.log('[Geolocation] Requesting position...');
    geolocationLoading.value = true;
    navigator.geolocation.getCurrentPosition(
      (position) => {
        coords.value = { latitude: position.coords.latitude, longitude: position.coords.longitude };
        geolocationLoading.value = false;
        console.log('[Geolocation] Success:', coords.value);
      },
      (error) => {
        geolocationLoading.value = false;
        console.error('[Geolocation] Error:', {
          code: error.code,
          message: error.message,
          codeDescription: error.code === 1 ? 'Permission denied' : error.code === 2 ? 'Position unavailable' : 'Timeout',
        });
        // Fall back to name sorting on failure
        sortBy.value = 'name-asc';
      },
      { enableHighAccuracy: false, timeout: 10000, maximumAge: 300000 }
    );
  }

  if (sortBy.value === ('last-check-desc' as RestaurantSort)) {
    sortBy.value = 'menu-newest';
  } else if (sortBy.value === ('last-check-asc' as RestaurantSort)) {
    sortBy.value = 'menu-oldest';
  }

  // If distance-asc was persisted from previous session, request geolocation on init
  if (sortBy.value === 'distance-asc') {
    requestGeolocation();
  }

  const collator = new Intl.Collator('de-DE', { sensitivity: 'base' });

  function toTimestamp(value?: string | null) {
    if (!value) return 0;
    const timestamp = new Date(value).getTime();
    return Number.isNaN(timestamp) ? 0 : timestamp;
  }

  function getLatestMenuTimestamp(restaurant: RestaurantRecord) {
    const latestMenu = restaurant.expand?.menus?.[0];
    return toTimestamp(latestMenu?.created ?? null);
  }

  function getUserCoordinates() {
    if (!coords.value) return null;
    const lat = coords.value.latitude;
    const lon = coords.value.longitude;
    if (!Number.isFinite(lat) || !Number.isFinite(lon)) {
      return null;
    }
    return { latitude: lat, longitude: lon };
  }

  function toRadians(value: number) {
    return (value * Math.PI) / 180;
  }

  function getDistanceMeters(left: { latitude: number; longitude: number }, right: { latitude: number; longitude: number }) {
    const earthRadiusMeters = 6371e3;
    const phi1 = toRadians(left.latitude);
    const phi2 = toRadians(right.latitude);
    const deltaPhi = toRadians(right.latitude - left.latitude);
    const deltaLambda = toRadians(right.longitude - left.longitude);

    const a = Math.sin(deltaPhi / 2) * Math.sin(deltaPhi / 2) + Math.cos(phi1) * Math.cos(phi2) * Math.sin(deltaLambda / 2) * Math.sin(deltaLambda / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return earthRadiusMeters * c;
  }

  function getRestaurantDistanceMeters(restaurant: RestaurantRecord, userCoordinates: { latitude: number; longitude: number } | null) {
    const latitude = restaurant.latitude;
    const longitude = restaurant.longitude;

    if (!userCoordinates || typeof latitude !== 'number' || typeof longitude !== 'number') {
      return Number.POSITIVE_INFINITY;
    }

    return getDistanceMeters(userCoordinates, { latitude, longitude });
  }

  function getRestaurantDistanceKm(restaurant: RestaurantRecord) {
    const distanceMeters = getRestaurantDistanceMeters(restaurant, getUserCoordinates());
    if (!Number.isFinite(distanceMeters)) {
      return null;
    }

    return distanceMeters / 1000;
  }

  function sortRestaurants(records: RestaurantRecord[]) {
    const nextRecords = [...records];
    const userCoordinates = getUserCoordinates();

    nextRecords.sort((a, b) => {
      if (sortBy.value === 'name-asc') {
        return collator.compare(a.name, b.name);
      }

      if (sortBy.value === 'name-desc') {
        return collator.compare(b.name, a.name);
      }

      if (sortBy.value === 'distance-asc') {
        const aDistance = getRestaurantDistanceMeters(a, userCoordinates);
        const bDistance = getRestaurantDistanceMeters(b, userCoordinates);

        if (aDistance !== bDistance) return aDistance - bDistance;
        return collator.compare(a.name, b.name);
      }

      const aLatestMenu = getLatestMenuTimestamp(a);
      const bLatestMenu = getLatestMenuTimestamp(b);

      if (sortBy.value === 'menu-newest') {
        if (bLatestMenu !== aLatestMenu) return bLatestMenu - aLatestMenu;
        return collator.compare(a.name, b.name);
      }

      if (aLatestMenu !== bLatestMenu) return aLatestMenu - bLatestMenu;
      return collator.compare(a.name, b.name);
    });

    return nextRecords;
  }

  function findRestaurantIndexById(id: string) {
    return restaurants.value.findIndex((restaurant) => restaurant.id === id);
  }

  const coolDownTimers = new Map<string, ReturnType<typeof setTimeout>>();

  function clearCoolDownTimer(id: string) {
    if (coolDownTimers.has(id)) {
      clearTimeout(coolDownTimers.get(id));
      coolDownTimers.delete(id);
    }
  }

  function startCoolDownTimer(id: string, seconds: number) {
    coolDownTimers.set(
      id,
      setTimeout(() => {
        const idx = findRestaurantIndexById(id);
        if (idx !== -1 && restaurants.value[idx].status === RestaurantStatus.COOLDOWN) {
          restaurants.value[idx].status = RestaurantStatus.IDLE;
        }
        coolDownTimers.delete(id);
      }, seconds * 1000)
    );
  }

  function setRestaurantStatus(id: string, status: RestaurantStatus, coolDownSeconds?: number) {
    const index = findRestaurantIndexById(id);
    if (index === -1) return;
    clearCoolDownTimer(id);
    restaurants.value[index].status = status;
    if (status === RestaurantStatus.COOLDOWN && typeof coolDownSeconds === 'number' && coolDownSeconds > 0) {
      startCoolDownTimer(id, coolDownSeconds);
    }
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
    clearCoolDownTimer(id);
    const index = findRestaurantIndexById(id);
    if (index === -1) return;
    restaurants.value.splice(index, 1);
  }

  let unsubscribeStatus: (() => void) | null = null;
  let unsubscribeRestaurants: (() => void) | null = null;

  async function subscribeRealtime() {
    unsubscribeStatus = await backendClient.subscribeRestaurantStatus((event: RestaurantStatusEvent & { cooldownSeconds?: number; coolDownSeconds?: number }) => {
      setRestaurantStatus(
        event.id,
        event.status as RestaurantStatus,
        typeof event.coolDownSeconds === 'number' ? event.coolDownSeconds : event.cooldownSeconds
      );
    });

    unsubscribeRestaurants = await backendClient.subscribeRestaurants((action, record) => {
      if (action === 'delete') {
        removeRestaurant(record.id);
        return;
      }

      if (action === 'create' || action === 'update') {
        mergeRestaurant(record);
      }
    });
  }

  function unsubscribeRealtime() {
    unsubscribeStatus?.();
    unsubscribeStatus = null;
    unsubscribeRestaurants?.();
    unsubscribeRestaurants = null;
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

  function applyRestaurants(records: RestaurantRecord[]) {
    const nextRestaurants = backendClient.prepareRestaurants(records);
    // Clear timers for restaurants that are no longer present
    const nextIds = new Set(nextRestaurants.map((r) => r.id));
    for (const id of Array.from(coolDownTimers.keys())) {
      if (!nextIds.has(id)) {
        clearCoolDownTimer(id);
      }
    }
    restaurants.value = nextRestaurants;
    return nextRestaurants;
  }

  async function fetchRestaurants() {
    isLoading.value = true;

    try {
      const nextRestaurants = applyRestaurants(await backendClient.fetchRestaurants());
      await preloadThumbnails(nextRestaurants);
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
    // Track dependencies so sorting updates when they change
    void coords.value;
    void sortBy.value;
    const query = searchQuery.value.trim().toLocaleLowerCase('de-DE');

    if (!query) {
      return sortRestaurants(restaurants.value);
    }

    return sortRestaurants(
      restaurants.value.filter((restaurant) => {
        const haystack = [restaurant.name, restaurant.group, restaurant.address, restaurant.website, ...restaurant.tags]
          .filter(Boolean)
          .join(' ')
          .toLocaleLowerCase('de-DE');

        return haystack.includes(query);
      })
    );
  });

  const groupedRestaurants = computed<Record<string, RestaurantRecord[]>>(() => {
    const groups: Record<string, RestaurantRecord[]> = {};
    const nextRestaurants = filteredRestaurants.value;

    if (groupBy.value === 'none') {
      groups.Alle = nextRestaurants;
      return groups;
    }


    const favoriteRestaurants = nextRestaurants.filter((restaurant) => favorites.value[restaurant.id]);
    if (favoriteRestaurants.length) {
      groups.Favoriten = favoriteRestaurants;
    }

    for (const restaurant of nextRestaurants) {
      if (favorites.value[restaurant.id]) continue;

      const key = restaurant.group || 'Ohne Gruppe';
      if (!groups[key]) {
        groups[key] = [];
      }

      groups[key].push(restaurant);
    }

    return groups;
  });

  function applySearch(query: string) {
    searchQuery.value = query;
  }

  return {
    restaurants,
    searchQuery,
    sortBy,
    groupBy,
    groupedRestaurants,
    filteredRestaurants,
    restaurantSortOptions,
    restaurantGroupingOptions,
    initialize,
    unsubscribeRealtime,
    isLoading,
    fetchRestaurants,
    getFileUrl,
    getMapUrl,
    getPhoneUrl,
    applySearch,
    requestGeolocation,
    coords,
    geolocationLoading,
    getRestaurantDistanceKm,
  };
});
