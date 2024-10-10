<script setup lang="ts">
import { services_DayOfWeek, services_CleanRestaurant } from 'src/openapi';
import FavStar from './FavStar.vue';
import { computed } from 'vue';

const props = defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
  search: { type: Boolean },
  showStar: { type: Boolean, default: false },
});

const isClosed = () => {
  const now = new Date();
  const currentDay = now.toLocaleString('en-us', { weekday: 'long' });
  return props.restaurant.rest_days.includes(currentDay as services_DayOfWeek);
};

const thumbnail = computed(
  () =>
    process.env.BASE_URL + 'data/thumbnails/' + props.restaurant.id + '.webp'
);
</script>

<template>
  <q-item
    :class="{ 'q-px-none q-py-sm': search }"
    clickable
    :disable="isClosed()"
    :to="'/restaurants/' + restaurant.id"
    active-class="text-primary"
  >
    <q-item-section avatar>
      <q-avatar rounded>
        <q-img :src="thumbnail" fit="cover" />
      </q-avatar>
    </q-item-section>

    <q-item-section>
      <q-item-label>{{ restaurant.name }}</q-item-label>
      <q-item-label
        v-if="isClosed()"
        class="q-item__label--caption text-caption"
      >
        Geschlossen
      </q-item-label>
      <q-item-label v-else class="q-item__label--caption text-caption">
        <div class="row items-baseline q-gutter-x-sm">
          <div>{{ restaurant.description }}</div>
        </div>
      </q-item-label>
    </q-item-section>

    <q-item-section v-if="showStar" side>
      <q-item-label>
        <FavStar :restaurant="restaurant" />
      </q-item-label>
    </q-item-section>
  </q-item>
</template>
