export interface Collection {
  id: string;
  name: string;
  description?: string;
  item_count: number;
  is_public: boolean;
  is_smart: boolean;
  primary_media_type: 'music' | 'video' | 'image' | 'document' | 'mixed';
  created_at: string;
  updated_at: string;
  thumbnail_url?: string;
  cover_image?: string;
  owner_id: string;
}