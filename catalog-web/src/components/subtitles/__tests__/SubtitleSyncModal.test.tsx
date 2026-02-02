import { render, screen, waitFor, act, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SubtitleSyncModal } from '../SubtitleSyncModal'
import { subtitleApi } from '@/lib/subtitleApi'

// Mock framer-motion to render plain divs
jest.mock('framer-motion', () => ({
  motion: {
    div: ({ children, onClick, className, ...props }: any) => {
      return <div onClick={onClick} className={className}>{children}</div>
    },
  },
}))

// Mock lucide-react icons as simple spans
jest.mock('lucide-react', () => ({
  X: (props: any) => <span data-testid="icon-x" {...props} />,
  CheckCircle: (props: any) => <span data-testid="icon-check-circle" {...props} />,
  AlertCircle: (props: any) => <span data-testid="icon-alert-circle" {...props} />,
  Clock: (props: any) => <span data-testid="icon-clock" {...props} />,
  RefreshCw: (props: any) => <span data-testid="icon-refresh" {...props} />,
  Play: (props: any) => <span data-testid="icon-play" {...props} />,
  Pause: (props: any) => <span data-testid="icon-pause" {...props} />,
}))

jest.mock('@/lib/subtitleApi', () => ({
  subtitleApi: {
    verifySync: jest.fn(),
  },
}))

const mockVerifySync = subtitleApi.verifySync as jest.MockedFunction<typeof subtitleApi.verifySync>

const defaultProps = {
  isOpen: true,
  onClose: jest.fn(),
  subtitleId: 'sub-123',
  mediaId: 42,
}

