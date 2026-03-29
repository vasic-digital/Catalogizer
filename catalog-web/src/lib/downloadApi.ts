import api from './api'

export interface ArchiveRequest {
  file_ids: number[]
  format?: 'zip' | 'tar.gz'
}

function triggerBlobDownload(data: Blob, filename: string) {
  const url = window.URL.createObjectURL(data)
  const link = document.createElement('a')
  link.href = url
  link.setAttribute('download', filename)
  document.body.appendChild(link)
  link.click()
  link.parentNode?.removeChild(link)
  window.URL.revokeObjectURL(url)
}

export const downloadApi = {
  downloadFile: async (fileId: number, filename?: string): Promise<void> => {
    const response = await api.get(`/download/file/${fileId}`, {
      responseType: 'blob',
    })
    triggerBlobDownload(
      new Blob([response.data]),
      filename || `file-${fileId}`
    )
  },

  downloadArchive: async (body: ArchiveRequest, filename?: string): Promise<void> => {
    const response = await api.post('/download/archive', body, {
      responseType: 'blob',
    })
    triggerBlobDownload(
      new Blob([response.data]),
      filename || `archive.${body.format || 'zip'}`
    )
  },

  downloadDirectory: async (path: string, filename?: string): Promise<void> => {
    const response = await api.get(`/download/directory/${path}`, {
      responseType: 'blob',
    })
    const dirName = path.split('/').filter(Boolean).pop() || 'directory'
    triggerBlobDownload(
      new Blob([response.data]),
      filename || `${dirName}.zip`
    )
  },

  downloadEntity: async (entityId: number, filename?: string): Promise<void> => {
    const response = await api.get(`/entities/${entityId}/download`, {
      responseType: 'blob',
    })
    triggerBlobDownload(
      new Blob([response.data]),
      filename || `entity-${entityId}`
    )
  },
}
