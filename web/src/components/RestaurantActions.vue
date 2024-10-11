<script setup lang="ts">
import { ref } from 'vue';
import { RestaurantsService, type services_CleanRestaurant } from 'src/openapi';
import UploadForm from './UploadForm.vue';
import { Loading, Notify, Dialog, Dark } from 'quasar';
import { useRouter } from 'vue-router';
import { useRestaurantStore } from 'src/stores/restaurants';

const store = useRestaurantStore();
const router = useRouter();
const props = defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
  iconSize: {
    type: String,
    required: true,
  },
});

function confirmRefresh() {
  Dialog.create({
    title: props.restaurant.name + ' aktualisieren',
    message: 'Diese aktualisierung kann mehrere Minuten dauern...',
    cancel: true,
    persistent: true,
  }).onOk(() => {
    refresh();
  });
}

const refresh = () => {
  Loading.show({
    message:
      'Eine Aktualisierung findet statt. Dies kann mehrere Minuten dauern...',
    boxClass: Dark.isActive ? 'nav-bg-dark' : 'nav-bg-light',
    spinnerColor: 'primary',
  });
  RestaurantsService.putRestaurants(props.restaurant.id)
    .then((resp) => {
      store.restaurants[props.restaurant.id] = resp;
      Notify.create({
        type: 'positive',
        message: 'Erfolgreich',
        caption: 'Das Menü wurde aktualisiert',
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

const upload = ref(false);
</script>

<template>
  <div class="flex q-gutter-x-sm" v-if="restaurant.id !== ''">
    <q-btn
      id="upload"
      :size="iconSize"
      flat
      round
      icon="fa-solid fa-upload"
      @click="upload = true"
    >
      <q-tooltip>Menü hochladen</q-tooltip>
    </q-btn>
    <q-btn
      id="refresh"
      :size="iconSize"
      flat
      round
      icon="fa-solid fa-rotate-right"
      @click="confirmRefresh()"
    >
      <q-tooltip>Menü aktualisieren</q-tooltip>
    </q-btn>

    <q-dialog v-model="upload" backdrop-filter="blur(4px) saturate(150%)">
      <UploadForm :restaurant="restaurant" @uploaded="upload = false" />
    </q-dialog>
  </div>
</template>
