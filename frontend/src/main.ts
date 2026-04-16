import { createApp } from 'vue';
import App from './App.vue';
import { router } from './router';
import './style.css';

if (typeof window !== 'undefined' && 'scrollRestoration' in window.history) {
  window.history.scrollRestoration = 'manual';
  window.scrollTo({ left: 0, top: 0, behavior: 'auto' });
}

createApp(App).use(router).mount('#app');
