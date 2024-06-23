<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import { ref } from 'vue';
import { Dark } from 'quasar';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};
</script>

<template>
  <q-layout view="lHh LpR lFf">
    <q-header
      class="bg-transparent"
      :class="{ 'text-white': Dark.isActive, 'text-black': !Dark.isActive }"
    >
      <q-toolbar>
        <q-btn flat round icon="fa-solid fa-bars" @click="toggleLeftDrawer" />

        <q-space />
      </q-toolbar>
    </q-header>

    <q-drawer
      show-if-above
      v-model="leftDrawerOpen"
      side="left"
      style="
        min-height: 100vh;
        height: 100%;
        display: flex;
        align-items: stretch;
        flex-direction: column;
        justify-content: space-between;
      "
    >
      <MainNavigation />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>
  </q-layout>
</template>
