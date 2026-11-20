import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { RestaurantRecord } from '../models/restaurant';
import { backendClient } from '../services/backendClient';
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

function installStorageStub() {
  const backing = new Map<string, string>();
  const localStorageStub = {
    getItem: (key: string) => (backing.has(key) ? backing.get(key)! : null),
    setItem: (key: string, value: string) => {
      backing.set(key, String(value));
    },
    removeItem: (key: string) => {
      backing.delete(key);
    },
    clear: () => {
      backing.clear();
    },
    key: (index: number) => Array.from(backing.keys())[index] ?? null,
    get length() {
      return backing.size;
    },
  };

  vi.stubGlobal('window', { ...globalThis, localStorage: localStorageStub } as unknown as Window & typeof globalThis);
  vi.stubGlobal('localStorage', localStorageStub);
  vi.stubGlobal('navigator', { geolocation: undefined });
  vi.stubGlobal(
    'Image',
    class {
      onload: (() => void) | null = null;
      onerror: (() => void) | null = null;
    }
  );
}

function seedRestaurants(store: ReturnType<typeof useRestaurants>, records: RestaurantRecord[]) {
  store.restaurants.value = backendClient.prepareRestaurants(records);
}

describe('useRestaurants groupedRestaurants empty states', () => {
  beforeEach(() => {
    installStorageStub();
    vi.resetModules();
  });

  it('flags favorites-only grouping without favorites as no-favorites instead of showing nothing', async () => {
    const { useFavorites } = await import('./useFavorites');
    const { useRestaurants: useRestaurantsStore } = await import('./useRestaurants');
    useFavorites();
    const store = useRestaurantsStore();
    seedRestaurants(store, [restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);

    store.groupBy.value = 'favorites';

    expect(store.hasVisibleRestaurants.value).toBe(false);
    expect(store.emptyViewReason.value).toBe(RestaurantEmptyReason.NO_FAVORITES);
  });

  it('shows favorites for favorites-only grouping when a favorite exists', async () => {
    const { useFavorites } = await import('./useFavorites');
    const { useRestaurants: useRestaurantsStore } = await import('./useRestaurants');
    const { favorites } = useFavorites();
    const store = useRestaurantsStore();
    seedRestaurants(store, [restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);
    favorites.value['2'] = true;

    store.groupBy.value = 'favorites';

    expect(store.hasVisibleRestaurants.value).toBe(true);
    expect(store.emptyViewReason.value).toBe(null);
    expect(store.groupedRestaurants.value.Favoriten.map((entry) => entry.id)).toEqual(['2']);
  });

  it('flags favorites-only grouping with favorites but a filtering search', async () => {
    const { useFavorites } = await import('./useFavorites');
    const { useRestaurants: useRestaurantsStore } = await import('./useRestaurants');
    const { favorites } = useFavorites();
    const store = useRestaurantsStore();
    seedRestaurants(store, [restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);
    favorites.value['2'] = true;

    store.groupBy.value = 'favorites';
    store.applySearch('zzz-no-match');

    expect(store.hasVisibleRestaurants.value).toBe(false);
    expect(store.emptyViewReason.value).toBe(RestaurantEmptyReason.SEARCH);
  });

  it('keeps favorites plus other groups for favorites-on-top grouping with no favorites', async () => {
    const { useRestaurants: useRestaurantsStore } = await import('./useRestaurants');
    const store = useRestaurantsStore();
    seedRestaurants(store, [restaurant('1', 'Alpha', 'Mitte'), restaurant('2', 'Beta', 'Sued')]);

    store.groupBy.value = 'none-favorites';

    expect(store.hasVisibleRestaurants.value).toBe(true);
    expect(store.emptyViewReason.value).toBe(null);
  });
});
