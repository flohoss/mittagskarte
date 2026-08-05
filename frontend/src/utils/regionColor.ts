export interface RegionBadgeColors {
  backgroundColor: string;
  color: string;
}

const HUES = [0, 18, 36, 54, 72, 90, 108, 126, 144, 162, 234, 252, 270, 288, 306, 324, 342, 360, 200, 216];

function hashIndex(input: string, modulo: number): number {
  let hash = 0;
  for (let i = 0; i < input.length; i++) {
    hash = (hash << 5) - hash + input.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash) % modulo;
}

export function getRegionBadgeColors(region: string, isDark: boolean): RegionBadgeColors {
  const hue = HUES[hashIndex(region, HUES.length)];
  const saturation = isDark ? 65 : 70;
  const lightness = isDark ? 32 : 85;
  const backgroundColor = `hsl(${hue} ${saturation}% ${lightness}%)`;
  const color = isDark ? '#fafafa' : '#1f2937';

  return { backgroundColor, color };
}
