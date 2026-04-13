import type { RecordModel } from 'pocketbase';

export interface MenuRecord extends RecordModel {
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

export interface RestaurantRecord extends RecordModel {
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
  last_check?: {
    at: string;
    status: 'success' | 'not_changed' | 'error';
    detail?: string;
  } | null;
  expand?: {
    menus?: MenuRecord[];
  };
}

export interface RestaurantStatusEvent {
  id: string;
  status: string;
}
