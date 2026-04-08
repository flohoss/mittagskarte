import { createRouter, createWebHistory } from 'vue-router';
import HomeView from './views/HomeView.vue';
import RestaurantView from './views/RestaurantView.vue';

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
  ],
});
