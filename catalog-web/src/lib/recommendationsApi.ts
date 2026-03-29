import api from './api'

export interface RecommendedItem {
  id: number
  title: string
  media_type: string
  year?: number
  rating?: number
  similarity_score?: number
  reason?: string
}

export interface TrendingItem {
  id: number
  title: string
  media_type: string
  year?: number
  access_count: number
  trending_score: number
}

export interface PersonalizedRecommendation {
  id: number
  title: string
  media_type: string
  year?: number
  rating?: number
  recommendation_reason: string
}

export const recommendationsApi = {
  getSimilar: (mediaId: number): Promise<RecommendedItem[]> =>
    api.get(`/recommendations/similar/${mediaId}`).then((res) => res.data),

  getTrending: (limit = 10): Promise<TrendingItem[]> =>
    api.get('/recommendations/trending', { params: { limit } }).then((res) => res.data),

  getPersonalized: (userId: number, limit = 10): Promise<PersonalizedRecommendation[]> =>
    api.get(`/recommendations/personalized/${userId}`, { params: { limit } }).then((res) => res.data),
}
