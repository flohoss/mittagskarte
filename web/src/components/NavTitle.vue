<template>
  <q-btn class="no-hover-effect" flat clickable dense to="/">
    <q-item-section avatar>
      <q-avatar>
        <img src="/favicon/android-chrome-192x192.png" />
      </q-avatar>
    </q-item-section>

    <q-item-section>
      <q-item-label class="text-h6" v-if="$q.screen.gt.xs"
        >Mittagstisch</q-item-label
      >
    </q-item-section>
  </q-btn>
</template>
