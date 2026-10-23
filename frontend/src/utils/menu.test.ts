import { describe, expect, it } from 'vitest';
import type { MenuRecord } from '../models/restaurant';
import { getLatestMenu, getMenuDimensions, sortMenusByCreatedDesc } from './menu';

function menu(id: string, created: string, dimensions?: MenuRecord['dimensions']): MenuRecord {
  return { id, file: `${id}.webp`, hash: id, created, dimensions } as unknown as MenuRecord;
}

describe('sortMenusByCreatedDesc', () => {
  it('returns an empty array for missing input', () => {
    expect(sortMenusByCreatedDesc()).toEqual([]);
    expect(sortMenusByCreatedDesc([])).toEqual([]);
  });

  it('sorts menus newest first', () => {
    const menus = [menu('a', '2026-06-15T10:00:00Z'), menu('b', '2026-06-17T10:00:00Z'), menu('c', '2026-06-16T10:00:00Z')];

    expect(sortMenusByCreatedDesc(menus).map((entry) => entry.id)).toEqual(['b', 'c', 'a']);
  });

  it('does not mutate the input array', () => {
    const menus = [menu('a', '2026-06-15T10:00:00Z'), menu('b', '2026-06-17T10:00:00Z')];
    const original = menus.map((entry) => entry.id);

    sortMenusByCreatedDesc(menus);

    expect(menus.map((entry) => entry.id)).toEqual(original);
  });
});

describe('getLatestMenu', () => {
  it('returns null for missing or empty menus', () => {
    expect(getLatestMenu(null)).toBeNull();
    expect(getLatestMenu(undefined)).toBeNull();
    expect(getLatestMenu([])).toBeNull();
  });

  it('returns the menu with the most recent creation date', () => {
    const menus = [menu('old', '2026-06-01T10:00:00Z'), menu('new', '2026-06-20T10:00:00Z'), menu('mid', '2026-06-10T10:00:00Z')];

    expect(getLatestMenu(menus)?.id).toBe('new');
  });

  it('returns the single menu when only one is present', () => {
    expect(getLatestMenu([menu('only', '2026-06-01T10:00:00Z')])?.id).toBe('only');
  });
});

describe('getMenuDimensions', () => {
  it('returns null dimensions for missing menus', () => {
    expect(getMenuDimensions(null)).toEqual({ width: null, height: null });
    expect(getMenuDimensions(undefined)).toEqual({ width: null, height: null });
  });

  it('returns null dimensions when dimensions are absent or not an object', () => {
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z'))).toEqual({ width: null, height: null });
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', null))).toEqual({ width: null, height: null });

    const withString = { dimensions: 'large' } as unknown as MenuRecord;
    expect(getMenuDimensions(withString)).toEqual({ width: null, height: null });
  });

  it('rejects non-positive and non-finite values', () => {
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', { width: 0, height: -5 }))).toEqual({ width: null, height: null });
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', { width: Number.NaN, height: Number.POSITIVE_INFINITY }))).toEqual({
      width: null,
      height: null,
    });

    const withStrings = { dimensions: { width: '800', height: '600' } } as unknown as MenuRecord;
    expect(getMenuDimensions(withStrings)).toEqual({ width: null, height: null });
  });

  it('passes valid dimensions through', () => {
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', { width: 1920, height: 1080 }))).toEqual({ width: 1920, height: 1080 });
  });

  it('keeps valid values when only one dimension is invalid', () => {
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', { width: 1920, height: 0 }))).toEqual({ width: 1920, height: null });
    expect(getMenuDimensions(menu('a', '2026-06-01T10:00:00Z', { width: -1, height: 1080 }))).toEqual({ width: null, height: 1080 });
  });
});
