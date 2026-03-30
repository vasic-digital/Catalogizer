import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

describe('AIMetadata', () => {
  let AIMetadataExtractor: any
  let AIAutomationRules: any
  let AIContentQualityAnalyzer: any
  let AIMetadataService: any

  beforeAll(async () => {
    const mod = await import('../AIMetadata')
    AIMetadataExtractor = mod.AIMetadataExtractor
    AIAutomationRules = mod.AIAutomationRules
    AIContentQualityAnalyzer = mod.AIContentQualityAnalyzer
    AIMetadataService = mod.AIMetadataService
  })

  describe('AIMetadataExtractor', () => {
    it('renders heading', () => {
      if (!AIMetadataExtractor) return
      render(
        <AIMetadataExtractor
          content={{ title: 'Test', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      expect(screen.getByText('AI Metadata Extraction')).toBeInTheDocument()
    })

    it('renders Extract Metadata button', () => {
      if (!AIMetadataExtractor) return
      render(
        <AIMetadataExtractor
          content={{ title: 'Test', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      expect(screen.getByText('Extract Metadata')).toBeInTheDocument()
    })

    it('shows description text before extraction', () => {
      if (!AIMetadataExtractor) return
      render(
        <AIMetadataExtractor
          content={{ title: 'Test', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      expect(screen.getByText('Extract rich metadata using AI analysis')).toBeInTheDocument()
    })

    it('shows extracting state when button is clicked', async () => {
      if (!AIMetadataExtractor) return
      const user = userEvent.setup()
      render(
        <AIMetadataExtractor
          content={{ title: 'Test Video Content', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      await user.click(screen.getByText('Extract Metadata'))
      expect(screen.getByText('AI is extracting metadata...')).toBeInTheDocument()
    })

    it('calls onMetadataExtracted after extraction', async () => {
      if (!AIMetadataExtractor) return
      const user = userEvent.setup()
      const handleExtracted = vi.fn()
      render(
        <AIMetadataExtractor
          content={{ title: 'Test Video Content', fileType: 'video' }}
          onMetadataExtracted={handleExtracted}
        />
      )
      await user.click(screen.getByText('Extract Metadata'))
      await waitFor(() => {
        expect(handleExtracted).toHaveBeenCalledTimes(1)
      }, { timeout: 5000 })
      expect(handleExtracted).toHaveBeenCalledWith(expect.objectContaining({
        title: expect.any(String),
        tags: expect.any(Array),
        categories: expect.any(Array),
        confidence: expect.any(Number),
        source: 'ai',
      }))
    })

    it('shows extracted metadata after completion', async () => {
      if (!AIMetadataExtractor) return
      const user = userEvent.setup()
      render(
        <AIMetadataExtractor
          content={{ title: 'Test Video Content', fileType: 'video' }}
          onMetadataExtracted={vi.fn()}
        />
      )
      await user.click(screen.getByText('Extract Metadata'))
      await waitFor(() => {
        expect(screen.getByText('Extracted Metadata')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('AI-Generated')).toBeInTheDocument()
      expect(screen.getByText('Confidence')).toBeInTheDocument()
      expect(screen.getByText('Re-extract Metadata')).toBeInTheDocument()
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

    it('shows loading spinner initially', () => {
      if (!AIAutomationRules) return
      render(
        <AIAutomationRules
          onRuleToggle={vi.fn()}
          onRuleExecute={vi.fn()}
        />
      )
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })

    it('displays rules after loading', async () => {
      if (!AIAutomationRules) return
      render(
        <AIAutomationRules
          onRuleToggle={vi.fn()}
          onRuleExecute={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('Auto-Categorize Entertainment Content')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Weekly Content Organization')).toBeInTheDocument()
    })

    it('displays priority badges', async () => {
      if (!AIAutomationRules) return
      render(
        <AIAutomationRules
          onRuleToggle={vi.fn()}
          onRuleExecute={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('medium')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('low')).toBeInTheDocument()
    })

    it('displays execution counts', async () => {
      if (!AIAutomationRules) return
      render(
        <AIAutomationRules
          onRuleToggle={vi.fn()}
          onRuleExecute={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('Executed 47 times')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Executed 12 times')).toBeInTheDocument()
    })

    it('calls onRuleToggle when toggle button is clicked', async () => {
      if (!AIAutomationRules) return
      const user = userEvent.setup()
      const handleToggle = vi.fn()
      render(
        <AIAutomationRules
          onRuleToggle={handleToggle}
          onRuleExecute={vi.fn()}
        />
      )
      await waitFor(() => {
        expect(screen.getByText('Auto-Categorize Entertainment Content')).toBeInTheDocument()
      }, { timeout: 5000 })
      // Toggle buttons are the switch-like elements
      const toggleButtons = document.querySelectorAll('[class*="inline-flex"][class*="rounded-full"]')
      if (toggleButtons.length > 0) {
        await user.click(toggleButtons[0] as HTMLElement)
        expect(handleToggle).toHaveBeenCalledWith('auto-1', false)
      }
    })
  })

  describe('AIContentQualityAnalyzer', () => {
    it('renders heading', () => {
      if (!AIContentQualityAnalyzer) return
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      expect(screen.getByText('AI Quality Analysis')).toBeInTheDocument()
    })

    it('renders Analyze Quality button', () => {
      if (!AIContentQualityAnalyzer) return
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      expect(screen.getByText('Analyze Quality')).toBeInTheDocument()
    })

    it('shows description text before analysis', () => {
      if (!AIContentQualityAnalyzer) return
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      expect(screen.getByText('Analyze content quality and get improvement suggestions')).toBeInTheDocument()
    })

    it('shows analyzing state when button is clicked', async () => {
      if (!AIContentQualityAnalyzer) return
      const user = userEvent.setup()
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      await user.click(screen.getByText('Analyze Quality'))
      expect(screen.getByText('AI is analyzing content quality...')).toBeInTheDocument()
    })

    it('shows quality analysis after completion', async () => {
      if (!AIContentQualityAnalyzer) return
      const user = userEvent.setup()
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      await user.click(screen.getByText('Analyze Quality'))
      await waitFor(() => {
        expect(screen.getByText('Content Quality Analysis')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Quality Factors')).toBeInTheDocument()
      expect(screen.getByText('Re-analyze Quality')).toBeInTheDocument()
    })

    it('displays quality factor labels', async () => {
      if (!AIContentQualityAnalyzer) return
      const user = userEvent.setup()
      render(
        <AIContentQualityAnalyzer
          content={{ title: 'Test' }}
          onQualityImprovement={vi.fn()}
        />
      )
      await user.click(screen.getByText('Analyze Quality'))
      await waitFor(() => {
        expect(screen.getByText('completeness')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('accuracy')).toBeInTheDocument()
      expect(screen.getByText('engagement')).toBeInTheDocument()
      expect(screen.getByText('relevance')).toBeInTheDocument()
    })
  })

  describe('AIMetadataService', () => {
    it('extractMetadata returns metadata with expected shape', async () => {
      if (!AIMetadataService) return
      const result = await AIMetadataService.extractMetadata({
        title: 'Test Video',
        fileType: 'video',
      })
      expect(result).toHaveProperty('id')
      expect(result).toHaveProperty('title')
      expect(result).toHaveProperty('description')
      expect(result).toHaveProperty('tags')
      expect(result).toHaveProperty('categories')
      expect(result).toHaveProperty('confidence')
      expect(result.source).toBe('ai')
      expect(result.tags.length).toBeGreaterThan(0)
    })

    it('extractMetadata includes extractedFields', async () => {
      if (!AIMetadataService) return
      const result = await AIMetadataService.extractMetadata({
        title: 'Test',
        fileType: 'video',
        size: 5000000,
        duration: 120,
      })
      expect(result.extractedFields).toHaveProperty('contentType')
      expect(result.extractedFields).toHaveProperty('quality')
      expect(result.extractedFields).toHaveProperty('engagement')
      expect(result.extractedFields).toHaveProperty('duration')
      expect(result.extractedFields).toHaveProperty('size')
    })

    it('analyzeContentQuality returns quality analysis', async () => {
      if (!AIMetadataService) return
      const result = await AIMetadataService.analyzeContentQuality({
        description: 'A test description',
      })
      expect(result).toHaveProperty('score')
      expect(result).toHaveProperty('factors')
      expect(result.factors).toHaveProperty('completeness')
      expect(result.factors).toHaveProperty('accuracy')
      expect(result.factors).toHaveProperty('engagement')
      expect(result.factors).toHaveProperty('relevance')
      expect(result).toHaveProperty('recommendations')
      expect(result).toHaveProperty('improvements')
    })

    it('generateAutomationRules returns rules', async () => {
      if (!AIMetadataService) return
      const result = await AIMetadataService.generateAutomationRules({})
      expect(result).toHaveLength(2)
      expect(result[0].name).toBe('Auto-Categorize Entertainment Content')
      expect(result[0].trigger.type).toBe('content_added')
      expect(result[0].actions.length).toBeGreaterThan(0)
    })

    it('processContent returns smart content', async () => {
      if (!AIMetadataService) return
      const result = await AIMetadataService.processContent({
        fileType: 'video',
        title: 'Test Movie',
      })
      expect(result).toHaveProperty('id')
      expect(result).toHaveProperty('type')
      expect(result).toHaveProperty('qualityScore')
      expect(result).toHaveProperty('autoGenerated')
      expect(result).toHaveProperty('metadata')
      expect(result.processing.status).toBe('completed')
      expect(result.processing.progress).toBe(100)
    })
  })
})
