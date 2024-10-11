<script setup lang="ts">
import { computed, ref } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { services_CleanRestaurant } from 'src/openapi';
import { useQuasar } from 'quasar';
import RestaurantInfo from './RestaurantInfo.vue';
import SettingsForm from './SettingsForm.vue';

const $q = useQuasar();
defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
});

const iconSize = computed(() => {
  if ($q.screen.lt.md) {
    return 'md';
  }
  return 'sm';
});

const settings = ref(false);
</script>

<template>
  <div class="row q-gutter-x-sm">
    <RestaurantInfo
      v-if="$q.screen.gt.sm && restaurant.image_url !== ''"
      :restaurant="restaurant"
      :icon-size="iconSize"
    />
    <RestaurantActions :restaurant="restaurant" :icon-size="iconSize" />
    <q-btn
      :size="iconSize"
      round
      flat
      icon="fa-solid fa-gear"
      @click="settings = true"
      id="settings-btn"
      aria-label="Einstellungen"
    >
      <q-tooltip>Einstellungen</q-tooltip>
    </q-btn>

    <q-dialog v-model="settings" backdrop-filter="blur(4px) saturate(150%)">
      <SettingsForm />
    </q-dialog>
  </div>
</template>
