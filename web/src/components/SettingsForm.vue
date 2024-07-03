<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { Notify } from 'quasar';
import moment from 'moment';

const store = useRestaurantStore();

const onReductionChanged = () => {
  Notify.create({
    type: 'positive',
    group: false,
    message: 'Preisreduzierung gespeichert: ' + store.reduction + '€',
  });
  store.setReduction(store.reduction);
};

const onMiddayChanged = () => {
  Notify.create({
    type: 'positive',
    group: false,
    message: 'Mittagszeit gespeichert',
  });
  store.setMidday(store.midday);
};

function generateMiddayOptions() {
  const tmp = [];
  const start = moment().hour(10).minute(30);
  for (let i = 0; i < 9; i++) {
    const newTime = start.add(30, 'm');
    tmp.push({
      value: newTime.format('HHmm'),
      label: newTime.format('HH:mm') + ' Uhr',
    });
  }
  return tmp;
}
</script>

<template>
  <q-card style="width: 700px; max-width: 90vw" class="q-pa-md">
    <q-card-section class="row items-center">
      <div class="text-h6">Einstellungen</div>
      <q-space />
      <q-btn icon="fa-solid fa-xmark" dense flat round v-close-popup />
    </q-card-section>
    <q-card-section class="q-gutter-md">
      <q-input
        filled
        v-model="store.reduction"
        mask="#.##"
        fill-mask="0"
        suffix="€"
        clearable
        input-class="text-right"
        reverse-fill-mask
        label="Preisreduzierung"
        @update:model-value="onReductionChanged"
        debounce="700"
      />

      <q-select
        filled
        v-model="store.midday"
        :options="generateMiddayOptions()"
        label="Mittagszeit"
        emit-value
        map-options
        @update:model-value="onMiddayChanged"
      />
    </q-card-section>
  </q-card>
</template>
