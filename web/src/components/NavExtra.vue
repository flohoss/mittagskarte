<script setup lang="ts">
import { ref } from 'vue';
import SettingsForm from 'components/SettingsForm.vue';
import UploadForm from 'src/components/UploadForm.vue';

const settings = ref(false);
const upload = ref(false);
</script>

<template>
  <q-list class="q-mt-lg">
    <q-item clickable @click="settings = true">
      <q-item-section avatar>
        <q-icon name="fa-solid fa-gear" size="1.1rem" class="q-ml-sm" />
      </q-item-section>

      <q-item-section class="text-caption">Einstellungen</q-item-section>
    </q-item>
    <q-item clickable @click="upload = true">
      <q-item-section avatar>
        <q-icon name="fa-solid fa-upload" size="1.1rem" class="q-ml-sm" />
      </q-item-section>

      <q-item-section class="text-caption">Menü hochladen</q-item-section>
    </q-item>
  </q-list>

  <q-dialog v-model="settings">
    <SettingsForm />
  </q-dialog>
  <q-dialog v-model="upload" no-esc-dismiss no-backdrop-dismiss>
    <UploadForm @uploaded="upload = false" />
  </q-dialog>
</template>
