const DATE_ONLY_LENGTH = 10;

export function isOnHoliday(holidayUntil: string | null | undefined, now: number): boolean {
  if (!holidayUntil) return false;
  const until = holidayUntil.slice(0, DATE_ONLY_LENGTH);
  if (!until) return false;
  const today = new Date(now).toISOString().slice(0, DATE_ONLY_LENGTH);
  return today <= until;
}

export function formatHolidayLabel(holidayUntil: string | null | undefined): string {
  if (!holidayUntil) return '';
  const date = new Date(holidayUntil.slice(0, DATE_ONLY_LENGTH) + 'T00:00:00');
  if (Number.isNaN(date.getTime())) return '';
  return `Urlaub bis ${date.toLocaleDateString('de-DE')}`;
}
