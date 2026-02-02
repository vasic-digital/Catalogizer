import React from 'react'
import { render, screen, within } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'
import { WizardProvider } from '../contexts/WizardContext'
import { ConfigurationProvider } from '../contexts/ConfigurationContext'

export const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <ConfigurationProvider>
          <WizardProvider>
            {children}
          </WizardProvider>
        </ConfigurationProvider>
      </BrowserRouter>
    </QueryClientProvider>
  )
}

/**
 * Find an input element by its label text.
 * Works even without htmlFor/id association by finding the label
 * and then looking for the nearest input sibling.
 */
export function getInputByLabel(labelText: string): HTMLInputElement {
  const label = screen.getByText(labelText, { selector: 'label' })
  const container = label.parentElement!
  const input = container.querySelector('input')
  if (!input) {
    throw new Error(`No input found near label "${labelText}"`)
  }
  return input
}
