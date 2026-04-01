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
  menu: string;
  menu_dimensions?: {
    width?: number;
    height?: number;
    landscape?: boolean;
  } | null;
  [key: string]: unknown;
}

export interface RestaurantStatusEvent {
  id: string;
  status: string;
}
