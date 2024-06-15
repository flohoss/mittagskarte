<template>
  <router-view />
</template>

<script setup lang="ts">
import { OpenAPI } from './openapi';

defineOptions({
  name: 'App',
});

OpenAPI.BASE = 'http://localhost:4000/api/v1';
</script>
