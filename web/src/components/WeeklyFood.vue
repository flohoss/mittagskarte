<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';

defineProps<{ restaurant: handler_Restaurant }>();

const euroFormatter = new Intl.NumberFormat('de-DE', {
  style: 'currency',
  currency: 'EUR',
});
</script>

<template>
  <q-timeline color="secondary" style="margin: 2rem 1rem">
    <q-timeline-entry heading>
      <span class="text-h4">{{ restaurant.menu.description }}</span>
    </q-timeline-entry>
    <q-timeline-entry
      v-for="(entry, id) in restaurant.menu.food"
      :key="id"
      :subtitle="entry.day"
    >
      <template v-slot:title>
        <div class="row justify-between">
          <div>{{ entry.name }}</div>
          <div>{{ euroFormatter.format(entry.price) }}</div>
        </div>
      </template>
      {{ entry.description }}
    </q-timeline-entry>
  </q-timeline>
</template>
