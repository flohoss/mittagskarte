export interface MenuRecord {
  id: string;
  file: string;
  hash: string;
  created: string;
  dimensions?: {
    width?: number;
    height?: number;
    landscape?: boolean;
  } | null;
}

export interface RestaurantRecord {
  id: string;
  name: string;
  group: string;
  address: string;
  website: string;
  phone: string;
  tags: string[];
  rest_days: string[];
  method: string;
  status: string;
  updated: string;
  thumbnail: string;
  expand?: {
    menus?: MenuRecord[];
  };
  [key: string]: unknown;
}

export interface RestaurantStatusEvent {
  id: string;
  status: string;
}
