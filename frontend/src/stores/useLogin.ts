import { createGlobalState } from '@vueuse/core';
import { computed, ref } from 'vue';
import { backendClient } from '../services/backendClient';

export const useLogin = createGlobalState(() => {
  const authToken = ref(backendClient.getAuthToken());
  const authIdentity = ref('');

  function resolveAuthIdentity() {
    const record = backendClient.getAuthRecord();
    if (!record) {
      authIdentity.value = '';
      return;
    }

    const email = typeof record.email === 'string' ? record.email : '';
    const username = typeof record.username === 'string' ? record.username : '';
    const id = typeof record.id === 'string' ? record.id : '';
    authIdentity.value = email || username || id;
  }

  resolveAuthIdentity();

  backendClient.onAuthChange((token) => {
    authToken.value = token;
    resolveAuthIdentity();
  });

  async function ensureValidAuthToken() {
    if (!authToken.value) {
      return;
    }

    await backendClient.validateAuthentication();
  }

  async function authenticate(identity: string, password: string) {
    await backendClient.authenticate(identity, password);
  }

  function clearAuthentication() {
    backendClient.clearAuthentication();
  }

  const isAuthenticated = computed(() => Boolean(authToken.value));

  void ensureValidAuthToken();

  return {
    authToken,
    authIdentity,
    isAuthenticated,
    authenticate,
    clearAuthentication,
  };
});
