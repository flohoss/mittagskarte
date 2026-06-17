import { describe, expect, it } from 'vitest';
import { cronValidity, isMenuCurrent, type LastCheck } from './menuFreshness';

type TestCase = {
  name: string;
  cron: string;
  menuDate: string;
  now: string;
  expected: boolean;
};

function makeCheck(at: string, status: LastCheck['status']): LastCheck {
  return { at, status };
}

function ts(date: string): number {
  // Use local-time strings in tests so day boundaries are consistent in any timezone.
  return new Date(date).getTime();
}

describe('cronValidity', () => {
  it.each([
    // daily: every day or every weekday
    { cron: '0 0 * * *', expected: 'daily' },
    { cron: '30 10,11 * * 1-5', expected: 'daily' },
    { cron: '0 0 * * 1-5', expected: 'daily' },
    { cron: '0 0 * * 0-6', expected: 'daily' },
    { cron: '0 0 * * 1,2,3,4,5', expected: 'daily' },
    // weekly: sparse weekday refreshes
    { cron: '30 10,11 * * 1,2', expected: 'weekly' },
    { cron: '30 10,11 * * 1,3', expected: 'weekly' },
    { cron: '30 10,11 * * 2,3', expected: 'weekly' },
    { cron: '30 10,11 * * 1,4', expected: 'weekly' },
    { cron: '0 0 * * 1,2', expected: 'weekly' },
    { cron: '0 0 * * 3', expected: 'weekly' },
    { cron: '0 0 * * 5,6', expected: 'weekly' },
    // monthly: specific day-of-month refreshes
    { cron: '30 10,11 1-3 * *', expected: 'monthly' },
    { cron: '30 10,11 1,15 * *', expected: 'monthly' },
    { cron: '0 0 1 * *', expected: 'monthly' },
    { cron: '0 0 1,15 * *', expected: 'monthly' },
    { cron: '0 0 1-7 * *', expected: 'monthly' },
    // unknown / missing
    { cron: '', expected: 'unknown' },
    { cron: 'invalid', expected: 'unknown' },
    { cron: '0 0', expected: 'unknown' },
  ])('classifies $cron as $expected', ({ cron, expected }) => {
    expect(cronValidity(cron)).toBe(expected);
  });
});

