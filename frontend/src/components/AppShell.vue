<script setup lang="ts">
import Github from '~icons/simple-icons/github';

defineProps<{
  title: string;
  description: string;
}>();
</script>

<template>
  <div class="flex flex-col min-h-screen">
    <header class="navbar bg-base-200 rounded-xl">
      <div class="container flex justify-between items-center">
        <a href="/" class="flex items-center gap-2 flex-shrink-0">
          <img src="/static/schniddzl.webp" :alt="title" class="size-10" />
          <div class="font-light font-stretch-semi-expanded min-w-0 hidden lg:block">
            <div class="text-xl font-semibold truncate">{{ title }}</div>
            <div class="text-xs truncate">{{ description }}</div>
          </div>
        </a>
      </div>
    </header>
    <main class="my-5 lg:my-10 flex-grow">
      <div class="container">
        <slot />
      </div>
    </main>
    <footer class="navbar bg-base-200 rounded-xl px-4 my-2">
      <div class="container flex justify-between items-center">
        <div>
          {{ title }} - {{ description }} -
          <a class="link" title="Source Code auf GitHub" target="_blank" data-lg-blank="" href="https://github.com/flohoss/mittagskarte">Source Code</a>
        </div>
        <a target="_blank" data-lg-blank="" class="btn btn-circle btn-ghost" title="GitHub" href="https://github.com/flohoss/">
          <Github class="size-5" />
        </a>
      </div>
    </footer>
  </div>
</template>
