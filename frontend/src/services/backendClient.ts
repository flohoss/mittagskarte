import PocketBase from 'pocketbase';
import type { RestaurantRecord, RestaurantStatusEvent } from '../models/restaurant';
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

async function authenticate(identity: string, password: string) {
  await client.collection('users').authWithPassword(identity, password);
}

function clearAuthentication() {
  client.authStore.clear();
}

async function fetchRestaurants() {
  const records = await client.collection('restaurants').getFullList({
    sort: 'group,name',
  });

  return records as unknown as RestaurantRecord[];
}

async function subscribeRestaurantStatus(handler: (event: RestaurantStatusEvent) => void) {
  await client.realtime.subscribe('restaurants/status', (event: RestaurantStatusEvent) => {
    handler(event);
  });
}

async function subscribeRestaurants(handler: (action: string, record: RestaurantRecord) => void) {
  await client.collection('restaurants').subscribe('*', (event: { action: string; record: RestaurantRecord }) => {
    handler(event.action, event.record);
  });
}

function getFileUrl(record: RestaurantRecord) {
  return client.files.getURL(record as Record<string, unknown>, String(record.thumbnail ?? ''));
}

export const backendClient = {
  onAuthChange,
  getAuthToken,
  getAuthRecord,
  authenticate,
  clearAuthentication,
  fetchRestaurants,
  subscribeRestaurantStatus,
  subscribeRestaurants,
  getFileUrl,
};
