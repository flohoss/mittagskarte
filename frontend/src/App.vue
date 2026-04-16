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
const { initialize, filteredRestaurants, isLoading } = useRestaurants();
const hasRestaurants = computed(() => filteredRestaurants.value.length > 0);

initialize();

watch(
  () => [isLoading.value, hasRestaurants.value, route.name] as const,
  ([loading, hasData, routeName]) => {
    if (loading || hasData || routeName === 'home') return;
    void router.replace({ name: 'home' });
  },
  { immediate: true }
);
</script>

<template>
  <AppShell title="Schniddzl.de" description="deine Mittagskarte für die Region Stuttgart">
    <div class="app-stage">
      <div class="app-content" :class="{ 'app-content--ready': !isLoading }" :aria-busy="isLoading">
        <RestaurantsEmpty v-if="!isLoading && !hasRestaurants" />
        <div v-else-if="hasRestaurants">
          <RouterView v-slot="{ Component }">
            <component :is="Component" />
          </RouterView>
        </div>
      </div>

      <Transition name="app-overlay-fade" appear>
        <div v-if="isLoading" class="app-loading-layer">
          <RestaurantsLoading :is-loading="isLoading" />
        </div>
      </Transition>
    </div>
  </AppShell>
</template>

<style scoped>
.app-stage {
  position: relative;
  min-height: 20vh;
}

.app-content {
  position: relative;
  opacity: 0;
  transition: opacity 180ms ease-out;
}

.app-content--ready {
  opacity: 1;
}

.app-loading-layer {
  position: absolute;
  inset: 0;
  z-index: 1;
  background: var(--color-base-100);
}

.app-overlay-fade-enter-active,
.app-overlay-fade-leave-active {
  transition: opacity 180ms ease-out;
}

.app-overlay-fade-enter-from,
.app-overlay-fade-leave-to {
  opacity: 0;
}
</style>
