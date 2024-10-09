<script setup lang="ts">
import { ref } from 'vue';
import SettingsForm from 'components/SettingsForm.vue';
import UploadForm from 'src/components/UploadForm.vue';
import RestaurantActions from './RestaurantActions.vue';

const settings = ref(false);
const upload = ref(false);
</script>

<template>
  <div class="flex q-gutter-x-md">
    <RestaurantActions v-if="$q.screen.gt.sm" />
    <div class="row q-gutter-x-sm">
      <q-btn
        size="sm"
        round
        flat
        icon="fa-solid fa-upload"
        @click="upload = true"
      >
        <q-tooltip>Menü hochladen</q-tooltip>
      </q-btn>
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
  <q-dialog v-model="upload" backdrop-filter="blur(4px) saturate(150%)">
    <UploadForm @uploaded="upload = false" />
  </q-dialog>
</template>
