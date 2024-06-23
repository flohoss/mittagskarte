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
      <q-btn
        v-if="!Screen.gt.sm"
        fab
        round
        color="primary"
        icon="fa-solid fa-bars"
        @click="toggleLeftDrawer"
        :class="{ 'text-white': Dark.isActive, 'text-black': !Dark.isActive }"
        class="drawer-btn"
      />
      <router-view />
    </q-page-container>
  </q-layout>
</template>

<style scoped>
.drawer-btn {
  z-index: 1;
  position: absolute;
  bottom: 1rem;
  right: 1rem;
}
</style>
