<script setup lang="ts">
import { ref } from 'vue';
import { useFloating, autoUpdate, shift, flip, size } from '@floating-ui/vue';
import Fa7SolidListAlt from '~icons/fa7-solid/list-alt';
import { BackendURL } from '../main';

const props = defineProps<{
  menuUrl: string;
}>();

const isOpen = ref(false);
const reference = ref<HTMLElement | null>(null);
const floating = ref<HTMLElement | null>(null);

const { floatingStyles } = useFloating(reference, floating, {
  whileElementsMounted: autoUpdate,
  middleware: [
    size({
      padding: 8,
      apply({ availableWidth, availableHeight, elements }) {
        Object.assign(elements.floating.style, {
          maxWidth: `${Math.max(0, availableWidth)}px`,
          maxHeight: `${Math.max(0, availableHeight)}px`,
        });
      },
    }),
    flip({ fallbackStrategy: 'initialPlacement' }),
    shift({ padding: 8 }),
  ],
});

function openPopover() {
  isOpen.value = true;
}
function closePopover() {
  isOpen.value = false;
}
</script>

<template>
  <div class="relative inline-block">
    <a
      ref="reference"
      class="btn btn-primary"
      title="Speisekarte"
      aria-label="Speisekarte öffnen"
      :href="BackendURL + menuUrl"
      target="_blank"
      rel="noopener noreferrer"
      @mouseenter="openPopover"
      @mouseleave="closePopover"
      @focus="openPopover"
      @blur="closePopover"
      @click="openPopover"
    >
      <Fa7SolidListAlt class="btn-icon" aria-hidden="true" />
    </a>
    <teleport to="body">
      <div
        v-if="isOpen"
        ref="floating"
        :style="floatingStyles"
        class="z-50 rounded-xl border border-base-300 bg-base-100 p-4 shadow-xl min-w-[320px] min-h-[400px] max-w-[90vw] max-h-[80vh] overflow-auto"
        @mouseenter="openPopover"
        @mouseleave="closePopover"
        tabindex="-1"
      >
        <img :src="BackendURL + menuUrl" alt="Speisekarte" class="w-full h-auto object-contain" />
      </div>
    </teleport>
  </div>
</template>
