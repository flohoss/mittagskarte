<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue';
import { useBreakpoints, breakpointsTailwind } from '@vueuse/core';
import { useFloating, autoUpdate, offset, shift, flip, size } from '@floating-ui/vue';
import { useRouter } from 'vue-router';
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
const isSmallerThanLg = breakpoints.smaller('lg');
const imageWidth = ref<number | null>(props.menuWidth ?? null);
const imageHeight = ref<number | null>(props.menuHeight ?? null);
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

const floatingClass = computed(() => {
  const base = 'z-50 max-h-[95vh] rounded-xl border border-base-300 bg-base-100 p-3 shadow-xl';
  return imageSize.value.isLandscape ? `${base} overflow-hidden` : `${base} overflow-auto`;
});

const imageClass = computed(() =>
  imageSize.value.isLandscape
    ? 'block h-auto w-full max-h-[calc(95vh-28px)] object-contain'
    : 'block h-auto w-full object-contain'
);

const { floatingStyles } = useFloating(reference, floating, {
  placement: 'right',
  strategy: 'fixed',
  whileElementsMounted: autoUpdate,
  middleware: [
    offset(8),
    flip(),
    shift({ padding: 8, crossAxis: true }),
    size({
      padding: 8,
      apply({ availableWidth, elements }) {
        Object.assign(elements.floating.style, {
          maxWidth: imageSize.value.isLandscape
            ? `${Math.max(0, availableWidth)}px`
            : `${Math.min(Math.max(0, availableWidth), 576)}px`,
        });
      },
    }),
  ],
});

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
      restaurantId: props.restaurant.id,
    },
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

watch(isSmallerThanLg, (isSmall) => {
  if (isSmall) {
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
