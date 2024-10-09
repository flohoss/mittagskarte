import { defineStore } from 'pinia';
import { RestaurantsService, services_CleanRestaurant } from 'src/openapi';
import { LocalStorage } from 'quasar';

export const FavoriteKey = 'mittag_favorites';
export const MiddayKey = 'mittag_midday';

export const useRestaurantStore = defineStore('restaurant', {
  state: () => ({
    restaurants: {} as Record<string, services_CleanRestaurant>,
    midday: LocalStorage.getItem(MiddayKey) || '1300',
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
  },
  actions: {
    async fetchRestaurants(): Promise<
      Record<string, services_CleanRestaurant>
    > {
      const response = await RestaurantsService.getRestaurants();
      this.$state.restaurants = response || {};
      return this.$state.restaurants;
    },
    toggleFavorite(restaurant: services_CleanRestaurant) {
      if (this.favorites.includes(restaurant.id)) {
        this.favorites.splice(this.favorites.indexOf(restaurant.id), 1);
      } else {
        this.favorites.push(restaurant.id);
      }
      LocalStorage.set(FavoriteKey, this.favorites);
    },
    setMidday(midday: string) {
      this.$state.midday = midday;
      LocalStorage.setItem(MiddayKey, midday);
    },
  },
});
