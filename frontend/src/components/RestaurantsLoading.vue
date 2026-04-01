<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  isLoading: boolean;
}>();

const showSkeleton = computed(() => props.isLoading);
const skeletonItems = Array.from({ length: 5 }, (_, index) => index);
</script>

<template>
  <section class="grid min-h-[20vh]" aria-live="polite" aria-busy="true">
    <div v-if="showSkeleton" class="grid gap-3">
      <div class="skeleton h-5 w-44 rounded-lg" aria-hidden="true" />

      <div class="card-grid" aria-label="Restaurants werden geladen">
        <article v-for="item in skeletonItems" :key="item" class="card card-border overflow-hidden rounded-xl bg-base-100 opacity-80 shadow-md">
          <div class="skeleton h-30 w-full" aria-hidden="true" />

          <div class="card-body gap-3 p-3">
            <div class="skeleton h-4 w-3/4 rounded-lg" aria-hidden="true" />

            <div class="grid grid-cols-5 gap-1.5" aria-hidden="true">
              <div class="skeleton h-10 rounded-lg" />
              <div class="skeleton h-10 rounded-lg" />
              <div class="skeleton h-10 rounded-lg" />
              <div class="skeleton h-10 rounded-lg" />
              <div class="skeleton h-10 rounded-lg" />
            </div>
          </div>
        </article>
      </div>
    </div>
  </section>
</template>
