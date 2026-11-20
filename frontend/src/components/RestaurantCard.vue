<script setup lang="ts">
import { computed } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { useFavorites } from '../stores/useFavorites';
import type { RestaurantRecord } from '../models/restaurant';
import { useRestaurants } from '../stores/useRestaurants';
import { getRegionBadgeColors } from '../utils/regionColor';
import { usePrefersDark } from '../composables/usePrefersDark';
import { useRestaurantMeta } from '../composables/useRestaurantMeta';

const props = defineProps<{
  restaurant: RestaurantRecord;
  showRegion?: boolean;
}>();

const { applySearch } = useRestaurants();
const { toggleFavorite } = useFavorites();
const prefersDark = usePrefersDark();
const regionBadgeColors = computed(() => getRegionBadgeColors(props.restaurant.group || '', prefersDark.value));
const meta = useRestaurantMeta(() => props.restaurant);
const thumbnailUrl = meta.thumbnailUrl;
const isDimmed = meta.isDimmed;
const initials = meta.initials;
const isOnHoliday = meta.isOnHoliday;
const holidayLabel = meta.holidayLabel;
const isClosed = meta.isClosed;
const menuFreshness = meta.menuFreshness;
const showDistance = meta.showDistance;
const distanceLabel = meta.distanceLabel;
const isFavorited = meta.isFavorited;
const lastCheckText = meta.lastCheckText;
const lastCheckTitle = meta.lastCheckTitle;
</script>

<template>
  <article
    class="group card card-border overflow-hidden rounded-xl bg-base-100 opacity-80 shadow-md transition-[shadow,opacity] duration-200 hover:opacity-100 hover:shadow-xl"
  >
    <figure class="relative h-30 overflow-hidden bg-base-300">
      <img
        v-if="thumbnailUrl"
        :src="thumbnailUrl"
        :alt="props.restaurant.name"
        :class="['h-full w-full object-cover transition-transform duration-500 group-hover:scale-105', isDimmed ? 'opacity-40 grayscale' : '']"
        loading="lazy"
      />
      <div
        v-else
        :class="['flex h-full w-full items-center justify-center text-2xl font-semibold', isDimmed ? 'opacity-40 grayscale' : 'text-base-content/70']"
      >
        {{ initials }}
      </div>

      <div class="absolute inset-x-0 top-0 flex items-start justify-between px-3 pt-3">
        <div class="flex flex-col items-start gap-1.5">
          <span v-if="isOnHoliday" class="badge badge-sm badge-info backdrop-blur">{{ holidayLabel }}</span>
          <span v-else-if="isClosed" class="badge badge-sm badge-info backdrop-blur">Heute geschlossen</span>
          <span v-else-if="menuFreshness" :class="['badge badge-sm', menuFreshness.className]">{{ menuFreshness.label }}</span>
          <span v-if="showDistance" class="badge badge-sm badge-neutral/85 backdrop-blur w-fit">{{ distanceLabel }}</span>
        </div>
        <button
          type="button"
          :class="[
            'cursor-pointer border-0 bg-transparent p-0 text-lg leading-none drop-shadow-sm transition-colors opacity-80 hover:opacity-100 focus:outline-none',
            isFavorited ? 'text-warning' : 'text-white hover:text-warning',
          ]"
          :aria-label="isFavorited ? 'Favorit entfernen' : 'Als Favorit markieren'"
          :aria-pressed="isFavorited"
          @click="toggleFavorite(props.restaurant.id)"
        >
          <span class="icon-[fa7-solid--star]" aria-hidden="true"></span>
        </button>
      </div>
      <!-- bottom row: tags -->
      <div class="absolute inset-x-0 bottom-0 bg-linear-to-t from-base-100/90 to-transparent px-3 pb-3 pt-8">
        <div class="no-scrollbar fade-right -mx-3 flex gap-1.5 overflow-x-auto px-3 pb-1">
          <button
            v-if="showRegion && props.restaurant.group"
            type="button"
            class="badge badge-sm shrink-0 cursor-pointer border-0 px-2 py-2.5 text-xs font-medium backdrop-blur transition hover:brightness-95"
            :style="{ backgroundColor: regionBadgeColors.backgroundColor, color: regionBadgeColors.color }"
            :title="`Nach Region ${props.restaurant.group} filtern`"
            @click="applySearch(props.restaurant.group)"
          >
            {{ props.restaurant.group }}
          </button>
          <button
            v-for="tag in props.restaurant.tags"
            :key="tag"
            type="button"
            class="badge badge-outline badge-xs shrink-0 cursor-pointer border-base-300/60 bg-base-100/70 px-2 py-2.5 text-xs font-medium backdrop-blur transition-colors hover:bg-base-100"
            :title="`Nach Tag ${tag} filtern`"
            @click="applySearch(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </figure>
    <div class="card-body gap-3 p-3">
      <div>
        <h3 class="text-base font-semibold leading-tight">
          {{ props.restaurant.name }}
        </h3>

        <p v-if="lastCheckText" class="text-xs text-base-content/65" :title="lastCheckTitle" aria-label="Letzter Pruefstatus">
          {{ lastCheckText }}
        </p>
        <p class="text-xs text-base-content/65" v-else>Noch nicht geprüft</p>
      </div>

      <RestaurantActions :restaurant="props.restaurant" />
    </div>
  </article>
</template>
