<script setup lang="ts">
import { computed, ref } from 'vue';
import type { RestaurantRecord } from '../models/restaurant';
import Fa7SolidArrowsRotate from '~icons/fa7-solid/arrows-rotate';
import Fa7SolidDownload from '~icons/fa7-solid/download';
import Fa7SolidUpload from '~icons/fa7-solid/upload';
import Fa7SolidHourglassHalf from '~icons/fa7-solid/hourglass-half';
import Fa7SolidClock from '~icons/fa7-solid/clock';
import { RestaurantMethod, RestaurantStatus } from '../stores/useRestaurants';
import { useLogin } from '../stores/useLogin';
import { BackendURL } from '../config';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const uploadDialog = ref<HTMLDialogElement | null>(null);
const uploadFileInput = ref<HTMLInputElement | null>(null);
const uploadFile = ref<File | null>(null);
const isUploading = ref(false);
const isAuthenticating = ref(false);
const uploadError = ref('');
const authError = ref('');
const loginIdentity = ref('');
const loginPassword = ref('');
const { getAuthToken, isAuthenticated, authIdentity, authenticate, clearAuthentication } = useLogin();
const hasAuthToken = computed(() => isAuthenticated.value);

const statusMeta = computed(() => {
  switch (props.restaurant.status) {
    case RestaurantStatus.UPDATING:
      return { icon: Fa7SolidArrowsRotate, label: 'Wird Aktualisiert', className: 'btn-neutral', iconClass: 'animate-spin' };
    case RestaurantStatus.QUEUED:
      return { icon: Fa7SolidHourglassHalf, label: 'In Warteschlange', className: 'btn-neutral', iconClass: '' };
    case RestaurantStatus.COOLDOWN:
      return { icon: Fa7SolidClock, label: 'Cooldown', className: 'btn-neutral', iconClass: '' };
    default:
      switch (props.restaurant.method) {
        case RestaurantMethod.SCRAPE:
          return { icon: Fa7SolidArrowsRotate, label: 'Leerlauf', className: 'hover:btn-primary', iconClass: '' };
        case RestaurantMethod.DOWNLOAD:
          return { icon: Fa7SolidDownload, label: 'Leerlauf', className: 'hover:btn-primary', iconClass: '' };
        default:
          return { icon: Fa7SolidUpload, label: 'Datei hochladen', className: 'hover:btn-primary', iconClass: '' };
      }
  }
});

const canTriggerRefresh = computed(() => props.restaurant.status === RestaurantStatus.IDLE);
const isUploadMethod = computed(() => props.restaurant.method === RestaurantMethod.UPLOAD);

function openUploadDialog() {
  uploadError.value = '';
  uploadFile.value = null;
  uploadDialog.value?.showModal();
}

function resetUploadDialogState() {
  uploadError.value = '';
  authError.value = '';
  uploadFile.value = null;
  if (uploadFileInput.value) {
    uploadFileInput.value.value = '';
  }
}

