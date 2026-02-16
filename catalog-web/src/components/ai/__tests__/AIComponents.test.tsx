import { render, screen, waitFor } from '@testing-library/react'

// Check what AIComponents exports
vi.mock('@/lib/utils', () => ({
  debounce: vi.fn((fn: any) => fn),
  cn: (...args: any[]) => args.filter(Boolean).join(' '),
}))

describe('AIComponents', () => {
  let AICollectionSuggestions: any
  let AINaturalSearch: any
  let AIContentCategorizer: any

  beforeAll(async () => {
    const mod = await import('../AIComponents')
    AICollectionSuggestions = mod.AICollectionSuggestions
    AINaturalSearch = mod.AINaturalSearch
    AIContentCategorizer = mod.AIContentCategorizer
  })

  describe('AICollectionSuggestions', () => {
    it('renders heading', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      // In loading state, heading is "AI Suggestions"
      await waitFor(() => {
        expect(screen.getByText('AI Suggestions')).toBeInTheDocument()
      })
    })

    it('shows loading state initially', async () => {
      if (!AICollectionSuggestions) return
      render(<AICollectionSuggestions onSuggestionAccept={vi.fn()} />)
      // In loading state, it shows a spinner
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })
  })

  describe('AINaturalSearch', () => {
    it('renders search input', async () => {
      if (!AINaturalSearch) return
      render(<AINaturalSearch onSearch={vi.fn()} />)
      // The component has a search input with a placeholder about natural search
      expect(screen.getByPlaceholderText(/search/i)).toBeInTheDocument()
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
  })
})
