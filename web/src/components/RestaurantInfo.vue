<script setup lang="ts">
import { computed } from 'vue';
import { type services_CleanRestaurant } from 'src/openapi';
import FavStar from './FavStar.vue';

const props = defineProps({
    restaurant: {
        type: Object as () => services_CleanRestaurant,
        required: true,
    },
    iconSize: {
        type: String,
        required: true,
    },
});

const googleSearch = computed(() => 'https://www.google.com/maps/search/?api=1&query=' + props.restaurant.address);
</script>

<template>
    <div class="flex q-gutter-x-sm" v-if="restaurant.id !== ''">
        <FavStar :icon-size="iconSize" :restaurant="restaurant" />
        <q-btn id="map" target="_blank" :size="iconSize" flat round color="primary" icon="fa-solid fa-map-marker-alt" :href="googleSearch">
            <q-tooltip class="bg-primary">Karte öffnen</q-tooltip>
        </q-btn>
        <q-btn id="phone" :size="iconSize" flat round color="primary" icon="fa-solid fa-phone" :href="'tel:' + restaurant.phone">
            <q-tooltip class="bg-primary">Anrufen</q-tooltip>
        </q-btn>
        <q-btn id="website" target="_blank" :size="iconSize" flat round color="primary" icon="fa-solid fa-globe" :href="restaurant.page_url">
            <q-tooltip class="bg-primary">Restaurant öffnen</q-tooltip>
        </q-btn>
    </div>
</template>
