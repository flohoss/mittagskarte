<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import { ref } from 'vue';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};
</script>

<template>
  <q-layout view="lHh LpR lFf">
    <q-header class="bg-dark text-white">
      <q-toolbar>
        <q-btn flat round icon="fa-solid fa-bars" @click="toggleLeftDrawer" />

        <q-space />

        <q-btn flat round icon="fa-solid fa-clock" to="/" />
        <q-btn flat round icon="fa-solid fa-gear" to="/settings" />
      </q-toolbar>
    </q-header>

    <q-drawer show-if-above v-model="leftDrawerOpen" side="left" bordered>
      <MainNavigation />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>
