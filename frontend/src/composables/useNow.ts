import { onScopeDispose, readonly, ref } from 'vue';

const now = ref(Date.now());
let consumerCount = 0;
let timer: ReturnType<typeof setInterval> | null = null;

function startTimer(intervalMs: number) {
  if (timer) return;

  timer = setInterval(() => {
    now.value = Date.now();
  }, intervalMs);
}

function stopTimer() {
  if (!timer) return;

  clearInterval(timer);
  timer = null;
}

export function useNow(intervalMs = 30_000) {
  consumerCount += 1;
  startTimer(intervalMs);

  onScopeDispose(() => {
    consumerCount = Math.max(0, consumerCount - 1);

    if (consumerCount === 0) {
      stopTimer();
    }
  });

  return readonly(now);
}
