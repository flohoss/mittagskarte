import { formatAgeLabel } from './date';

export type LastCheckStatus = 'success' | 'not_changed' | 'error';

export type LastCheck = {
  at: string;
  status: LastCheckStatus;
  detail?: string;
};

export type MenuFreshnessMeta = {
  className: string;
  label: string;
  isCurrent: boolean;
};

const VISUAL_CURRENT = { className: 'badge-success', prefix: 'aktuell' };
const VISUAL_OUTDATED = { className: 'badge-warning', prefix: 'veraltet' };
const VISUAL_ERROR = { className: 'badge-error', prefix: 'Fehler' };

const MS_PER_DAY = 24 * 60 * 60 * 1000;

export function toTimestamp(value: string | null | undefined): number | null {
  if (!value) return null;
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? null : timestamp;
}

function parseIsoDate(value: string): { year: number; month: number; day: number; weekday: number } | null {
  const match = value.match(/^(\d{4})-(\d{2})-(\d{2})/);
  if (!match) return null;

  const year = parseInt(match[1], 10);
  const month = parseInt(match[2], 10);
  const day = parseInt(match[3], 10);
  if (Number.isNaN(year) || Number.isNaN(month) || Number.isNaN(day)) return null;

  const weekday = new Date(year, month - 1, day).getDay();
  return { year, month, day, weekday };
}

function daysBetweenDates(start: { year: number; month: number; day: number }, end: { year: number; month: number; day: number }): number {
  const startMs = Date.UTC(start.year, start.month - 1, start.day);
  const endMs = Date.UTC(end.year, end.month - 1, end.day);
  return Math.floor((endMs - startMs) / MS_PER_DAY);
}

export type MenuValidity = 'daily' | 'weekly' | 'monthly' | 'unknown';

function parseCronRangePart(part: string, min: number, max: number): number[] | null {
  if (!part || part === '*') return null;

  const values = new Set<number>();
  for (const segment of part.split(',')) {
    if (segment.includes('-')) {
      const [start, end] = segment.split('-').map((v) => parseInt(v, 10));
      if (Number.isNaN(start) || Number.isNaN(end)) return null;
      for (let v = Math.max(min, start); v <= Math.min(max, end); v += 1) {
        values.add(v);
      }
    } else {
      const value = parseInt(segment, 10);
      if (!Number.isNaN(value) && value >= min && value <= max) {
        values.add(value);
      }
    }
  }

  return values.size > 0 ? [...values].sort((a, b) => a - b) : null;
}

function cronWeekdays(cron: string | null | undefined): number[] | null {
  if (!cron || !cron.trim()) return null;
  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return null;
  return parseCronRangePart(parts[4], 0, 6);
}

function coversAllWeekdays(cron: string | null | undefined): boolean {
  const days = cronWeekdays(cron);
  if (!days) return false;
  const weekdaySet = new Set(days);
  return [1, 2, 3, 4, 5].every((d) => weekdaySet.has(d));
}

export function cronValidity(cron: string | null | undefined): MenuValidity {
  if (!cron || !cron.trim()) return 'unknown';

  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return 'unknown';

  const [, , dayOfMonth, month, dayOfWeek] = parts;

  const isEveryDayOfMonth = dayOfMonth === '*';
  const isEveryMonth = month === '*';
  const isEveryDayOfWeek = dayOfWeek === '*';

  if (isEveryDayOfMonth && isEveryMonth && (isEveryDayOfWeek || coversAllWeekdays(cron))) {
    return 'daily';
  }

  if (!isEveryDayOfMonth && isEveryMonth) {
    return 'monthly';
  }

  if (isEveryDayOfMonth && isEveryMonth && !isEveryDayOfWeek) {
    return 'weekly';
  }

  return 'unknown';
}

function daysUntilSaturday(fromWeekday: number): number {
  // Saturday = 6. A menu created on any weekday is valid until Saturday 00:00.
  // If created on Saturday/Sunday, treat it as valid for the upcoming week.
  return fromWeekday >= 6 ? 7 : 6 - fromWeekday;
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

function maxMenuAgeDays(cron: string | null | undefined, menuDate: { year: number; month: number; day: number; weekday: number }): number | null {
  const validity = cronValidity(cron);

  if (validity === 'daily') {
    return 1;
  }

  if (validity === 'weekly') {
    // Weekly menus are valid for the calendar week they were created in,
    // regardless of which weekdays the cron refreshes on.
    return daysUntilSaturday(menuDate.weekday);
  }

  if (validity === 'monthly') {
    // Monthly menus are valid until the first day of the next month.
    return daysInMonth(menuDate.year, menuDate.month - 1) - menuDate.day + 1;
  }

  return null;
}

function parseNowDate(now: number): { year: number; month: number; day: number } | null {
  const d = new Date(now);
  return {
    year: d.getUTCFullYear(),
    month: d.getUTCMonth() + 1,
    day: d.getUTCDate(),
  };
}

export function isMenuCurrent(input: { menuDate: string | null; lastCheck: LastCheck | null; cron: string | null; now: number }): boolean {
  const { menuDate, lastCheck, cron, now } = input;

  const menuTs = toTimestamp(menuDate);
  const checkTs = toTimestamp(lastCheck?.at);
  const checkSucceeded = lastCheck?.status === 'success' || lastCheck?.status === 'not_changed';

  if (lastCheck?.status === 'error') {
    return false;
  }

  if (menuDate) {
    const parsedMenuDate = parseIsoDate(menuDate);
    const parsedNowDate = parseNowDate(now);

    if (parsedMenuDate && parsedNowDate) {
      const maxAge = maxMenuAgeDays(cron, parsedMenuDate);
      if (maxAge !== null) {
        return daysBetweenDates(parsedMenuDate, parsedNowDate) < maxAge;
      }
    }
  }

  return menuTs !== null || (checkSucceeded && checkTs !== null);
}

export function getMenuFreshnessMeta(input: { menuDate: string; cron: string; lastCheck: LastCheck | null; now: number }): MenuFreshnessMeta {
  const { menuDate, cron, lastCheck, now } = input;
  const menuTs = toTimestamp(menuDate);

  if (menuTs === null) {
    return {
      className: 'badge-neutral',
      label: 'Unbekannt',
      isCurrent: false,
    };
  }

  const relative = formatAgeLabel(menuDate, now);

  if (lastCheck?.status === 'error') {
    return {
      className: VISUAL_ERROR.className,
      label: `${VISUAL_ERROR.prefix} • ${relative}`,
      isCurrent: false,
    };
  }

  const current = isMenuCurrent({ menuDate, lastCheck, cron, now });

  const visual = current ? VISUAL_CURRENT : VISUAL_OUTDATED;
  return {
    className: visual.className,
    label: `${visual.prefix} • ${relative}`,
    isCurrent: current,
  };
}
