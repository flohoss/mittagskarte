<script setup lang="ts">
import { computed } from 'vue';
import type { RestaurantRecord } from '../models/restaurant';
import MenuPopover from './buttons/RestaurantMenuButton.vue';
import RestaurantRefreshButton from './buttons/RestaurantRefreshButton.vue';
import RestaurantLinkButtons from './buttons/RestaurantLinkButtons.vue';
import ActionIconButton from './buttons/ActionIconButton.vue';
import { useLogin } from '../stores/useLogin';
import { backendClient } from '../services/backendClient';
import { getLatestMenu, getMenuDimensions } from '../utils/menu';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { isAuthenticated } = useLogin();

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
    <ActionIconButton v-else class-name="btn btn-primary" title="Keine Speisekarte verfügbar" ariaLabel="Keine Speisekarte verfügbar" :disabled="true">
      <span class="icon-[fa7-solid--list-alt] btn-icon" aria-hidden="true"></span>
    </ActionIconButton>

    <RestaurantRefreshButton v-if="isAuthenticated" :restaurant="props.restaurant" />

    <RestaurantLinkButtons :restaurant="props.restaurant" />
  </div>
</template>
