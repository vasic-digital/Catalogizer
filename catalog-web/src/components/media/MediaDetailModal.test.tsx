import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MediaDetailModal } from './MediaDetailModal'
import type { MediaItem } from '@/types/media'

describe('MediaDetailModal', () => {
  const mockMedia: MediaItem = {
    id: 123,
    title: 'Test Movie',
    media_type: 'movie',
    year: 2024,
    rating: 8.5,
    quality: '1080p',
    description: 'A great test movie for testing purposes.',
    file_size: 1073741824, // 1 GB
    duration: 7200, // 2 hours
    storage_root_name: 'main_storage',
    storage_root_protocol: 'smb',
    directory_path: '/movies/test.mp4',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    external_metadata: [
      {
        provider: 'tmdb',
        title: 'Test Movie External',
        description: 'External description from TMDB',
        poster_url: 'https://example.com/poster.jpg',
        backdrop_url: 'https://example.com/backdrop.jpg',
        genres: ['Action', 'Adventure', 'Sci-Fi'],
        cast: ['Actor One', 'Actor Two', 'Actor Three'],
      },
    ],
    versions: [
      {
        id: 1,
        quality: '1080p',
        resolution: '1920x1080',
        codec: 'H.264',
        file_size: 1073741824,
        language: 'en',
      },
      {
        id: 2,
        quality: '720p',
        resolution: '1280x720',
        codec: 'H.264',
        file_size: 536870912,
        language: 'en',
      },
    ],
  }

  const mockOnClose = vi.fn()
  const mockOnDownload = vi.fn()
  const mockOnPlay = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('Rendering', () => {
    it('should not render when media is null', () => {
      const { container } = render(
        <MediaDetailModal
          media={null}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(container).toBeEmptyDOMElement()
    })

    it('should not render when isOpen is false', () => {
      const { container } = render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={false}
          onClose={mockOnClose}
        />
      )
      // Headless UI Dialog doesn't render children when closed
      expect(container.querySelector('[role="dialog"]')).not.toBeInTheDocument()
    })

    it('should render when media is provided and isOpen is true', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })

    it('should render external metadata title when available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Test Movie External')).toBeInTheDocument()
    })

    it('should fallback to media title when external metadata not available', () => {
      const mediaWithoutExternal = { ...mockMedia, external_metadata: undefined }
      render(
        <MediaDetailModal
          media={mediaWithoutExternal}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Test Movie')).toBeInTheDocument()
    })
  })

  describe('Close Functionality', () => {
    it('should call onClose when close button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )

      const closeButton = screen.getByRole('button', { name: /close/i })
      await user.click(closeButton)

      expect(mockOnClose).toHaveBeenCalledTimes(1)
    })

    it('should call onClose when clicking outside (backdrop)', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )

      // Headless UI handles backdrop clicks automatically
      // This is handled by the Dialog component itself
      expect(screen.getByRole('dialog')).toBeInTheDocument()
    })
  })

  describe('Media Information Display', () => {
    it('should display year when available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('2024')).toBeInTheDocument()
    })

    it('should display rating when available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('8.5')).toBeInTheDocument()
    })

    it('should display media type', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('movie')).toBeInTheDocument()
    })

    it('should display quality badge', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('1080p')).toBeInTheDocument()
    })

    it('should display description', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('External description from TMDB')).toBeInTheDocument()
    })

    it('should display genres', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Action')).toBeInTheDocument()
      expect(screen.getByText('Adventure')).toBeInTheDocument()
      expect(screen.getByText('Sci-Fi')).toBeInTheDocument()
    })

    it('should not display genres when not available', () => {
      const mediaWithoutGenres = {
        ...mockMedia,
        external_metadata: [{ ...mockMedia.external_metadata![0], genres: undefined }],
      }
      render(
        <MediaDetailModal
          media={mediaWithoutGenres}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('Action')).not.toBeInTheDocument()
    })
  })

  describe('Technical Details', () => {
    it('should display formatted file size', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('File Size')).toBeInTheDocument()
      expect(screen.getByText('1.00 GB')).toBeInTheDocument()
    })

    it('should display formatted duration', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Duration')).toBeInTheDocument()
      expect(screen.getByText('2h 0m')).toBeInTheDocument()
    })

    it('should format duration without hours when less than 60 minutes', () => {
      const mediaWithShortDuration = { ...mockMedia, duration: 1800 } // 30 minutes
      render(
        <MediaDetailModal
          media={mediaWithShortDuration}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('30m')).toBeInTheDocument()
    })

    it('should display storage root name', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Storage')).toBeInTheDocument()
      expect(screen.getByText('main_storage')).toBeInTheDocument()
    })

    it('should display storage protocol', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Protocol')).toBeInTheDocument()
      expect(screen.getByText('SMB')).toBeInTheDocument() // Should be uppercase
    })
  })

  describe('Format Helper Functions', () => {
    it('should format bytes to KB', () => {
      const mediaWithSmallFile = { ...mockMedia, file_size: 2048 }
      render(
        <MediaDetailModal
          media={mediaWithSmallFile}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('2.00 KB')).toBeInTheDocument()
    })

    it('should format bytes to MB', () => {
      const mediaWithMediumFile = { ...mockMedia, file_size: 10485760 } // 10 MB
      render(
        <MediaDetailModal
          media={mediaWithMediumFile}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('10.00 MB')).toBeInTheDocument()
    })

    it('should handle unknown file size', () => {
      const mediaWithoutFileSize = { ...mockMedia, file_size: undefined }
      render(
        <MediaDetailModal
          media={mediaWithoutFileSize}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('File Size')).not.toBeInTheDocument()
    })

    it('should handle unknown duration', () => {
      const mediaWithoutDuration = { ...mockMedia, duration: undefined }
      render(
        <MediaDetailModal
          media={mediaWithoutDuration}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('Duration')).not.toBeInTheDocument()
    })
  })

  describe('Action Buttons', () => {
    it('should render download button when onDownload is provided', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
          onDownload={mockOnDownload}
        />
      )
      expect(screen.getByRole('button', { name: /download/i })).toBeInTheDocument()
    })

    it('should not render download button when onDownload is not provided', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByRole('button', { name: /download/i })).not.toBeInTheDocument()
    })

    it('should call onDownload when download button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
          onDownload={mockOnDownload}
        />
      )

      const downloadButton = screen.getByRole('button', { name: /download/i })
      await user.click(downloadButton)

      expect(mockOnDownload).toHaveBeenCalledTimes(1)
      expect(mockOnDownload).toHaveBeenCalledWith(mockMedia)
    })

    it('should render play button when onPlay is provided', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
          onPlay={mockOnPlay}
        />
      )
      expect(screen.getByRole('button', { name: /play/i })).toBeInTheDocument()
    })

    it('should not render play button when onPlay is not provided', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByRole('button', { name: /play/i })).not.toBeInTheDocument()
    })

    it('should call onPlay when play button is clicked', async () => {
      const user = userEvent.setup()
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
          onPlay={mockOnPlay}
        />
      )

      const playButton = screen.getByRole('button', { name: /play/i })
      await user.click(playButton)

      expect(mockOnPlay).toHaveBeenCalledTimes(1)
      expect(mockOnPlay).toHaveBeenCalledWith(mockMedia)
    })

    it('should render both play and download buttons when both callbacks are provided', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
          onPlay={mockOnPlay}
          onDownload={mockOnDownload}
        />
      )
      expect(screen.getByRole('button', { name: /play/i })).toBeInTheDocument()
      expect(screen.getByRole('button', { name: /download/i })).toBeInTheDocument()
    })
  })

  describe('Images', () => {
    it('should render backdrop image when available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      const backdrop = screen.getByAltText('Test Movie')
      expect(backdrop).toHaveAttribute('src', 'https://example.com/backdrop.jpg')
    })

    it('should render poster image when available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      // There will be two images with same alt text (backdrop and poster)
      const images = screen.getAllByAltText('Test Movie')
      expect(images).toHaveLength(2)
      expect(images[1]).toHaveAttribute('src', 'https://example.com/poster.jpg')
    })

    it('should fallback to cover_image for poster when external metadata not available', () => {
      const mediaWithCoverImage = {
        ...mockMedia,
        cover_image: 'https://example.com/cover.jpg',
        external_metadata: undefined,
      }
      render(
        <MediaDetailModal
          media={mediaWithCoverImage}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      const poster = screen.getByAltText('Test Movie')
      expect(poster).toHaveAttribute('src', 'https://example.com/cover.jpg')
    })

    it('should not render backdrop when not available', () => {
      const mediaWithoutBackdrop = {
        ...mockMedia,
        external_metadata: [{ ...mockMedia.external_metadata![0], backdrop_url: undefined }],
      }
      render(
        <MediaDetailModal
          media={mediaWithoutBackdrop}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      const images = screen.queryAllByAltText('Test Movie')
      expect(images).toHaveLength(1) // Only poster
    })
  })

  describe('Cast Information', () => {
    it('should display cast section when cast is available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Cast')).toBeInTheDocument()
      expect(screen.getByText('Actor One')).toBeInTheDocument()
      expect(screen.getByText('Actor Two')).toBeInTheDocument()
      expect(screen.getByText('Actor Three')).toBeInTheDocument()
    })

    it('should limit cast display to 10 actors', () => {
      const mediaWithManyCast = {
        ...mockMedia,
        external_metadata: [
          {
            ...mockMedia.external_metadata![0],
            cast: Array.from({ length: 15 }, (_, i) => `Actor ${i + 1}`),
          },
        ],
      }
      render(
        <MediaDetailModal
          media={mediaWithManyCast}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Actor 1')).toBeInTheDocument()
      expect(screen.getByText('Actor 10')).toBeInTheDocument()
      expect(screen.queryByText('Actor 11')).not.toBeInTheDocument()
    })

    it('should not display cast section when cast is not available', () => {
      const mediaWithoutCast = {
        ...mockMedia,
        external_metadata: [{ ...mockMedia.external_metadata![0], cast: undefined }],
      }
      render(
        <MediaDetailModal
          media={mediaWithoutCast}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('Cast')).not.toBeInTheDocument()
    })
  })

  describe('Versions', () => {
    it('should display versions section when versions are available', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('Available Versions')).toBeInTheDocument()
    })

    it('should display version details', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('1080p - 1920x1080')).toBeInTheDocument()
      expect(screen.getByText(/H\.264 • 1\.00 GB/)).toBeInTheDocument()
      expect(screen.getByText('720p - 1280x720')).toBeInTheDocument()
      expect(screen.getByText(/H\.264 • 512\.00 MB/)).toBeInTheDocument()
    })

    it('should display version language badges', () => {
      render(
        <MediaDetailModal
          media={mockMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      const languageBadges = screen.getAllByText('en')
      expect(languageBadges).toHaveLength(2) // Two versions with 'en' language
    })

    it('should not display versions section when versions are not available', () => {
      const mediaWithoutVersions = { ...mockMedia, versions: undefined }
      render(
        <MediaDetailModal
          media={mediaWithoutVersions}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('Available Versions')).not.toBeInTheDocument()
    })

    it('should not display versions section when versions array is empty', () => {
      const mediaWithEmptyVersions = { ...mockMedia, versions: [] }
      render(
        <MediaDetailModal
          media={mediaWithEmptyVersions}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.queryByText('Available Versions')).not.toBeInTheDocument()
    })
  })

  describe('Edge Cases', () => {
    it('should handle media without any external metadata', () => {
      const minimalMedia: MediaItem = {
        id: 1,
        title: 'Minimal Movie',
        media_type: 'movie',
        directory_path: '/movies/minimal.mp4',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      }

      render(
        <MediaDetailModal
          media={minimalMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )

      expect(screen.getByText('Minimal Movie')).toBeInTheDocument()
      expect(screen.queryByText('Cast')).not.toBeInTheDocument()
      expect(screen.queryByText('Available Versions')).not.toBeInTheDocument()
    })

    it('should handle media type with underscores', () => {
      const tvShowMedia = { ...mockMedia, media_type: 'tv_show' }
      render(
        <MediaDetailModal
          media={tvShowMedia}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      expect(screen.getByText('tv show')).toBeInTheDocument()
    })

    it('should handle zero duration', () => {
      const mediaWithZeroDuration = { ...mockMedia, duration: 0 }
      render(
        <MediaDetailModal
          media={mediaWithZeroDuration}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      // Duration should not be displayed when 0
      expect(screen.queryByText('Duration')).not.toBeInTheDocument()
    })

    it('should handle zero file size', () => {
      const mediaWithZeroSize = { ...mockMedia, file_size: 0 }
      render(
        <MediaDetailModal
          media={mediaWithZeroSize}
          isOpen={true}
          onClose={mockOnClose}
        />
      )
      // File size should not be displayed when 0
      expect(screen.queryByText('File Size')).not.toBeInTheDocument()
    })
  })
})
