import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { RestaurantRecord } from '../models/restaurant';
import { RestaurantEmptyReason, useRestaurants } from './useRestaurants';

function restaurant(id: string, name: string, group: string): RestaurantRecord {
  return {
    id,
    name,
    slug: name.toLowerCase().replace(/\s+/g, '-'),
    group,
    latitude: 0,
    longitude: 0,
    address: '',
    website: '',
    phone: '',
    tags: [],
    rest_days: [],
    method: '',
    cron: '',
    holiday_until: '',
    status: 'idle',
    updated: '',
    thumbnail: '',
    last_check: null,
    expand: { menus: [] },
  } as unknown as RestaurantRecord;
}

type StorageShim = Record<string, string>;

function installStorageStub() {
  const backing: StorageShim = {};
  const localStorageStub = {
    getItem: (key: string) => (key in backing ? backing[key] : null),
    setItem: (key: string, value: string) => {
      backing[key] = String(value);
    },
    removeItem: (key: string) => {
      delete backing[key];
    },
    clear: () => {
      for (const key of Object.keys(backing)) delete backing[key];
    },
    key: (index: number) => Object.keys(backing)[index] ?? null,
    get length() {
      return Object.keys(backing).length;
    },
  };

  vi.stubGlobal('window', { ...globalThis, localStorage: localStorageStub } as unknown as Window & typeof globalThis);
  vi.stubGlobal('localStorage', localStorageStub);
  vi.stubGlobal('navigator', { geolocation: undefined });
  vi.stubGlobal('Image', class {
    onload: (() => void) | null = null;
    onerror: (() => void) | null = null;
  });
}

describe('useRestaurants groupedRestaurants empty states', () => {
  beforeEach(() => {
    installStorageStub();
    vi.resetModules();
  });

  it('flags favorites-only grouping without favorites as no-favorites instead of showing nothing', async () => {
    const { useFavorites } = await import('./useFavorites');
    const { backendClient } = await import('../services/backendClient');
    backendClient.prepareRestaurants = (records: RestaurantRecord[]) => records;

    const useRestaurants = (await import('./useRestaurants')).useRestaurants;
    const store = useRestaurants();
    store.applyRestaurants([restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);

    store.groupBy = 'favorites';

    expect(store.hasVisibleRestaurants).toBe(false);
    expect(store.emptyViewReason).toBe(RestaurantEmptyReason.NO_FAVORITES);
  });

  it('shows favorites for favorites-only grouping when favorites exist', async () => {
    const { useFavorites } = await import('./useFavorites');
    const { backendClient } = await import('../services/backendClient');
    backendClient.prepareRestaurants = (records: RestaurantRecord[]) => records;

    const useRestaurants = (await import('./useRestaurants')).useRestaurants;
    const favorites = useFavorites();
    const store = useRestaurants();
    store.applyRestaurants([restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);
    favorites.value['2'] = true;

    store.groupBy = 'favorites';

    expect(store.hasVisibleRestaurants).toBe(true);
    expect(store.emptyViewReason).toBe(null);
    expect(store.groupedRestaurants.value.Favoriten.map((entry) => entry.id)).toEqual(['2']);
  });

  it('keeps favorites plus other groups for favorites-on-top grouping with no favorites', async () => {
    const { backendClient } = await import('../services/backendClient');
    backendClient.prepareRestaurants = (records: RestaurantRecord[]) => records;

    const useRestaurants = (await import('./useRestaurants')).useRestaurants;
    const store = useRestaurants();
    store.applyRestaurants([restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);

    store.groupBy = 'none-favorites';

    expect(store.hasVisibleRestaurants).toBe(true);
    expect(store.emptyViewReason).toBe(null);
  });
});
