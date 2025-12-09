export interface Collection {
  id: string;
  name: string;
  description?: string;
  mediaCount: number;
  duration: number;
  thumbnail?: string;
  isSmart: boolean;
  criteria?: {
    genres?: string[];
    yearRange?: [number, number];
    ratingRange?: [number, number];
    tags?: string[];
  };
  createdAt: string;
  updatedAt: string;
}