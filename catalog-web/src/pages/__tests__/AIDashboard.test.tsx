import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import AIDashboard from '../AIDashboard'

// Mock AI component modules
vi.mock('../../components/ai/AIComponents', () => ({
  AICollectionSuggestions: ({ onSuggestionAccept }: any) => (
    <div data-testid="ai-suggestions">AI Collection Suggestions</div>
  ),
  AINaturalSearch: ({ onSearch }: any) => (
    <div data-testid="ai-search">AI Natural Search</div>
  ),
  AIContentCategorizer: ({ onCategorizationComplete }: any) => (
    <div data-testid="ai-categorizer">AI Content Categorizer</div>
  ),
  AIService: {},
}))

vi.mock('../../components/ai/AIAnalytics', () => ({
  AIUserBehaviorAnalytics: ({ onActionImplement }: any) => (
    <div data-testid="ai-behavior">AI User Behavior Analytics</div>
  ),
  AIPredictions: ({ onPredictionAction }: any) => (
    <div data-testid="ai-predictions">AI Predictions</div>
  ),
  AISmartOrganization: ({ onSuggestionApply }: any) => (
    <div data-testid="ai-organization">AI Smart Organization</div>
  ),
  AIAnalyticsService: {},
}))

vi.mock('../../components/ai/AIMetadata', () => ({
  AIMetadataExtractor: ({ onMetadataExtracted }: any) => (
    <div data-testid="ai-metadata">AI Metadata Extractor</div>
  ),
  AIAutomationRules: ({ onRuleToggle, onRuleExecute }: any) => (
    <div data-testid="ai-automation">AI Automation Rules</div>
  ),
  AIContentQualityAnalyzer: ({ onQualityImprovement }: any) => (
    <div data-testid="ai-quality">AI Content Quality Analyzer</div>
  ),
  AIMetadataService: {},
}))

describe('AIDashboard Page', () => {
  it('renders AI Dashboard heading', () => {
    render(<AIDashboard />)
    expect(screen.getByText('AI Dashboard')).toBeInTheDocument()
  })

  it('renders description text', () => {
    render(<AIDashboard />)
    expect(
      screen.getByText(/Advanced AI-powered features for intelligent content management/)
    ).toBeInTheDocument()
  })

  it('renders navigation tabs', () => {
    render(<AIDashboard />)
    expect(screen.getByText('Overview')).toBeInTheDocument()
    expect(screen.getByText('AI Suggestions')).toBeInTheDocument()
    expect(screen.getByText('Natural Search')).toBeInTheDocument()
    expect(screen.getByText('Analytics')).toBeInTheDocument()
    expect(screen.getByText('Metadata')).toBeInTheDocument()
    expect(screen.getByText('Automation')).toBeInTheDocument()
  })

  it('shows overview section by default', () => {
    render(<AIDashboard />)
    expect(screen.getByText('Processed Items')).toBeInTheDocument()
    expect(screen.getByText('AI Accuracy')).toBeInTheDocument()
    expect(screen.getByText('Time Saved')).toBeInTheDocument()
  })

  it('displays metrics values', () => {
    render(<AIDashboard />)
    expect(screen.getByText('2,847')).toBeInTheDocument()
    expect(screen.getByText('92%')).toBeInTheDocument()
    expect(screen.getByText('12.5 hours')).toBeInTheDocument()
  })

  it('switches to suggestions tab on click', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('AI Suggestions'))

    expect(screen.getByText('AI-Powered Suggestions')).toBeInTheDocument()
    expect(screen.getByTestId('ai-suggestions')).toBeInTheDocument()
  })

  it('switches to search tab on click', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('Natural Search'))

    expect(screen.getByText('Natural Language Search')).toBeInTheDocument()
    expect(screen.getByTestId('ai-search')).toBeInTheDocument()
  })

  it('switches to analytics tab on click', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('Analytics'))

    expect(screen.getByText('AI Analytics')).toBeInTheDocument()
    expect(screen.getByTestId('ai-behavior')).toBeInTheDocument()
    expect(screen.getByTestId('ai-predictions')).toBeInTheDocument()
  })

  it('switches to metadata tab on click', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('Metadata'))

    expect(screen.getByText('AI Metadata Services')).toBeInTheDocument()
    expect(screen.getByTestId('ai-metadata')).toBeInTheDocument()
    expect(screen.getByTestId('ai-quality')).toBeInTheDocument()
  })

  it('switches to automation tab on click', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('Automation'))

    expect(screen.getByText('AI Automation')).toBeInTheDocument()
    expect(screen.getByTestId('ai-automation')).toBeInTheDocument()
  })

  it('shows quick actions in overview', () => {
    render(<AIDashboard />)
    expect(screen.getByText('Quick Actions')).toBeInTheDocument()
    expect(screen.getByText('Get Suggestions')).toBeInTheDocument()
    expect(screen.getByText('AI Search')).toBeInTheDocument()
    expect(screen.getByText('View Analytics')).toBeInTheDocument()
    expect(screen.getByText('Manage Rules')).toBeInTheDocument()
  })

  it('navigates to suggestions from quick action button', async () => {
    const user = userEvent.setup()
    render(<AIDashboard />)

    await user.click(screen.getByText('Get Suggestions'))

    expect(screen.getByText('AI-Powered Suggestions')).toBeInTheDocument()
  })
})
