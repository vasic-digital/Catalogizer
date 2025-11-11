import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { MediaDetailModal } from '../MediaDetailModal'
import type { MediaItem } from '@/types/media'

// Mock headlessui Dialog to avoid portal issues in tests
jest.mock('@headlessui/react', () => ({
  Dialog: Object.assign(
    ({ children, onClose }: any) => <div onClick={onClose}>{children}</div>,
    {
      Panel: ({ children, className }: any) => <div className={className}>{children}</div>,
      Title: ({ children, as: Component = 'h2', className }: any) => (
        <Component className={className}>{children}</Component>
      ),
    }
  ),
  Transition: Object.assign(
    ({ children, show }: any) => (show ? <div>{children}</div> : null),
    {
      Child: ({ children }: any) => <div>{children}</div>,
    }
  ),
}))

const mockMediaItem: MediaItem = {
  id: 1,
  title: 'Test Movie',
  media_type: 'movie',
  year: 2023,
  rating: 8.5,
  quality: '1080p',
  file_size: 2147483648, // 2 GB
  duration: 7200, // 2 hours
  storage_root_name: 'Main Storage',
  storage_root_protocol: 'smb',
  description: 'A great test movie for testing purposes',
  cover_image: '/test-poster.jpg',
  directory_path: '/movies/test-movie.mp4',
  created_at: '2023-01-01T00:00:00Z',
  updated_at: '2023-01-01T00:00:00Z',
  external_metadata: [
    {
      id: 1,
      media_id: 1,
      provider: 'tmdb',
      external_id: 'tmdb-123',
      title: 'Test Movie (External)',
      description: 'External description',
      poster_url: '/external-poster.jpg',
      backdrop_url: '/backdrop.jpg',
      genres: ['Action', 'Adventure', 'Sci-Fi'],
      cast: ['Actor One', 'Actor Two', 'Actor Three'],
      rating: 8.7,
      metadata: {},
      last_updated: '2023-01-01T00:00:00Z',
    },
  ],
  versions: [
    {
      id: 1,
      media_id: 1,
      version: 'v1',
      quality: '1080p',
      file_path: '/movies/test-movie-1080p.mp4',
      resolution: '1920x1080',
      codec: 'H.264',
      file_size: 2147483648,
      language: 'en',
    },
    {
      id: 2,
      media_id: 1,
      version: 'v2',
      quality: '720p',
      file_path: '/movies/test-movie-720p.mp4',
      resolution: '1280x720',
      codec: 'H.264',
      file_size: 1073741824,
      language: 'en',
    },
  ],
}

