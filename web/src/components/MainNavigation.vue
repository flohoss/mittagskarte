<script setup lang="ts">
import { GroupsService, handler_Restaurant } from 'src/openapi';
import { ref } from 'vue';
import { useRouter } from 'vue-router';

const router = useRouter();

const active = (link: string) => {
  return router.currentRoute.value.path.includes(link);
};

const groups = ref<Record<string, handler_Restaurant[]>>();
GroupsService.getGroups()
  .then((res) => {
    groups.value = res;
  })
  .catch((err) => {
    console.log(err);
  });
</script>

<template>
  <div>
    <q-list padding class="text-primary">
      <template v-for="(restaurants, key) in groups" :key="key">
        <q-item-label header>{{ key }}</q-item-label>
        <q-item
          v-for="(restaurant, index) in restaurants"
          :key="index"
          clickable
          :active="active(restaurant.id)"
          :to="'/' + restaurant.id"
          active-class="my-menu-link"
        >
          <q-item-section top avatar>
            <q-avatar><q-icon :name="restaurant.icon" /></q-avatar>
          </q-item-section>

          <q-item-section>
            <q-item-label>{{ restaurant.name }}</q-item-label>
            <q-item-label caption lines="1">
              {{ restaurant.description }}
            </q-item-label>
          </q-item-section>

          <q-item-section side>
            <q-item-label caption
              ><q-icon name="fa-solid fa-circle-check"
            /></q-item-label>
          </q-item-section>
        </q-item>
      </template>
    </q-list>
  </div>
</template>
