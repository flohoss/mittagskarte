import { describe, expect, it } from 'vitest';
import { formatAgeLabel, formatRelativePastLabel } from './date';

function now(): number {
  return new Date('2026-06-17T12:00:00').getTime();
}

describe('formatAgeLabel', () => {
  it('returns Unbekannt for invalid dates', () => {
    expect(formatAgeLabel('not-a-date', now())).toBe('Unbekannt');
  });

  it('returns gerade eben for times within the last minute', () => {
    expect(formatAgeLabel('2026-06-17T11:59:30', now())).toBe('gerade eben');
  });

  it('returns X Minuten alt for times within the last hour', () => {
    expect(formatAgeLabel('2026-06-17T11:30:00', now())).toBe('30 Minuten alt');
    expect(formatAgeLabel('2026-06-17T11:45:00', now())).toBe('15 Minuten alt');
  });

  it('returns X Stunden alt for times within the last day', () => {
    expect(formatAgeLabel('2026-06-17T08:00:00', now())).toBe('4 Stunden alt');
    expect(formatAgeLabel('2026-06-16T13:00:00', now())).toBe('23 Stunden alt');
  });

  it('returns X Tage alt for times within the last week', () => {
    expect(formatAgeLabel('2026-06-15T12:00:00', now())).toBe('2 Tage alt');
    expect(formatAgeLabel('2026-06-11T12:00:00', now())).toBe('6 Tage alt');
  });

  it('returns X Wochen alt for times within the last month', () => {
    expect(formatAgeLabel('2026-06-10T12:00:00', now())).toBe('1 Woche alt');
    expect(formatAgeLabel('2026-06-03T12:00:00', now())).toBe('2 Wochen alt');
  });

  it('returns X Monate alt for times within the last year', () => {
    expect(formatAgeLabel('2026-05-17T12:00:00', now())).toBe('1 Monat alt');
    expect(formatAgeLabel('2026-04-17T12:00:00', now())).toBe('2 Monate alt');
  });

  it('returns X Jahre alt for times over a year ago', () => {
    expect(formatAgeLabel('2025-06-17T12:00:00', now())).toBe('1 Jahr alt');
    expect(formatAgeLabel('2024-06-17T12:00:00', now())).toBe('2 Jahre alt');
  });

  it('clamps future dates to gerade eben', () => {
    expect(formatAgeLabel('2026-06-17T12:00:01', now())).toBe('gerade eben');
  });
});

describe('formatRelativePastLabel', () => {
  it('returns Unbekannt for invalid dates', () => {
    expect(formatRelativePastLabel('not-a-date', now())).toBe('Unbekannt');
  });

  it('returns gerade eben for times within the last minute', () => {
    expect(formatRelativePastLabel('2026-06-17T11:59:30', now())).toBe('gerade eben');
  });

  it('returns vor X Minuten for times within the last hour', () => {
    expect(formatRelativePastLabel('2026-06-17T11:30:00', now())).toBe('vor 30 Minuten');
  });

  it('returns vor X Stunden for times within the last day', () => {
    expect(formatRelativePastLabel('2026-06-17T08:00:00', now())).toBe('vor 4 Stunden');
  });

  it('returns vor X Tagen for times within the last week', () => {
    expect(formatRelativePastLabel('2026-06-15T12:00:00', now())).toBe('vorgestern');
    expect(formatRelativePastLabel('2026-06-11T12:00:00', now())).toBe('vor 6 Tagen');
  });

  it('returns vor X Wochen for times within the last month', () => {
    expect(formatRelativePastLabel('2026-06-03T12:00:00', now())).toBe('vor 2 Wochen');
  });

  it('returns vor X Monaten for times within the last year', () => {
    expect(formatRelativePastLabel('2026-04-17T12:00:00', now())).toBe('vor 2 Monaten');
  });

  it('returns vor X Jahren for times over a year ago', () => {
    expect(formatRelativePastLabel('2025-06-17T12:00:00', now())).toBe('letztes Jahr');
    expect(formatRelativePastLabel('2024-06-17T12:00:00', now())).toBe('vor 2 Jahren');
  });

  it('clamps future dates to gerade eben', () => {
    expect(formatRelativePastLabel('2026-06-17T12:00:01', now())).toBe('gerade eben');
  });
});
