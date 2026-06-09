<script setup lang="ts">
import Fa7SolidStar from '~icons/fa7-solid/star';
import { computed } from 'vue';
import RestaurantActions from './RestaurantActions.vue';
import { useFavorites } from '../stores/useFavorites';
import type { RestaurantRecord } from '../models/restaurant';
import { useRestaurants } from '../stores/useRestaurants';
import { useNow } from '../composables/useNow';
import { getLatestMenu } from '../utils/menu';

const props = defineProps<{
  restaurant: RestaurantRecord;
}>();

const { getFileUrl, applySearch, getRestaurantDistanceKm, sortBy, coords } = useRestaurants();
const { isFavorite, toggleFavorite } = useFavorites();

const WEEKDAYS = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];
const RELATIVE_TIME_UNITS: Array<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 60 * 60 * 24 * 365],
  ['month', 60 * 60 * 24 * 30],
  ['week', 60 * 60 * 24 * 7],
  ['day', 60 * 60 * 24],
  ['hour', 60 * 60],
  ['minute', 60],
];

const relativeTimeFormatter = new Intl.RelativeTimeFormat('de', {
  numeric: 'auto',
  style: 'long',
});

type LastCheckStatus = NonNullable<NonNullable<RestaurantRecord['last_check']>['status']>;
type FreshnessStage = 'fresh' | 'aging' | 'warning' | 'stale' | 'expired';

const freshnessVisuals: Record<FreshnessStage, { className: string; prefix: string; title: string }> = {
  fresh: {
    className: 'badge-success',
    prefix: 'neu',
    title: 'Aktuell',
  },
  aging: {
    className: 'border border-success bg-success text-success-content',
    prefix: 'älter',
    title: 'Noch brauchbar, aber altert',
  },
  warning: {
    className: 'border border-warning bg-warning text-warning-content',
    prefix: 'alt',
    title: 'Sollte bald aktualisiert werden',
  },
  stale: {
    className: 'border badge-stale',
    prefix: 'sehr alt',
    title: 'Wahrscheinlich nicht mehr aktuell',
  },
  expired: {
    className: 'badge-error',
    prefix: 'zu alt',
    title: 'Nicht mehr aktuell',
  },
};

const nowMs = useNow(30_000);
const currentWeekday = computed(() => WEEKDAYS[new Date(nowMs.value).getDay()]);
const isClosed = computed(() => props.restaurant.rest_days.includes(currentWeekday.value));
const isFavorited = computed(() => isFavorite(props.restaurant.id));
const thumbnailUrl = computed(() => getFileUrl(props.restaurant));
const latestMenuCreated = computed(() => {
  return getLatestMenu(props.restaurant.expand?.menus)?.created ?? null;
});
const lastCheck = computed(() => {
  const lc = props.restaurant.last_check;
  return lc ?? null;
});
const freshnessReferenceDate = computed(() => {
  const menuDate = latestMenuCreated.value;
  const check = lastCheck.value;

  if (!check?.at || (check.status !== 'success' && check.status !== 'not_changed')) {
    return menuDate;
  }

  if (!menuDate) return check.at;

  const menuTs = toTimestamp(menuDate);
  const checkTs = toTimestamp(check.at);
  if (menuTs === null) return check.at;
  if (checkTs === null) return menuDate;

  return checkTs > menuTs ? check.at : menuDate;
});
const menuFreshness = computed(() => {
  if (!latestMenuCreated.value || !freshnessReferenceDate.value) return null;
  return getMenuFreshnessMeta(
    latestMenuCreated.value,
    freshnessReferenceDate.value,
    props.restaurant.cron,
    props.restaurant.method,
    lastCheck.value?.status
  );
});
const lastCheckText = computed(() => {
  if (!lastCheck.value) return '';
  return `${formatRelativeDate(lastCheck.value.at)} zuletzt versucht`;
});
const distanceKm = computed(() => getRestaurantDistanceKm(props.restaurant));
const showDistance = computed(() => sortBy.value === 'distance-asc' && coords.value && distanceKm.value !== null);
const distanceLabel = computed(() => {
  if (distanceKm.value === null) return '';

  if (distanceKm.value < 1) {
    return `~${Math.round(distanceKm.value * 1000)} m`;
  }

  if (distanceKm.value < 10) {
    return `~${distanceKm.value.toFixed(1)} km`;
  }

  return `~${Math.round(distanceKm.value)} km`;
});
const lastCheckTitle = computed(() => {
  if (!lastCheck.value) return '';
  if (lastCheck.value.status === 'success') {
    return 'Erfolgreich aktualisiert';
  }
  if (lastCheck.value.status === 'not_changed') {
    return 'Keine Änderung gefunden';
  }
  if (lastCheck.value.status === 'error') {
    return `Fehler${lastCheck.value.detail ? `: ${lastCheck.value.detail}` : ''}`;
  }
  return '';
});

