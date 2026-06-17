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

export function formatAgeLabel(value: string, nowMs: number) {
  const diffSeconds = clampedDiffSeconds(value, nowMs);
  if (diffSeconds === null) return 'Unbekannt';

  if (diffSeconds > -60) return 'gerade eben';

  const absSeconds = Math.abs(diffSeconds);

  for (const [unit, secondsInUnit] of RELATIVE_TIME_UNITS) {
    if (absSeconds >= secondsInUnit || unit === 'minute') {
      return relativeTimeFormatter.format(Math.round(diffSeconds / secondsInUnit), unit);
    }
  }

  return 'gerade eben';
}

export function formatRelativePastLabel(value: string, nowMs: number) {
  const diffSeconds = clampedDiffSeconds(value, nowMs);
  if (diffSeconds === null) return 'Unbekannt';

  if (diffSeconds > -60) return 'gerade eben';

  const absSeconds = Math.abs(diffSeconds);

  for (const [unit, secondsInUnit] of RELATIVE_TIME_UNITS) {
    if (absSeconds >= secondsInUnit || unit === 'minute') {
      return relativeTimeFormatter.format(Math.round(diffSeconds / secondsInUnit), unit);
    }
  }

  return 'gerade eben';
}
