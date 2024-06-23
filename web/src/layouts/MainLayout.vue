<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import GlobalSearch from 'src/components/GlobalSearch.vue';
import NavTitle from 'src/components/NavTitle.vue';
import { computed, ref } from 'vue';
import { Dark } from 'quasar';
import NavExtra from 'src/components/NavExtra.vue';

const leftDrawerOpen = ref(false);

const toggleLeftDrawer = () => {
  leftDrawerOpen.value = !leftDrawerOpen.value;
};

const bgAndText = computed(() => {
  return Dark.isActive ? 'nav-bg-dark' : 'nav-bg-light';
});

const text = computed(() => {
  return Dark.isActive ? 'text-white' : 'text-black';
});
</script>

<template>
  <q-layout view="hHr Lpr fFr">
    <q-header :class="bgAndText">
      <q-toolbar>
        <q-toolbar-title>
          <NavTitle />
        </q-toolbar-title>
        <q-space />
        <NavExtra />
      </q-toolbar>
    </q-header>

    <q-drawer bordered show-if-above v-model="leftDrawerOpen" side="left">
      <MainNavigation />
    </q-drawer>

    <q-page-container>
      <router-view />
    </q-page-container>

    <q-footer v-if="$q.screen.lt.md" :class="bgAndText">
      <q-toolbar>
        <q-btn
          flat
          icon="fa-solid fa-bars"
          @click="toggleLeftDrawer"
          :class="text"
        />
        <q-space />
        <GlobalSearch />
      </q-toolbar>
    </q-footer>
  </q-layout>

  <Teleport to="body"> </Teleport>
</template>

<style>
.nav-bg-dark {
  background-color: #242526;
  color: white;
}
.nav-bg-light {
  background-color: #ffffff;
  color: black;
}
</style>
