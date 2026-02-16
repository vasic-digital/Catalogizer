import { MediaService } from '../MediaService';
import { HttpClient } from '../../utils/http';

// Mock HttpClient
jest.mock('../../utils/http');

describe('MediaService', () => {
  let mediaService: MediaService;
  let mockHttp: jest.Mocked<HttpClient>;

  beforeEach(() => {
    jest.clearAllMocks();

    mockHttp = {
      get: jest.fn(),
      post: jest.fn(),
      put: jest.fn(),
      patch: jest.fn(),
      delete: jest.fn(),
      downloadStream: jest.fn(),
    } as any;

    mediaService = new MediaService(mockHttp);
  });

  describe('search', () => {
    it('searches with query parameters', async () => {
      const response = { items: [{ id: 1, title: 'Movie' }], total: 1, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      const result = await mediaService.search({ query: 'movie', media_type: 'movie', limit: 10 });

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?query=movie&media_type=movie&limit=10');
      expect(result).toEqual(response);
    });

    it('searches with empty params', async () => {
      const response = { items: [], total: 0, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      await mediaService.search({});

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search');
    });

    it('searches with default empty request', async () => {
      const response = { items: [], total: 0, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      await mediaService.search();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search');
    });

    it('skips undefined and null values in query params', async () => {
      const response = { items: [], total: 0, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      await mediaService.search({ query: 'test', media_type: undefined });

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?query=test');
    });
  });

  describe('getById', () => {
    it('gets media by ID', async () => {
      const media = { id: 1, title: 'Test Movie', media_type: 'movie' };
      mockHttp.get.mockResolvedValueOnce(media);

      const result = await mediaService.getById(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1');
      expect(result).toEqual(media);
    });
  });

  describe('getStats', () => {
    it('gets media statistics', async () => {
      const stats = { total_items: 100, by_type: { movie: 50 }, by_quality: {}, total_size: 1000, recent_additions: 5 };
      mockHttp.get.mockResolvedValueOnce(stats);

      const result = await mediaService.getStats();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/stats');
      expect(result).toEqual(stats);
    });
  });

  describe('getRecentlyAdded', () => {
    it('gets recently added with default limit', async () => {
      const response = { items: [{ id: 1 }], total: 1, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      const result = await mediaService.getRecentlyAdded();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?sort_by=created_at&sort_order=desc&limit=20');
      expect(result).toEqual(response.items);
    });

    it('gets recently added with custom limit', async () => {
      const response = { items: [], total: 0, limit: 5, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      await mediaService.getRecentlyAdded(5);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?sort_by=created_at&sort_order=desc&limit=5');
    });
  });

  describe('getTrending', () => {
    it('gets trending media sorted by rating', async () => {
      const response = { items: [{ id: 1, rating: 9.0 }], total: 1, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      const result = await mediaService.getTrending();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?sort_by=rating&sort_order=desc&limit=20');
      expect(result).toEqual(response.items);
    });
  });

  describe('getByType', () => {
    it('gets media filtered by type', async () => {
      const response = { items: [{ id: 1, media_type: 'tv' }], total: 1, limit: 20, offset: 0, has_next: false, has_previous: false };
      mockHttp.get.mockResolvedValueOnce(response);

      const result = await mediaService.getByType('tv');

      expect(mockHttp.get).toHaveBeenCalledWith('/media/search?media_type=tv&sort_by=updated_at&sort_order=desc&limit=20');
      expect(result).toEqual(response.items);
    });
  });

  describe('favorites', () => {
    it('gets user favorites', async () => {
      const favorites = [{ id: 1, title: 'Fav Movie' }];
      mockHttp.get.mockResolvedValueOnce(favorites);

      const result = await mediaService.getFavorites();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/favorites?limit=50');
      expect(result).toEqual(favorites);
    });

    it('toggles favorite status', async () => {
      const toggleResult = { is_favorite: true };
      mockHttp.post.mockResolvedValueOnce(toggleResult);

      const result = await mediaService.toggleFavorite(42);

      expect(mockHttp.post).toHaveBeenCalledWith('/media/42/favorite');
      expect(result).toEqual(toggleResult);
    });
  });

  describe('playback progress', () => {
    it('updates playback progress', async () => {
      mockHttp.put.mockResolvedValueOnce(undefined);

      const progress = { media_id: 1, position: 300, duration: 7200, timestamp: Date.now() };
      await mediaService.updateProgress(1, progress);

      expect(mockHttp.put).toHaveBeenCalledWith('/media/1/progress', progress);
    });

    it('gets playback progress', async () => {
      const progress = { media_id: 1, position: 300, duration: 7200 };
      mockHttp.get.mockResolvedValueOnce(progress);

      const result = await mediaService.getProgress(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1/progress');
      expect(result).toEqual(progress);
    });

    it('returns null when no progress found', async () => {
      mockHttp.get.mockRejectedValueOnce(new Error('Not found'));

      const result = await mediaService.getProgress(1);

      expect(result).toBeNull();
    });

    it('marks media as watched', async () => {
      mockHttp.put.mockResolvedValueOnce(undefined);

      await mediaService.markAsWatched(1);

      expect(mockHttp.put).toHaveBeenCalledWith('/media/1/progress', expect.objectContaining({
        media_id: 1,
        position: 100,
        duration: 100,
      }));
    });

    it('gets continue watching items', async () => {
      const items = [{ id: 1, title: 'In Progress' }];
      mockHttp.get.mockResolvedValueOnce(items);

      const result = await mediaService.getContinueWatching();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/continue-watching?limit=20');
      expect(result).toEqual(items);
    });
  });

  describe('streaming and downloads', () => {
    it('gets stream URL', async () => {
      const streamInfo = { url: 'http://example.com/stream', mime_type: 'video/mp4', file_size: 1000000 };
      mockHttp.get.mockResolvedValueOnce(streamInfo);

      const result = await mediaService.getStreamUrl(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1/stream');
      expect(result).toEqual(streamInfo);
    });

    it('gets download URL', async () => {
      const downloadInfo = { url: 'http://example.com/dl', expires_at: '2024-01-01' };
      mockHttp.get.mockResolvedValueOnce(downloadInfo);

      const result = await mediaService.getDownloadUrl(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1/download');
      expect(result).toEqual(downloadInfo);
    });

    it('queues media for download', async () => {
      const job = { id: 1, media_id: 1, status: 'pending', progress: 0 };
      mockHttp.post.mockResolvedValueOnce(job);

      const result = await mediaService.queueDownload(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/media/1/download');
      expect(result).toEqual(job);
    });

    it('gets download jobs', async () => {
      const jobs = [{ id: 1, status: 'downloading', progress: 50 }];
      mockHttp.get.mockResolvedValueOnce(jobs);

      const result = await mediaService.getDownloadJobs();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/downloads');
      expect(result).toEqual(jobs);
    });

    it('cancels a download job', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);
      await mediaService.cancelDownload(1);
      expect(mockHttp.post).toHaveBeenCalledWith('/media/downloads/1/cancel');
    });

    it('pauses a download job', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);
      await mediaService.pauseDownload(1);
      expect(mockHttp.post).toHaveBeenCalledWith('/media/downloads/1/pause');
    });

    it('resumes a download job', async () => {
      mockHttp.post.mockResolvedValueOnce(undefined);
      await mediaService.resumeDownload(1);
      expect(mockHttp.post).toHaveBeenCalledWith('/media/downloads/1/resume');
    });
  });

  describe('media images', () => {
    it('gets thumbnail without size', async () => {
      const buffer = new ArrayBuffer(100);
      mockHttp.downloadStream.mockResolvedValueOnce(buffer);

      const result = await mediaService.getThumbnail(1);

      expect(mockHttp.downloadStream).toHaveBeenCalledWith('/media/1/thumbnail');
      expect(result).toBe(buffer);
    });

    it('gets thumbnail with size', async () => {
      const buffer = new ArrayBuffer(100);
      mockHttp.downloadStream.mockResolvedValueOnce(buffer);

      await mediaService.getThumbnail(1, 'small');

      expect(mockHttp.downloadStream).toHaveBeenCalledWith('/media/1/thumbnail?size=small');
    });

    it('gets poster', async () => {
      const buffer = new ArrayBuffer(100);
      mockHttp.downloadStream.mockResolvedValueOnce(buffer);

      await mediaService.getPoster(1, 'large');

      expect(mockHttp.downloadStream).toHaveBeenCalledWith('/media/1/poster?size=large');
    });

    it('gets backdrop', async () => {
      const buffer = new ArrayBuffer(100);
      mockHttp.downloadStream.mockResolvedValueOnce(buffer);

      await mediaService.getBackdrop(1, 'medium');

      expect(mockHttp.downloadStream).toHaveBeenCalledWith('/media/1/backdrop?size=medium');
    });
  });

  describe('metadata operations', () => {
    it('updates media metadata', async () => {
      const updated = { id: 1, title: 'Updated Title' };
      mockHttp.put.mockResolvedValueOnce(updated);

      const result = await mediaService.updateMetadata(1, { title: 'Updated Title' } as any);

      expect(mockHttp.put).toHaveBeenCalledWith('/media/1', { title: 'Updated Title' });
      expect(result).toEqual(updated);
    });

    it('refreshes media metadata', async () => {
      const refreshed = { id: 1, title: 'Refreshed' };
      mockHttp.post.mockResolvedValueOnce(refreshed);

      const result = await mediaService.refreshMetadata(1);

      expect(mockHttp.post).toHaveBeenCalledWith('/media/1/refresh');
      expect(result).toEqual(refreshed);
    });

    it('deletes media without files', async () => {
      mockHttp.delete.mockResolvedValueOnce(undefined);

      await mediaService.delete(1);

      expect(mockHttp.delete).toHaveBeenCalledWith('/media/1');
    });

    it('deletes media with files', async () => {
      mockHttp.delete.mockResolvedValueOnce(undefined);

      await mediaService.delete(1, true);

      expect(mockHttp.delete).toHaveBeenCalledWith('/media/1?delete_files=true');
    });
  });

  describe('recommendations and ratings', () => {
    it('gets similar media', async () => {
      const similar = [{ id: 2, title: 'Similar' }];
      mockHttp.get.mockResolvedValueOnce(similar);

      const result = await mediaService.getSimilar(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1/similar?limit=10');
      expect(result).toEqual(similar);
    });

    it('gets recommendations', async () => {
      const recs = [{ id: 3, title: 'Recommended' }];
      mockHttp.get.mockResolvedValueOnce(recs);

      const result = await mediaService.getRecommendations();

      expect(mockHttp.get).toHaveBeenCalledWith('/media/recommendations?limit=20');
      expect(result).toEqual(recs);
    });

    it('rates a media item', async () => {
      mockHttp.post.mockResolvedValueOnce({ rating: 8 });

      const result = await mediaService.rate(1, 8);

      expect(mockHttp.post).toHaveBeenCalledWith('/media/1/rate', { rating: 8 });
      expect(result).toEqual({ rating: 8 });
    });

    it('gets user rating', async () => {
      mockHttp.get.mockResolvedValueOnce({ rating: 7 });

      const result = await mediaService.getRating(1);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/1/rating');
      expect(result).toEqual({ rating: 7 });
    });

    it('returns null when no rating exists', async () => {
      mockHttp.get.mockRejectedValueOnce(new Error('Not found'));

      const result = await mediaService.getRating(1);

      expect(result).toBeNull();
    });
  });

  describe('download job details', () => {
    it('gets specific download job', async () => {
      const job = { id: 5, media_id: 1, status: 'completed', progress: 100 };
      mockHttp.get.mockResolvedValueOnce(job);

      const result = await mediaService.getDownloadJob(5);

      expect(mockHttp.get).toHaveBeenCalledWith('/media/downloads/5');
      expect(result).toEqual(job);
    });
  });
});
