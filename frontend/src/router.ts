import { createRouter, createWebHistory } from 'vue-router';
import HomeView from './views/HomeView.vue';
import RestaurantView from './views/RestaurantView.vue';
import PrivacyView from './views/PrivacyView.vue';

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/restaurants/:restaurantId',
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
