<script setup lang="ts">
import { computed } from 'vue';
import type { RestaurantRecord } from '../../models/restaurant';
import ActionIconButton from './ActionIconButton.vue';
import { useRestaurants } from '../../stores/useRestaurants';

const props = withDefaults(
  defineProps<{
    restaurant: RestaurantRecord;
    square?: boolean;
  }>(),
  {
    square: false,
  }
);

const { getMapUrl, getPhoneUrl } = useRestaurants();

const mapUrl = computed(() => (props.restaurant.address ? getMapUrl(props.restaurant) : ''));
const phoneUrl = computed(() => (props.restaurant.phone ? getPhoneUrl(props.restaurant) : ''));
const mapActionTitle = computed(() => (mapUrl.value ? `${props.restaurant.address}` : 'Keine Karte verfügbar'));
const phoneActionTitle = computed(() => (phoneUrl.value ? 'Anrufen' : 'Keine Telefonnummer verfügbar'));
const websiteUrl = computed(() => props.restaurant.website || '');
const websiteActionTitle = computed(() => (websiteUrl.value ? 'Website öffnen' : 'Keine Website verfügbar'));

const classNames = computed(() => (props.square ? 'btn btn-soft btn-square shrink-0' : 'btn btn-soft'));
</script>

<template>
  <ActionIconButton :href="mapUrl" :class-name="classNames + ' hover:btn-warning'" :title="mapActionTitle" ariaLabel="Karte öffnen">
    <span class="icon-[fa7-solid--map] btn-icon" aria-hidden="true"></span>
  </ActionIconButton>

  <ActionIconButton :href="phoneUrl" :class-name="classNames + ' hover:btn-success'" :title="phoneActionTitle" ariaLabel="Restaurant anrufen" target="_self">
    <span class="icon-[fa7-solid--phone] btn-icon" aria-hidden="true"></span>
  </ActionIconButton>

  <ActionIconButton :href="websiteUrl" :class-name="classNames + ' hover:btn-info'" :title="websiteActionTitle" ariaLabel="Website öffnen">
    <span class="icon-[fa7-solid--globe] btn-icon" aria-hidden="true"></span>
  </ActionIconButton>
</template>
