<script setup lang="ts">
import { computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { RouterView } from 'vue-router';
import AppShell from './components/AppShell.vue';
import RestaurantsEmpty from './components/state/RestaurantsEmpty.vue';
import RestaurantsLoading from './components/state/RestaurantsLoading.vue';
import { useRestaurants } from './stores/useRestaurants';

const route = useRoute();
const router = useRouter();
const { initialize, restaurants, isLoaded, isLoading } = useRestaurants();
const hasRestaurants = computed(() => restaurants.value.length > 0);

initialize();

watch(
  () => [isLoaded.value, hasRestaurants.value, route.name] as const,
  ([loaded, hasData, routeName]) => {
    if (!loaded || hasData || routeName === 'home') return;
    void router.replace({ name: 'home' });
  },
  { immediate: true }
);
</script>

<template>
  <AppShell title="Schniddzl.de" description="deine Mittagskarte für die Region Stuttgart">
    <Transition name="app-crossfade" mode="out-in">
      <RestaurantsLoading v-if="!isLoaded" key="loading" :is-loading="isLoading" />
      <RestaurantsEmpty v-else-if="!hasRestaurants" key="empty" />
      <div v-else key="content">
        <RouterView v-slot="{ Component }">
          <component :is="Component" />
        </RouterView>
      </div>
    </Transition>
  </AppShell>
</template>

<style scoped>
.app-crossfade-enter-active,
.app-crossfade-leave-active {
  transition: opacity 260ms ease;
}

.app-crossfade-enter-from,
.app-crossfade-leave-to {
  opacity: 0;
}
</style>
