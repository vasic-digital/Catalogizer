import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { SubtitleUploadModal } from '../SubtitleUploadModal'
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
  Upload: (props: any) => <span data-testid="icon-upload" {...props} />,
  FileText: (props: any) => <span data-testid="icon-file-text" {...props} />,
  Globe: (props: any) => <span data-testid="icon-globe" {...props} />,
  AlertCircle: (props: any) => <span data-testid="icon-alert-circle" {...props} />,
  CheckCircle: (props: any) => <span data-testid="icon-check-circle" {...props} />,
}))

jest.mock('@/lib/subtitleApi', () => ({
  subtitleApi: {
    uploadSubtitle: jest.fn(),
  },
}))

const mockUploadSubtitle = subtitleApi.uploadSubtitle as jest.MockedFunction<typeof subtitleApi.uploadSubtitle>

const defaultProps = {
  isOpen: true,
  onClose: jest.fn(),
  mediaId: 42,
}

function createFile(name: string, size: number = 1024): File {
  const content = new Array(size).fill('a').join('')
  return new File([content], name, { type: 'text/plain' })
}

/** Helper to get the main upload action button (not the heading) */
function getUploadButton(): HTMLButtonElement {
  // The upload action button is inside a div.flex.gap-3 and contains "Upload Subtitle" text
  const buttons = screen.getAllByRole('button')
  const uploadBtn = buttons.find(
    (btn) => btn.textContent?.includes('Upload Subtitle') && btn.classList.contains('flex-1')
  )
  return uploadBtn as HTMLButtonElement
}

