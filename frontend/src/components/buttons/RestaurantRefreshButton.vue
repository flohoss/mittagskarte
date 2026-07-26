<script setup lang="ts">
import { computed, ref } from 'vue';
import type { RestaurantRecord } from '../../models/restaurant';
import { RestaurantMethod, RestaurantStatus } from '../../stores/useRestaurants';
import { backendClient } from '../../services/backendClient';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const uploadDialog = ref<HTMLDialogElement | null>(null);
const uploadFileInput = ref<HTMLInputElement | null>(null);
const uploadFile = ref<File | null>(null);
const isUploading = ref(false);
const uploadError = ref('');
const isDragOver = ref(false);
const dragCounter = ref(0);

const ACCEPTED_TYPES = ['image/', 'application/pdf'];
const ACCEPTED_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.webp', '.gif', '.bmp', '.tiff', '.tif', '.heic', '.heif', '.avif', '.svg', '.ico', '.pdf'];
const MAX_FILE_SIZE = 25 * 1024 * 1024; // 25 MB

const acceptedExtensions = ACCEPTED_EXTENSIONS.join(', ');

const formattedUploadFile = computed(() => {
  const file = uploadFile.value;
  if (!file) return null;
  return {
    name: file.name,
    size: formatFileSize(file.size),
  };
});

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function isAcceptedFile(file: File): boolean {
  if (file.size > MAX_FILE_SIZE) return false;
  if (ACCEPTED_TYPES.some((type) => file.type.startsWith(type))) return true;
  const ext = file.name.toLowerCase().match(/\.[^.]+$/)?.[0] ?? '';
  return ACCEPTED_EXTENSIONS.includes(ext);
}

function setUploadFile(file: File | null) {
  if (!file) {
    uploadFile.value = null;
    return;
  }

  if (!isAcceptedFile(file)) {
    uploadError.value = 'Bitte eine Bilddatei oder ein PDF auswählen (max. 25 MB).';
    return;
  }

  uploadFile.value = file;
  uploadError.value = '';
}

