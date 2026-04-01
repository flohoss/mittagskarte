import { createGlobalState, useStorage } from '@vueuse/core';
import { computed } from 'vue';

type FavoritesLookup = Record<string, boolean>;

export const useFavorites = createGlobalState(() => {
  const favorites = useStorage<FavoritesLookup>('mittag-favorites', {});

  const favoriteIds = computed(() =>
    Object.entries(favorites.value)
      .filter(([, isFavorite]) => isFavorite)
      .map(([id]) => id)
  );

  function isFavorite(id: string) {
    return Boolean(favorites.value[id]);
  }

  function toggleFavorite(id: string) {
    const nextFavorites = { ...favorites.value };

    if (nextFavorites[id]) {
      delete nextFavorites[id];
    } else {
      nextFavorites[id] = true;
    }

    favorites.value = nextFavorites;
  }

  return {
    favorites,
    favoriteIds,
    isFavorite,
    toggleFavorite,
  };
});
