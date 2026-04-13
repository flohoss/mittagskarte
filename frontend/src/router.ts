import { createRouter, createWebHistory } from 'vue-router';
import { nextTick } from 'vue';
import HomeView from './views/HomeView.vue';
import RestaurantView from './views/RestaurantView.vue';

export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    if (savedPosition) return nextTick().then(() => savedPosition);
    return { top: 0 };
  },
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
