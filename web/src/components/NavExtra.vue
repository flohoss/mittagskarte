<script setup lang="ts">
import { ref } from 'vue';
import SettingsForm from 'components/SettingsForm.vue';
import UploadForm from 'src/components/UploadForm.vue';
import GlobalSearch from './GlobalSearch.vue';

const settings = ref(false);
const upload = ref(false);
</script>

<template>
  <div class="row q-gutter-x-xs">
    <q-btn size="sm" flat icon="fa-solid fa-upload" @click="upload = true" />
    <q-btn size="sm" flat icon="fa-solid fa-gear" @click="settings = true" />
    <GlobalSearch v-if="$q.screen.gt.sm" />
  </div>

  <q-dialog v-model="settings" backdrop-filter="blur(4px) saturate(150%)">
    <SettingsForm />
  </q-dialog>
  <q-dialog v-model="upload" backdrop-filter="blur(4px) saturate(150%)">
    <UploadForm @uploaded="upload = false" />
  </q-dialog>
</template>
