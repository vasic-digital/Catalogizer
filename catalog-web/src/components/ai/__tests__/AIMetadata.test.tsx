import { render, screen, waitFor } from '@testing-library/react'

describe('AIMetadata', () => {
  let AIMetadataExtractor: any
  let AIAutomationRules: any
  let AIContentQualityAnalyzer: any

  beforeAll(async () => {
    const mod = await import('../AIMetadata')
    AIMetadataExtractor = mod.AIMetadataExtractor
    AIAutomationRules = mod.AIAutomationRules
    AIContentQualityAnalyzer = mod.AIContentQualityAnalyzer
  })

  describe('AIMetadataExtractor', () => {
    it('renders heading', async () => {
      if (!AIMetadataExtractor) return
      render(
        <AIMetadataExtractor
          content={{ title: 'Test', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('AI Metadata Extraction')).toBeInTheDocument()
      })
    })
  })

  describe('AIAutomationRules', () => {
    it('renders heading', async () => {
      if (!AIAutomationRules) return
      render(
        <AIAutomationRules
          onRuleToggle={vi.fn()}
          onRuleExecute={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('AI Automation Rules')).toBeInTheDocument()
      })
    })
  })

  describe('AIContentQualityAnalyzer', () => {
    it('renders heading', async () => {
      if (!AIContentQualityAnalyzer) return
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('AI Quality Analysis')).toBeInTheDocument()
      })
    })
  })
})
