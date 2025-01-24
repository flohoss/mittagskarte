<script setup lang="ts">
import MainNavigation from 'components/MainNavigation.vue';
import NavTitle from 'src/components/NavTitle.vue';
import { computed, ref } from 'vue';
import { Dark, useQuasar } from 'quasar';
import NavExtra from 'src/components/NavExtra.vue';
import { useRestaurantStore, emptyRestaurant } from 'src/stores/restaurants';
import { useRoute } from 'vue-router';
import RestaurantInfo from 'src/components/RestaurantInfo.vue';

const $q = useQuasar();
const route = useRoute();
const store = useRestaurantStore();
const restaurant = computed(() => store.restaurants[route.params.name as string] ?? emptyRestaurant);

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
const iconSize = computed(() => {
    if ($q.screen.lt.md) {
        return 'md';
    }
    return 'sm';
});
</script>

<template>
    <q-layout view="hHr Lpr fFr">
        <q-header :class="bgAndText" :bordered="!Dark.isActive">
            <q-toolbar>
                <NavTitle />
                <q-space />
                <NavExtra :restaurant="restaurant" />
            </q-toolbar>
        </q-header>

        <q-drawer bordered show-if-above v-model="leftDrawerOpen" side="left">
            <MainNavigation />
        </q-drawer>

        <q-page-container>
            <router-view />
        </q-page-container>

        <q-footer v-if="$q.screen.lt.md" :class="bgAndText" :bordered="!Dark.isActive">
            <q-toolbar>
                <q-btn id="menu" flat round icon="fa-solid fa-bars" @click="toggleLeftDrawer" :class="text" />
                <q-space />
                <RestaurantInfo v-if="restaurant.image_url !== ''" :restaurant="restaurant" :icon-size="iconSize" />
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
