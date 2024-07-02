<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { ref } from 'vue';
import { Notify } from 'quasar';

const store = useRestaurantStore();
const reduction = ref(store.reduction || 0);

const onSubmit = () => {
  Notify.create({
    type: 'positive',
    group: false,
    message: 'Preisreduzierung gespeichert',
  });
  store.setReduction(reduction.value);
};
</script>

<template>
  <q-card style="min-width: 50vw; width: 100%" class="q-pa-md">
    <q-card-section class="row items-center">
      <div class="text-h6">Preisreduzierung</div>
      <q-space />
      <q-btn icon="fa-solid fa-xmark" dense flat round v-close-popup />
    </q-card-section>
    <q-card-section>
      <q-input
        filled
        v-model="reduction"
        mask="#.##"
        fill-mask="0"
        suffix="€"
        clearable
        input-class="text-right"
        reverse-fill-mask
        @blur="onSubmit"
      />
    </q-card-section>
  </q-card>
</template>
