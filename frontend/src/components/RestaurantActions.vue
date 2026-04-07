<script setup lang="ts">
import { computed } from 'vue';
import type { RestaurantRecord } from '../models/restaurant';
import MenuPopover from './buttons/RestaurantMenuButton.vue';
import RestaurantRefreshButton from './buttons/RestaurantRefreshButton.vue';
import ActionIconButton from './buttons/ActionIconButton.vue';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import Fa7SolidPhone from '~icons/fa7-solid/phone';
import Fa7SolidMap from '~icons/fa7-solid/map';
import Fa7SolidGlobe from '~icons/fa7-solid/globe';
import { useRestaurants } from '../stores/useRestaurants';
import { useLogin } from '../stores/useLogin';
import { backendClient } from '../services/backendClient';
import { getLatestMenu, getMenuDimensions } from '../utils/menu';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { getMapUrl, getPhoneUrl } = useRestaurants();
const { isAuthenticated } = useLogin();
const mapUrl = computed(() => (props.restaurant.address ? getMapUrl(props.restaurant) : ''));
const phoneUrl = computed(() => (props.restaurant.phone ? getPhoneUrl(props.restaurant) : ''));
const mapActionTitle = computed(() => (mapUrl.value ? 'In Google Maps öffnen' : 'Keine Karte verfügbar'));
const mapActionLabel = computed(() => (mapUrl.value ? 'Karte öffnen' : 'Keine Karte verfügbar'));
const phoneActionTitle = computed(() => (phoneUrl.value ? 'Anrufen' : 'Keine Telefonnummer verfügbar'));
const phoneActionLabel = computed(() => (phoneUrl.value ? 'Restaurant anrufen' : 'Keine Telefonnummer verfügbar'));
const websiteUrl = computed(() => props.restaurant.website || '');
const websiteActionTitle = computed(() => (websiteUrl.value ? 'Website öffnen' : 'Keine Website verfügbar'));
const websiteActionLabel = computed(() => (websiteUrl.value ? 'Website öffnen' : 'Keine Website verfügbar'));

const latestMenu = computed(() => {
  return getLatestMenu(props.restaurant.expand?.menus);
});

const latestMenuUrl = computed(() => (latestMenu.value ? backendClient.getMenuFileUrl(latestMenu.value) : null));

const menuDimensions = computed(() => {
  return getMenuDimensions(latestMenu.value);
});
</script>

<template>
  <div :class="['grid gap-1.5', isAuthenticated ? 'grid-cols-5' : 'grid-cols-4']">
    <MenuPopover
      v-if="latestMenuUrl"
      :restaurant="props.restaurant"
      :menu-url="latestMenuUrl"
      :menu-width="menuDimensions.width"
      :menu-height="menuDimensions.height"
    />
    <ActionIconButton v-else class-name="btn btn-primary" title="Keine Speisekarte verfügbar" aria-label="Keine Speisekarte verfügbar" :disabled="true">
      <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
    </ActionIconButton>

    <RestaurantRefreshButton v-if="isAuthenticated" :restaurant="props.restaurant" />

    <ActionIconButton
      :href="mapUrl"
      class-name="btn btn-soft hover:btn-warning"
      :title="mapActionTitle"
      :aria-label="mapActionLabel"
    >
      <Fa7SolidMap class="btn-icon" aria-hidden="true" />
    </ActionIconButton>

    <ActionIconButton :href="phoneUrl" class-name="btn btn-soft hover:btn-success" :title="phoneActionTitle" :aria-label="phoneActionLabel" target="_self">
      <Fa7SolidPhone class="btn-icon" aria-hidden="true" />
    </ActionIconButton>

    <ActionIconButton :href="websiteUrl" class-name="btn btn-soft hover:btn-info" :title="websiteActionTitle" :aria-label="websiteActionLabel">
      <Fa7SolidGlobe class="btn-icon" aria-hidden="true" />
    </ActionIconButton>
  </div>
</template>
