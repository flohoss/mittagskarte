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

function startOfDayMs(ts: number): number {
  const d = new Date(ts);
  d.setHours(0, 0, 0, 0);
  return d.getTime();
}

function daysBetween(startTs: number, endTs: number): number {
  return Math.floor((startOfDayMs(endTs) - startOfDayMs(startTs)) / MS_PER_DAY);
}

export type MenuValidity = 'daily' | 'weekly' | 'monthly' | 'unknown';

export function cronValidity(cron: string | null | undefined): MenuValidity {
  if (!cron || !cron.trim()) return 'unknown';

  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return 'unknown';

  const [, , dayOfMonth, month, dayOfWeek] = parts;

  const isEveryDayOfMonth = dayOfMonth === '*';
  const isEveryMonth = month === '*';
  const isEveryDayOfWeek = dayOfWeek === '*';

  if (isEveryDayOfMonth && isEveryMonth && isEveryDayOfWeek) {
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

function cronWeekdays(cron: string | null | undefined): number[] | null {
  if (!cron || !cron.trim()) return null;

  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return null;

  const dayOfWeek = parts[4];
  if (dayOfWeek === '*') return null;

  return dayOfWeek
    .split(',')
    .map((d) => parseInt(d, 10))
    .filter((d) => !Number.isNaN(d) && d >= 0 && d <= 6);
}

function cronDaysOfMonth(cron: string | null | undefined): number[] | null {
  if (!cron || !cron.trim()) return null;

  const parts = cron.trim().split(/\s+/);
  if (parts.length < 5) return null;

  const dayOfMonth = parts[2];
  if (dayOfMonth === '*') return null;

  return dayOfMonth
    .split(',')
    .map((d) => parseInt(d, 10))
    .filter((d) => !Number.isNaN(d) && d >= 1 && d <= 31);
}

function weekdayOf(ts: number): number {
  return new Date(ts).getDay();
}

function daysUntilNextWeekday(fromWeekday: number, targetWeekdays: number[]): number {
  const sorted = [...targetWeekdays].sort((a, b) => a - b);
  for (let offset = 1; offset <= 7; offset += 1) {
    const weekday = (fromWeekday + offset) % 7;
    if (sorted.includes(weekday)) return offset;
  }
  return 7;
}

function daysInMonth(year: number, month: number): number {
  return new Date(year, month + 1, 0).getDate();
}

function maxMenuAgeDays(cron: string | null | undefined, menuTs: number): number | null {
  const validity = cronValidity(cron);

  if (validity === 'daily') {
    return 1;
  }

  if (validity === 'weekly') {
    const weekdays = cronWeekdays(cron);
    if (!weekdays || weekdays.length === 0) return 7;
    const menuWeekday = weekdayOf(menuTs);
    return daysUntilNextWeekday(menuWeekday, weekdays);
  }

  if (validity === 'monthly') {
    const days = cronDaysOfMonth(cron);
    if (!days || days.length === 0) return 30;

    const menuDate = new Date(menuTs);
    const menuYear = menuDate.getFullYear();
    const menuMonth = menuDate.getMonth();
    const menuDay = menuDate.getDate();
    const sorted = [...days].sort((a, b) => a - b);

    const nextDayInMonth = sorted.find((d) => d > menuDay);
    if (nextDayInMonth !== undefined) {
      return nextDayInMonth - menuDay;
    }

    const nextMonthLength = daysInMonth(menuYear, menuMonth + 1);
    const firstTargetNextMonth = sorted[0];
    const rolloverDay = Math.min(firstTargetNextMonth, nextMonthLength);
    return daysInMonth(menuYear, menuMonth) - menuDay + rolloverDay;
  }

  return null;
}

export function isMenuCurrent(input: { menuDate: string | null; lastCheck: LastCheck | null; cron: string | null; now: number }): boolean {
  const { menuDate, lastCheck, cron, now } = input;

  const menuTs = toTimestamp(menuDate);
  const checkTs = toTimestamp(lastCheck?.at);
  const checkSucceeded = lastCheck?.status === 'success' || lastCheck?.status === 'not_changed';

  if (lastCheck?.status === 'error') {
    return false;
  }

  if (menuTs !== null) {
    const maxAge = maxMenuAgeDays(cron, menuTs);
    if (maxAge !== null) {
      return daysBetween(menuTs, now) < maxAge;
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
