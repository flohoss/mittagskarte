import { describe, it, expect } from 'vitest';
import { isOnHoliday, formatHolidayLabel } from './holiday';

function ts(date: string): number {
  return new Date(date).getTime();
}

describe('isOnHoliday', () => {
  it('returns false when holiday_until is empty', () => {
    expect(isOnHoliday('', ts('2026-08-14T12:00:00'))).toBe(false);
  });

  it('returns false when holiday_until is null or undefined', () => {
    expect(isOnHoliday(null, ts('2026-08-14T12:00:00'))).toBe(false);
    expect(isOnHoliday(undefined, ts('2026-08-14T12:00:00'))).toBe(false);
  });

  it('returns true when today is before holiday_until', () => {
    expect(isOnHoliday('2026-08-28', ts('2026-08-14T12:00:00'))).toBe(true);
  });

  it('returns true when today is exactly holiday_until', () => {
    expect(isOnHoliday('2026-08-14', ts('2026-08-14T23:59:00'))).toBe(true);
  });

  it('returns false when today is after holiday_until', () => {
    expect(isOnHoliday('2026-08-10', ts('2026-08-14T12:00:00'))).toBe(false);
  });

  it('ignores time portion in holiday_until', () => {
    expect(isOnHoliday('2026-08-14T00:00:00.000Z', ts('2026-08-14T12:00:00'))).toBe(true);
  });
});

describe('formatHolidayLabel', () => {
  it('returns empty string when holiday_until is empty', () => {
    expect(formatHolidayLabel('')).toBe('');
  });

  it('returns empty string when holiday_until is null or undefined', () => {
    expect(formatHolidayLabel(null)).toBe('');
    expect(formatHolidayLabel(undefined)).toBe('');
  });

  it('returns German-formatted label', () => {
    expect(formatHolidayLabel('2026-08-23')).toBe('Urlaub bis 23.8.2026');
  });

  it('ignores time portion in holiday_until', () => {
    expect(formatHolidayLabel('2026-08-23T00:00:00.000Z')).toBe('Urlaub bis 23.8.2026');
  });

  it('returns empty string for invalid date', () => {
    expect(formatHolidayLabel('not-a-date')).toBe('');
  });
});
