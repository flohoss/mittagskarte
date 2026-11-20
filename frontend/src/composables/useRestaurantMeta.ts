import { computed, type ComputedRef, type Ref } from 'vue';
import type { RestaurantRecord } from '../models/restaurant';
import { useFavorites } from '../stores/useFavorites';
import { useNow } from './useNow';
import { getLatestMenu } from '../utils/menu';
import { formatRelativePastLabel } from '../utils/date';
import { getMenuFreshnessMeta, type MenuFreshnessMeta } from '../utils/menuFreshness';
import { isOnHoliday as checkIsOnHoliday, formatHolidayLabel } from '../utils/holiday';
import { useRestaurants } from '../stores/useRestaurants';

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

type RestaurantSource = Ref<RestaurantRecord> | ComputedRef<RestaurantRecord>;

function toRestaurantSource(restaurant: RestaurantSource | (() => RestaurantRecord)): ComputedRef<RestaurantRecord> {
  if (typeof restaurant === 'function') {
    return computed(() => restaurant());
  }
  return computed(() => restaurant.value);
}

export type RestaurantMeta = {
  isClosed: ComputedRef<boolean>;
  isOnHoliday: ComputedRef<boolean>;
  holidayLabel: ComputedRef<string>;
  isFavorited: ComputedRef<boolean>;
  thumbnailUrl: ComputedRef<string>;
  latestMenuCreated: ComputedRef<string | null>;
  menuFreshness: ComputedRef<MenuFreshnessMeta | null>;
  isManualUpload: ComputedRef<boolean>;
  lastCheckText: ComputedRef<string>;
  lastCheckTitle: ComputedRef<string>;
  distanceKm: ComputedRef<number | null>;
  showDistance: ComputedRef<boolean>;
  distanceLabel: ComputedRef<string>;
  initials: ComputedRef<string>;
  isDimmed: ComputedRef<boolean>;
};

export function useRestaurantMeta(restaurantSource: RestaurantSource | (() => RestaurantRecord)): RestaurantMeta {
  const { sortBy, coords, getFileUrl, getRestaurantDistanceKm } = useRestaurants();
  const { isFavorite } = useFavorites();

  const restaurant = toRestaurantSource(restaurantSource);

  const nowMs = useNow(30_000);
  const currentWeekday = computed(() => WEEKDAYS[new Date(nowMs.value).getDay()]);

  const isClosed = computed(() => restaurant.value.rest_days.includes(currentWeekday.value));
  const isOnHoliday = computed(() => checkIsOnHoliday(restaurant.value.holiday_until, nowMs.value));
  const holidayLabel = computed(() => {
    if (!isOnHoliday.value) return '';
    return formatHolidayLabel(restaurant.value.holiday_until);
  });
  const isFavorited = computed(() => isFavorite(restaurant.value.id));
  const thumbnailUrl = computed(() => getFileUrl(restaurant.value));
  const latestMenuCreated = computed(() => {
    return getLatestMenu(restaurant.value.expand?.menus)?.created ?? null;
  });
  const lastCheck = computed(() => {
    return restaurant.value.last_check ?? null;
  });
  const menuFreshness = computed(() => {
    if (!latestMenuCreated.value) return null;
    return getMenuFreshnessMeta({
      menuDate: latestMenuCreated.value,
      cron: restaurant.value.cron,
      lastCheck: lastCheck.value,
      now: nowMs.value,
    });
  });
  const isManualUpload = computed(() => restaurant.value.method === 'upload');
  const lastCheckText = computed(() => {
    if (isManualUpload.value) {
      if (!latestMenuCreated.value) return '';
      return `${formatRelativePastLabel(latestMenuCreated.value, nowMs.value)} zuletzt aktualisiert`;
    }
    if (!lastCheck.value) return '';
    return `${formatRelativePastLabel(lastCheck.value.at, nowMs.value)} zuletzt versucht`;
  });
  const distanceKm = computed(() => getRestaurantDistanceKm(restaurant.value));
  const showDistance = computed(() => Boolean(sortBy.value === 'distance-asc' && coords.value && distanceKm.value !== null));
  const distanceLabel = computed(() => {
    if (distanceKm.value === null) return '';

    if (distanceKm.value < 1) {
      return `~${Math.round(distanceKm.value * 1000)} m`;
    }

    if (distanceKm.value < 10) {
      return `~${distanceKm.value.toFixed(1)} km`;
    }

    return `~${Math.round(distanceKm.value)} km`;
  });
  const lastCheckTitle = computed(() => {
    if (isManualUpload.value) {
      return latestMenuCreated.value ? 'Zuletzt manuell aktualisiert' : '';
    }
    if (!lastCheck.value) return '';
    if (lastCheck.value.status === 'success') {
      return 'Erfolgreich aktualisiert';
    }
    if (lastCheck.value.status === 'not_changed') {
      return 'Keine Änderung gefunden';
    }
    if (lastCheck.value.status === 'error') {
      return `Fehler${lastCheck.value.detail ? `: ${lastCheck.value.detail}` : ''}`;
    }
    return '';
  });
  const initials = computed(() => {
    return getInitials(restaurant.value.name);
  });
  const isDimmed = computed(() => isClosed.value || isOnHoliday.value);

  return {
    isClosed,
    isOnHoliday,
    holidayLabel,
    isFavorited,
    thumbnailUrl,
    latestMenuCreated,
    menuFreshness,
    isManualUpload,
    lastCheckText,
    lastCheckTitle,
    distanceKm,
    showDistance,
    distanceLabel,
    initials,
    isDimmed,
  };
}

export function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('');
}
