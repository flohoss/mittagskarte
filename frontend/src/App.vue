<script setup lang="ts">
import { RouterView } from 'vue-router';
import AppShell from './components/AppShell.vue';
</script>

<template>
  <AppShell title="Schniddzl.de" description="deine Mittagskarte für die Region Stuttgart">
    <RouterView />
  </AppShell>
</template>
