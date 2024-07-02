<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { ref } from 'vue';
import { Notify } from 'quasar';
import moment from 'moment';

const store = useRestaurantStore();
const reduction = ref(store.reduction);
const midday = ref(store.midday);

const onReductionChanged = () => {
  Notify.create({
    type: 'positive',
    group: false,
    message: 'Preisreduzierung gespeichert',
  });
  store.setReduction(reduction.value);
};

const onMiddayChanged = () => {
  Notify.create({
    type: 'positive',
    group: false,
    message: 'Mittagszeit gespeichert',
  });
  store.setMidday(midday.value);
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
        v-model="reduction"
        mask="#.##"
        fill-mask="0"
        suffix="€"
        clearable
        input-class="text-right"
        reverse-fill-mask
        label="Preisreduzierung"
        @blur="onReductionChanged"
        @keyup.enter="onReductionChanged"
      />

      <q-select
        filled
        v-model="midday"
        :options="generateMiddayOptions()"
        label="Mittagszeit"
        emit-value
        map-options
        @update:model-value="onMiddayChanged"
      />
    </q-card-section>
  </q-card>
</template>
