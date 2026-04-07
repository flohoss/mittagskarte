<script setup lang="ts">
import { computed, ref } from 'vue';
import Fa7SolidEnvelope from '~icons/fa7-solid/envelope';
import Fa7SolidLock from '~icons/fa7-solid/lock';
import Fa7SolidLockOpen from '~icons/fa7-solid/lock-open';
import Fa7SolidRightToBracket from '~icons/fa7-solid/right-to-bracket';
import Fa7SolidRightFromBracket from '~icons/fa7-solid/right-from-bracket';
import { useLogin } from '../stores/useLogin';

const loginDialog = ref<HTMLDialogElement | null>(null);
const isAuthenticating = ref(false);
const authError = ref('');
const loginIdentity = ref('');
const loginPassword = ref('');

const { isAuthenticated, authIdentity, authenticate, clearAuthentication } = useLogin();
const hasAuthToken = computed(() => isAuthenticated.value);

const initials = computed(() => {
  const identity = authIdentity.value || '';
  const name = identity.includes('@') ? identity.split('@')[0] : identity;
  const parts = name.split(/[.\-_\s]+/).filter(Boolean);
  if (parts.length === 0) return '?';
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase();
  return (parts[0][0] + parts[1][0]).toUpperCase();
});

function openLoginDialog() {
  authError.value = '';
  loginDialog.value?.showModal();
}

function closeLoginDialog() {
  if (isAuthenticating.value) return;

  authError.value = '';
  loginPassword.value = '';
  loginDialog.value?.close();
}

async function login() {
  if (!loginIdentity.value.trim() || !loginPassword.value) {
    authError.value = 'Bitte E-Mail und Passwort eingeben.';
    return;
  }

  isAuthenticating.value = true;
  authError.value = '';

  try {
    await authenticate(loginIdentity.value.trim(), loginPassword.value);
    loginPassword.value = '';
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Anmeldung fehlgeschlagen.';
  } finally {
    isAuthenticating.value = false;
  }
}

function logout() {
  clearAuthentication();
}
</script>

<template>
  <button v-if="!hasAuthToken" type="button" class="btn btn-soft btn-square rounded-lg" title="Anmelden" aria-label="Anmelden" @click="openLoginDialog">
    <Fa7SolidLockOpen class="size-4" aria-hidden="true" />
  </button>

  <button
    v-else
    type="button"
    class="btn btn-soft btn-success btn-square rounded-lg"
    :title="`Angemeldet als ${authIdentity || 'Benutzer'}`"
    :aria-label="`Angemeldet als ${authIdentity || 'Benutzer'}`"
    @click="openLoginDialog"
  >
    <Fa7SolidLock class="size-4" aria-hidden="true" />
  </button>

  <dialog ref="loginDialog" class="modal">
    <div class="modal-box max-w-md p-5 sm:p-6">
      <template v-if="hasAuthToken">
        <div class="flex flex-col items-center gap-4 py-4">
          <div class="avatar avatar-placeholder">
            <div class="bg-success text-success-content size-16 rounded-full">
              <span class="text-xl font-semibold">{{ initials }}</span>
            </div>
          </div>
          <p class="text-sm font-medium">{{ authIdentity }}</p>
        </div>
        <div class="modal-action mt-2 flex-col gap-2 sm:flex-row sm:justify-end">
          <button type="button" class="btn w-full sm:w-auto" @click="closeLoginDialog">Schließen</button>
          <button type="button" class="btn btn-error w-full sm:w-auto" @click="logout">
            <Fa7SolidRightFromBracket class="size-4" aria-hidden="true" />
            <span>Abmelden</span>
          </button>
        </div>
      </template>
      <template v-else>
        <div class="text-center">
          <h3 class="text-lg font-semibold">Anmeldung</h3>
          <p class="mt-1 text-sm text-base-content/70">Bitte mit einem Konto einloggen, um Aktionen auszuführen.</p>
        </div>

        <div class="mt-5 grid gap-3">
          <label class="input input-bordered flex w-full items-center gap-2">
            <Fa7SolidEnvelope class="size-4 opacity-70" aria-hidden="true" />
            <input v-model="loginIdentity" type="email" class="grow" placeholder="E-Mail" autocomplete="username" :disabled="isAuthenticating" />
          </label>

          <label class="input input-bordered flex w-full items-center gap-2">
            <Fa7SolidLock class="size-4 opacity-70" aria-hidden="true" />
            <input
              v-model="loginPassword"
              type="password"
              class="grow"
              placeholder="Passwort"
              autocomplete="current-password"
              :disabled="isAuthenticating"
              @keydown.enter.prevent="login"
            />
          </label>

          <p v-if="authError" class="text-sm text-error">{{ authError }}</p>
        </div>

        <div class="modal-action mt-6 flex-col-reverse gap-2 sm:flex-row sm:items-center sm:justify-end">
          <button type="button" class="btn w-full sm:w-auto" :disabled="isAuthenticating" @click="closeLoginDialog">Schließen</button>
          <button type="button" class="btn btn-primary w-full sm:w-auto" :disabled="isAuthenticating" @click="login">
            <span v-if="isAuthenticating" class="loading loading-spinner loading-xs" aria-hidden="true" />
            <Fa7SolidRightToBracket v-else class="size-4" aria-hidden="true" />
            <span>{{ isAuthenticating ? 'Anmeldung läuft...' : 'Einloggen' }}</span>
          </button>
        </div>
      </template>
    </div>
    <form method="dialog" class="modal-backdrop" @submit.prevent="closeLoginDialog">
      <button type="submit">close</button>
    </form>
  </dialog>
</template>
