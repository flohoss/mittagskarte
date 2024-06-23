<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';

const search = ref('');

const store = useRestaurantStore();
const res = computed(() => store.result);
</script>

<template>
  <q-page class="row justify-center items-start q-pt-md">
    <q-input
      filled
      v-model="store.search"
      label="Suchen"
      @keyup.esc="store.search = ''"
    >
      <template v-slot:append>
        <q-icon v-if="search === ''" name="fa-solid fa-search" />
      </template>
    </q-input>
    <div v-if="res.length > 0">
      {{ res }}
    </div>
  </q-page>
</template>
