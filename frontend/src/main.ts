import { createApp } from 'vue';
import App from './App.vue';
import './style.css';

export const BackendURL = import.meta.env.MODE === 'development' ? 'http://localhost:8090' : '';

createApp(App).mount('#app');
