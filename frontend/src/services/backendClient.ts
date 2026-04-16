import PocketBase from 'pocketbase';
import type { RecordModel } from 'pocketbase';
import type { MenuRecord, RestaurantRecord, RestaurantStatusEvent } from '../models/restaurant';
import { sortMenusByCreatedDesc } from '../utils/menu';
import { BackendURL } from '../config';

const client = new PocketBase(BackendURL);

const RESTAURANT_LIST_OPTIONS = {
  sort: 'group,name',
  expand: 'menus',
} as const;

type AuthChangeHandler = (token: string) => void;
type RestaurantInput = Partial<RestaurantRecord>;
type RestaurantSubscriptionEvent = { action: string; record: RestaurantRecord };

function normalizeMenu(menu: Partial<MenuRecord> | null | undefined): MenuRecord {
  return {
    ...(menu as MenuRecord),
    id: String(menu?.id ?? ''),
    file: String(menu?.file ?? ''),
    hash: String(menu?.hash ?? ''),
    created: String(menu?.created ?? ''),
  };
}

function normalizeRestaurant(record: RestaurantInput): RestaurantRecord {
  const menus = sortMenusByCreatedDesc((record.expand?.menus ?? []).map((menu) => normalizeMenu(menu)));

  return {
    ...(record as RestaurantRecord),
    id: String(record.id ?? ''),
    name: String(record.name ?? ''),
    group: String(record.group ?? ''),
    address: String(record.address ?? ''),
    website: String(record.website ?? ''),
    phone: String(record.phone ?? ''),
    tags: Array.isArray(record.tags) ? record.tags.map(String) : [],
    rest_days: Array.isArray(record.rest_days) ? record.rest_days.map(String) : [],
    method: String(record.method ?? ''),
    status: String(record.status ?? 'idle'),
    updated: String(record.updated ?? ''),
    thumbnail: String(record.thumbnail ?? ''),
    last_check: record.last_check ?? null,
    expand: { menus },
  };
}

function prepareRestaurants(records: RestaurantInput[]) {
  return [...records]
    .map((record) => normalizeRestaurant(record))
    .sort((left, right) => {
      const leftKey = `${left.group} ${left.name}`.trim().toLocaleLowerCase('de-DE');
      const rightKey = `${right.group} ${right.name}`.trim().toLocaleLowerCase('de-DE');
      return leftKey.localeCompare(rightKey, 'de-DE');
    });
}

function onAuthChange(handler: AuthChangeHandler) {
  return client.authStore.onChange((token) => {
    handler(token);
  });
}

function getAuthToken() {
  return client.authStore.token;
}

function getAuthRecord() {
  return client.authStore.record as RecordModel | null;
}

function clearAuthAndReturnFalse() {
  client.authStore.clear();
  return false;
}

async function authenticate(identity: string, password: string) {
  await client.collection('users').authWithPassword(identity, password);
}

async function validateAuthentication() {
  const token = client.authStore.token;
  if (!token) {
    return false;
  }

  if (!client.authStore.isValid) {
    return clearAuthAndReturnFalse();
  }

  try {
    await client.collection('users').authRefresh();
    if (!client.authStore.isValid) {
      return clearAuthAndReturnFalse();
    }

    return true;
  } catch {
    return clearAuthAndReturnFalse();
  }
}

function clearAuthentication() {
  client.authStore.clear();
}

function buildFileUrl(record: RecordModel, fileName: string) {
  if (!fileName) {
    return '';
  }

  const url = client.files.getURL(record, fileName);
  return normalizeFileUrl(url);
}

async function fetchRestaurants() {
  return client.collection('restaurants').getFullList<RestaurantRecord>(RESTAURANT_LIST_OPTIONS);
}

async function subscribeRestaurantStatus(handler: (event: RestaurantStatusEvent) => void) {
  await client.realtime.subscribe('restaurants/status', (event: RestaurantStatusEvent) => {
    handler(event);
  });
}

async function subscribeRestaurants(handler: (action: string, record: RestaurantRecord) => void) {
  await client.collection('restaurants').subscribe(
    '*',
    (event: RestaurantSubscriptionEvent) => {
      handler(event.action, normalizeRestaurant(event.record));
    },
    { expand: RESTAURANT_LIST_OPTIONS.expand }
  );
}

function getFileUrl(record: RestaurantRecord) {
  return buildFileUrl(record, String(record.thumbnail ?? ''));
}

function getMenuFileUrl(menu: MenuRecord) {
  return buildFileUrl(menu, menu.file);
}

function normalizeFileUrl(url: string) {
  if (!url) return url;
  if (url.startsWith('/')) return url;
  if (/^https?:\/\//i.test(url)) {
    try {
      const parsed = new URL(url);
      if (typeof window !== 'undefined' && parsed.origin === window.location.origin) {
        const apiPathIndex = parsed.pathname.indexOf('/api/');
        if (apiPathIndex >= 0) {
          return `${parsed.pathname.slice(apiPathIndex)}${parsed.search}${parsed.hash}`;
        }
      }
    } catch {
      return url;
    }

    return url;
  }

  return `/${url}`;
}

async function uploadMenu(restaurantId: string, file: File) {
  const formData = new FormData();
  formData.append('restaurant', restaurantId);
  formData.append('file', file);
  await client.collection('menus').create(formData);
}

async function triggerScrape(restaurantId: string) {
  await client.send('/api/restaurants/scrape', {
    method: 'POST',
    body: JSON.stringify({ id: restaurantId }),
    headers: { 'Content-Type': 'application/json' },
  });
}

export const backendClient = {
  onAuthChange,
  getAuthToken,
  getAuthRecord,
  authenticate,
  validateAuthentication,
  clearAuthentication,
  fetchRestaurants,
  subscribeRestaurantStatus,
  subscribeRestaurants,
  normalizeRestaurant,
  prepareRestaurants,
  getFileUrl,
  getMenuFileUrl,
  uploadMenu,
  triggerScrape,
};
