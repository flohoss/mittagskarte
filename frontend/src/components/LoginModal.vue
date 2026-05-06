<script setup lang="ts">
import { computed, ref } from 'vue';
import Fa7SolidEnvelope from '~icons/fa7-solid/envelope';
import Fa7SolidLockOpen from '~icons/fa7-solid/lock-open';
import Fa7SolidRightToBracket from '~icons/fa7-solid/right-to-bracket';
import Fa7SolidRightFromBracket from '~icons/fa7-solid/right-from-bracket';
import { useLogin } from '../stores/useLogin';

const props = withDefaults(
  defineProps<{
    showLabel?: boolean;
    showTrigger?: boolean;
  }>(),
  {
    showLabel: false,
    showTrigger: true,
  }
);

const emit = defineEmits<{
  closed: [];
  authenticated: [];
}>();

const loginDialog = ref<HTMLDialogElement | null>(null);
const isAuthenticating = ref(false);
const authError = ref('');
const loginIdentity = ref('');
const loginPassword = ref('');

const { isAuthenticated, authenticate, clearAuthentication } = useLogin();
const hasAuthToken = computed(() => isAuthenticated.value);

function openLoginDialog() {
  authError.value = '';
  loginDialog.value?.showModal();
}

function closeLoginDialog(notify = true, force = false) {
  if (isAuthenticating.value && !force) return;

  authError.value = '';
  loginPassword.value = '';
  loginDialog.value?.close();
  if (notify) {
    emit('closed');
  }
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
    closeLoginDialog(false, true);
    emit('authenticated');
  } catch (error) {
    authError.value = error instanceof Error ? error.message : 'Anmeldung fehlgeschlagen.';
  } finally {
    isAuthenticating.value = false;
  }
}

function logout() {
  clearAuthentication();
}

defineExpose({
  open: openLoginDialog,
});
</script>

<template>
  <button
    v-if="props.showTrigger && !hasAuthToken"
    type="button"
    :class="['btn btn-soft rounded-lg', props.showLabel ? 'gap-2' : 'btn-square']"
    title="Anmelden"
    aria-label="Anmelden"
    @click="openLoginDialog"
  >
    <Fa7SolidLockOpen class="size-4" aria-hidden="true" />
    <span v-if="props.showLabel">Anmelden</span>
  </button>

  <button
    v-else-if="props.showTrigger"
    type="button"
    :class="['btn btn-soft btn-error rounded-lg', props.showLabel ? 'gap-2' : 'btn-square']"
    title="Abmelden"
    aria-label="Abmelden"
    @click="logout"
  >
    <Fa7SolidRightFromBracket class="size-4" aria-hidden="true" />
    <span v-if="props.showLabel">Abmelden</span>
  </button>

  <teleport to="body">
    <dialog ref="loginDialog" class="modal">
      <div class="modal-box max-w-md p-5 sm:p-6">
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
            <Fa7SolidLockOpen class="size-4 opacity-70" aria-hidden="true" />
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
      </div>
      <form method="dialog" class="modal-backdrop" @submit.prevent="closeLoginDialog">
        <button type="submit">close</button>
      </form>
    </dialog>
  </teleport>
</template>
