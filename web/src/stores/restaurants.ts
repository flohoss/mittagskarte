import { defineStore } from 'pinia';
import {
  RestaurantsService,
  services_Group,
  services_CleanRestaurant,
} from 'src/openapi';
import { LocalStorage } from 'quasar';

const emptyRestaurant: services_CleanRestaurant = {
  address: '',
  description: '',
  group: services_Group.Degerloch,
  icon: '',
  id: '',
  image_url: '',
  name: '',
  page_url: '',
  phone: '',
  price: 0,
  rest_days: [],
};

export const ReductionKey = 'mittag_reduction';
export const FavoriteKey = 'mittag_favorites';
export const MiddayKey = 'mittag_midday';

export const useRestaurantStore = defineStore('restaurant', {
  state: () => ({
    restaurant: emptyRestaurant as services_CleanRestaurant,
    restaurants: {} as Record<string, services_CleanRestaurant>,
    reduction: LocalStorage.getItem(ReductionKey || 0),
    midday: LocalStorage.getItem(MiddayKey) || '1300',
    search: '',
    favorites: LocalStorage.getItem(FavoriteKey) || ([] as string[]),
  }),
  getters: {
    favoriteRestaurants() {
      const res = [] as services_CleanRestaurant[];
      for (const value of Object.values(this.restaurants)) {
        if (this.favorites.includes(value.id)) res.push(value);
      }
      return res;
    },
    grouped() {
      const groupMap: Record<string, services_CleanRestaurant[]> = {};
      for (const value of Object.values(this.restaurants)) {
        const group = value.group;
        groupMap[group] = groupMap[group] || [];
        groupMap[group].push(value);
      }
      const sortedGroupMap: Record<string, services_CleanRestaurant[]> = {};
      for (const [group, restaurants] of Object.entries(groupMap).sort()) {
        sortedGroupMap[group] = restaurants;
      }
      return sortedGroupMap;
    },
    result(): services_CleanRestaurant[] {
      if (this.search === '') {
        return [];
      }
      const lowerCaseSearch = this.search.toLowerCase();
      return Object.values(this.restaurants).filter((restaurant) => {
        return (
          restaurant.id.toLowerCase().includes(lowerCaseSearch) ||
          restaurant.name.toLowerCase().includes(lowerCaseSearch) ||
          restaurant.description.toLowerCase().includes(lowerCaseSearch) ||
          restaurant.address.toLowerCase().includes(lowerCaseSearch) ||
          restaurant.group.toLowerCase().includes(lowerCaseSearch)
        );
      });
    },
  },
  actions: {
    toggleFavorite(restaurant: services_CleanRestaurant) {
      if (this.favorites.includes(restaurant.id)) {
        this.favorites.splice(this.favorites.indexOf(restaurant.id), 1);
      } else {
        this.favorites.push(restaurant.id);
      }
      LocalStorage.set(FavoriteKey, this.favorites);
    },
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
    setMidday(midday: string) {
      this.$state.midday = midday;
      LocalStorage.setItem(MiddayKey, midday);
    },
  },
});
