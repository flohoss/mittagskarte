import PocketBase from 'pocketbase';
import type { RecordModel } from 'pocketbase';
import type { MenuRecord, RestaurantRecord, RestaurantStatusEvent } from '../models/restaurant';
import { BackendURL } from '../config';

const client = new PocketBase(BackendURL);

type AuthChangeHandler = (token: string) => void;

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
  const url = client.files.getURL(record, fileName);
  return normalizeFileUrl(url);
}

async function fetchRestaurants() {
  const records = await client.collection('restaurants').getFullList<RestaurantRecord>({
    sort: 'group,name',
    expand: 'menus',
  });

  return records;
}

async function subscribeRestaurantStatus(handler: (event: RestaurantStatusEvent) => void) {
  await client.realtime.subscribe('restaurants/status', (event: RestaurantStatusEvent) => {
    handler(event);
  });
}

async function subscribeRestaurants(handler: (action: string, record: RestaurantRecord) => void) {
  await client.collection('restaurants').subscribe(
    '*',
    (event: { action: string; record: RestaurantRecord }) => {
      handler(event.action, event.record);
    },
    { expand: 'menus' }
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
  getFileUrl,
  getMenuFileUrl,
  uploadMenu,
  triggerScrape,
};
