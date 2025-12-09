import React, { useState, useRef } from 'react'
import { motion } from 'framer-motion'
import { X, Upload, FileText, Globe, AlertCircle, CheckCircle } from 'lucide-react'
import { subtitleApi } from '@/lib/subtitleApi'
import { COMMON_LANGUAGES, SUBTITLE_FORMATS } from '@/types/subtitles'

interface SubtitleUploadModalProps {
  isOpen: boolean
  onClose: () => void
  mediaId: number
  mediaTitle?: string
  onUploadSuccess?: () => void
}

export const SubtitleUploadModal: React.FC<SubtitleUploadModalProps> = ({
  isOpen,
  onClose,
  mediaId,
  mediaTitle,
  onUploadSuccess
}) => {
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [selectedLanguage, setSelectedLanguage] = useState('')
  const [selectedFormat, setSelectedFormat] = useState('')
  const [isUploading, setIsUploading] = useState(false)
  const [uploadResult, setUploadResult] = useState<{ success: boolean; message: string } | null>(null)
  const [dragActive, setDragActive] = useState(false)
  
  const fileInputRef = useRef<HTMLInputElement>(null)

  const handleFileSelect = (file: File) => {
    if (file && (file.name.endsWith('.srt') || file.name.endsWith('.vtt') || 
                 file.name.endsWith('.ass') || file.name.endsWith('.ssa') || 
                 file.name.endsWith('.sub'))) {
      setSelectedFile(file)
      
      // Auto-detect format from file extension
      const extension = file.name.split('.').pop()?.toLowerCase()
      if (extension && SUBTITLE_FORMATS.includes(extension as any)) {
        setSelectedFormat(extension)
      }
      
      setUploadResult(null)
    } else {
      setUploadResult({
        success: false,
        message: 'Invalid file format. Please select a subtitle file (.srt, .vtt, .ass, .ssa, .sub)'
      })
    }
  }

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
    
    const files = e.dataTransfer.files
    if (files.length > 0) {
      handleFileSelect(files[0])
    }
  }

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(true)
  }

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault()
    setDragActive(false)
  }

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files
    if (files && files.length > 0) {
      handleFileSelect(files[0])
    }
  }

  const handleUpload = async () => {
    if (!selectedFile || !selectedLanguage) {
      setUploadResult({
        success: false,
        message: 'Please select a file and language'
      })
      return
    }

    setIsUploading(true)
    try {
      const result = await subtitleApi.uploadSubtitle(
        mediaId,
        selectedFile,
        selectedLanguage,
        selectedFormat || undefined
      )
      
      if (result.success) {
        setUploadResult({
          success: true,
          message: result.message || 'Subtitle uploaded successfully!'
        })
        onUploadSuccess?.()
        
        // Reset form after successful upload
        setTimeout(() => {
          setSelectedFile(null)
          setSelectedLanguage('')
          setSelectedFormat('')
          setUploadResult(null)
          if (fileInputRef.current) {
            fileInputRef.current.value = ''
          }
        }, 2000)
      } else {
        setUploadResult({
          success: false,
          message: result.error || 'Upload failed'
        })
      }
    } catch (error) {
      setUploadResult({
        success: false,
        message: error instanceof Error ? error.message : 'Upload failed'
      })
    } finally {
      setIsUploading(false)
    }
  }

  const resetForm = () => {
    setSelectedFile(null)
    setSelectedLanguage('')
    setSelectedFormat('')
    setUploadResult(null)
    if (fileInputRef.current) {
      fileInputRef.current.value = ''
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  if (!isOpen) return null

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      exit={{ opacity: 0 }}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <motion.div
        initial={{ scale: 0.95, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.95, opacity: 0 }}
        className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-lg w-full p-6"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex justify-between items-center mb-6">
          <h2 className="text-xl font-semibold flex items-center gap-2">
            <Upload className="w-5 h-5" />
            Upload Subtitle
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {mediaTitle && (
          <div className="mb-4 text-sm text-gray-600 dark:text-gray-400">
            Uploading for <span className="font-medium">{mediaTitle}</span>
          </div>
        )}

        {/* File Drop Area */}
        <div
          className={`border-2 border-dashed rounded-lg p-6 mb-4 transition-colors ${
            dragActive 
              ? 'border-blue-400 bg-blue-50 dark:bg-blue-900/20' 
              : 'border-gray-300 dark:border-gray-600 hover:border-gray-400 dark:hover:border-gray-500'
          }`}
          onDrop={handleDrop}
          onDragOver={handleDragOver}
          onDragLeave={handleDragLeave}
        >
          <div className="text-center">
            <FileText className="w-12 h-12 mx-auto mb-3 text-gray-400" />
            <p className="text-gray-600 dark:text-gray-400 mb-2">
              Drag and drop your subtitle file here, or click to browse
            </p>
            <input
              ref={fileInputRef}
              type="file"
              accept=".srt,.vtt,.ass,.ssa,.sub"
              onChange={handleFileInputChange}
              className="hidden"
            />
            <button
              onClick={() => fileInputRef.current?.click()}
              className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
            >
              Browse Files
            </button>
          </div>
        </div>

        {/* Selected File Info */}
        {selectedFile && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className="bg-gray-50 dark:bg-gray-700 rounded-lg p-3 mb-4"
          >
            <div className="flex items-center gap-3">
              <FileText className="w-8 h-8 text-blue-500" />
              <div className="flex-1">
                <div className="font-medium text-gray-900 dark:text-white">
                  {selectedFile.name}
                </div>
                <div className="text-sm text-gray-600 dark:text-gray-400">
                  {formatFileSize(selectedFile.size)}
                </div>
              </div>
              <button
                onClick={resetForm}
                className="p-1 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          </motion.div>
        )}

        {/* Language Selection */}
        <div className="mb-4">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            <Globe className="w-4 h-4 inline mr-1" />
            Language
          </label>
          <select
            value={selectedLanguage}
            onChange={(e) => setSelectedLanguage(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="">Select language...</option>
            {COMMON_LANGUAGES.map((lang) => (
              <option key={lang.code} value={lang.code}>
                {lang.native_name} ({lang.name})
              </option>
            ))}
          </select>
        </div>

        {/* Format Selection */}
        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
            Format
          </label>
          <select
            value={selectedFormat}
            onChange={(e) => setSelectedFormat(e.target.value)}
            className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="">Auto-detect...</option>
            {SUBTITLE_FORMATS.map((format) => (
              <option key={format} value={format}>
                {format.toUpperCase()}
              </option>
            ))}
          </select>
        </div>

        {/* Upload Result */}
        {uploadResult && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className={`rounded-lg p-3 mb-6 flex items-center gap-3 ${
              uploadResult.success
                ? 'bg-green-100 dark:bg-green-900/20 text-green-800 dark:text-green-200'
                : 'bg-red-100 dark:bg-red-900/20 text-red-800 dark:text-red-200'
            }`}
          >
            {uploadResult.success ? (
              <CheckCircle className="w-5 h-5" />
            ) : (
              <AlertCircle className="w-5 h-5" />
            )}
            <span className="text-sm">{uploadResult.message}</span>
          </motion.div>
        )}

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={handleUpload}
            disabled={isUploading || !selectedFile || !selectedLanguage}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed flex items-center justify-center gap-2"
          >
            {isUploading ? (
              <>
                <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                Uploading...
              </>
            ) : (
              <>
                <Upload className="w-4 h-4" />
                Upload Subtitle
              </>
            )}
          </button>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            Cancel
          </button>
        </div>

        <div className="mt-4 text-xs text-gray-500 dark:text-gray-500 text-center">
          Supported formats: SRT, VTT, ASS, SSA, SUB
        </div>
      </motion.div>
    </motion.div>
  )
}