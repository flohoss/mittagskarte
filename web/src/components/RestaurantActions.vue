<script setup lang="ts">
import { computed, ref } from 'vue';
import FavStar from './FavStar.vue';
import type { services_CleanRestaurant } from 'src/openapi';
import UploadForm from './UploadForm.vue';

const props = defineProps({
  restaurant: {
    type: Object as () => services_CleanRestaurant,
    required: true,
  },
  iconSize: {
    type: String,
    default: 'sm',
    required: false,
  },
});

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    props.restaurant.address
);

const upload = ref(false);
</script>

<template>
  <div class="row q-gutter-x-sm" v-if="restaurant.id !== ''">
    <FavStar :restaurant="restaurant" />
    <q-btn
      target="_blank"
      :size="iconSize"
      flat
      round
      color="secondary"
      icon="fa-solid fa-map-marker-alt"
      :href="googleSearch"
    >
      <q-tooltip class="bg-accent">Karte öffnen</q-tooltip>
    </q-btn>
    <q-btn
      :size="iconSize"
      flat
      round
      color="secondary"
      icon="fa-solid fa-phone"
      :href="'tel:' + restaurant.phone"
    >
      <q-tooltip class="bg-accent">Anrufen</q-tooltip>
    </q-btn>
    <q-btn
      target="_blank"
      v-if="restaurant.page_url"
      :size="iconSize"
      flat
      round
      color="secondary"
      icon="fa-solid fa-globe"
      :href="restaurant.page_url"
    >
      <q-tooltip class="bg-accent">Restaurant öffnen</q-tooltip>
    </q-btn>
    <q-btn
      :size="iconSize"
      flat
      round
      color="secondary"
      icon="fa-solid fa-upload"
      @click="upload = true"
    >
      <q-tooltip class="bg-accent">Menü hochladen</q-tooltip>
    </q-btn>

    <q-dialog v-model="upload" backdrop-filter="blur(4px) saturate(150%)">
      <UploadForm :restaurant="restaurant" @uploaded="upload = false" />
    </q-dialog>
  </div>
</template>
