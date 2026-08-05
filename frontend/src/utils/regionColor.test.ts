import { describe, it, expect } from 'vitest';
import { getRegionBadgeColors } from './regionColor';

describe('getRegionBadgeColors', () => {
  it('returns an hsl background and a hex text color', () => {
    const light = getRegionBadgeColors('Stuttgart-Mitte', false);
    expect(light.backgroundColor).toMatch(/^hsl\(/);
    expect(light.color).toMatch(/^#[0-9a-f]{6}$/);
  });

  it('produces different colors for different regions', () => {
    const a = getRegionBadgeColors('Stuttgart-Mitte', false);
    const b = getRegionBadgeColors('Böblingen', false);
    expect(a.backgroundColor).not.toBe(b.backgroundColor);
  });

  it('is stable for the same region', () => {
    const a = getRegionBadgeColors('Stuttgart-Mitte', false);
    const b = getRegionBadgeColors('Stuttgart-Mitte', false);
    expect(a).toEqual(b);
  });

  it('uses dark text in light mode', () => {
    const light = getRegionBadgeColors('Stuttgart-Mitte', false);
    expect(light.color).toBe('#1f2937');
  });

  it('uses light text in dark mode', () => {
    const dark = getRegionBadgeColors('Stuttgart-Mitte', true);
    expect(dark.color).toBe('#fafafa');
  });

  it('never produces a blue-only hue overlapping the primary theme color', () => {
    for (const region of ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']) {
      const { backgroundColor } = getRegionBadgeColors(region, false);
      const hue = Number(backgroundColor.match(/hsl\((\d+)/)?.[1]);
      expect(hue < 190 || hue > 225).toBe(true);
    }
  });
});
