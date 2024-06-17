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
      for (const value of Object.values(this.restaurants)) {
        const group = value.group;
        groupMap[group] = groupMap[group] || [];
        groupMap[group].push(value);
      }
      const sortedGroupMap: Record<string, handler_Restaurant[]> = {};
      for (const [group, restaurants] of Object.entries(groupMap).sort()) {
        sortedGroupMap[group] = restaurants;
      }
      return sortedGroupMap;
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
