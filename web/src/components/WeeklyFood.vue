<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { LocalStorage } from 'quasar';
import { ref } from 'vue';
import { ReductionKey } from 'src/stores/restaurants';

defineProps<{ restaurant: handler_Restaurant }>();

const euroFormatter = new Intl.NumberFormat('de-DE', {
  style: 'currency',
  currency: 'EUR',
});

const reduction = ref<number>(LocalStorage.getItem(ReductionKey) || 0);

const calcPrice = (price: number) => {
  const result = price + reduction.value;
  return euroFormatter.format(result);
};
</script>

<template>
  <q-timeline color="secondary" class="q-pa-md">
    <q-timeline-entry heading>
      <span class="text-h4">{{ restaurant.menu.description }}</span>
    </q-timeline-entry>
    <q-timeline-entry
      v-for="(entry, id) in restaurant.menu.food"
      :key="id"
      :subtitle="entry.day"
    >
      <template v-slot:title>
        <div
          style="
            display: flex;
            justify-content: space-between;
            min-width: 0;
            gap: 1rem;
          "
        >
          <div style="white-space: nowrap" class="ellipsis">
            {{ entry.name }}
          </div>
          <div style="flex-shrink: 0">
            <q-chip color="primary" text-color="white">
              {{ calcPrice(entry.price) }}
            </q-chip>
          </div>
        </div>
      </template>
      {{ entry.description }}
    </q-timeline-entry>
  </q-timeline>
</template>
