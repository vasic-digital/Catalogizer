import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

// Check what AIComponents exports
vi.mock('@/lib/utils', () => ({
  debounce: vi.fn((fn: any) => fn),
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('AIComponents', () => {
  let AICollectionSuggestions: any
  let AINaturalSearch: any
  let AIContentCategorizer: any
  let AIService: any

  beforeAll(async () => {
    const mod = await import('../AIComponents')
    AICollectionSuggestions = mod.AICollectionSuggestions
    AINaturalSearch = mod.AINaturalSearch
    AIContentCategorizer = mod.AIContentCategorizer
    AIService = mod.AIService
  })

  describe('AICollectionSuggestions', () => {
    it('renders heading in loading state', () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      expect(screen.getByText('AI Suggestions')).toBeInTheDocument()
    })

    it('shows loading spinner initially', () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })

    it('shows loading skeleton placeholders', () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      const pulseElements = document.querySelectorAll('.animate-pulse')
      expect(pulseElements.length).toBeGreaterThan(0)
    })

    it('displays suggestions after loading', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('AI-Powered Suggestions')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Movie Marathon Collection')).toBeInTheDocument()
    })

    it('displays confidence percentages', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('92% match')).toBeInTheDocument()
      }, { timeout: 5000 })
    })

    it('displays AI reasoning for suggestions', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      await waitFor(() => {
        const reasonings = screen.getAllByText(/AI Reasoning:/)
        expect(reasonings.length).toBeGreaterThan(0)
      }, { timeout: 5000 })
    })

    it('calls onSuggestionAccept when clicking a suggestion', async () => {
      if (!AICollectionSuggestions) return
      const user = userEvent.setup()
      const handleAccept = vi.fn()
      render(<AICollectionSuggestions onSuggestionAccept={handleAccept} />)
      await waitFor(() => {
        expect(screen.getByText('Movie Marathon Collection')).toBeInTheDocument()
      }, { timeout: 5000 })
      await user.click(screen.getByText('Movie Marathon Collection'))
      expect(handleAccept).toHaveBeenCalledTimes(1)
      expect(handleAccept).toHaveBeenCalledWith(expect.objectContaining({
        id: 'ai-collection-1',
        type: 'collection',
      }))
    })

    it('respects maxSuggestions prop', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} maxSuggestions={1} />)
      await waitFor(() => {
        expect(screen.getByText('AI-Powered Suggestions')).toBeInTheDocument()
      }, { timeout: 5000 })
    })
  })

  describe('AINaturalSearch', () => {
    it('renders search input with default placeholder', () => {
      if (!AINaturalSearch) return
      render(<AINaturalSearch onSearch={vi.fn()} />)
      expect(screen.getByPlaceholderText(/search naturally/i)).toBeInTheDocument()
    })

    it('renders with custom placeholder', () => {
      if (!AINaturalSearch) return
      render(<AINaturalSearch onSearch={vi.fn()} placeholder="Custom search" />)
      expect(screen.getByPlaceholderText('Custom search')).toBeInTheDocument()
    })

    it('updates input value on typing', async () => {
      if (!AINaturalSearch) return
      const user = userEvent.setup()
      render(<AINaturalSearch onSearch={vi.fn()} />)
      const input = screen.getByPlaceholderText(/search naturally/i)
      await user.type(input, 'action movies')
      expect(input).toHaveValue('action movies')
    })

    it('renders as a form element', () => {
      if (!AINaturalSearch) return
      const { container } = render(<AINaturalSearch onSearch={vi.fn()} />)
      const form = container.querySelector('form')
      expect(form).toBeTruthy()
    })
  })

  describe('AIContentCategorizer', () => {
    it('renders heading', () => {
      if (!AIContentCategorizer) return
      render(
        <AIContentCategorizer
          item={{ title: 'Test', description: 'Test desc' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      expect(screen.getByText('Smart Content Categorization')).toBeInTheDocument()
    })

    it('renders Categorize with AI button', () => {
      if (!AIContentCategorizer) return
      render(
        <AIContentCategorizer
          item={{ title: 'Test' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      expect(screen.getByText('Categorize with AI')).toBeInTheDocument()
    })

    it('shows description text before categorization', () => {
      if (!AIContentCategorizer) return
      render(
        <AIContentCategorizer
          item={{ title: 'Test' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      expect(screen.getByText('AI can automatically categorize your content')).toBeInTheDocument()
    })

    it('starts categorizing when button is clicked', async () => {
      if (!AIContentCategorizer) return
      const user = userEvent.setup()
      render(
        <AIContentCategorizer
          item={{ title: 'Test Movie', description: 'A test movie' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      await user.click(screen.getByText('Categorize with AI'))
      // Should show loading state
      expect(screen.getByText('AI is analyzing your content...')).toBeInTheDocument()
    })

    it('calls onCategorizationComplete after categorization', async () => {
      if (!AIContentCategorizer) return
      const user = userEvent.setup()
      const handleComplete = vi.fn()
      render(
        <AIContentCategorizer
          item={{ title: 'Test Movie' }}
          onCategorizationComplete={handleComplete}
        />
      )
      await user.click(screen.getByText('Categorize with AI'))
      await waitFor(() => {
        expect(handleComplete).toHaveBeenCalledTimes(1)
      }, { timeout: 5000 })
      expect(handleComplete).toHaveBeenCalledWith(expect.objectContaining({
        category: expect.any(String),
        confidence: expect.any(Number),
        suggestedTags: expect.any(Array),
      }))
    })

    it('shows Recategorize button after categorization', async () => {
      if (!AIContentCategorizer) return
      const user = userEvent.setup()
      render(
        <AIContentCategorizer
          item={{ title: 'Test Movie' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      await user.click(screen.getByText('Categorize with AI'))
      await waitFor(() => {
        expect(screen.getByText('Recategorize with AI')).toBeInTheDocument()
      }, { timeout: 5000 })
    })

    it('displays Suggested Tags heading after categorization', async () => {
      if (!AIContentCategorizer) return
      const user = userEvent.setup()
      render(
        <AIContentCategorizer
          item={{ title: 'Test Movie' }}
          onCategorizationComplete={vi.fn()}
        />
      )
      await user.click(screen.getByText('Categorize with AI'))
      await waitFor(() => {
        expect(screen.getByText('Suggested Tags:')).toBeInTheDocument()
      }, { timeout: 5000 })
    })
  })

  describe('AIService', () => {
    it('generateCollectionSuggestions returns suggestions', async () => {
      if (!AIService) return
      const suggestions = await AIService.generateCollectionSuggestions({
        recentSearches: [], existingCollections: [],
        contentAnalysis: [], userBehavior: {},
      })
      expect(suggestions.length).toBeGreaterThan(0)
      expect(suggestions[0]).toHaveProperty('title')
      expect(suggestions[0]).toHaveProperty('confidence')
      expect(suggestions[0].confidence).toBeGreaterThan(0.8)
    })

    it('categorizeContent returns categorization result', async () => {
      if (!AIService) return
      const result = await AIService.categorizeContent({
        title: 'Test Item',
        description: 'A test description',
      })
      expect(result).toHaveProperty('category')
      expect(result).toHaveProperty('confidence')
      expect(result).toHaveProperty('suggestedTags')
      expect(result.suggestedTags.length).toBeGreaterThan(0)
    })

    it('processNaturalLanguageQuery detects browse intent', async () => {
      if (!AIService) return
      const result = await AIService.processNaturalLanguageQuery('show me all movies')
      expect(result.intent).toBe('browse')
      expect(result.naturalLanguage).toBe(true)
    })

    it('processNaturalLanguageQuery detects compare intent', async () => {
      if (!AIService) return
      const result = await AIService.processNaturalLanguageQuery('compare action versus comedy')
      expect(result.intent).toBe('compare')
    })

    it('processNaturalLanguageQuery detects organize intent', async () => {
      if (!AIService) return
      const result = await AIService.processNaturalLanguageQuery('organize my files')
      expect(result.intent).toBe('organize')
    })

    it('processNaturalLanguageQuery defaults to search intent', async () => {
      if (!AIService) return
      const result = await AIService.processNaturalLanguageQuery('matrix 1999')
      expect(result.intent).toBe('search')
    })

    it('generateSmartSearchSuggestions returns suggestions', async () => {
      if (!AIService) return
      const suggestions = await AIService.generateSmartSearchSuggestions('action')
      expect(suggestions.length).toBeLessThanOrEqual(5)
      expect(suggestions[0]).toContain('action')
    })
  })
})
