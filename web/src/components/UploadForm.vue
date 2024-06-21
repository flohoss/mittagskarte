<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';
import { RestaurantsService } from 'src/openapi';

const store = useRestaurantStore();
const options = computed(() =>
  Object.keys(store.restaurants).filter((key) => key !== '')
);

const file = ref();
const id = ref();
const token = ref();
const onSubmit = () => {
  RestaurantsService.postRestaurants(id.value, file.value, token.value);
};
</script>

<template>
  <q-card style="min-width: 50vw; width: 100%" class="q-pa-md">
    <q-card-section class="row items-center q-pb-none">
      <div class="text-h6">Preisreduzierung</div>
      <q-space />
      <q-btn icon="fa-solid fa-xmark" flat round dense v-close-popup />
    </q-card-section>
    <q-card-section class='column q-gutter-md'
      <q-select filled v-model="id" :options="options" label="Filled" />
      <q-file outlined v-model="file">
        <template v-slot:prepend>
          <q-icon name="fa-solid " />
        </template>
      </q-file>
    </q-card-section>
  </q-card>
</template>
