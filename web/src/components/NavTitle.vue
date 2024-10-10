<template>
  <q-btn class="btn--no-hover" flat :ripple="false" dense to="/">
    <q-item-section avatar>
      <q-avatar>
        <img src="/favicon/android-chrome-192x192.png" />
      </q-avatar>
    </q-item-section>

    <q-item-section>
      <q-item-label class="text-h6">
        Mittag<span v-if="$q.screen.gt.xs" class="text-primary">stisch</span>
      </q-item-label>
    </q-item-section>
  </q-btn>
</template>
