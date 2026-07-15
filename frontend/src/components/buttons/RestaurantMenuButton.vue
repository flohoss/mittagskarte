<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core';
import { useFloating, autoUpdate, offset, shift, autoPlacement, size } from '@floating-ui/vue';
import { useRoute, useRouter } from 'vue-router';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import type { RestaurantRecord } from '../../models/restaurant';

const props = defineProps<{
  restaurant: RestaurantRecord;
  menuUrl: string | null;
  menuWidth?: number | null;
  menuHeight?: number | null;
}>();

const isOpen = ref(false);
const reference = ref<HTMLElement | null>(null);
const floating = ref<HTMLElement | null>(null);
const hideTimeout = ref<number | null>(null);
const breakpoints = useBreakpoints(breakpointsTailwind);
const isLargeScreen = breakpoints.greaterOrEqual('lg');
const imageWidth = ref<number | null>(props.menuWidth ?? null);
const imageHeight = ref<number | null>(props.menuHeight ?? null);
const route = useRoute();
const router = useRouter();
const imageSize = computed(() => {
  const width = imageWidth.value && imageWidth.value > 0 ? imageWidth.value : undefined;
  const height = imageHeight.value && imageHeight.value > 0 ? imageHeight.value : undefined;

  return {
    width,
    height,
    isLandscape: !!width && !!height ? width >= height : false,
  };
});

// Portrait menus (left/right): appear beside the button — button row stays accessible.
// Landscape menus (top/bottom): appear above/below — button is never covered by a wide image.
const middleware = computed(() => [
  offset(8),
  autoPlacement({
    padding: 8,
    allowedPlacements: imageSize.value.isLandscape ? ['top', 'bottom'] : ['left', 'right'],
  }),
  shift({ padding: 8 }),
  size({
    padding: 8,
    apply({ availableWidth, availableHeight, elements }) {
      // Landscape (top/bottom): cap at 1400px so wide menus have plenty of space without spanning the full screen.
      // Portrait (left/right): cap at 640px — tall menus stay readable beside the button.
      const cap = imageSize.value.isLandscape ? 1400 : 640;
      Object.assign(elements.floating.style, {
        maxWidth: `${Math.min(Math.max(0, availableWidth), cap)}px`,
        maxHeight: `${Math.max(0, availableHeight)}px`,
      });
    },
  }),
]);

const { floatingStyles } = useFloating(reference, floating, {
  strategy: 'fixed',
  whileElementsMounted: autoUpdate,
  middleware,
});

const floatingClass = 'z-50 rounded-xl border border-base-300 bg-base-100 p-3 shadow-xl overflow-auto';
const imageClass = 'block w-full h-auto object-contain';

function openPopover() {
  if (!isLargeScreen.value || !props.menuUrl) return;

  clearHideTimeout();
  isOpen.value = true;
}

function openMenuHistory() {
  if (!props.menuUrl) return;
  router.push({
    name: 'restaurant-view',
    params: {
      restaurantSlug: props.restaurant.slug,
    },
    query: route.query,
  });
}

function hidePopoverImmediate() {
  clearHideTimeout();
  isOpen.value = false;
}

function clearHideTimeout() {
  if (hideTimeout.value !== null) {
    window.clearTimeout(hideTimeout.value);
    hideTimeout.value = null;
  }
}

function scheduleClosePopover() {
  if (!isLargeScreen.value) {
    hidePopoverImmediate();
    return;
  }

  clearHideTimeout();
  hideTimeout.value = window.setTimeout(() => {
    isOpen.value = false;
    hideTimeout.value = null;
  }, 150);
}

function onImageLoad(event: Event) {
  const target = event.target;
  if (!(target instanceof HTMLImageElement)) return;

  if (target.naturalWidth > 0 && target.naturalHeight > 0) {
    imageWidth.value = target.naturalWidth;
    imageHeight.value = target.naturalHeight;
  }
}

watch(isLargeScreen, (isLarge) => {
  if (!isLarge) {
    hidePopoverImmediate();
  }
});

watch(
  () => [props.menuWidth, props.menuHeight],
  ([width, height]) => {
    imageWidth.value = typeof width === 'number' && width > 0 ? width : null;
    imageHeight.value = typeof height === 'number' && height > 0 ? height : null;
  }
);

onBeforeUnmount(() => {
  clearHideTimeout();
});
</script>

<template>
  <div class="relative block w-full">
    <button
      ref="reference"
      class="btn btn-primary w-full"
      type="button"
      title="Menüverlauf öffnen"
      aria-label="Menüverlauf öffnen"
      @mouseenter="openPopover"
      @mouseleave="scheduleClosePopover"
      @focus="openPopover"
      @blur="hidePopoverImmediate"
      @click="openMenuHistory"
    >
      <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
    </button>
    <teleport to="body">
      <div
        v-if="isOpen && menuUrl"
        ref="floating"
        :style="floatingStyles"
        :class="floatingClass"
        @mouseenter="openPopover"
        @mouseleave="scheduleClosePopover"
        tabindex="-1"
      >
        <img :src="menuUrl" alt="Speisekarte" :width="imageSize.width" :height="imageSize.height" :class="imageClass" loading="lazy" @load="onImageLoad" />
      </div>
    </teleport>
  </div>
</template>