describe('SubtitleUploadModal', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    jest.useFakeTimers()
  })

  afterEach(() => {
    jest.useRealTimers()
  })

  describe('Rendering', () => {
    it('renders nothing when isOpen is false', () => {
      const { container } = render(
        <SubtitleUploadModal {...defaultProps} isOpen={false} />
      )
      expect(container.firstChild).toBeNull()
    })

    it('renders the modal when isOpen is true', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Upload Subtitle', { selector: 'h2' })).toBeInTheDocument()
    })

    it('renders the drag and drop area', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(
        screen.getByText('Drag and drop your subtitle file here, or click to browse')
      ).toBeInTheDocument()
    })

    it('renders the Browse Files button', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Browse Files')).toBeInTheDocument()
    })

    it('renders the language selector', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Select language...')).toBeInTheDocument()
    })

    it('renders the format selector with auto-detect default', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Auto-detect...')).toBeInTheDocument()
    })

    it('renders the Cancel button', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Cancel')).toBeInTheDocument()
    })

    it('renders media title when provided', () => {
      render(<SubtitleUploadModal {...defaultProps} mediaTitle="The Matrix" />)
      expect(screen.getByText('The Matrix')).toBeInTheDocument()
      expect(screen.getByText(/Uploading for/)).toBeInTheDocument()
    })

    it('does not render media title section when not provided', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.queryByText(/Uploading for/)).not.toBeInTheDocument()
    })

    it('renders supported formats info text', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('Supported formats: SRT, VTT, ASS, SSA, SUB')).toBeInTheDocument()
    })

    it('renders language options from COMMON_LANGUAGES', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('English (English)')).toBeInTheDocument()
      expect(screen.getByText(/Español/)).toBeInTheDocument()
      expect(screen.getByText(/Français/)).toBeInTheDocument()
    })

    it('renders format options from SUBTITLE_FORMATS', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(screen.getByText('SRT')).toBeInTheDocument()
      expect(screen.getByText('VTT')).toBeInTheDocument()
      expect(screen.getByText('ASS')).toBeInTheDocument()
    })
  })

  describe('User interactions', () => {
    it('calls onClose when Cancel button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      const onClose = jest.fn()

      render(<SubtitleUploadModal {...defaultProps} onClose={onClose} />)
      await user.click(screen.getByText('Cancel'))

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when X button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      const onClose = jest.fn()

      render(<SubtitleUploadModal {...defaultProps} onClose={onClose} />)
      const xIcon = screen.getAllByTestId('icon-x')[0]
      await user.click(xIcon.closest('button')!)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('calls onClose when clicking the backdrop', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      const onClose = jest.fn()

      render(<SubtitleUploadModal {...defaultProps} onClose={onClose} />)
      const backdrop = screen.getByText('Upload Subtitle', { selector: 'h2' }).closest('.fixed')!
      await user.click(backdrop)

      expect(onClose).toHaveBeenCalledTimes(1)
    })

    it('does not close when clicking inside the modal content', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      const onClose = jest.fn()

      render(<SubtitleUploadModal {...defaultProps} onClose={onClose} />)
      await user.click(screen.getByText('Upload Subtitle', { selector: 'h2' }))

      expect(onClose).not.toHaveBeenCalled()
    })
  })

  describe('File selection', () => {
    it('shows file info after selecting a valid .srt file', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt', 2048)

      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.getByText('movie.srt')).toBeInTheDocument()
      expect(screen.getByText('2 KB')).toBeInTheDocument()
    })

    it('auto-detects format from file extension', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.vtt')

      fireEvent.change(fileInput, { target: { files: [file] } })

      const formatSelect = screen.getAllByRole('combobox')[1]
      expect(formatSelect).toHaveValue('vtt')
    })

    it('shows error for invalid file format', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.txt')

      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(
        screen.getByText('Invalid file format. Please select a subtitle file (.srt, .vtt, .ass, .ssa, .sub)')
      ).toBeInTheDocument()
    })

    it('handles drag and drop of valid file', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const dropZone = screen.getByText('Drag and drop your subtitle file here, or click to browse').closest('[class*="border-"]')!
      const file = createFile('movie.srt', 512)

      fireEvent.dragOver(dropZone)
      fireEvent.drop(dropZone, {
        dataTransfer: { files: [file] },
      })

      expect(screen.getByText('movie.srt')).toBeInTheDocument()
    })

    it('handles drag leave event', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const dropZone = screen.getByText('Drag and drop your subtitle file here, or click to browse').closest('[class*="border-"]')!

      fireEvent.dragOver(dropZone)
      fireEvent.dragLeave(dropZone)

      expect(dropZone).toBeInTheDocument()
    })

    it('clears file when remove button is clicked', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.getByText('movie.srt')).toBeInTheDocument()

      // The remove button is inside the file info area
      const fileInfoSection = screen.getByText('movie.srt').closest('[class*="bg-gray"]')!
      const clearButton = fileInfoSection.querySelector('button')!
      await user.click(clearButton)

      expect(screen.queryByText('movie.srt')).not.toBeInTheDocument()
    })
  })

  describe('Upload flow', () => {
    it('upload button is disabled when no file is selected', () => {
      render(<SubtitleUploadModal {...defaultProps} />)
      expect(getUploadButton()).toBeDisabled()
    })

    it('upload button is disabled when no language is selected', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(getUploadButton()).toBeDisabled()
    })

    it('upload button is enabled when file and language are selected', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      expect(getUploadButton()).not.toBeDisabled()
    })

    it('calls uploadSubtitle API on successful upload', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockResolvedValue({
        success: true,
        message: 'Subtitle uploaded successfully!',
      })

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(mockUploadSubtitle).toHaveBeenCalledWith(42, file, 'en', 'srt')
      })
    })

    it('shows success message after upload', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockResolvedValue({
        success: true,
        message: 'Subtitle uploaded successfully!',
      })

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(screen.getByText('Subtitle uploaded successfully!')).toBeInTheDocument()
      })
    })

    it('calls onUploadSuccess callback on successful upload', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      const onUploadSuccess = jest.fn()
      mockUploadSubtitle.mockResolvedValue({
        success: true,
        message: 'Done!',
      })

      render(<SubtitleUploadModal {...defaultProps} onUploadSuccess={onUploadSuccess} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(onUploadSuccess).toHaveBeenCalledTimes(1)
      })
    })

    it('resets form after successful upload timeout', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockResolvedValue({
        success: true,
        message: 'Done!',
      })

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(screen.getByText('Done!')).toBeInTheDocument()
      })

      // Advance timers to trigger the form reset (2000ms timeout)
      act(() => {
        jest.advanceTimersByTime(2000)
      })

      await waitFor(() => {
        expect(screen.queryByText('movie.srt')).not.toBeInTheDocument()
      })
    })

    it('shows error message when upload fails with error response', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockResolvedValue({
        success: false,
        error: 'File too large',
      })

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(screen.getByText('File too large')).toBeInTheDocument()
      })
    })

    it('shows error message when upload throws an exception', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockRejectedValue(new Error('Network error'))

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(screen.getByText('Network error')).toBeInTheDocument()
      })
    })

    it('shows generic error for non-Error exceptions', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockRejectedValue('string error')

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(screen.getByText('Upload failed')).toBeInTheDocument()
      })
    })

    it('shows "Uploading..." during upload', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      let resolvePromise: (value: any) => void
      mockUploadSubtitle.mockReturnValue(
        new Promise((resolve) => { resolvePromise = resolve })
      )

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      expect(screen.getByText('Uploading...')).toBeInTheDocument()

      await act(async () => {
        resolvePromise!({ success: true, message: 'Done!' })
      })

      await waitFor(() => {
        expect(screen.queryByText('Uploading...')).not.toBeInTheDocument()
      })
    })

    it('passes format as undefined when auto-detect is selected', async () => {
      const user = userEvent.setup({ advanceTimers: jest.advanceTimersByTime })
      mockUploadSubtitle.mockResolvedValue({
        success: true,
        message: 'Done!',
      })

      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt')
      fireEvent.change(fileInput, { target: { files: [file] } })

      // File auto-detected as 'srt', now change format back to auto-detect
      const formatSelect = screen.getAllByRole('combobox')[1]
      await user.selectOptions(formatSelect, '')

      const languageSelect = screen.getAllByRole('combobox')[0]
      await user.selectOptions(languageSelect, 'en')

      await user.click(getUploadButton())

      await waitFor(() => {
        expect(mockUploadSubtitle).toHaveBeenCalledWith(42, file, 'en', undefined)
      })
    })
  })

  describe('File size formatting', () => {
    it('displays file size in Bytes for small files', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt', 500)
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.getByText('500 Bytes')).toBeInTheDocument()
    })

    it('displays file size in KB', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt', 2048)
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.getByText('2 KB')).toBeInTheDocument()
    })

    it('displays file size in MB', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.srt', 1048576)
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.getByText('1 MB')).toBeInTheDocument()
    })
  })

  describe('Supported file types', () => {
    const validExtensions = ['.srt', '.vtt', '.ass', '.ssa', '.sub']

    validExtensions.forEach((ext) => {
      it(`accepts ${ext} files`, () => {
        render(<SubtitleUploadModal {...defaultProps} />)

        const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
        const file = createFile(`movie${ext}`)
        fireEvent.change(fileInput, { target: { files: [file] } })

        expect(screen.getByText(`movie${ext}`)).toBeInTheDocument()
      })
    })

    it('rejects .mp4 files', () => {
      render(<SubtitleUploadModal {...defaultProps} />)

      const fileInput = document.querySelector('input[type="file"]') as HTMLInputElement
      const file = createFile('movie.mp4')
      fireEvent.change(fileInput, { target: { files: [file] } })

      expect(screen.queryByText('movie.mp4')).not.toBeInTheDocument()
      expect(screen.getByText(/Invalid file format/)).toBeInTheDocument()
    })
  })
})
