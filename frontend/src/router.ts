import { createRouter, createWebHistory } from 'vue-router';
import HomeView from './views/HomeView.vue';
import RestaurantView from './views/RestaurantView.vue';
import PrivacyView from './views/PrivacyView.vue';

let lastHomeScrollTop = 0;

export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(to, from, savedPosition) {
    if (from.name === 'home' && typeof window !== 'undefined') {
      lastHomeScrollTop = window.scrollY;
    }

    if (savedPosition) {
      return savedPosition;
    }

    if (to.name === 'restaurant-view') {
      return { left: 0, top: 0 };
    }

    if (to.name === 'home') {
      return { left: 0, top: Math.max(0, lastHomeScrollTop) };
    }

    return { left: 0, top: 0 };
  },
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/restaurants/:restaurantSlug',
      name: 'restaurant-view',
      component: RestaurantView,
    },
    {
      path: '/datenschutz',
      name: 'privacy',
      component: PrivacyView,
    },
  ],
});
