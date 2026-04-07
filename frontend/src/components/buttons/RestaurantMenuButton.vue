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
  const aspectRatio = width && height ? width / height : null;

  return {
    width,
    height,
    aspectRatio,
    isLandscape: !!width && !!height ? width >= height : false,
  };
});

const isScrollablePortrait = computed(() => {
  if (!imageSize.value.aspectRatio || imageSize.value.isLandscape) return false;

  const ratio = imageSize.value.aspectRatio;
  const height = imageSize.value.height ?? 0;

  // Keep standard portrait pages (e.g. 9:16) fitting in view.
  // Switch to scroll mode for narrower/taller menu captures to keep text readable.
  return ratio < 0.55 || (height >= 3200 && ratio < 0.72);
});

const popoverSizingStyle = computed(() => {
  if (imageSize.value.isLandscape) {
    return {
      width: 'min(78vw, 1100px)',
      minWidth: '420px',
      minHeight: '320px',
    };
  }

  if (imageSize.value.aspectRatio && isScrollablePortrait.value) {
    // Convert portrait ratio to a viewport width target based on max tooltip height.
    const portraitWidthVw = Math.max(22, Math.min(32, imageSize.value.aspectRatio * 85));
    return {
      width: `${portraitWidthVw}vw`,
      minWidth: '0',
      minHeight: '0',
    };
  }

  return {
    width: 'auto',
    minWidth: '0',
    minHeight: '0',
  };
});

const imageClass = computed(() => {
  if (imageSize.value.isLandscape) {
    return 'block h-auto w-full max-h-[calc(95vh-28px)] object-contain';
  }

  if (isScrollablePortrait.value) {
    return 'block h-auto w-full object-contain';
  }

  return 'block h-auto w-auto max-h-[calc(95vh-28px)] max-w-[62vw] object-contain';
});

const popoverOverflowClass = computed(() => (isScrollablePortrait.value ? 'overflow-auto' : 'overflow-hidden'));

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
          maxWidth: `${Math.max(0, availableWidth)}px`,
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
    name: 'restaurant-menu-history',
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
        :style="[floatingStyles, popoverSizingStyle]"
        :class="['z-50 max-h-[95vh] max-w-[90vw] rounded-xl border border-base-300 bg-base-100 p-3 shadow-xl', popoverOverflowClass]"
        @mouseenter="openPopover"
        @mouseleave="scheduleClosePopover"
        tabindex="-1"
      >
        <img :src="menuUrl" alt="Speisekarte" :width="imageSize.width" :height="imageSize.height" :class="imageClass" loading="lazy" @load="onImageLoad" />
      </div>
    </teleport>
  </div>
</template>
