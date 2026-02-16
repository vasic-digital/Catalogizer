import { render, screen, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ConversionTools } from '../ConversionTools'

// Mock conversionApi
vi.mock('@/lib/conversionApi', () => ({
  conversionApi: {
    getConversionJobs: vi.fn(() =>
      Promise.resolve([
        {
          id: '1',
          sourceFile: { path: '/media/test.mkv', name: 'test.mkv', format: 'mkv', size: 1073741824 },
          targetFormat: 'mp4',
          quality: 'high',
          status: 'completed',
          progress: 100,
          startTime: '2024-01-01T10:00:00Z',
          endTime: '2024-01-01T10:15:00Z',
          outputFile: '/media/converted/test.mp4',
          options: { resolution: '1080p', bitrate: 5000, framerate: 30, audioCodec: 'aac', videoCodec: 'h264' },
        },
      ])
    ),
    startConversion: vi.fn(),
    cancelConversion: vi.fn(),
    retryConversion: vi.fn(),
    downloadFile: vi.fn(),
  },
}))

vi.mock('react-hot-toast', () => ({
  __esModule: true,
  default: {
    success: vi.fn(),
    error: vi.fn(),
  },
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

describe('ConversionTools Page', () => {
  it('renders page heading', () => {
    render(<ConversionTools />, { wrapper: createWrapper() })
    // "Format Converter" appears both as page heading and component heading
    const headings = screen.getAllByText('Format Converter')
    expect(headings.length).toBeGreaterThanOrEqual(1)
  })

  it('renders page description', () => {
    render(<ConversionTools />, { wrapper: createWrapper() })
    expect(
      screen.getByText(/Convert media files to different formats/)
    ).toBeInTheDocument()
  })

  it('renders FormatConverter component', () => {
    render(<ConversionTools />, { wrapper: createWrapper() })
    expect(screen.getByText('New Conversion')).toBeInTheDocument()
  })

  it('loads and displays conversion jobs', async () => {
    render(<ConversionTools />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('test.mkv')).toBeInTheDocument()
    })
  })

  it('wraps content in max-width container', () => {
    const { container } = render(<ConversionTools />, { wrapper: createWrapper() })
    const wrapper = container.firstChild as HTMLElement
    expect(wrapper).toHaveClass('max-w-7xl')
  })

  it('displays supported format converter', async () => {
    render(<ConversionTools />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Completed')).toBeInTheDocument()
    })
  })

  it('shows download button for completed jobs', async () => {
    render(<ConversionTools />, { wrapper: createWrapper() })

    await waitFor(() => {
      expect(screen.getByText('Download')).toBeInTheDocument()
    })
  })
})
