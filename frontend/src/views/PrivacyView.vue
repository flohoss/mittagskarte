<script setup lang="ts">
const email = (window as any)?.__EMAIL__;
</script>

<template>
  <section class="mx-auto w-full max-w-3xl">
    <h1 class="text-2xl font-semibold md:text-3xl">Datenschutz</h1>
    <p class="mt-3 text-sm leading-relaxed">Diese Seite ist ein privates Hobbyprojekt ohne Werbung und ohne kommerzielle Nutzung.</p>

    <template v-if="email">
      <h2 class="mt-8 text-base font-semibold">Kontakt</h2>
      <p class="mt-2 text-sm leading-relaxed">
        Bei Fragen zum Datenschutz erreichst du mich per E-Mail:
        <a class="link" :href="`mailto:${email}`">{{ email }}</a>
      </p>
    </template>

    <h2 class="mt-8 text-base font-semibold">Welche Daten wir nutzen</h2>
    <ul class="mt-2 list-disc space-y-1 pl-5 text-sm leading-relaxed">
      <li>Server-Logdaten bei Seitenaufruf (z. B. IP-Adresse, Uhrzeit, aufgerufene Seite)</li>
      <li>Favoriten im LocalStorage deines Browsers</li>
      <li>Anmeldedaten nur bei optionalem Login für Admin-Funktionen</li>
    </ul>

    <h2 class="mt-8 text-base font-semibold">Wofür wir die Daten nutzen</h2>
    <p class="mt-2 text-sm leading-relaxed">
      Wir nutzen Daten nur für den sicheren Betrieb der Website, zur Fehleranalyse und für die Login-Funktion. Es gibt keine Analyse-, Tracking- oder
      Werbedienste.
    </p>

    <h2 class="mt-8 text-base font-semibold">Weitergabe und Speicherdauer</h2>
    <p class="mt-2 text-sm leading-relaxed">
      Es erfolgt keine Weitergabe an Dritte. Server-Logs werden nach spätestens 7 Tagen gelöscht. LocalStorage-Daten kannst du jederzeit in deinem Browser
      entfernen.
    </p>

    <h2 class="mt-8 text-base font-semibold">Deine Rechte</h2>
    <p class="mt-2 text-sm leading-relaxed">
      Du hast die Rechte auf Auskunft, Berichtigung, Löschung, Einschränkung, Datenübertragbarkeit und Widerspruch.
      <span v-if="email">
        Schreib mir dazu an
        <a class="link" :href="`mailto:${email}`"> {{ email }} </a>.
      </span>
    </p>

    <p class="mt-8 text-xs text-base-content/60">Stand: Mai 2026</p>
  </section>
</template>
