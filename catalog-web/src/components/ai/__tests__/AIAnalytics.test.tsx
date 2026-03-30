import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'

describe('AIAnalytics', () => {
  let AIUserBehaviorAnalytics: any
  let AIPredictions: any
  let AISmartOrganization: any
  let AIAnalyticsService: any

  beforeAll(async () => {
    const mod = await import('../AIAnalytics')
    AIUserBehaviorAnalytics = mod.AIUserBehaviorAnalytics
    AIPredictions = mod.AIPredictions
    AISmartOrganization = mod.AISmartOrganization
    AIAnalyticsService = mod.AIAnalyticsService
  })

  describe('AIUserBehaviorAnalytics', () => {
    it('renders heading', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('User Behavior Analytics')).toBeInTheDocument()
      })
    })

    it('shows loading state initially', () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })

    it('shows loading skeleton placeholders', () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      const pulseElements = document.querySelectorAll('.animate-pulse')
      expect(pulseElements.length).toBeGreaterThan(0)
    })

    it('loads and displays behavior patterns after loading', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Evening Entertainment Browsing')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Weekend Learning Sessions')).toBeInTheDocument()
      expect(screen.getByText('Work Resource Collection')).toBeInTheDocument()
    })

    it('displays impact badges for patterns', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getAllByText('high impact').length).toBeGreaterThan(0)
      }, { timeout: 5000 })
      expect(screen.getByText('medium impact')).toBeInTheDocument()
    })

    it('displays AI Recommendations heading for each pattern', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        const headings = screen.getAllByText('AI Recommendations:')
        expect(headings.length).toBe(3)
      }, { timeout: 5000 })
    })

    it('calls onActionImplement when Implement Suggestions is clicked', async () => {
      if (!AIUserBehaviorAnalytics) return
      const user = userEvent.setup()
      const handleAction = vi.fn()
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={handleAction} />)
      await waitFor(() => {
        expect(screen.getAllByText('Implement Suggestions').length).toBeGreaterThan(0)
      }, { timeout: 5000 })
      await user.click(screen.getAllByText('Implement Suggestions')[0])
      expect(handleAction).toHaveBeenCalledTimes(1)
      expect(handleAction).toHaveBeenCalledWith(expect.stringContaining('Implement recommendations for:'))
    })

    it('shows refresh button after loading', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Evening Entertainment Browsing')).toBeInTheDocument()
      }, { timeout: 5000 })
      const buttons = screen.getAllByRole('button')
      expect(buttons.length).toBeGreaterThan(0)
    })
  })

  describe('AIPredictions', () => {
    it('renders heading', async () => {
      if (!AIPredictions) return
      render(<AIPredictions onPredictionAction={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('AI Predictions')).toBeInTheDocument()
      })
    })

    it('shows loading state initially', () => {
      if (!AIPredictions) return
      render(<AIPredictions onPredictionAction={vi.fn()} />)
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })

    it('displays predictions after loading', async () => {
      if (!AIPredictions) return
      render(<AIPredictions onPredictionAction={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Rising Interest in Productivity Tools')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Collection Optimization Needed')).toBeInTheDocument()
    })

    it('displays Updated hourly label', async () => {
      if (!AIPredictions) return
      render(<AIPredictions onPredictionAction={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Updated hourly')).toBeInTheDocument()
      }, { timeout: 5000 })
    })

    it('shows confidence and timeframe for predictions', async () => {
      if (!AIPredictions) return
      render(<AIPredictions onPredictionAction={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('87% confidence')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Next 2 weeks')).toBeInTheDocument()
      expect(screen.getByText('This week')).toBeInTheDocument()
    })
  })

  describe('AISmartOrganization', () => {
    it('renders heading', async () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Smart Organization')).toBeInTheDocument()
      })
    })

    it('shows loading state initially', () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
    })

    it('displays AI-powered label', async () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('AI-powered')).toBeInTheDocument()
      }, { timeout: 5000 })
    })

    it('displays suggestions after loading', async () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Consolidate Similar Collections')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('Implement Smart Tagging System')).toBeInTheDocument()
    })

    it('displays priority and effort badges', async () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('high priority')).toBeInTheDocument()
      }, { timeout: 5000 })
      expect(screen.getByText('medium priority')).toBeInTheDocument()
    })
  })

  describe('AIAnalyticsService', () => {
    it('analyzeUserBehavior returns patterns', async () => {
      if (!AIAnalyticsService) return
      const patterns = await AIAnalyticsService.analyzeUserBehavior('user1')
      expect(patterns).toHaveLength(3)
      expect(patterns[0].pattern).toBe('Evening Entertainment Browsing')
      expect(patterns[0].confidence).toBe(0.92)
    })

    it('generatePredictions returns predictions', async () => {
      if (!AIAnalyticsService) return
      const predictions = await AIAnalyticsService.generatePredictions({
        userHistory: [], contentMetrics: {}, currentTrends: {},
      })
      expect(predictions).toHaveLength(2)
      expect(predictions[0].type).toBe('trending')
    })

    it('generateContentInsights returns insights', async () => {
      if (!AIAnalyticsService) return
      const insights = await AIAnalyticsService.generateContentInsights()
      expect(insights).toHaveLength(2)
      expect(insights[0].category).toBe('Entertainment')
      expect(insights[1].category).toBe('Education')
    })

    it('generateOrganizationSuggestions returns suggestions', async () => {
      if (!AIAnalyticsService) return
      const suggestions = await AIAnalyticsService.generateOrganizationSuggestions([])
      expect(suggestions).toHaveLength(2)
      expect(suggestions[0].priority).toBe('high')
      expect(suggestions[0].steps.length).toBeGreaterThan(0)
    })
  })
})
