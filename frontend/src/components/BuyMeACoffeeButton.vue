<script setup lang="ts">
defineProps<{
  href: string;
}>();
</script>

<template>
  <a
    :href="href"
    target="_blank"
    rel="noreferrer"
    aria-label="Buy me a coffee"
    title="Buy me a coffee"
    class="group fixed right-8 bottom-8 md:right-14 md:bottom-14 z-40 flex h-12 max-w-[calc(100vw-2rem)] items-center rounded-full bg-[#ffdd00] px-3 text-black shadow-lg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
  >
    <span class="icon-[simple-icons--buymeacoffee] size-6 shrink-0" aria-hidden="true"></span>
    <span
      class="max-w-0 overflow-hidden text-sm font-semibold whitespace-nowrap opacity-0 transition-[max-width,opacity,margin] duration-300 ease-out group-hover:ml-2 group-hover:max-w-40 group-hover:opacity-100 group-focus-visible:ml-2 group-focus-visible:max-w-40 group-focus-visible:opacity-100 motion-reduce:transition-none"
    >
      Buy me a coffee
    </span>
  </a>
</template>