describe('MediaDetailModal', () => {
  describe('Rendering', () => {
    it('renders nothing when media is null', () => {
      const { container } = render(
        <MediaDetailModal media={null} isOpen={true} onClose={jest.fn()} />
      )

      expect(container.firstChild).toBeNull()
    })

    it('renders modal when isOpen is true and media is provided', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Test Movie (External)')).toBeInTheDocument()
    })

    it('displays external metadata title when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Test Movie (External)')).toBeInTheDocument()
    })

    it('displays fallback title when external metadata is not available', () => {
      const mediaWithoutExternal = { ...mockMediaItem, external_metadata: [] }
      render(
        <MediaDetailModal media={mediaWithoutExternal} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Test Movie')).toBeInTheDocument()
    })

    it('displays backdrop image when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      const images = screen.getAllByAltText('Test Movie')
      const backdropImage = images.find((img) => img.getAttribute('src') === '/backdrop.jpg')
      expect(backdropImage).toBeDefined()
      expect(backdropImage).toHaveAttribute('src', '/backdrop.jpg')
    })

    it('displays poster image when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      const images = screen.getAllByAltText('Test Movie')
      const posterImage = images.find((img) => img.getAttribute('src') === '/external-poster.jpg')
      expect(posterImage).toBeInTheDocument()
    })
  })

  describe('Meta Information', () => {
    it('displays year when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('2023')).toBeInTheDocument()
    })

    it('displays rating when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('8.5')).toBeInTheDocument()
    })

    it('displays media type when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      const movieMatches = screen.getAllByText(/movie/i)
      expect(movieMatches.length).toBeGreaterThan(0)
    })

    it('displays quality badge when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      const qualityMatches = screen.getAllByText(/1080p/i)
      expect(qualityMatches.length).toBeGreaterThan(0)
    })

    it('does not display optional meta info when not available', () => {
      const minimalMedia: MediaItem = {
        id: 1,
        title: 'Minimal Movie',
        media_type: 'movie',
        directory_path: '/movies/minimal.mp4',
        created_at: '2023-01-01T00:00:00Z',
        updated_at: '2023-01-01T00:00:00Z',
      }
      render(
        <MediaDetailModal media={minimalMedia} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText('2023')).not.toBeInTheDocument()
      expect(screen.queryByText('8.5')).not.toBeInTheDocument()
    })
  })

  describe('Genres', () => {
    it('displays genres when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Action')).toBeInTheDocument()
      expect(screen.getByText('Adventure')).toBeInTheDocument()
      expect(screen.getByText('Sci-Fi')).toBeInTheDocument()
    })

    it('does not display genres section when not available', () => {
      const mediaWithoutGenres = {
        ...mockMediaItem,
        external_metadata: [{ ...mockMediaItem.external_metadata![0], genres: [] }],
      }
      render(
        <MediaDetailModal media={mediaWithoutGenres} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText('Action')).not.toBeInTheDocument()
    })
  })

  describe('Description', () => {
    it('displays external metadata description when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('External description')).toBeInTheDocument()
    })

    it('displays fallback description when external metadata is not available', () => {
      const mediaWithoutExternal = { ...mockMediaItem, external_metadata: [] }
      render(
        <MediaDetailModal media={mediaWithoutExternal} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('A great test movie for testing purposes')).toBeInTheDocument()
    })

    it('does not display description section when not available', () => {
      const mediaWithoutDescription = {
        ...mockMediaItem,
        description: undefined,
        external_metadata: [{ ...mockMediaItem.external_metadata![0], description: undefined }],
      }
      render(
        <MediaDetailModal media={mediaWithoutDescription} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText(/description/i)).not.toBeInTheDocument()
    })
  })

  describe('Action Buttons', () => {
    it('displays Play button when onPlay is provided', () => {
      const onPlay = jest.fn()
      render(
        <MediaDetailModal
          media={mockMediaItem}
          isOpen={true}
          onClose={jest.fn()}
          onPlay={onPlay}
        />
      )

      expect(screen.getByRole('button', { name: /play/i })).toBeInTheDocument()
    })

    it('calls onPlay when Play button is clicked', async () => {
      const user = userEvent.setup()
      const onPlay = jest.fn()
      render(
        <MediaDetailModal
          media={mockMediaItem}
          isOpen={true}
          onClose={jest.fn()}
          onPlay={onPlay}
        />
      )

      await user.click(screen.getByRole('button', { name: /play/i }))
      expect(onPlay).toHaveBeenCalledWith(mockMediaItem)
    })

    it('does not display Play button when onPlay is not provided', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByRole('button', { name: /play/i })).not.toBeInTheDocument()
    })

    it('displays Download button when onDownload is provided', () => {
      const onDownload = jest.fn()
      render(
        <MediaDetailModal
          media={mockMediaItem}
          isOpen={true}
          onClose={jest.fn()}
          onDownload={onDownload}
        />
      )

      expect(screen.getByRole('button', { name: /download/i })).toBeInTheDocument()
    })

    it('calls onDownload when Download button is clicked', async () => {
      const user = userEvent.setup()
      const onDownload = jest.fn()
      render(
        <MediaDetailModal
          media={mockMediaItem}
          isOpen={true}
          onClose={jest.fn()}
          onDownload={onDownload}
        />
      )

      await user.click(screen.getByRole('button', { name: /download/i }))
      expect(onDownload).toHaveBeenCalledWith(mockMediaItem)
    })

    it('does not display Download button when onDownload is not provided', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByRole('button', { name: /download/i })).not.toBeInTheDocument()
    })
  })

  describe('Technical Details', () => {
    it('displays file size when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('File Size')).toBeInTheDocument()
      expect(screen.getByText('2.00 GB')).toBeInTheDocument()
    })

    it('displays duration when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Duration')).toBeInTheDocument()
      expect(screen.getByText('2h 0m')).toBeInTheDocument()
    })

    it('displays storage name when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Storage')).toBeInTheDocument()
      expect(screen.getByText('Main Storage')).toBeInTheDocument()
    })

    it('displays protocol when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Protocol')).toBeInTheDocument()
      expect(screen.getByText(/smb/i)).toBeInTheDocument()
    })

    it('formats file size correctly for different sizes', () => {
      const mediaWithSmallFile = { ...mockMediaItem, file_size: 1024 }
      const { rerender } = render(
        <MediaDetailModal media={mediaWithSmallFile} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('1.00 KB')).toBeInTheDocument()

      const mediaWithMediumFile = { ...mockMediaItem, file_size: 1048576 }
      rerender(
        <MediaDetailModal media={mediaWithMediumFile} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('1.00 MB')).toBeInTheDocument()
    })

    it('formats duration correctly for minutes only', () => {
      const mediaWithShortDuration = { ...mockMediaItem, duration: 1800 } // 30 minutes
      render(
        <MediaDetailModal media={mediaWithShortDuration} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('30m')).toBeInTheDocument()
    })

    it('does not display technical details when not available', () => {
      const minimalMedia = {
        ...mockMediaItem,
        file_size: undefined,
        duration: undefined,
        storage_root_name: undefined,
        storage_root_protocol: undefined,
      }
      render(
        <MediaDetailModal media={minimalMedia} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText('File Size')).not.toBeInTheDocument()
      expect(screen.queryByText('Duration')).not.toBeInTheDocument()
      expect(screen.queryByText('Storage')).not.toBeInTheDocument()
      expect(screen.queryByText('Protocol')).not.toBeInTheDocument()
    })
  })

  describe('Cast', () => {
    it('displays cast section when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Cast')).toBeInTheDocument()
      expect(screen.getByText('Actor One')).toBeInTheDocument()
      expect(screen.getByText('Actor Two')).toBeInTheDocument()
      expect(screen.getByText('Actor Three')).toBeInTheDocument()
    })

    it('limits cast display to 10 actors', () => {
      const mediaWithManyCast = {
        ...mockMediaItem,
        external_metadata: [
          {
            ...mockMediaItem.external_metadata![0],
            cast: Array.from({ length: 15 }, (_, i) => `Actor ${i + 1}`),
          },
        ],
      }
      render(
        <MediaDetailModal media={mediaWithManyCast} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Actor 10')).toBeInTheDocument()
      expect(screen.queryByText('Actor 11')).not.toBeInTheDocument()
    })

    it('does not display cast section when not available', () => {
      const mediaWithoutCast = {
        ...mockMediaItem,
        external_metadata: [{ ...mockMediaItem.external_metadata![0], cast: [] }],
      }
      render(
        <MediaDetailModal media={mediaWithoutCast} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText('Cast')).not.toBeInTheDocument()
    })
  })

  describe('Versions', () => {
    it('displays versions section when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('Available Versions')).toBeInTheDocument()
    })

    it('displays version details correctly', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.getByText('1080p - 1920x1080')).toBeInTheDocument()
      expect(screen.getByText('720p - 1280x720')).toBeInTheDocument()
      const codecMatches = screen.getAllByText(/H\.264/)
      expect(codecMatches.length).toBeGreaterThan(0)
    })

    it('displays version language when available', () => {
      render(
        <MediaDetailModal media={mockMediaItem} isOpen={true} onClose={jest.fn()} />
      )

      const languageBadges = screen.getAllByText('en')
      expect(languageBadges.length).toBeGreaterThan(0)
    })

    it('does not display versions section when not available', () => {
      const mediaWithoutVersions = { ...mockMediaItem, versions: [] }
      render(
        <MediaDetailModal media={mediaWithoutVersions} isOpen={true} onClose={jest.fn()} />
      )

      expect(screen.queryByText('Available Versions')).not.toBeInTheDocument()
    })
  })
})
