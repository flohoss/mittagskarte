<script setup lang="ts" generic="ContentType extends string | number">
import RestaurantInfo from 'src/components/RestaurantInfo.vue';
import { emptyRestaurant, useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';
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

const magnification: number = 3; // Magnification factor
const lens = ref<HTMLDivElement | null>(null);
const image = ref<ContentType | null>(null);

const magnify = (event: MouseEvent): void => {
    const imgEl = image.value?.$el.querySelector('img') as HTMLImageElement;
    const lensEl = lens.value;

    if (!imgEl || !lensEl) return;

    const rect = imgEl.getBoundingClientRect();

    // Get cursor position relative to the image
    const x = event.clientX - rect.left;
    const y = event.clientY - rect.top;

    // Prevent the lens from overflowing
    const lensWidth = lensEl.offsetWidth / 2;
    const lensHeight = lensEl.offsetHeight / 2;

    let lensX = x - lensWidth;
    let lensY = y - lensHeight;

    if (lensX < 0) lensX = 0; // Left boundary
    if (lensY < 0) lensY = 0; // Top boundary
    if (lensX > rect.width - lensEl.offsetWidth) lensX = rect.width - lensEl.offsetWidth; // Right boundary
    if (lensY > rect.height - lensEl.offsetHeight) lensY = rect.height - lensEl.offsetHeight; // Bottom boundary

    // Position the lens
    lensEl.style.left = `${lensX}px`;
    lensEl.style.top = `${lensY}px`;

    // Calculate background position for the magnified image
    const bgX = (x / rect.width) * 100; // Cursor position as a percentage of the image width
    const bgY = (y / rect.height) * 100; // Cursor position as a percentage of the image height

    lensEl.style.backgroundPosition = `${bgX}% ${bgY}%`;
};

const startMagnify = (event: MouseEvent): void => {
    const lensEl = lens.value;
    const imgEl = image.value?.$el.querySelector('img') as HTMLImageElement;

    if (!imgEl || !lensEl) return;

    lensEl.style.display = 'block';
    lensEl.style.backgroundImage = `url(${imgEl.src})`;
    lensEl.style.backgroundSize = `${imgEl.width * magnification}px ${imgEl.height * magnification}px`;

    magnify(event);
};

const stopMagnify = (): void => {
    if (lens.value) {
        lens.value.style.display = 'none';
    }
};
</script>

<template>
    <q-page class="flex column items-center">
        <div
            v-if="cardUrl"
            :class="[$q.screen.gt.sm ? 'q-py-md' : 'q-pa-sm', 'image-magnifier']"
            :style="{
                'border-radius': '0.5rem',
                width: '100%',
                'max-width': $q.screen.sizes.sm + 'px',
            }"
        >
            <q-img
                ref="image"
                :src="cardUrl"
                :style="{
                    'border-radius': '0.5rem',
                    width: '100%',
                }"
                alt="Magnifiable"
                class="magnify-image"
                @mousemove="magnify"
                @mouseenter="startMagnify"
                @mouseleave="stopMagnify"
            />
            <div class="magnifying-lens" ref="lens"></div>
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

.image-magnifier {
    position: relative;
    display: inline-block;
}

.magnify-image {
    width: 100%; /* Adjust as needed */
}

.magnifying-lens {
    position: absolute;
    border: 1px solid #000;
    width: 20rem;
    height: 12rem;
    border-radius: 2rem;
    display: none;
    pointer-events: none;
    background-repeat: no-repeat;
    background-color: rgba(255, 255, 255, 0.5);
    cursor: none;
}
</style>
