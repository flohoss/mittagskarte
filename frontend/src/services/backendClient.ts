import PocketBase from 'pocketbase';
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
  return client.authStore.record as Record<string, unknown> | null;
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

  try {
    const parts = token.split('.');
    if (parts.length < 2) {
      return clearAuthAndReturnFalse();
    }

    // Decode JWT payload (base64url) and check exp locally to avoid refresh loops.
    const payload = JSON.parse(atob(parts[1].replace(/-/g, '+').replace(/_/g, '/'))) as { exp?: number };
    if (typeof payload.exp !== 'number') {
      return clearAuthAndReturnFalse();
    }

    const now = Math.floor(Date.now() / 1000);
    if (payload.exp <= now) {
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

function buildFileUrl(record: Record<string, unknown>, fileName: string) {
  const url = client.files.getURL(record, fileName);
  return normalizeFileUrl(url);
}

async function fetchRestaurants() {
  const records = await client.collection('restaurants').getFullList({
    sort: 'group,name',
    expand: 'menus',
  });

  return records as unknown as RestaurantRecord[];
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
  return buildFileUrl(record as Record<string, unknown>, String(record.thumbnail ?? ''));
}

function getMenuFileUrl(menu: MenuRecord) {
  return buildFileUrl(menu as unknown as Record<string, unknown>, menu.file);
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