describe('SubtitleSyncModal', () => {
  beforeEach(() => {
    jest.clearAllMocks()
  })

  describe('Rendering', () => {
    it('renders nothing when isOpen is false', () => {
      const { container } = render(
        <SubtitleSyncModal {...defaultProps} isOpen={false} />
      )
      expect(container.firstChild).toBeNull()
    })

    it('renders the modal when isOpen is true', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.getByText('Verify Subtitle Sync')).toBeInTheDocument()
    })

    it('renders the Start Verification button', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.getByText('Start Verification')).toBeInTheDocument()
    })

    it('renders the Close button', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.getByText('Close')).toBeInTheDocument()
    })

    it('renders sample duration slider with default value', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.getByText('Sample Duration (seconds)')).toBeInTheDocument()
      expect(screen.getByText('60s')).toBeInTheDocument()
    })

    it('renders sensitivity slider with default value', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.getByText('Sensitivity')).toBeInTheDocument()
      // Default sensitivity is 5, displayed as just "5" in the span
      const sensitivitySpan = screen.getByText('5')
      expect(sensitivitySpan).toBeInTheDocument()
    })

    it('renders subtitle language when provided', () => {
      render(<SubtitleSyncModal {...defaultProps} subtitleLanguage="English" />)
      expect(screen.getByText('English')).toBeInTheDocument()
      expect(screen.getByText(/Verifying sync for/)).toBeInTheDocument()
    })

    it('does not render subtitle language section when not provided', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(screen.queryByText(/Verifying sync for/)).not.toBeInTheDocument()
    })

    it('renders the info text about verification', () => {
      render(<SubtitleSyncModal {...defaultProps} />)
      expect(
        screen.getByText('Verification analyzes audio patterns to detect subtitle synchronization issues')
      ).toBeInTheDocument()
    })
  })

  describe('User interactions', () => {
    it('calls onClose when Close button is clicked', async () => {
      const user = userEvent.setup()
      const onClose = jest.fn()

      render(<SubtitleSyncModal {...defaultProps} onClose={onClose} />)
      await user.click(screen.getByText('Close'))

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when X button is clicked', async () => {
      const user = userEvent.setup()
      const onClose = jest.fn()

      render(<SubtitleSyncModal {...defaultProps} onClose={onClose} />)
      // The X icon button is the one containing the icon-x testid
      const xIcon = screen.getAllByTestId('icon-x')[0]
      await user.click(xIcon.closest('button')!)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when clicking the backdrop', async () => {
      const user = userEvent.setup()
      const onClose = jest.fn()

      render(<SubtitleSyncModal {...defaultProps} onClose={onClose} />)
      const backdrop = screen.getByText('Verify Subtitle Sync').closest('.fixed')!
      await user.click(backdrop)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('does not close when clicking inside the modal content', async () => {
      const user = userEvent.setup()
      const onClose = jest.fn()

      render(<SubtitleSyncModal {...defaultProps} onClose={onClose} />)
      await user.click(screen.getByText('Verify Subtitle Sync'))

      expect(onClose).not.toHaveBeenCalled()
    })
  })

  describe('Verification flow', () => {
    it('calls subtitleApi.verifySync with correct parameters', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockResolvedValue({
        success: true,
        status: 'perfect',
        message: 'Subtitle is perfectly synced',
      })

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(mockVerifySync).toHaveBeenCalledWith('sub-123', 42, {
          sample_duration: 60,
          sensitivity: 5,
        })
      })
    })

    it('shows "Verifying..." while verification is in progress', async () => {
      const user = userEvent.setup()
      let resolvePromise: (value: any) => void
      mockVerifySync.mockReturnValue(
        new Promise((resolve) => { resolvePromise = resolve })
      )

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      expect(screen.getByText('Verifying...')).toBeInTheDocument()

      // Resolve to clean up
      await act(async () => {
        resolvePromise!({
          success: true,
          status: 'good',
          message: 'Good sync',
        })
      })

      await waitFor(() => {
        expect(screen.queryByText('Verifying...')).not.toBeInTheDocument()
      })
    })

    it('disables the verify button while verifying', async () => {
      const user = userEvent.setup()
      let resolvePromise: (value: any) => void
      mockVerifySync.mockReturnValue(
        new Promise((resolve) => { resolvePromise = resolve })
      )

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      // During verification, button should show "Verifying..." and be disabled
      const button = screen.getByText('Verifying...').closest('button')!
      expect(button).toBeDisabled()

      await act(async () => {
        resolvePromise!({ success: true, status: 'good' })
      })

      await waitFor(() => {
        expect(screen.getByText('Start Verification').closest('button')).not.toBeDisabled()
      })
    })

    it('displays perfect sync result', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockResolvedValue({
        success: true,
        status: 'perfect',
        message: 'Subtitle is perfectly synced',
        sync_offset: 0,
        confidence: 0.98,
        sync_score: 0.99,
      })

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(screen.getByText('Subtitle is perfectly synced')).toBeInTheDocument()
      })
      expect(screen.getByText(/perfect Sync/)).toBeInTheDocument()
      expect(screen.getByText('0ms')).toBeInTheDocument()
      expect(screen.getByText('98.0%')).toBeInTheDocument()
      expect(screen.getByText('0.99')).toBeInTheDocument()
    })

    it('displays good sync result with positive offset', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockResolvedValue({
        success: true,
        status: 'good',
        message: 'Good sync quality',
        sync_offset: 150,
        confidence: 0.85,
      })

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(screen.getByText(/good/)).toBeInTheDocument()
      })
      expect(screen.getByText('+150ms')).toBeInTheDocument()
      expect(screen.getByText('85.0%')).toBeInTheDocument()
    })

    it('displays unusable sync result on API error', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockRejectedValue(new Error('Network failure'))

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(screen.getByText('Network failure')).toBeInTheDocument()
      })
    })

    it('displays generic error message for non-Error exceptions', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockRejectedValue('some string error')

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(screen.getByText('Verification failed')).toBeInTheDocument()
      })
    })

    it('displays error info from failed verification response', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockResolvedValue({
        success: false,
        status: 'unusable',
        error: 'Audio track not found',
      })

      render(<SubtitleSyncModal {...defaultProps} />)
      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(screen.getByText('Audio track not found')).toBeInTheDocument()
      })
    })
  })

  describe('Slider controls', () => {
    it('updates sample duration when slider changes', () => {
      render(<SubtitleSyncModal {...defaultProps} />)

      const sliders = screen.getAllByRole('slider')
      const durationSlider = sliders[0]

      fireEvent.change(durationSlider, { target: { value: '120' } })

      expect(screen.getByText('120s')).toBeInTheDocument()
    })

    it('updates sensitivity when slider changes', () => {
      render(<SubtitleSyncModal {...defaultProps} />)

      const sliders = screen.getAllByRole('slider')
      const sensitivitySlider = sliders[1]

      fireEvent.change(sensitivitySlider, { target: { value: '8' } })

      expect(screen.getByText('8')).toBeInTheDocument()
    })

    it('passes updated slider values to verifySync', async () => {
      const user = userEvent.setup()
      mockVerifySync.mockResolvedValue({
        success: true,
        status: 'good',
        message: 'Good',
      })

      render(<SubtitleSyncModal {...defaultProps} />)

      const sliders = screen.getAllByRole('slider')
      fireEvent.change(sliders[0], { target: { value: '120' } })
      fireEvent.change(sliders[1], { target: { value: '8' } })

      await user.click(screen.getByText('Start Verification'))

      await waitFor(() => {
        expect(mockVerifySync).toHaveBeenCalledWith('sub-123', 42, {
          sample_duration: 120,
          sensitivity: 8,
        })
      })
    })
  })
})
