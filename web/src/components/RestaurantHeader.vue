<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { computed } from 'vue';

const props = defineProps<{ restaurant: handler_Restaurant }>();
const emit = defineEmits(['openMenu']);

const thumbnail = computed(
  () =>
    process.env.BASE_URL + 'config/thumbnails/' + props.restaurant.id + '.webp'
);

const googleSearch = computed(
  () =>
    'https://www.google.com/maps/search/?api=1&query=' +
    props.restaurant.address
);
</script>

<template>
  <div class="row justify-between items-center q-px-md q-gutter-y-md">
    <div class="row items-center q-gutter-x-md">
      <q-img
        :src="thumbnail"
        fit="cover"
        style="height: 5rem; width: 5rem; border-radius: 0.5rem"
      />
      <div>
        <div class="text-h4 ellipsis">{{ restaurant.name }}</div>
        <div class="text-caption">
          {{ restaurant.description }}
          <span v-for="i in restaurant.price" :key="i">€</span>
        </div>
      </div>
    </div>
    <div class="row q-gutter-x-sm">
      <q-btn
        outline
        round
        color="secondary"
        icon="fa-solid fa-map-marker-alt"
        :href="googleSearch"
      />
      <q-btn
        outline
        round
        color="secondary"
        icon="fa-solid fa-phone"
        :href="'tel:' + restaurant.phone"
      />
      <q-btn
        v-if="restaurant.page_url"
        outline
        round
        color="secondary"
        icon="fa-solid fa-globe"
        :href="restaurant.page_url"
      />
      <q-btn
        v-if="restaurant.menu.card && restaurant.menu.food.length > 0"
        outline
        round
        color="accent"
        icon="fa-solid fa-file"
        @click="emit('openMenu')"
      />
    </div>
  </div>
</template>
