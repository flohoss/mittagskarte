<script setup lang="ts">
import { config_DayOfWeek, handler_Restaurant } from 'src/openapi';

const props = defineProps<{
  restaurant: handler_Restaurant;
  search: boolean;
}>();

const isClosed = () => {
  const now = new Date();
  const currentDay = now.toLocaleString('en-us', { weekday: 'long' });
  return props.restaurant.rest_days.includes(currentDay as config_DayOfWeek);
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
      <q-item-label class="text-caption">Geschlossen</q-item-label>
    </q-item-section>
  </q-item>
</template>
