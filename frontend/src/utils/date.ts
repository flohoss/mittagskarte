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

function clampedDiffSeconds(value: string, nowMs: number): number | null {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return null;

  const rawDiffSeconds = Math.round((date.getTime() - nowMs) / 1000);
  // Backend and client clocks can drift slightly; avoid future times in the UI.
  return Math.min(0, rawDiffSeconds);
}

function formatRelativeTime(diffSeconds: number): string {
  if (diffSeconds > -60) return 'gerade eben';

  const absSeconds = Math.abs(diffSeconds);

  for (const [unit, secondsInUnit] of RELATIVE_TIME_UNITS) {
    if (absSeconds >= secondsInUnit || unit === 'minute') {
      return relativeTimeFormatter.format(Math.round(diffSeconds / secondsInUnit), unit);
    }
  }

  return 'gerade eben';
}

export function formatAgeLabel(value: string, nowMs: number) {
  const diffSeconds = clampedDiffSeconds(value, nowMs);
  if (diffSeconds === null) return 'Unbekannt';

  if (diffSeconds > -60) return 'gerade eben';

  const absSeconds = Math.abs(diffSeconds);

  if (absSeconds < 60 * 60) {
    const minutes = Math.round(absSeconds / 60);
    return minutes === 1 ? '1 Minute alt' : `${minutes} Minuten alt`;
  }

  if (absSeconds < 60 * 60 * 24) {
    const hours = Math.round(absSeconds / (60 * 60));
    return hours === 1 ? '1 Stunde alt' : `${hours} Stunden alt`;
  }

  const days = Math.floor(absSeconds / (60 * 60 * 24));

  if (days >= 365) {
    const years = Math.round(days / 365);
    return years === 1 ? '1 Jahr alt' : `${years} Jahre alt`;
  }

  if (days >= 30) {
    const months = Math.round(days / 30);
    return months === 1 ? '1 Monat alt' : `${months} Monate alt`;
  }

  if (days >= 7) {
    const weeks = Math.round(days / 7);
    return weeks === 1 ? '1 Woche alt' : `${weeks} Wochen alt`;
  }

  return days === 1 ? '1 Tag alt' : `${days} Tage alt`;
}

export function formatRelativePastLabel(value: string, nowMs: number) {
  const diffSeconds = clampedDiffSeconds(value, nowMs);
  if (diffSeconds === null) return 'Unbekannt';

  return formatRelativeTime(diffSeconds);
}
