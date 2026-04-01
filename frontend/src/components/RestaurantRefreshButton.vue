<script setup lang="ts">
import { computed } from 'vue';
import type { RecordModel } from 'pocketbase';
import Fa7SolidArrowsRotate from '~icons/fa7-solid/arrows-rotate';
import Fa7SolidDownload from '~icons/fa7-solid/download';
import Fa7SolidUpload from '~icons/fa7-solid/upload';
import Fa7SolidHourglassHalf from '~icons/fa7-solid/hourglass-half';
import Fa7SolidClock from '~icons/fa7-solid/clock';
import { RestaurantMethod, RestaurantStatus } from '../stores/useRestaurants';
import { BackendURL } from '../main';

const props = defineProps<{
  restaurant: RecordModel;
}>();

const statusMeta = computed(() => {
  switch (props.restaurant.status) {
    case RestaurantStatus.UPDATING:
      return { icon: Fa7SolidArrowsRotate, label: 'Wird Aktualisiert', className: 'btn-neutral', iconClass: 'animate-spin' };
    case RestaurantStatus.QUEUED:
      return { icon: Fa7SolidHourglassHalf, label: 'In Warteschlange', className: 'btn-neutral', iconClass: '' };
    case RestaurantStatus.COOLDOWN:
      return { icon: Fa7SolidClock, label: 'Cooldown', className: 'btn-neutral', iconClass: '' };
    default:
      switch (props.restaurant.method) {
        case RestaurantMethod.SCRAPE:
          return { icon: Fa7SolidArrowsRotate, label: 'Leerlauf', className: 'hover:btn-primary', iconClass: '' };
        case RestaurantMethod.DOWNLOAD:
          return { icon: Fa7SolidDownload, label: 'Leerlauf', className: 'hover:btn-primary', iconClass: '' };
        default:
          return { icon: Fa7SolidUpload, label: 'Leerlauf', className: 'hover:btn-primary', iconClass: '' };
      }
  }
});

const canTriggerRefresh = computed(() => props.restaurant.status === RestaurantStatus.IDLE);

async function triggerRefresh() {
  try {
    await fetch(`${BackendURL}/scrape`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: props.restaurant.id,
      }),
    });
  } catch (error) {
    console.error('Failed to trigger scrape', error);
  }
}
</script>

<template>
  <button
    type="button"
    :class="['btn btn-soft', statusMeta.className]"
    :title="canTriggerRefresh ? 'Jetzt aktualisieren' : `Status: ${statusMeta.label}`"
    :aria-label="canTriggerRefresh ? 'Jetzt aktualisieren' : `Status: ${statusMeta.label}`"
    :disabled="!canTriggerRefresh"
    @click="triggerRefresh"
  >
    <component :is="statusMeta.icon" :class="['btn-icon', statusMeta.iconClass]" aria-hidden="true" />
  </button>
</template>
