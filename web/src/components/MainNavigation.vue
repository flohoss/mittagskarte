<script setup lang="ts">
import type { services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import NavRestaurant from './NavRestaurant.vue';
import RestaurantSearch from './RestaurantSearch.vue';
import { useRoute } from 'vue-router';
import { computed } from 'vue';

const store = useRestaurantStore();
const route = useRoute();
const id = computed(() => route.params.name as string);

const groupedRestaurants = computed(() => store.groupedRestaurants);
const favoriteRestaurants = computed(() => store.favoriteRestaurants);

const amountOfRestaurants = (restaurants: services_CleanRestaurant[]) => {
    const amount = restaurants.length;
    return amount === 1 ? amount + ' Restaurant' : amount + ' Restaurants';
};

function isIdIncluded(records: services_CleanRestaurant[], targetId?: string): boolean {
    if (!targetId || store.favorites.includes(targetId)) {
        return false;
    }

    records.forEach((record) => {
        if (record.id === targetId) {
            return true;
        }
    });
    return false;
}
</script>

<template>
    <q-list>
        <q-item class="q-mb-sm">
            <RestaurantSearch />
        </q-item>
        <q-expansion-item class="q-pb-md" label="Favoriten" default-opened hide-expand-icon expand-icon-toggle :caption="amountOfRestaurants(store.favorites)">
            <transition-group v-if="store.favorites.length > 0" appear enter-active-class="animated fadeIn" leave-active-class="animated fadeOutLeft">
                <NavRestaurant v-for="restaurant in favoriteRestaurants" :key="'fav-' + restaurant.id" :restaurant="restaurant" />
            </transition-group>
        </q-expansion-item>
        <q-expansion-item
            v-for="(restaurants, key) in groupedRestaurants"
            :key="key"
            :label="key"
            :caption="amountOfRestaurants(restaurants)"
            :default-opened="isIdIncluded(restaurants, id)"
        >
            <NavRestaurant v-for="restaurant in restaurants" :key="restaurant.id" :restaurant="restaurant" />
        </q-expansion-item>
    </q-list>
</template>