async function loginForUpload() {
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

function logoutUploadAuth() {
  clearAuthentication();
  authError.value = '';
}

function closeUploadDialog() {
  if (isUploading.value) return;

  resetUploadDialogState();
  uploadDialog.value?.close();
}

function forceCloseUploadDialog() {
  resetUploadDialogState();
  uploadDialog.value?.close();
}

function onFileChange(event: Event) {
  const input = event.target;
  if (!(input instanceof HTMLInputElement)) return;

  uploadFile.value = input.files?.[0] ?? null;
  uploadError.value = '';
}

async function submitUpload() {
  if (!uploadFile.value) {
    uploadError.value = 'Bitte eine Datei auswählen.';
    return;
  }

  const authToken = getAuthToken();
  if (!authToken) {
    uploadError.value = 'Bitte zuerst anmelden.';
    return;
  }

  isUploading.value = true;
  uploadError.value = '';

  try {
    const formData = new FormData();
    formData.append('id', props.restaurant.id);
    formData.append('file', uploadFile.value);

    const response = await fetch(`${BackendURL}/api/restaurants/upload`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${authToken}`,
      },
      body: formData,
    });

    if (!response.ok) {
      const message = (await response.text()) || 'Upload fehlgeschlagen.';
      throw new Error(message);
    }

    forceCloseUploadDialog();
  } catch (error) {
    uploadError.value = error instanceof Error ? error.message : 'Upload fehlgeschlagen.';
  } finally {
    isUploading.value = false;
  }
}

async function triggerRefresh() {
  if (isUploadMethod.value) {
    openUploadDialog();
    return;
  }

  try {
    await fetch(`${BackendURL}/api/restaurants/scrape`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: props.restaurant.id,
      }),
    });
  } catch (error) {
    console.error('Failed to trigger scrape', error);
  }
}
</script>

<template>
  <button
    type="button"
    :class="['btn btn-soft', statusMeta.className]"
    :title="canTriggerRefresh ? 'Jetzt aktualisieren' : `Status: ${statusMeta.label}`"
    :aria-label="canTriggerRefresh ? 'Jetzt aktualisieren' : `Status: ${statusMeta.label}`"
    :disabled="!canTriggerRefresh"
    @click="triggerRefresh"
  >
    <component :is="statusMeta.icon" :class="['btn-icon', statusMeta.iconClass]" aria-hidden="true" />
  </button>

  <dialog ref="uploadDialog" class="modal">
    <div class="modal-box">
      <h3 class="text-lg font-semibold">Datei hochladen</h3>
      <p class="mt-1 text-sm text-base-content/70">Für {{ props.restaurant.name }} eine Datei hochladen und als neue Speisekarte verarbeiten.</p>

      <div class="mt-4 grid gap-2">
        <div class="grid gap-2 rounded-lg border border-base-300 p-3">
          <p class="text-sm font-medium">Anmeldung</p>

          <template v-if="!hasAuthToken">
            <input
              v-model="loginIdentity"
              type="email"
              class="input input-bordered w-full"
              placeholder="E-Mail"
              autocomplete="username"
              :disabled="isUploading || isAuthenticating"
            />
            <input
              v-model="loginPassword"
              type="password"
              class="input input-bordered w-full"
              placeholder="Passwort"
              autocomplete="current-password"
              :disabled="isUploading || isAuthenticating"
              @keydown.enter.prevent="loginForUpload"
            />
            <button
              type="button"
              class="btn btn-outline"
              :disabled="isUploading || isAuthenticating"
              @click="loginForUpload"
            >
              <span v-if="isAuthenticating" class="loading loading-spinner loading-xs" aria-hidden="true" />
              <span>{{ isAuthenticating ? 'Anmeldung läuft...' : 'Anmelden' }}</span>
            </button>
          </template>

          <div v-else class="flex items-center justify-between gap-2">
            <div class="alert alert-success py-2">
              <span class="text-xs">
                Angemeldet als <span class="font-semibold">{{ authIdentity || 'Benutzer' }}</span>
              </span>
            </div>
            <button type="button" class="btn btn-xs btn-ghost" :disabled="isUploading || isAuthenticating" @click="logoutUploadAuth">Abmelden</button>
          </div>

          <p v-if="authError" class="text-sm text-error">{{ authError }}</p>
        </div>

        <label class="label px-0 pb-1">
          <span class="label-text">Datei (PDF oder Bild)</span>
        </label>
        <input ref="uploadFileInput" type="file" class="file-input file-input-bordered w-full" accept=".pdf,image/*" :disabled="isUploading" @change="onFileChange" />
        <p v-if="uploadError" class="text-sm text-error">{{ uploadError }}</p>
      </div>

      <div class="modal-action">
        <button type="button" class="btn" :disabled="isUploading" @click="closeUploadDialog">Abbrechen</button>
        <button type="button" class="btn btn-primary" :class="{ 'btn-disabled': isUploading }" :disabled="isUploading" @click="submitUpload">
          <span v-if="isUploading" class="loading loading-spinner loading-xs" aria-hidden="true" />
          <span>{{ isUploading ? 'Lädt hoch...' : 'Hochladen' }}</span>
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop" @submit.prevent="closeUploadDialog">
      <button type="submit">close</button>
    </form>
  </dialog>
</template>
