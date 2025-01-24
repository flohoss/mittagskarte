<script setup lang="ts">
import type { services_DayOfWeek, services_CleanRestaurant } from 'src/openapi';
import FavStar from './FavStar.vue';
import { computed } from 'vue';

const props = defineProps({
    restaurant: {
        type: Object as () => services_CleanRestaurant,
        required: true,
    },
    search: { type: Boolean, default: false },
});

const isClosed = () => {
    const now = new Date();
    const currentDay = now.toLocaleString('en-us', { weekday: 'long' });
    return props.restaurant.rest_days.includes(currentDay as services_DayOfWeek);
};

const thumbnail = computed(() => process.env.BASE_URL + 'data/thumbnails/' + props.restaurant.id + '.webp');
</script>

<template>
    <div style="position: relative">
        <q-item :class="{ 'q-px-none q-py-sm': search }" clickable :disable="isClosed()" :to="'/restaurants/' + restaurant.id" active-class="text-primary">
            <q-item-section avatar>
                <q-avatar rounded>
                    <q-img :src="thumbnail" fit="cover" />
                </q-avatar>
            </q-item-section>

            <q-item-section>
                <q-item-label>{{ restaurant.name }}</q-item-label>
                <q-item-label v-if="isClosed()" class="q-item__label--caption text-caption"> Geschlossen </q-item-label>
                <q-item-label v-else class="q-item__label--caption text-caption">
                    <div class="row items-baseline q-gutter-x-sm">
                        <div>{{ restaurant.description }}</div>
                    </div>
                </q-item-label>
            </q-item-section>
        </q-item>
        <div class="favStar" v-if="!search">
            <FavStar :restaurant="restaurant" />
        </div>
    </div>
</template>

<style scoped>
.favStar {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    right: 0.6rem;
}
</style>
