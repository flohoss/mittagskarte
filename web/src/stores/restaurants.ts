import { defineStore } from 'pinia';
import { RestaurantsService, handler_Restaurant } from 'src/openapi';

export const useRestaurantStore = defineStore('restaurant', {
  state: () => ({
    restaurant: {} as handler_Restaurant,
    restaurants: {} as Record<string, handler_Restaurant>,
  }),
  getters: {
    grouped() {
      const groupMap: Record<string, handler_Restaurant[]> = {};
      for (const [key, value] of Object.entries(this.restaurants)) {
        if (!groupMap[key]) {
          groupMap[key] = [];
        }
        groupMap[key].push(value);
      }
      return groupMap;
    },
  },
  actions: {
    async getRestaurant(name: string) {
      const response = await RestaurantsService.getRestaurants1(name);
      this.$state.restaurant = response;
    },
    async getRestaurants() {
      const response = await RestaurantsService.getRestaurants();
      this.$state.restaurants = response;
    },
  },
});
