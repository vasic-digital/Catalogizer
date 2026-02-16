import React, { Suspense } from 'react'
import { render, screen } from '@testing-library/react'
import { ComponentLoader, preloadComponent } from '../LazyComponents'

vi.mock('lucide-react', () => ({
  Loader2: (props: any) => (
    <span data-testid="loader" className={props.className}>
      Loading...
    </span>
  ),
}))

// Mock all lazy-loaded component imports
vi.mock('../../collections/CollectionTemplates', () => ({
  default: () => <div data-testid="collection-templates">CollectionTemplates</div>,
}))

vi.mock('../../collections/AdvancedSearch', () => ({
  default: () => <div data-testid="advanced-search">AdvancedSearch</div>,
}))

vi.mock('../../collections/CollectionAutomation', () => ({
  default: () => (
    <div data-testid="collection-automation">CollectionAutomation</div>
  ),
}))

vi.mock('../../collections/ExternalIntegrations', () => ({
  default: () => (
    <div data-testid="external-integrations">ExternalIntegrations</div>
  ),
}))

vi.mock('../../collections/SmartCollectionBuilder', () => ({
  SmartCollectionBuilder: () => (
    <div data-testid="smart-collection-builder">SmartCollectionBuilder</div>
  ),
}))

vi.mock('../../collections/CollectionAnalytics', () => ({
  CollectionAnalytics: () => (
    <div data-testid="collection-analytics">CollectionAnalytics</div>
  ),
}))

vi.mock('../../collections/BulkOperations', () => ({
  default: () => <div data-testid="bulk-operations">BulkOperations</div>,
}))

describe('ComponentLoader', () => {
  it('renders children content', () => {
    render(
      <ComponentLoader componentName="TestComponent">
        <div>Child Content</div>
      </ComponentLoader>
    )

    expect(screen.getByText('Child Content')).toBeInTheDocument()
  })

  it('wraps children in Suspense', () => {
    const { container } = render(
      <ComponentLoader componentName="TestComponent">
        <div>Wrapped Content</div>
      </ComponentLoader>
    )

    expect(container.textContent).toContain('Wrapped Content')
  })

  it('shows custom fallback when provided', () => {
    // Use a component that suspends
    const SuspendingComponent = React.lazy(
      () => new Promise(() => {}) // Never resolves
    )

    render(
      <ComponentLoader
        componentName="TestComponent"
        fallback={<div data-testid="custom-fallback">Custom Loading...</div>}
      >
        <SuspendingComponent />
      </ComponentLoader>
    )

    expect(screen.getByTestId('custom-fallback')).toBeInTheDocument()
    expect(screen.getByText('Custom Loading...')).toBeInTheDocument()
  })

  it('uses default fallback with loader icon', () => {
    const SuspendingComponent = React.lazy(
      () => new Promise(() => {}) // Never resolves
    )

    render(
      <ComponentLoader componentName="TestComponent">
        <SuspendingComponent />
      </ComponentLoader>
    )

    expect(screen.getByTestId('loader')).toBeInTheDocument()
  })

  it('renders without children', () => {
    const { container } = render(
      <ComponentLoader componentName="EmptyComponent" />
    )

    expect(container).toBeInTheDocument()
  })
})

describe('preloadComponent', () => {
  it('does not throw for known component names', () => {
    expect(() => preloadComponent('CollectionTemplates')).not.toThrow()
    expect(() => preloadComponent('AdvancedSearch')).not.toThrow()
    expect(() => preloadComponent('CollectionAutomation')).not.toThrow()
    expect(() => preloadComponent('ExternalIntegrations')).not.toThrow()
    expect(() => preloadComponent('SmartCollectionBuilder')).not.toThrow()
    expect(() => preloadComponent('CollectionAnalytics')).not.toThrow()
    expect(() => preloadComponent('BulkOperations')).not.toThrow()
  })

  it('handles unknown component names gracefully', () => {
    expect(() => preloadComponent('UnknownComponent')).not.toThrow()
  })

  it('handles empty string', () => {
    expect(() => preloadComponent('')).not.toThrow()
  })
})

describe('Lazy component exports', () => {
  it('exports CollectionTemplates as lazy component', async () => {
    const { CollectionTemplates } = await import('../LazyComponents')
    expect(CollectionTemplates).toBeDefined()
  })

  it('exports AdvancedSearch as lazy component', async () => {
    const { AdvancedSearch } = await import('../LazyComponents')
    expect(AdvancedSearch).toBeDefined()
  })

  it('exports CollectionAutomation as lazy component', async () => {
    const { CollectionAutomation } = await import('../LazyComponents')
    expect(CollectionAutomation).toBeDefined()
  })

  it('exports ExternalIntegrations as lazy component', async () => {
    const { ExternalIntegrations } = await import('../LazyComponents')
    expect(ExternalIntegrations).toBeDefined()
  })

  it('exports SmartCollectionBuilder as lazy component', async () => {
    const { SmartCollectionBuilder } = await import('../LazyComponents')
    expect(SmartCollectionBuilder).toBeDefined()
  })

  it('exports CollectionAnalytics as lazy component', async () => {
    const { CollectionAnalytics } = await import('../LazyComponents')
    expect(CollectionAnalytics).toBeDefined()
  })

  it('exports BulkOperations as lazy component', async () => {
    const { BulkOperations } = await import('../LazyComponents')
    expect(BulkOperations).toBeDefined()
  })
})
