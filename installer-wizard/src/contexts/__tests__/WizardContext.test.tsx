import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { act } from 'react'
import { WizardProvider, useWizard } from '../WizardContext'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <WizardProvider>{children}</WizardProvider>
)

describe('WizardContext', () => {
  it('initializes with correct default state', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    expect(result.current.state).toEqual({
      currentStep: 0,
      totalSteps: 5,
      canGoNext: true,
      canGoPrevious: false,
      isComplete: false,
    })
  })

  it('advances to next step', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    act(() => {
      result.current.nextStep()
    })

    expect(result.current.state.currentStep).toBe(1)
    expect(result.current.state.canGoPrevious).toBe(true)
    expect(result.current.state.isComplete).toBe(false)
  })

  it('goes to previous step', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    // First go to step 1
    act(() => {
      result.current.nextStep()
    })

    // Then go back to step 0
    act(() => {
      result.current.previousStep()
    })

    expect(result.current.state.currentStep).toBe(0)
    expect(result.current.state.canGoPrevious).toBe(false)
  })

  it('sets specific step', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    act(() => {
      result.current.setStep(3)
    })

    expect(result.current.state.currentStep).toBe(3)
    expect(result.current.state.canGoPrevious).toBe(true)
    expect(result.current.state.isComplete).toBe(false)
  })

  it('marks as complete when reaching final step', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    act(() => {
      result.current.setStep(4) // Last step (totalSteps - 1)
    })

    expect(result.current.state.currentStep).toBe(4)
    expect(result.current.state.isComplete).toBe(true)
  })

  it('prevents going beyond boundaries', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    // Try to go to step beyond total steps
    act(() => {
      result.current.setStep(10)
    })

    expect(result.current.state.currentStep).toBe(4) // Should be clamped to totalSteps - 1

    // Try to go to negative step
    act(() => {
      result.current.setStep(-1)
    })

    expect(result.current.state.currentStep).toBe(0) // Should be clamped to 0
  })

  it('updates canNext state', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    act(() => {
      result.current.setCanNext(false)
    })

    expect(result.current.state.canGoNext).toBe(false)

    act(() => {
      result.current.setCanNext(true)
    })

    expect(result.current.state.canGoNext).toBe(true)
  })

  it('resets to initial state', () => {
    const { result } = renderHook(() => useWizard(), { wrapper })

    // Modify state
    act(() => {
      result.current.setStep(3)
      result.current.setCanNext(false)
    })

    // Reset
    act(() => {
      result.current.reset()
    })

    expect(result.current.state).toEqual({
      currentStep: 0,
      totalSteps: 5,
      canGoNext: true,
      canGoPrevious: false,
      isComplete: false,
    })
  })
})