function formatRelativeDate(value: string) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return 'Unbekannt';

  const rawDiffSeconds = Math.round((date.getTime() - nowMs.value) / 1000);
  // Backend and client clocks can drift slightly; avoid showing future times in the UI.
  const diffSeconds = Math.min(0, rawDiffSeconds);

  if (diffSeconds > -60) return 'gerade eben';

  for (const [unit, seconds] of RELATIVE_TIME_UNITS) {
    if (Math.abs(diffSeconds) >= seconds) {
      return relativeTimeFormatter.format(Math.round(diffSeconds / seconds), unit);
    }
  }

  return 'gerade eben';
}

function toTimestamp(value?: string | null) {
  if (!value) return null;
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? null : timestamp;
}

function countCronWeekdays(field: string): number {
  if (field === '*') return 7;
  let count = 0;
  for (const part of field.split(',')) {
    if (part.includes('-')) {
      const [start, end] = part.split('-').map(Number);
      count += end - start + 1;
    } else if (!Number.isNaN(Number(part))) {
      count += 1;
    }
  }
  return count || 1;
}

function getExpectedIntervalMs(cron: string, method: string): number {
  const dayMs = 24 * 60 * 60 * 1000;
  const normalizedMethod = method.trim().toLowerCase();
  if (normalizedMethod === 'upload') return 7 * dayMs;
  if (!cron) return 7 * dayMs;

  const fields = cron.trim().split(/\s+/);
  if (fields.length < 5) return 7 * dayMs;

  const [, , dayOfMonth, , dayOfWeek] = fields;

  // Monthly: runs on a specific day of the month
  if (dayOfMonth !== '*') return 30 * dayMs;

  // Weekly or daily depending on how many weekdays
  if (dayOfWeek !== '*') {
    const days = countCronWeekdays(dayOfWeek);
    if (days >= 5) return dayMs; // Mon–Fri treated as daily
    return Math.round((7 / days) * dayMs);
  }

  // Every day (or more frequent)
  return dayMs;
}

function getFreshnessStage(ageMs: number, intervalMs: number): FreshnessStage {
  const ageRatio = ageMs / intervalMs;
  if (ageRatio <= 1) return 'fresh';
  if (ageRatio <= 1.5) return 'aging';
  if (ageRatio <= 2.5) return 'warning';
  if (ageRatio <= 4) return 'stale';
  return 'expired';
}

function getFreshnessPrefix(stage: FreshnessStage, checkStatus?: LastCheckStatus) {
  if ((stage === 'fresh' || stage === 'aging') && checkStatus === 'not_changed') {
    return 'unverändert';
  }

  return freshnessVisuals[stage].prefix;
}

