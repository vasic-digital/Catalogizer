import { describe, it, expect, vi } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import MediaDetailPage from '../MediaDetailPage'

// Mock apiService
vi.mock('../../services/apiService', () => ({
  apiService: {
    getMediaById: vi.fn(),
  },
}))

// Mock lucide-react
vi.mock('lucide-react', () => ({
  ArrowLeft: (props: any) => <span data-testid="icon-arrow-left" {...props} />,
  Play: (props: any) => <span data-testid="icon-play" {...props} />,
  Download: (props: any) => <span data-testid="icon-download" {...props} />,
  Heart: (props: any) => <span data-testid="icon-heart" {...props} />,
  Star: (props: any) => <span data-testid="icon-star" {...props} />,
  Calendar: (props: any) => <span data-testid="icon-calendar" {...props} />,
  HardDrive: (props: any) => <span data-testid="icon-hard-drive" {...props} />,
}))

import { apiService } from '../../services/apiService'

const createQueryClient = () =>
  new QueryClient({
    defaultOptions: {
      queries: { retry: false },
    },
  })

const renderWithRoute = (mediaId: string = '1') => {
  const queryClient = createQueryClient()
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/media/${mediaId}`]}>
        <Routes>
          <Route path="/media/:id" element={<MediaDetailPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>
  )
}

describe('MediaDetailPage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('shows loading skeleton initially', () => {
    vi.mocked(apiService.getMediaById).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )

    const { container } = renderWithRoute()

    const pulseElements = container.querySelectorAll('.animate-pulse')
    expect(pulseElements.length).toBeGreaterThan(0)
  })

  it('shows media not found when media is null', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue(null as any)

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('Media not found')).toBeInTheDocument()
    })
  })

  it('displays media title', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      year: 1999,
      rating: 8.7,
      description: 'A sci-fi classic',
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('The Matrix')).toBeInTheDocument()
    })
  })

  it('displays media year', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      year: 1999,
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('1999')).toBeInTheDocument()
    })
  })

  it('displays media rating', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      rating: 8.7,
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('8.7')).toBeInTheDocument()
    })
  })

  it('displays media type badge', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('movie')).toBeInTheDocument()
    })
  })

  it('displays media description', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      description: 'A sci-fi classic about the nature of reality',
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('A sci-fi classic about the nature of reality')).toBeInTheDocument()
    })
  })

  it('renders action buttons (Play, Download)', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('Play')).toBeInTheDocument()
    })

    expect(screen.getByText('Download')).toBeInTheDocument()
  })

  it('renders Back button', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText('Back')).toBeInTheDocument()
    })
  })

  it('displays file size when available', async () => {
    vi.mocked(apiService.getMediaById).mockResolvedValue({
      id: 1,
      title: 'The Matrix',
      media_type: 'movie',
      file_size: 4294967296, // 4 GB
      directory_path: '/media/movies',
      created_at: '2023-01-01',
      updated_at: '2023-01-01',
    })

    renderWithRoute()

    await waitFor(() => {
      expect(screen.getByText(/File Size: 4.00 GB/)).toBeInTheDocument()
    })
  })
})
