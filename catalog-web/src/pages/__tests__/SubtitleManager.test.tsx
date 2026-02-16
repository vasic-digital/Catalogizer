import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SubtitleManager } from '../SubtitleManager'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'

// Mock framer-motion
vi.mock('framer-motion', () => ({
  motion: {
    div: ({ children, className, ...props }: any) => <div className={className}>{children}</div>,
  },
  AnimatePresence: ({ children }: any) => <>{children}</>,
}))

// Mock subtitle API
vi.mock('@/lib/subtitleApi', () => ({
  subtitleApi: {
    searchSubtitles: vi.fn(),
    downloadSubtitle: vi.fn(),
    getMediaSubtitles: vi.fn(),
    verifySync: vi.fn(),
    translateSubtitle: vi.fn(),
    getSupportedLanguages: vi.fn(() => Promise.resolve([])),
    getSupportedProviders: vi.fn(() => Promise.resolve([])),
    deleteSubtitle: vi.fn(),
    updateSubtitle: vi.fn(),
    uploadSubtitle: vi.fn(),
  },
}))

vi.mock('@/lib/mediaApi', () => ({
  mediaApi: {
    searchMedia: vi.fn(() => Promise.resolve({ items: [] })),
  },
}))

// Mock subtitle components
vi.mock('@/components/subtitles/SubtitleSyncModal', () => ({
  SubtitleSyncModal: () => <div data-testid="sync-modal">Sync Modal</div>,
}))

vi.mock('@/components/subtitles/SubtitleUploadModal', () => ({
  SubtitleUploadModal: () => <div data-testid="upload-modal">Upload Modal</div>,
}))

vi.mock('@/types/subtitles', () => ({
  COMMON_LANGUAGES: [
    { code: 'en', name: 'English', native_name: 'English' },
    { code: 'es', name: 'Spanish', native_name: 'Espanol' },
  ],
}))

const createWrapper = () => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, gcTime: 0 },
    },
  })
  return ({ children }: { children: React.ReactNode }) => (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}

describe('SubtitleManager Page', () => {
  it('renders page heading', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Subtitle Manager')).toBeInTheDocument()
  })

  it('renders page description', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(
      screen.getByText('Search, download, and manage subtitles for your media collection')
    ).toBeInTheDocument()
  })

  it('renders Selected Media section', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Selected Media')).toBeInTheDocument()
  })

  it('shows empty media state when no media selected', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('No media selected')).toBeInTheDocument()
    expect(screen.getByText('Select a media item to manage its subtitles')).toBeInTheDocument()
  })

  it('renders Select Media button', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Select Media')).toBeInTheDocument()
  })

  it('renders Search Subtitles section', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Search Subtitles')).toBeInTheDocument()
  })

  it('renders subtitle search input', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(
      screen.getByPlaceholderText(/Search subtitles by title/)
    ).toBeInTheDocument()
  })

  it('renders Filters button', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Filters')).toBeInTheDocument()
  })

  it('renders Search button', () => {
    render(<SubtitleManager />, { wrapper: createWrapper() })
    expect(screen.getByText('Search')).toBeInTheDocument()
  })

  it('opens media selector when Select Media is clicked', async () => {
    const user = userEvent.setup()
    render(<SubtitleManager />, { wrapper: createWrapper() })

    await user.click(screen.getByText('Select Media'))

    expect(screen.getByText('Search for Media')).toBeInTheDocument()
    expect(screen.getByPlaceholderText(/Search for movies/)).toBeInTheDocument()
  })

  it('shows Cancel button in media selector', async () => {
    const user = userEvent.setup()
    render(<SubtitleManager />, { wrapper: createWrapper() })

    await user.click(screen.getByText('Select Media'))

    expect(screen.getByText('Cancel')).toBeInTheDocument()
  })
})
