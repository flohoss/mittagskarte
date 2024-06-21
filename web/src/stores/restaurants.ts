import { defineStore } from 'pinia';
import {
  RestaurantsService,
  config_Group,
  handler_Restaurant,
} from 'src/openapi';
import { LocalStorage } from 'quasar';

const emptyRestaurant: handler_Restaurant = {
  address: '',
  description: '',
  group: config_Group.Degerloch,
  icon: '',
  id: '',
  menu: {
    card: '',
    description: '',
    food: [],
  },
  name: '',
  page_url: '',
  phone: '',
  price: 0,
  rest_days: [],
};

export const ReductionKey = 'mittag_reduction';

export const useRestaurantStore = defineStore('restaurant', {
  state: () => ({
    restaurant: emptyRestaurant as handler_Restaurant,
    restaurants: {} as Record<string, handler_Restaurant>,
    reduction: LocalStorage.getItem(ReductionKey),
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
    setReduction(reduction: number) {
      this.$state.reduction = reduction;
      LocalStorage.setItem(ReductionKey, reduction);
    },
  },
});
