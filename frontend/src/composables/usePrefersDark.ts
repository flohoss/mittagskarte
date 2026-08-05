import { onMounted, onUnmounted, ref } from 'vue';

export function usePrefersDark() {
  const prefersDark = ref(false);

  let mediaQuery: MediaQueryList | null = null;
  const handler = (event: MediaQueryListEvent) => {
    prefersDark.value = event.matches;
  };

  onMounted(() => {
    if (typeof window === 'undefined' || !window.matchMedia) return;
    mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    prefersDark.value = mediaQuery.matches;
    mediaQuery.addEventListener('change', handler);
  });

  onUnmounted(() => {
    mediaQuery?.removeEventListener('change', handler);
  });

  return prefersDark;
}
