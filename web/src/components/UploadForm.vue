<script setup lang="ts">
import { Dark, Loading, Notify } from 'quasar';
import { RestaurantsService, services_CleanRestaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const props = defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
});

const emit = defineEmits(['uploaded']);
const store = useRestaurantStore();
const router = useRouter();

const file = ref();
const token = ref('');
const onSubmit = () => {
  Loading.show({
    message: 'Die Datei wurd hochgeladen. Dies kann mehrere Minuten dauern...',
    boxClass: Dark.isActive ? 'nav-bg-dark' : 'nav-bg-light',
    spinnerColor: 'primary',
  });
  RestaurantsService.postRestaurants(
    'Bearer ' + token.value,
    props.restaurant.id,
    file.value
  )
    .then((resp) => {
      store.restaurants[props.restaurant.id] = resp;
      Notify.create({
        type: 'positive',
        message: 'Erfolgreich',
        caption: 'Das Menü wurde hochgeladen',
        actions: [
          {
            icon: 'fa-solid fa-xmark',
            color: 'white',
            round: true,
            title: 'Schließen',
          },
        ],
        timeout: 0,
      });
      router.push({
        name: 'restaurants',
        params: { name: props.restaurant.id },
        query: { cache: Date.now() },
      });
      emit('uploaded');
    })
    .catch((err) => {
      Notify.create({
        type: 'negative',
        message: 'Fehler beim Laden der Daten',
        caption:
          err?.body?.message ??
          err?.message ??
          'Unbekannter Fehler, bitte prüfen Sie die Konsole für weitere Informationen',
        actions: [
          {
            icon: 'fa-solid fa-xmark',
            color: 'white',
            round: true,
            title: 'Schließen',
          },
        ],
        timeout: 0,
      });
      console.log(err);
    })
    .finally(() => Loading.hide());
};
</script>

<template>
  <q-card style="width: 700px; max-width: 90vw" class="q-pa-md">
    <q-card-section class="row items-center">
      <div class="text-h6">Neues Menü für {{ restaurant.name }} hochladen</div>
      <q-space />
      <q-btn id="close" icon="fa-solid fa-xmark" dense flat round v-close-popup />
    </q-card-section>
    <q-card-section class="column q-gutter-md">
      <q-input type="password" filled v-model="token" label="API-Token" />
      <q-file v-model="file" filled>
        <template v-slot:prepend>
          <q-icon name="fa-solid fa-file" />
        </template>
      </q-file>
    </q-card-section>
    <q-card-actions align="right">
      <q-btn id="upload" label="Hochladen" color="primary" @click="onSubmit" />
    </q-card-actions>
  </q-card>
</template>