describe('isMenuCurrent', () => {
  const cases: TestCase[] = [
    // daily: valid only the same calendar day
    {
      name: 'daily menu from today is current',
      cron: '30 10,11 * * 1-5',
      menuDate: '2026-06-17T10:30:00+02:00',
      now: '2026-06-17T12:00:00+02:00',
      expected: true,
    },
    {
      name: 'daily menu from yesterday is old',
      cron: '30 10,11 * * 1-5',
      menuDate: '2026-06-16T10:30:00+02:00',
      now: '2026-06-17T09:00:00+02:00',
      expected: false,
    },
    {
      name: 'daily every-day cron menu from yesterday is old',
      cron: '0 0 * * *',
      menuDate: '2026-06-16T10:30:00',
      now: '2026-06-17T09:00:00',
      expected: false,
    },
    {
      name: 'daily every-day cron menu from today is current',
      cron: '0 0 * * *',
      menuDate: '2026-06-17T10:30:00',
      now: '2026-06-17T23:59:00',
      expected: true,
    },

    // weekly: valid until Saturday 00:00
    {
      name: 'weekly menu created Monday is current on Wednesday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-15T10:30:00+02:00',
      now: '2026-06-17T12:00:00+02:00',
      expected: true,
    },
    {
      name: 'weekly menu created Monday is current on Friday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-15T10:30:00+02:00',
      now: '2026-06-19T23:59:00+02:00',
      expected: true,
    },
    {
      name: 'weekly menu created Monday is old on Saturday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-15T10:30:00',
      now: '2026-06-20T00:01:00',
      expected: false,
    },
    {
      name: 'weekly menu created Tuesday is current on Wednesday',
      cron: '30 10,11 * * 2,3',
      menuDate: '2026-06-16T10:30:00',
      now: '2026-06-17T12:00:00',
      expected: true,
    },
    {
      name: 'weekly menu created Tuesday is old on Saturday',
      cron: '30 10,11 * * 2,3',
      menuDate: '2026-06-16T10:30:00',
      now: '2026-06-20T00:01:00',
      expected: false,
    },
    {
      name: 'weekly menu created Monday with Mon/Wed cron is current on Wednesday',
      cron: '30 10,11 * * 1,3',
      menuDate: '2026-06-15T10:30:00',
      now: '2026-06-17T12:00:00',
      expected: true,
    },
    {
      name: 'weekly menu created Thursday with Mon/Thu cron is current on Friday',
      cron: '30 10,11 * * 1,4',
      menuDate: '2026-06-18T10:30:00',
      now: '2026-06-19T23:59:00',
      expected: true,
    },
    {
      name: 'weekly menu created Thursday with Mon/Thu cron is old on Saturday',
      cron: '30 10,11 * * 1,4',
      menuDate: '2026-06-18T10:30:00',
      now: '2026-06-20T00:01:00',
      expected: false,
    },
    {
      name: 'weekly menu created Friday is current on Friday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-19T10:30:00',
      now: '2026-06-19T23:59:00',
      expected: true,
    },
    {
      name: 'weekly menu created Friday is old on Saturday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-19T10:30:00',
      now: '2026-06-20T00:01:00',
      expected: false,
    },
    {
      name: 'weekly menu created Saturday is current for upcoming week',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-20T10:30:00',
      now: '2026-06-21T12:00:00',
      expected: true,
    },
    {
      name: 'weekly menu created Saturday is old next Saturday',
      cron: '30 10,11 * * 1,2',
      menuDate: '2026-06-20T10:30:00',
      now: '2026-06-27T00:01:00',
      expected: false,
    },
    {
      name: 'single weekday cron menu created Wednesday is current on Friday',
      cron: '0 0 * * 3',
      menuDate: '2026-06-17T10:30:00',
      now: '2026-06-19T12:00:00',
      expected: true,
    },

    // monthly: valid until the first day of the next month
    {
      name: 'monthly menu created on 1st is current later in month',
      cron: '30 10,11 1-3 * *',
      menuDate: '2026-06-01T10:30:00',
      now: '2026-06-15T12:00:00',
      expected: true,
    },
    {
      name: 'monthly menu created on 1st is old on 1st of next month',
      cron: '30 10,11 1-3 * *',
      menuDate: '2026-06-01T10:30:00',
      now: '2026-07-01T00:01:00',
      expected: false,
    },
    {
      name: 'monthly menu created on 3rd is current at end of month',
      cron: '30 10,11 1-3 * *',
      menuDate: '2026-06-03T10:30:00',
      now: '2026-06-30T23:59:00',
      expected: true,
    },
    {
      name: 'monthly menu created on 3rd is old on 1st of next month',
      cron: '30 10,11 1-3 * *',
      menuDate: '2026-06-03T10:30:00',
      now: '2026-07-01T00:01:00',
      expected: false,
    },
    {
      name: 'monthly menu created on 15th is current at end of month',
      cron: '30 10,11 1,15 * *',
      menuDate: '2026-06-15T10:30:00',
      now: '2026-06-30T23:59:00',
      expected: true,
    },
    {
      name: 'monthly menu created on 15th is old on 1st of next month',
      cron: '30 10,11 1,15 * *',
      menuDate: '2026-06-15T10:30:00',
      now: '2026-07-01T00:01:00',
      expected: false,
    },
    {
      name: 'monthly menu created on 1st is old on 1st of next month with single-day cron',
      cron: '0 0 1 * *',
      menuDate: '2026-06-01T10:30:00',
      now: '2026-07-01T00:01:00',
      expected: false,
    },
    {
      name: 'monthly menu handles February rollover',
      cron: '0 0 1 * *',
      menuDate: '2026-02-01T10:30:00',
      now: '2026-02-28T23:59:00',
      expected: true,
    },
    {
      name: 'monthly menu becomes old on 1st of March',
      cron: '0 0 1 * *',
      menuDate: '2026-02-01T10:30:00',
      now: '2026-03-01T00:01:00',
      expected: false,
    },
  ];

  it.each(cases)('$name', ({ cron, menuDate, now, expected }) => {
    expect(
      isMenuCurrent({
        menuDate,
        cron,
        lastCheck: null,
        now: ts(now),
      })
    ).toBe(expected);
  });

  it('returns false when last check failed', () => {
    expect(
      isMenuCurrent({
        menuDate: '2026-06-17T10:30:00+02:00',
        cron: '30 10,11 * * 1-5',
        lastCheck: makeCheck('2026-06-17T10:30:00+02:00', 'error'),
        now: ts('2026-06-17T12:00:00+02:00'),
      })
    ).toBe(false);
  });

  it('falls back to menu existence when cron is missing', () => {
    expect(
      isMenuCurrent({
        menuDate: '2026-06-10T10:30:00+02:00',
        cron: '',
        lastCheck: null,
        now: ts('2026-06-17T12:00:00+02:00'),
      })
    ).toBe(true);
  });
});
