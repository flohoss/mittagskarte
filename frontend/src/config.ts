export const BackendURL = import.meta.env.MODE === 'development' ? 'http://localhost:8090' : '';

export const RepoURL = import.meta.env.VITE_REPO_URL || '';

export const AppVersion = import.meta.env.VITE_APP_VERSION || '';
