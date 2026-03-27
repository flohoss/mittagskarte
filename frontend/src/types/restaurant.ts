export type ScrapeStatus = 'idle' | 'queued' | 'updating' | 'cooldown';

export interface Restaurant {
  id: string;
  collectionId: string;
  collectionName: string;
  name: string;
  group: string;
  address: string;
  phone: string;
  website: string;
  menu: string;
  menu_hash: string;
  thumbnail: string;
  content_type: string;
  method: string;
  cron: string;
  created: string;
  updated: string;
  rest_days: string[];
  tags: string[];
  navigate: string[];
  scrape_status: ScrapeStatus;
}