function getMenuFreshnessMeta(
  menuDateValue: string,
  referenceDateValue: string,
  cron: string,
  method: string,
  checkStatus?: LastCheckStatus
) {
  const menuDate = new Date(menuDateValue);
  const referenceDate = new Date(referenceDateValue);
  if (Number.isNaN(menuDate.getTime()) || Number.isNaN(referenceDate.getTime())) {
    return {
      className: 'badge-neutral',
      label: 'Unbekannt',
      title: 'Datum unbekannt',
    };
  }

  const ageMs = nowMs.value - referenceDate.getTime();
  const intervalMs = getExpectedIntervalMs(cron, method);
  const relative = formatRelativeDate(menuDateValue);
  const escapedCron = cron.trim() || 'manuell (Upload)';

  if (checkStatus === 'error') {
    return {
      className: 'badge-error',
      label: `Fehler • ${relative}`,
      title: `Letzte Prüfung fehlgeschlagen (Menü: ${relative}, Intervall: ${escapedCron})`,
    };
  }

  const stage = getFreshnessStage(ageMs, intervalMs);
  const visual = freshnessVisuals[stage];
  const prefix = getFreshnessPrefix(stage, checkStatus);

  return {
    className: visual.className,
    label: `${prefix} • ${relative}`,
    title: `${visual.title} (${relative}, Intervall: ${escapedCron})`,
  };
}

function getInitials(name: string) {
  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map((part) => part[0]?.toUpperCase() ?? '')
    .join('');
}
</script>

<template>
  <article
    class="group card card-border overflow-hidden rounded-xl bg-base-100 opacity-80 shadow-md transition-[shadow,opacity] duration-200 hover:opacity-100 hover:shadow-xl"
  >
    <figure class="relative h-30 overflow-hidden bg-base-300">
      <img
        v-if="thumbnailUrl"
        :src="thumbnailUrl"
        :alt="props.restaurant.name"
        :class="['h-full w-full object-cover transition-transform duration-500 group-hover:scale-105', isClosed ? 'opacity-40 grayscale' : '']"
        loading="lazy"
      />
      <div
        v-else
        :class="['flex h-full w-full items-center justify-center text-2xl font-semibold', isClosed ? 'opacity-40 grayscale' : 'text-base-content/70']"
      >
        {{ getInitials(props.restaurant.name) }}
      </div>

      <div class="absolute inset-x-0 top-0 flex items-start justify-between px-3 pt-3">
        <div class="flex flex-col items-start gap-1.5">
          <span v-if="isClosed" class="badge badge-sm badge-info backdrop-blur">Heute geschlossen</span>
          <span
            v-else-if="menuFreshness"
            :class="['badge badge-sm', menuFreshness.className]"
            :title="menuFreshness.title"
            >{{ menuFreshness.label }}</span
          >
          <span v-if="showDistance" class="badge badge-sm badge-neutral/85 backdrop-blur w-fit">{{ distanceLabel }}</span>
        </div>
        <button
          type="button"
          :class="[
            'cursor-pointer border-0 bg-transparent p-0 text-lg leading-none drop-shadow-sm transition-colors opacity-80 hover:opacity-100 focus:outline-none',
            isFavorited ? 'text-warning' : 'text-white hover:text-warning',
          ]"
          :aria-label="isFavorited ? 'Favorit entfernen' : 'Als Favorit markieren'"
          :aria-pressed="isFavorited"
          @click="toggleFavorite(props.restaurant.id)"
        >
          <Fa7SolidStar aria-hidden="true" />
        </button>
      </div>
      <!-- bottom row: tags -->
      <div class="absolute inset-x-0 bottom-0 bg-linear-to-t from-base-100/90 to-transparent px-3 pb-3 pt-8">
        <div class="flex flex-wrap gap-1.5">
          <button
            v-for="tag in props.restaurant.tags"
            :key="tag"
            type="button"
            class="badge badge-outline badge-xs cursor-pointer border-base-300/60 bg-base-100/70 px-2 py-2.5 text-xs font-medium backdrop-blur transition-colors hover:bg-base-100"
            :title="`Nach Tag ${tag} filtern`"
            @click="applySearch(tag)"
          >
            {{ tag }}
          </button>
        </div>
      </div>
    </figure>

    <div class="card-body gap-3 p-3">
      <div>
        <h3 class="text-base font-semibold leading-tight">
          {{ props.restaurant.name }}
        </h3>

        <p v-if="lastCheck" class="text-xs text-base-content/65" :title="lastCheckTitle" aria-label="Letzter Pruefstatus">
          {{ lastCheckText }}
        </p>
        <p class="text-xs text-base-content/65" v-else>Noch nicht geprüft</p>
      </div>

      <RestaurantActions :restaurant="props.restaurant" />
    </div>
  </article>
</template>
