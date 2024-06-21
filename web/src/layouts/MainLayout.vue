<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import SettingsForm from 'components/SettingsForm.vue';
import UploadForm from 'src/components/UploadForm.vue';
import { ref } from 'vue';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};

const settings = ref(false);
const upload = ref(false);
</script>

<template>
  <q-layout view="lHh LpR lFf">
    <q-header class="bg-dark text-white">
      <q-toolbar>
        <q-btn flat round icon="fa-solid fa-bars" @click="toggleLeftDrawer" />

        <q-space />

        <q-btn flat round icon="fa-solid fa-upload" @click="upload = true" />
        <q-btn flat round icon="fa-solid fa-gear" @click="settings = true" />
      </q-toolbar>
    </q-header>

    <q-drawer show-if-above v-model="leftDrawerOpen" side="left" bordered>
      <MainNavigation />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

    <q-dialog v-model="settings">
      <SettingsForm @close="settings = false" />
    </q-dialog>
    <q-dialog v-model="upload">
      <UploadForm @close="upload = false" />
    </q-dialog>
  </q-layout>
</template>
