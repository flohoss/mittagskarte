<script setup lang="ts">
import { computed } from 'vue';
import type { RestaurantRecord } from '../models/restaurant';
import MenuPopover from './RestaurantMenuButton.vue';
import RestaurantRefreshButton from './RestaurantRefreshButton.vue';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import Fa7SolidPhone from '~icons/fa7-solid/phone';
import Fa7SolidMap from '~icons/fa7-solid/map';
import Fa7SolidGlobe from '~icons/fa7-solid/globe';
import { useRestaurants } from '../stores/useRestaurants';
import { useLogin } from '../stores/useLogin';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { getMapUrl, getPhoneUrl } = useRestaurants();
const { isAuthenticated } = useLogin();
const mapUrl = computed(() => (props.restaurant.address ? getMapUrl(props.restaurant) : ''));
const phoneUrl = computed(() => (props.restaurant.phone ? getPhoneUrl(props.restaurant) : ''));

const menuDimensions = computed(() => {
  const raw = props.restaurant.menu_dimensions;

  if (!raw || typeof raw !== 'object') {
    return {
      width: null,
      height: null,
    };
  }

  const parsed = raw as Record<string, unknown>;
  const width = typeof parsed.width === 'number' && Number.isFinite(parsed.width) && parsed.width > 0 ? parsed.width : null;
  const height = typeof parsed.height === 'number' && Number.isFinite(parsed.height) && parsed.height > 0 ? parsed.height : null;

  return {
    width,
    height,
  };
});
</script>

<template>
  <div :class="['grid gap-1.5', isAuthenticated ? 'grid-cols-5' : 'grid-cols-4']">
    <MenuPopover v-if="props.restaurant.menu" :menu-url="props.restaurant.menu" :menu-width="menuDimensions.width" :menu-height="menuDimensions.height" />
    <button v-else type="button" class="btn btn-primary" title="Keine Speisekarte verfügbar" aria-label="Keine Speisekarte verfügbar" disabled>
      <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
    </button>

    <RestaurantRefreshButton v-if="isAuthenticated" :restaurant="props.restaurant" />

    <a
      v-if="mapUrl"
      :href="mapUrl"
      target="_blank"
      rel="noreferrer"
      class="btn btn-soft hover:btn-warning"
      title="In Google Maps öffnen"
      aria-label="Karte öffnen"
    >
      <Fa7SolidMap class="btn-icon" aria-hidden="true" />
    </a>
    <button v-else type="button" class="btn btn-soft hover:btn-warning" title="Keine Karte verfügbar" aria-label="Keine Karte verfügbar" disabled>
      <Fa7SolidMap class="btn-icon" aria-hidden="true" />
    </button>

    <a v-if="props.restaurant.phone" :href="phoneUrl" class="btn btn-soft hover:btn-success" title="Anrufen" aria-label="Restaurant anrufen">
      <Fa7SolidPhone class="btn-icon" aria-hidden="true" />
    </a>
    <button
      v-else
      type="button"
      class="btn btn-soft hover:btn-success"
      title="Keine Telefonnummer verfügbar"
      aria-label="Keine Telefonnummer verfügbar"
      disabled
    >
      <Fa7SolidPhone class="btn-icon" aria-hidden="true" />
    </button>

    <a
      v-if="props.restaurant.website"
      :href="props.restaurant.website"
      target="_blank"
      rel="noreferrer"
      class="btn btn-soft hover:btn-info"
      title="Website öffnen"
      aria-label="Website öffnen"
    >
      <Fa7SolidGlobe class="btn-icon" aria-hidden="true" />
    </a>
    <button v-else type="button" class="btn btn-soft hover:btn-info" title="Keine Website verfügbar" aria-label="Keine Website verfügbar" disabled>
      <Fa7SolidGlobe class="btn-icon" aria-hidden="true" />
    </button>
  </div>
</template>
