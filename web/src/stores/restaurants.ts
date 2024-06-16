import { defineStore } from 'pinia';
import { RestaurantsService, handler_Restaurant } from 'src/openapi';

export const useRestaurantStore = defineStore('restaurant', {
  state: () => ({
    restaurant: {} as handler_Restaurant,
  }),
  actions: {
    async getRestaurant(name: string) {
      const response = await RestaurantsService.getRestaurants1(name);
      this.$state.restaurant = response;
    },
  },
});
