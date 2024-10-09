<script setup lang="ts">
import { Loading, Notify } from 'quasar';
import { RestaurantsService } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

const emit = defineEmits(['uploaded']);
const store = useRestaurantStore();
const router = useRouter();
const options = computed(() =>
  Object.keys(store.restaurants)
    .filter((key) => store.restaurants[key].image_url === '')
    .map((key) => ({ label: store.restaurants[key].name, value: key }))
);

const file = ref();
const id = ref();
const token = ref('');
const onSubmit = () => {
  Loading.show();
  RestaurantsService.postRestaurants(
    'Bearer ' + token.value,
    id.value,
    file.value
  )
    .then((resp) => {
      store.restaurants[id.value] = resp;
      Notify.create({
        type: 'positive',
        group: false,
        message: 'Menü hochgeladen',
      });
      router.push({
        name: 'restaurants',
        params: { name: id.value },
        query: { cache: Date.now() },
      });
      emit('uploaded');
    })
    .catch((err) => {
      Notify.create({
        type: 'negative',
        group: false,
        message: 'Fehler: ' + err.status + ' - ' + err.message,
      });
    })
    .finally(() => Loading.hide());
};
</script>

<template>
  <q-card style="width: 700px; max-width: 90vw" class="q-pa-md">
    <q-card-section class="row items-center">
      <div class="text-h6">Neues Menü hochladen</div>
      <q-space />
      <q-btn icon="fa-solid fa-xmark" dense flat round v-close-popup />
    </q-card-section>
    <q-card-section class="column q-gutter-md">
      <q-select
        filled
        v-model="id"
        :options="options"
        label="Restaurant"
        emit-value
        map-options
      />
      <q-input type="password" filled v-model="token" label="API-Token" />
      <q-file v-model="file" filled>
        <template v-slot:prepend>
          <q-icon name="fa-solid fa-file" />
        </template>
      </q-file>
    </q-card-section>
    <q-card-actions align="right">
      <q-btn label="Hochladen" color="primary" @click="onSubmit" />
    </q-card-actions>
  </q-card>
</template>
