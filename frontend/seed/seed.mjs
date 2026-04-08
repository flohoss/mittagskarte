#!/usr/bin/env node
/**
 * Seed script — creates all restaurants (with selectors and thumbnails) in PocketBase.
 *
 * Run via the existing yarn service (backend must be running):
 * docker compose run -e PB_EMAIL=REPLACED -e PB_PASSWORD=REPLACED --entrypoint node --rm yarn /app/seed/seed.mjs
 */

import PocketBase from 'pocketbase';
import { readFileSync, existsSync } from 'fs';
import { resolve, dirname, join } from 'path';
import { fileURLToPath } from 'url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const root = join(__dirname, 'thumbnails');

// ---------------------------------------------------------------------------
// Config
// ---------------------------------------------------------------------------
const PB_URL = process.env.PB_URL ?? 'http://host.docker.internal:8090';
const PB_EMAIL = process.env.PB_EMAIL;
const PB_PASSWORD = process.env.PB_PASSWORD;

if (!PB_EMAIL || !PB_PASSWORD) {
  console.error('PB_EMAIL and PB_PASSWORD environment variables are required.');
  process.exit(1);
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

/** Read a local file as a Blob for PocketBase SDK uploads */
function fileBlob(filePath, mimeType = 'image/webp') {
  const buffer = readFileSync(filePath);
  return new File([buffer], filePath.split('/').pop(), { type: mimeType });
}

/**
 * Derive the thumbnail filename from the stored thumbnail field.
 * The JSON has names like "avada_gevrx1cgg2.webp" — we strip PocketBase's
 * random suffix and match against images/thumbnails/<stem>.webp.
 */
function resolveThumbnail(thumbnailField) {
  const candidate = join(root, thumbnailField);
  return existsSync(candidate) ? candidate : null;
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------
const client = new PocketBase(PB_URL);

console.log(`Connecting to ${PB_URL} as ${PB_EMAIL}...`);
await client.collection('_superusers').authWithPassword(PB_EMAIL, PB_PASSWORD);
console.log('Authenticated.\n');

const data = JSON.parse(readFileSync(join(__dirname, 'restaurants.json'), 'utf-8'));
const restaurants = data.restaurants;

for (const r of restaurants) {
  console.log(`→ ${r.name}`);

  // 1. Create restaurant without navigate first
  const formData = new FormData();
  formData.append('name', r.name);
  formData.append('group', r.group ?? '');
  formData.append('address', r.address ?? '');
  formData.append('website', r.website ?? '');
  formData.append('phone', r.phone ?? '');
  formData.append('method', r.method);
  formData.append('content_type', r.content_type ?? '');
  formData.append('cron', r.cron ?? '');

  for (const day of r.rest_days ?? []) {
    formData.append('rest_days', day);
  }

  formData.append('tags', JSON.stringify(r.tags ?? []));

  if (r.thumbnail) {
    const thumbPath = resolveThumbnail(r.thumbnail);
    if (thumbPath) {
      formData.append('thumbnail', fileBlob(thumbPath));
      console.log(`   thumbnail: ${thumbPath.split('/').pop()}`);
    } else {
      console.warn(`   ⚠ thumbnail not found for "${r.thumbnail}"`);
    }
  }

  const created = await client.collection('restaurants').create(formData);
  console.log(`   created: ${created.id}`);

  // 2. Create selectors with restaurant reference already set
  const navigateIds = [];
  for (const nav of r.navigate ?? []) {
    const selector = await client.collection('selectors').create({
      order: nav.order,
      locator: nav.locator,
      attribute: nav.attribute ?? '',
      style: nav.style ?? '',
      restaurant: created.id,
    });
    navigateIds.push(selector.id);
    console.log(`   selector #${nav.order}: ${selector.id}`);
  }

  // 3. Patch restaurant with selector IDs
  if (navigateIds.length > 0) {
    await client.collection('restaurants').update(created.id, { navigate: navigateIds });
  }
  console.log();
}

console.log(`Done — seeded ${restaurants.length} restaurants.`);
