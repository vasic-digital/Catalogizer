import { downloadApi } from '../downloadApi'
import apiDefault from '../api'

vi.mock('../api', async () => {
  const mockApi = {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  }
  return {
    __esModule: true,
    default: mockApi,
    api: mockApi,
  }
})

const mockApi = vi.mocked(apiDefault)

// Mock URL.createObjectURL and revokeObjectURL
const mockCreateObjectURL = vi.fn().mockReturnValue('blob:http://localhost/mock-url')
const mockRevokeObjectURL = vi.fn()
Object.defineProperty(window, 'URL', {
  value: {
    createObjectURL: mockCreateObjectURL,
    revokeObjectURL: mockRevokeObjectURL,
  },
  writable: true,
})

describe('downloadApi', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // Mock document.createElement and body.appendChild for download trigger
    vi.spyOn(document, 'createElement').mockReturnValue({
      href: '',
      setAttribute: vi.fn(),
      click: vi.fn(),
      parentNode: { removeChild: vi.fn() },
    } as unknown as HTMLAnchorElement)
    vi.spyOn(document.body, 'appendChild').mockImplementation((node) => node)
  })

  describe('downloadFile', () => {
    it('calls GET /download/file/:fileId with blob responseType', async () => {
      const mockBlob = new Blob(['file content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadFile(42)

      expect(mockApi.get).toHaveBeenCalledWith('/download/file/42', {
        responseType: 'blob',
      })
    })

    it('uses provided filename', async () => {
      const mockBlob = new Blob(['file content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadFile(42, 'movie.mp4')

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'movie.mp4')
    })

    it('uses default filename when not provided', async () => {
      const mockBlob = new Blob(['file content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadFile(42)

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'file-42')
    })

    it('propagates errors on download failure', async () => {
      mockApi.get.mockRejectedValue(new Error('Download failed'))

      await expect(downloadApi.downloadFile(1)).rejects.toThrow('Download failed')
    })
  })

  describe('downloadArchive', () => {
    it('calls POST /download/archive with body and blob responseType', async () => {
      const mockBlob = new Blob(['archive content'])
      mockApi.post.mockResolvedValue({ data: mockBlob })

      const request = { file_ids: [1, 2, 3], format: 'zip' as const }
      await downloadApi.downloadArchive(request)

      expect(mockApi.post).toHaveBeenCalledWith('/download/archive', request, {
        responseType: 'blob',
      })
    })

    it('uses tar.gz format in default filename', async () => {
      const mockBlob = new Blob(['archive content'])
      mockApi.post.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadArchive({ file_ids: [1], format: 'tar.gz' })

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'archive.tar.gz')
    })

    it('defaults to zip format in filename', async () => {
      const mockBlob = new Blob(['archive content'])
      mockApi.post.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadArchive({ file_ids: [1] })

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'archive.zip')
    })
  })

  describe('downloadDirectory', () => {
    it('calls GET /download/directory/:path with blob responseType', async () => {
      const mockBlob = new Blob(['dir content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadDirectory('movies/action')

      expect(mockApi.get).toHaveBeenCalledWith('/download/directory/movies/action', {
        responseType: 'blob',
      })
    })

    it('extracts directory name for default filename', async () => {
      const mockBlob = new Blob(['dir content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadDirectory('/movies/action')

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'action.zip')
    })
  })

  describe('downloadEntity', () => {
    it('calls GET /entities/:entityId/download with blob responseType', async () => {
      const mockBlob = new Blob(['entity content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadEntity(10)

      expect(mockApi.get).toHaveBeenCalledWith('/entities/10/download', {
        responseType: 'blob',
      })
    })

    it('uses provided filename', async () => {
      const mockBlob = new Blob(['entity content'])
      mockApi.get.mockResolvedValue({ data: mockBlob })

      await downloadApi.downloadEntity(10, 'my-entity.zip')

      const anchor = document.createElement('a')
      expect(anchor.setAttribute).toHaveBeenCalledWith('download', 'my-entity.zip')
    })
  })
})
