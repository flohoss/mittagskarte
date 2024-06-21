<script setup lang="ts">
import { handler_Restaurant } from 'src/openapi';
import { useRestaurantStore } from 'src/stores/restaurants';
import { ComputedRef, computed } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();
const active = (link: string) => {
  return router.currentRoute.value.path.includes(link);
};

const store = useRestaurantStore();
const groups: ComputedRef<Record<string, handler_Restaurant[]>> = computed(
  () => store.grouped
);

const isIpen = (closedDays: string[]) => {
  const now = new Date();
  const currentDay = now.toLocaleString('en-us', { weekday: 'long' });
  return closedDays.includes(currentDay);
};
</script>

<template>
  <div>
    <q-list padding>
      <q-item clickable dense to="/" exact active-class="text-secondary">
        <q-item-section top avatar>
          <q-avatar>
            <img src="/favicon/android-chrome-192x192.png" />
          </q-avatar>
        </q-item-section>

        <q-item-section>
          <q-item-label style="font-weight: bold">Mittagstisch</q-item-label>
        </q-item-section>
      </q-item>
      <template v-for="(restaurants, key) in groups" :key="key">
        <q-item-label header>{{ key }}</q-item-label>
        <q-item
          v-for="(restaurant, index) in restaurants"
          :key="index"
          clickable
          dense
          :disable="isIpen(restaurant.rest_days)"
          :active="active(restaurant.id)"
          :to="'/restaurants/' + restaurant.id"
          active-class="text-secondary"
        >
          <q-item-section top avatar>
            <q-avatar><q-icon :name="restaurant.icon" /></q-avatar>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{ restaurant.name }}</q-item-label>
          </q-item-section>

          <q-item-section side>
            <q-item-label caption>
              <q-icon
                v-if="isIpen(restaurant.rest_days)"
                name="fa-solid fa-shop-lock"
              />
            </q-item-label>
          </q-item-section>
        </q-item>
      </template>
    </q-list>
  </div>
</template>
