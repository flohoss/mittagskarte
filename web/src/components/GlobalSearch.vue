<script setup lang="ts">
import { useRestaurantStore } from 'src/stores/restaurants';
import { computed, ref } from 'vue';
import NavRestaurant from './NavRestaurant.vue';

const store = useRestaurantStore();
const res = computed(() => store.result);
const dialog = ref(false);
const search = ref('');
</script>

<template>
  <q-input
    v-model="search"
    standout
    readonly
    dense
    square
    placeholder="Suchen"
    :autofocus="false"
    @click="dialog = true"
  >
    <template v-slot:prepend><q-icon name="fa-solid fa-search" /></template>
  </q-input>

  <q-dialog v-model="dialog" backdrop-filter="blur(4px) saturate(150%)">
    <q-card style="width: 700px; max-width: 90vw">
      <q-card-section>
        <q-input
          v-model="store.search"
          filled
          autofocus
          type="search"
          placeholder="Suchen"
          @keyup.esc="store.search = ''"
        >
          <template v-slot:prepend>
            <q-icon name="fa-solid fa-search" />
          </template>
        </q-input>
      </q-card-section>

      <q-card-section>
        <div v-if="res.length > 0" class="wor items-start jusify-center">
          <NavRestaurant
            v-for="restaurant in res"
            :key="restaurant.id"
            :restaurant="restaurant"
            :search="true"
          />
        </div>
      </q-card-section>

      <q-separator />
      <q-card-actions class="row items-center q-gutter-x-md">
        <div class="row justify-center items-center q-gutter-x-sm">
          <svg width="15" height="15" aria-label="Escape key" role="img">
            <g
              fill="none"
              stroke="currentColor"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.2"
            >
              <path
                d="M13.6167 8.936c-.1065.3583-.6883.962-1.4875.962-.7993 0-1.653-.9165-1.653-2.1258v-.5678c0-1.2548.7896-2.1016 1.653-2.1016.8634 0 1.3601.4778 1.4875 1.0724M9 6c-.1352-.4735-.7506-.9219-1.46-.8972-.7092.0246-1.344.57-1.344 1.2166s.4198.8812 1.3445.9805C8.465 7.3992 8.968 7.9337 9 8.5c.032.5663-.454 1.398-1.4595 1.398C6.6593 9.898 6 9 5.963 8.4851m-1.4748.5368c-.2635.5941-.8099.876-1.5443.876s-1.7073-.6248-1.7073-2.204v-.4603c0-1.0416.721-2.131 1.7073-2.131.9864 0 1.6425 1.031 1.5443 2.2492h-2.956"
              ></path>
            </g>
          </svg>
          <div class="text-caption">um Suche zu schließen</div>
        </div>
      </q-card-actions>
    </q-card>
  </q-dialog>
</template>
