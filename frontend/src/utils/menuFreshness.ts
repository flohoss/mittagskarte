import { CronExpressionParser } from 'cron-parser';
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

export function toTimestamp(value: string | null | undefined): number | null {
  if (!value) return null;
  const timestamp = new Date(value).getTime();
  return Number.isNaN(timestamp) ? null : timestamp;
}

export function getPreviousCronRunMs(cron: string | null | undefined, now: number): number | null {
  if (!cron || !cron.trim()) return null;

  try {
    const expression = CronExpressionParser.parse(cron, { currentDate: new Date(now) });
    if (!expression.hasPrev()) return null;
    return expression.prev().getTime();
  } catch {
    return null;
  }
}

export function isMenuCurrent(input: { menuDate: string | null; lastCheck: LastCheck | null; cron: string | null; now: number }): boolean {
  const { menuDate, lastCheck, cron, now } = input;

  const menuTs = toTimestamp(menuDate);
  const checkTs = toTimestamp(lastCheck?.at);
  const checkSucceeded = lastCheck?.status === 'success' || lastCheck?.status === 'not_changed';
  const runMs = getPreviousCronRunMs(cron, now);

  if (runMs === null) {
    return menuTs !== null || (checkSucceeded && checkTs !== null);
  }

  return (menuTs !== null && menuTs >= runMs) || (checkSucceeded && checkTs !== null && checkTs >= runMs);
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
      className: 'badge-error',
      label: `Fehler • ${relative}`,
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
