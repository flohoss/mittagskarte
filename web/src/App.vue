<script setup lang="ts">
import { Loading, Notify } from 'quasar';
import { OpenAPI } from './openapi';
import { useRestaurantStore } from './stores/restaurants';

const BASE_URL = process.env.BASE_URL || '';
OpenAPI.BASE = BASE_URL + 'api/v1';

defineOptions({
    name: 'App',
    preFetch() {
        Loading.show();
        const store = useRestaurantStore();

        store
            .fetchRestaurants()
            .catch((err) => {
                Notify.create({
                    type: 'negative',
                    message: 'Fehler beim Laden der Daten',
                    caption: err?.body?.message ?? err?.message ?? 'Unbekannter Fehler, bitte prüfen Sie die Konsole für weitere Informationen',
                    actions: [
                        {
                            icon: 'fa-solid fa-xmark',
                            color: 'white',
                            round: true,
                            title: 'Schließen',
                        },
                    ],
                    timeout: 0,
                });
                console.log(err);
            })
            .finally(() => Loading.hide());
    },
});
</script>

<template>
    <router-view />
</template>

<style lang="scss">
.container {
    width: 100%;
    max-width: $breakpoint-sm-max;
    padding-top: 1rem;
}
.no-hover-effect:hover {
    background-color: inherit !important;
    color: inherit !important;
    box-shadow: none !important;
    border-color: inherit !important;
}
.q-btn.btn--no-hover .q-focus-helper {
    display: none;
}
</style>
