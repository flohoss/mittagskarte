import { createRouter, createWebHistory } from 'vue-router';
import { nextTick } from 'vue';
import HomeView from './views/HomeView.vue';
import RestaurantView from './views/RestaurantView.vue';

export const router = createRouter({
  history: createWebHistory(),
  scrollBehavior(_to, _from, savedPosition) {
    return nextTick().then(() => savedPosition ?? { left: 0, top: 0 });
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
