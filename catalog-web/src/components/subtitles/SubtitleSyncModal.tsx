import React, { useState } from 'react'
import { motion } from 'framer-motion'
import { X, CheckCircle, AlertCircle, Clock, RefreshCw, Play, Pause } from 'lucide-react'
import { subtitleApi } from '@/lib/subtitleApi'
import type { SubtitleSyncVerificationResponse } from '@/types/subtitles'

interface SubtitleSyncModalProps {
  isOpen: boolean
  onClose: () => void
  subtitleId: string
  mediaId: number
  subtitleLanguage?: string
}

export const SubtitleSyncModal: React.FC<SubtitleSyncModalProps> = ({
  isOpen,
  onClose,
  subtitleId,
  mediaId,
  subtitleLanguage
}) => {
  const [isVerifying, setIsVerifying] = useState(false)
  const [verificationResult, setVerificationResult] = useState<SubtitleSyncVerificationResponse | null>(null)
  const [sampleDuration, setSampleDuration] = useState(60) // seconds
  const [sensitivity, setSensitivity] = useState(5) // 1-10

  const handleVerify = async () => {
    setIsVerifying(true)
    try {
      const result = await subtitleApi.verifySync(subtitleId, mediaId, {
        sample_duration: sampleDuration,
        sensitivity,
      })
      setVerificationResult(result)
    } catch (error) {
      console.error('Verification failed:', error)
      setVerificationResult({
        success: false,
        status: 'unusable',
        error: error instanceof Error ? error.message : 'Verification failed'
      })
    } finally {
      setIsVerifying(false)
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'perfect':
        return 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-900/20'
      case 'good':
        return 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-900/20'
      case 'acceptable':
        return 'text-yellow-600 bg-yellow-100 dark:text-yellow-400 dark:bg-yellow-900/20'
      case 'poor':
        return 'text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-900/20'
      case 'unusable':
        return 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-900/20'
      default:
        return 'text-gray-600 bg-gray-100 dark:text-gray-400 dark:bg-gray-900/20'
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'perfect':
      case 'good':
        return <CheckCircle className="w-5 h-5" />
      default:
        return <AlertCircle className="w-5 h-5" />
    }
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
            <Clock className="w-5 h-5" />
            Verify Subtitle Sync
          </h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {subtitleLanguage && (
          <div className="mb-4 text-sm text-gray-600 dark:text-gray-400">
            Verifying sync for <span className="font-medium">{subtitleLanguage}</span> subtitle
          </div>
        )}

        {/* Settings */}
        <div className="space-y-4 mb-6">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Sample Duration (seconds)
            </label>
            <div className="flex items-center gap-3">
              <input
                type="range"
                min="30"
                max="300"
                step="30"
                value={sampleDuration}
                onChange={(e) => setSampleDuration(Number(e.target.value))}
                className="flex-1"
              />
              <span className="text-sm font-medium w-20 text-center">
                {sampleDuration}s
              </span>
            </div>
            <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
              Longer durations provide more accurate results but take longer
            </p>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              Sensitivity
            </label>
            <div className="flex items-center gap-3">
              <input
                type="range"
                min="1"
                max="10"
                value={sensitivity}
                onChange={(e) => setSensitivity(Number(e.target.value))}
                className="flex-1"
              />
              <span className="text-sm font-medium w-8 text-center">
                {sensitivity}
              </span>
            </div>
            <p className="text-xs text-gray-500 dark:text-gray-500 mt-1">
              Higher sensitivity detects smaller sync issues but may produce false positives
            </p>
          </div>
        </div>

        {/* Verification Result */}
        {verificationResult && (
          <motion.div
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            className={`rounded-lg p-4 mb-6 ${getStatusColor(verificationResult.status)}`}
          >
            <div className="flex items-center gap-3 mb-3">
              {getStatusIcon(verificationResult.status)}
              <div>
                <div className="font-medium capitalize">
                  {verificationResult.status} Sync
                </div>
                <div className="text-sm opacity-75">
                  {verificationResult.message}
                </div>
              </div>
            </div>

            {verificationResult.success && (
              <div className="grid grid-cols-2 gap-4 text-sm">
                {verificationResult.sync_offset !== undefined && (
                  <div>
                    <span className="opacity-75">Sync Offset:</span>
                    <div className="font-medium">
                      {verificationResult.sync_offset > 0 ? '+' : ''}{verificationResult.sync_offset}ms
                    </div>
                  </div>
                )}
                {verificationResult.confidence !== undefined && (
                  <div>
                    <span className="opacity-75">Confidence:</span>
                    <div className="font-medium">
                      {(verificationResult.confidence * 100).toFixed(1)}%
                    </div>
                  </div>
                )}
                {verificationResult.sync_score !== undefined && (
                  <div>
                    <span className="opacity-75">Sync Score:</span>
                    <div className="font-medium">
                      {verificationResult.sync_score.toFixed(2)}
                    </div>
                  </div>
                )}
              </div>
            )}

            {verificationResult.error && (
              <div className="mt-3 text-sm">
                <span className="opacity-75">Error:</span>
                <div className="font-medium">
                  {verificationResult.error}
                </div>
              </div>
            )}
          </motion.div>
        )}

        {/* Actions */}
        <div className="flex gap-3">
          <button
            onClick={handleVerify}
            disabled={isVerifying}
            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 flex items-center justify-center gap-2"
          >
            {isVerifying ? (
              <>
                <RefreshCw className="w-4 h-4 animate-spin" />
                Verifying...
              </>
            ) : (
              <>
                <Play className="w-4 h-4" />
                Start Verification
              </>
            )}
          </button>
          <button
            onClick={onClose}
            className="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600"
          >
            Close
          </button>
        </div>

        <div className="mt-4 text-xs text-gray-500 dark:text-gray-500 text-center">
          Verification analyzes audio patterns to detect subtitle synchronization issues
        </div>
      </motion.div>
    </motion.div>
  )
}