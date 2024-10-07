<script setup lang="ts">
import { services_DayOfWeek, services_CleanRestaurant } from 'src/openapi';
import FavStar from './FavStar.vue';

const props = defineProps<{
  restaurant: services_CleanRestaurant;
  search: boolean;
  showStar?: boolean;
}>();

const isClosed = () => {
  const now = new Date();
  const currentDay = now.toLocaleString('en-us', { weekday: 'long' });
  return props.restaurant.rest_days.includes(currentDay as services_DayOfWeek);
};
</script>

<template>
  <q-item
    :class="{ 'q-px-none q-py-sm': search }"
    dense
    clickable
    :disable="isClosed()"
    :to="'/restaurants/' + restaurant.id"
    active-class="text-secondary"
  >
    <q-item-section avatar>
      <q-avatar><q-icon :name="restaurant.icon" /></q-avatar>
    </q-item-section>

    <q-item-section>
      <q-item-label>{{ restaurant.name }}</q-item-label>
      <q-item-label v-if="isClosed()" class="text-caption"
        >Geschlossen</q-item-label
      >
    </q-item-section>

    <q-item-section v-if="showStar" side>
      <q-item-label>
        <FavStar :restaurant="restaurant" />
      </q-item-label>
    </q-item-section>
  </q-item>
</template>
