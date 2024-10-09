<script setup lang="ts">
import { ref } from 'vue';
import SettingsForm from 'components/SettingsForm.vue';
import RestaurantActions from './RestaurantActions.vue';
import { services_CleanRestaurant } from 'src/openapi';

defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
});

const settings = ref(false);
</script>

<template>
  <div class="flex q-gutter-x-md">
    <RestaurantActions v-if="$q.screen.gt.sm" :restaurant="restaurant" />
    <div class="row q-gutter-x-sm">
      <q-btn
        size="sm"
        round
        flat
        icon="fa-solid fa-gear"
        @click="settings = true"
      >
        <q-tooltip>Einstellungen</q-tooltip>
      </q-btn>
    </div>
  </div>

  <q-dialog v-model="settings" backdrop-filter="blur(4px) saturate(150%)">
    <SettingsForm />
  </q-dialog>
</template>