const statusMeta = computed(() => {
  switch (props.restaurant.status) {
    case RestaurantStatus.UPDATING:
      return {
        iconClass: 'icon-[fa7-solid--arrows-rotate]',
        label: 'Wird Aktualisiert',
        className: 'btn-neutral',
        iconClassModifier: 'animate-spin',
      };
    case RestaurantStatus.QUEUED:
      return {
        iconClass: 'icon-[fa7-solid--hourglass-half]',
        label: 'In Warteschlange',
        className: 'btn-neutral',
        iconClassModifier: '',
      };
    case RestaurantStatus.COOLDOWN:
      return {
        iconClass: 'icon-[fa7-solid--clock]',
        label: 'Cooldown',
        className: 'btn-neutral',
        iconClassModifier: '',
      };
    default:
      switch (props.restaurant.method) {
        case RestaurantMethod.SCRAPE:
          return {
            iconClass: 'icon-[fa7-solid--arrows-rotate]',
            label: 'Leerlauf',
            className: 'hover:btn-primary',
            iconClassModifier: '',
          };
        case RestaurantMethod.DOWNLOAD:
          return {
            iconClass: 'icon-[fa7-solid--download]',
            label: 'Leerlauf',
            className: 'hover:btn-primary',
            iconClassModifier: '',
          };
        default:
          return {
            iconClass: 'icon-[fa7-solid--upload]',
            label: 'Datei hochladen',
            className: 'hover:btn-primary',
            iconClassModifier: '',
          };
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
  uploadFile.value = null;
  if (uploadFileInput.value) {
    uploadFileInput.value.value = '';
  }
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

  setUploadFile(input.files?.[0] ?? null);
}

function openFilePicker() {
  if (isUploading.value) return;
  uploadFileInput.value?.click();
}

function onDrop(event: DragEvent) {
  event.preventDefault();
  isDragOver.value = false;
  dragCounter.value = 0;
  if (isUploading.value) return;

  const file = event.dataTransfer?.files?.[0];
  if (file) {
    setUploadFile(file);
  }
}

function onDragEnter(event: DragEvent) {
  event.preventDefault();
  if (isUploading.value) return;
  dragCounter.value += 1;
  isDragOver.value = true;
}

function onDragOver(event: DragEvent) {
  event.preventDefault();
  if (isUploading.value) return;
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy';
  }
}

function onDragLeave(event: DragEvent) {
  event.preventDefault();
  dragCounter.value -= 1;
  if (dragCounter.value <= 0) {
    isDragOver.value = false;
    dragCounter.value = 0;
  }
}

function clearUploadFile() {
  uploadFile.value = null;
  uploadError.value = '';
  if (uploadFileInput.value) {
    uploadFileInput.value.value = '';
  }
}

async function submitUpload() {
  if (!uploadFile.value) {
    uploadError.value = 'Bitte eine Datei auswählen.';
    return;
  }

  isUploading.value = true;
  uploadError.value = '';

  try {
    await backendClient.uploadMenu(props.restaurant.id, uploadFile.value);
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
    await backendClient.triggerScrape(props.restaurant.id);
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
    <span :class="['btn-icon', statusMeta.iconClass, statusMeta.iconClassModifier]" aria-hidden="true"></span>
  </button>

  <dialog ref="uploadDialog" class="modal">
    <div class="modal-box">
      <h3 class="text-lg font-semibold">Datei hochladen</h3>
      <p class="mt-1 text-sm text-base-content/70">Für {{ props.restaurant.name }} eine Datei hochladen und als neue Speisekarte verarbeiten.</p>

      <div class="mt-4 grid gap-2">
        <label class="label px-0 pb-1">
          <span class="label-text">Datei (Bild oder PDF)</span>
        </label>
        <input ref="uploadFileInput" type="file" class="hidden" :accept="acceptedExtensions" :disabled="isUploading" @change="onFileChange" />

        <div
          v-if="!uploadFile"
          class="dropzone flex cursor-pointer flex-col items-center justify-center gap-3 rounded-lg border-2 border-dashed p-8 text-center transition-colors"
          :class="[
            isDragOver ? 'border-primary bg-primary/10' : 'border-base-300 hover:border-primary/50 hover:bg-base-200/50',
            isUploading && 'pointer-events-none opacity-60',
          ]"
          role="button"
          tabindex="0"
          @click="openFilePicker"
          @keydown.enter.prevent="openFilePicker"
          @keydown.space.prevent="openFilePicker"
          @drop="onDrop"
          @dragenter="onDragEnter"
          @dragover="onDragOver"
          @dragleave="onDragLeave"
        >
          <span class="icon-[fa7-solid--cloud-arrow-up] size-10 text-base-content/40" aria-hidden="true"></span>
          <div class="text-sm text-base-content/80">
            <span class="font-semibold">Datei hierher ziehen</span>
            <span> oder klicken zum Auswählen</span>
          </div>
          <div class="text-xs text-base-content/60">Bilder, TIFF, HEIC, AVIF oder PDF · max. 25 MB</div>
        </div>

        <div v-else class="flex items-center gap-3 rounded-lg border border-base-300 bg-base-200/50 p-3">
          <span class="icon-[fa7-solid--file] size-8 shrink-0 text-primary" aria-hidden="true"></span>
          <div class="min-w-0 flex-1">
            <div class="truncate text-sm font-medium text-base-content">{{ formattedUploadFile?.name }}</div>
            <div class="text-xs text-base-content/60">{{ formattedUploadFile?.size }}</div>
          </div>
          <button
            v-if="!isUploading"
            type="button"
            class="btn btn-ghost btn-sm btn-square"
            title="Datei entfernen"
            aria-label="Datei entfernen"
            @click="clearUploadFile"
          >
            <span class="icon-[fa7-solid--xmark] size-4" aria-hidden="true"></span>
          </button>
        </div>

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
