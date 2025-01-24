<script setup lang="ts">
import RestaurantInfo from 'src/components/RestaurantInfo.vue';
import { emptyRestaurant, useRestaurantStore } from 'src/stores/restaurants';
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import parser from 'cron-parser';

const route = useRoute();
const store = useRestaurantStore();
const restaurant = computed(() => store.restaurants[route.params.name as string] ?? emptyRestaurant);

const cardUrl = computed(() => {
    if (restaurant.value.image_url === '') {
        return '';
    }
    let url = process.env.BASE_URL + restaurant.value.image_url;
    if (route.query.cache !== undefined) {
        url += '?rnd=' + route.query.cache?.toString();
    }
    return url;
});

const thumbnail = computed(() => process.env.BASE_URL + 'data/thumbnails/' + restaurant.value.id + '.webp');

const nextUpdate = computed(() => {
    if (restaurant.value.update_cron === '') {
        return '';
    }
    const date = parser.parseExpression(restaurant.value.update_cron).next();
    return date.getDate() + '.' + (date.getMonth() + 1) + '.' + date.getFullYear() + ' ' + date.getHours() + ':' + date.getMinutes() + ' Uhr';
});
</script>

<template>
    <q-page class="flex column items-center">
        <div
            v-if="cardUrl"
            :class="[$q.screen.gt.sm ? 'q-py-md' : 'q-pa-sm']"
            :style="{
                'border-radius': '0.5rem',
                width: '100%',
                'max-width': $q.screen.sizes.sm + 'px',
            }"
        >
            <q-img
                :src="cardUrl"
                :style="{
                    'border-radius': '0.5rem',
                    width: '100%',
                }"
            />
        </div>

        <q-card
            v-else-if="restaurant.id !== ''"
            :class="[$q.screen.gt.sm ? 'q-my-md' : 'q-ma-sm']"
            class="my-card"
            flat
            bordered
            :style="{ 'max-width': +'px' }"
        >
            <q-img :src="thumbnail" style="max-height: 15rem" />

            <q-card-section>
                <div class="text-h6">Kein Menü gefunden...</div>
            </q-card-section>

            <q-card-actions align="right">
                <RestaurantInfo icon-size="md" :restaurant="restaurant" />
            </q-card-actions>
        </q-card>

        <div v-if="restaurant.update_cron !== ''" class="q-pb-md">Nächste Aktualisierung: {{ nextUpdate }}</div>
    </q-page>
</template>

<style scoped>
.my-card {
    width: 100%;
    max-width: 20rem;
}
</style>
