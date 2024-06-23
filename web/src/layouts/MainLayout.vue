<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import { ref } from 'vue';
import { Dark } from 'quasar';
import { Screen } from 'quasar';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};
</script>

<template>
  <q-layout view="hHh LpR fFf">
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

  <Teleport to="body">
    <q-btn
      v-if="Screen.lt.md"
      fab
      round
      color="primary"
      icon="fa-solid fa-bars"
      @click="toggleLeftDrawer"
      :class="{ 'text-white': Dark.isActive, 'text-black': !Dark.isActive }"
      style="z-index: 1; position: fixed; bottom: 1rem; right: 1rem"
    />
  </Teleport>
</template>
