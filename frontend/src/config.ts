export const BackendURL = import.meta.env.MODE === 'development' ? 'http://localhost:8090/' : '/';

export const RepoURL = import.meta.env.VITE_REPO_URL || '';

declare global {
  interface Window {
    __BUYMEACOFFEE__?: string;
  }
}

export const BuyMeACoffeeURL = (typeof window === 'undefined' ? '' : window.__BUYMEACOFFEE__) || import.meta.env.VITE_BUYMEACOFFEE_URL || '';

export const AppVersion = import.meta.env.VITE_APP_VERSION || '';
