import type { MenuRecord } from '../models/restaurant';

export type MenuDimensions = {
  width: number | null;
  height: number | null;
};

export function sortMenusByCreatedDesc(menus: MenuRecord[] = []) {
  return [...menus].sort((a, b) => (a.created > b.created ? -1 : 1));
}

export function getLatestMenu(menus: MenuRecord[] | null | undefined) {
  if (!menus?.length) return null;
  return sortMenusByCreatedDesc(menus)[0];
}

export function getMenuDimensions(menu: MenuRecord | null | undefined): MenuDimensions {
  const raw = menu?.dimensions;

  if (!raw || typeof raw !== 'object') {
    return { width: null, height: null };
  }

  const width = typeof raw.width === 'number' && Number.isFinite(raw.width) && raw.width > 0 ? raw.width : null;
  const height = typeof raw.height === 'number' && Number.isFinite(raw.height) && raw.height > 0 ? raw.height : null;

  return { width, height };
}
