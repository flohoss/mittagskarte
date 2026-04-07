<script setup lang="ts">
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { useRestaurants } from '../stores/useRestaurants';
import MenuHistoryGrid from '../components/menu/MenuHistoryGrid.vue';
import { sortMenusByCreatedDesc } from '../utils/menu';

const route = useRoute();
const router = useRouter();
const { restaurants } = useRestaurants();

const restaurantId = computed(() => String(route.params.restaurantId ?? '').trim());
const restaurant = computed(() => restaurants.value.find((entry) => entry.id === restaurantId.value) ?? null);

const sortedMenus = computed(() => sortMenusByCreatedDesc(restaurant.value?.expand?.menus ?? []));

const shouldRedirectToHome = computed(() => {
  if (!restaurantId.value) return true;
  if (!restaurant.value) return true;
  if (!sortedMenus.value.length) return true;
  return false;
});

watch(
  () => shouldRedirectToHome.value,
  (shouldRedirect) => {
    if (!shouldRedirect) return;
    void router.replace({ name: 'home' });
  },
  { immediate: true }
);
</script>

<template>
  <div class="grid gap-4 md:gap-6">
    <MenuHistoryGrid :menus="sortedMenus" />
  </div>
</template>
