<script setup lang="ts">
import { watch, onMounted } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import RestaurantGroup from '../components/RestaurantGroup.vue';
import { useRestaurants } from '../stores/useRestaurants';

const props = defineProps({ q: { type: String, default: '' } });
const { groupedRestaurants, searchQuery, applySearch } = useRestaurants();
const route = useRoute();
const router = useRouter();

onMounted(() => {
  if (props.q && props.q !== searchQuery.value) {
    applySearch(props.q);
  }
});

watch(searchQuery, (val) => {
  if (val !== route.query.q) {
    router.replace({ query: { ...route.query, q: val || undefined } });
  }
});

watch(
  () => route.query.q,
  (newQ) => {
    let qStr = '';
    if (Array.isArray(newQ)) {
      qStr = newQ.join(' ');
    } else if (typeof newQ === 'string') {
      qStr = newQ;
    }
    if (qStr !== searchQuery.value) {
      applySearch(qStr);
    }
  }
);
</script>

<template>
  <div class="grid gap-8">
    <RestaurantGroup v-for="(restaurants, group) in groupedRestaurants" :key="group" :restaurants="restaurants" :group="group" />
  </div>
</template>
