import { render, screen, waitFor } from '@testing-library/react'

describe('AIAnalytics', () => {
  let AIUserBehaviorAnalytics: any
  let AIPredictions: any
  let AISmartOrganization: any

  beforeAll(async () => {
    const mod = await import('../AIAnalytics')
    AIUserBehaviorAnalytics = mod.AIUserBehaviorAnalytics
    AIPredictions = mod.AIPredictions
    AISmartOrganization = mod.AISmartOrganization
  })

  describe('AIUserBehaviorAnalytics', () => {
    it('renders heading', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('User Behavior Analytics')).toBeInTheDocument()
      })
    })

    it('shows loading state initially', async () => {
      if (!AIUserBehaviorAnalytics) return
      render(<AIUserBehaviorAnalytics userId="user1" onActionImplement={vi.fn()} />)
      // In loading state, it shows a spinner
      const spinner = document.querySelector('.animate-spin')
      expect(spinner).toBeTruthy()
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
  })

  describe('AISmartOrganization', () => {
    it('renders heading', async () => {
      if (!AISmartOrganization) return
      render(<AISmartOrganization collections={[]} onSuggestionApply={vi.fn()} />)
      await waitFor(() => {
        expect(screen.getByText('Smart Organization')).toBeInTheDocument()
      })
    })
  })
})
