<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import { ref } from 'vue';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};
</script>

<template>
  <q-layout view="hHh LpR fFf">
    <q-header class="bg-dark text-white">
      <q-toolbar>
        <q-btn dense flat round icon="fa-solid fa-bars" @click="toggleLeftDrawer" />

        <q-toolbar-title>
          <q-avatar>
            <img src="/favicon/android-chrome-192x192.png" />
          </q-avatar>
          Mittagstisch
        </q-toolbar-title>
      </q-toolbar>
    </q-header>

    <q-drawer show-if-above v-model="leftDrawerOpen" side="left">
      <MainNavigation />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>
