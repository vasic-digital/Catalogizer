import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { act } from 'react'
import { WizardProvider, useWizard } from '../WizardContext'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <WizardProvider>{children}</WizardProvider>
)

describe('WizardContext (extended edge cases)', () => {
  describe('boundary navigation', () => {
    it('does not go below step 0 on previousStep', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      expect(result.current.state.currentStep).toBe(0)

      act(() => {
        result.current.previousStep()
      })

      expect(result.current.state.currentStep).toBe(0)
      expect(result.current.state.canGoPrevious).toBe(false)
    })

    it('does not go above totalSteps - 1 on nextStep', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      // Navigate to last step
      for (let i = 0; i < 10; i++) {
        act(() => {
          result.current.nextStep()
        })
      }

      expect(result.current.state.currentStep).toBe(4)
      expect(result.current.state.isComplete).toBe(true)
    })

    it('clamps setStep to 0 for negative values', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(-100)
      })

      expect(result.current.state.currentStep).toBe(0)
    })

    it('clamps setStep to totalSteps - 1 for large values', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(999)
      })

      expect(result.current.state.currentStep).toBe(4)
    })

    it('setStep to 0 is valid', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(3)
      })

      act(() => {
        result.current.setStep(0)
      })

      expect(result.current.state.currentStep).toBe(0)
      expect(result.current.state.canGoPrevious).toBe(false)
    })
  })

  describe('setTotalSteps', () => {
    it('changes the total number of steps', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setTotalSteps(7)
      })

      expect(result.current.state.totalSteps).toBe(7)
    })

    it('allows navigating to new last step after increasing total', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setTotalSteps(10)
      })

      act(() => {
        result.current.setStep(9)
      })

      expect(result.current.state.currentStep).toBe(9)
      expect(result.current.state.isComplete).toBe(true)
    })

    it('does not auto-clamp current step when total decreases', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(4)
      })

      act(() => {
        result.current.setTotalSteps(3)
      })

      // Current step is not clamped automatically by setTotalSteps
      // Only setStep, nextStep, previousStep clamp
      expect(result.current.state.totalSteps).toBe(3)
    })
  })

  describe('canGoPrevious derivation', () => {
    it('is false at step 0', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      expect(result.current.state.currentStep).toBe(0)
      expect(result.current.state.canGoPrevious).toBe(false)
    })

    it('is true at step 1', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.nextStep()
      })

      expect(result.current.state.currentStep).toBe(1)
      expect(result.current.state.canGoPrevious).toBe(true)
    })

    it('is true at any step above 0', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      for (let i = 1; i <= 4; i++) {
        act(() => {
          result.current.setStep(i)
        })

        expect(result.current.state.canGoPrevious).toBe(true)
      }
    })

    it('becomes false after navigating back to step 0', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.nextStep()
      })

      expect(result.current.state.canGoPrevious).toBe(true)

      act(() => {
        result.current.previousStep()
      })

      expect(result.current.state.canGoPrevious).toBe(false)
    })
  })

  describe('isComplete derivation', () => {
    it('is false on first step', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      expect(result.current.state.isComplete).toBe(false)
    })

    it('is true only on last step', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      // Steps 0 through 3 should not be complete
      for (let i = 0; i < 4; i++) {
        act(() => {
          result.current.setStep(i)
        })
        expect(result.current.state.isComplete).toBe(false)
      }

      // Step 4 (totalSteps - 1) should be complete
      act(() => {
        result.current.setStep(4)
      })
      expect(result.current.state.isComplete).toBe(true)
    })

    it('becomes false when stepping back from last step', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(4)
      })
      expect(result.current.state.isComplete).toBe(true)

      act(() => {
        result.current.previousStep()
      })
      expect(result.current.state.isComplete).toBe(false)
    })
  })

  describe('setCanNext and setCanPrevious', () => {
    it('setCanNext toggles canGoNext', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      expect(result.current.state.canGoNext).toBe(true)

      act(() => {
        result.current.setCanNext(false)
      })

      expect(result.current.state.canGoNext).toBe(false)

      act(() => {
        result.current.setCanNext(true)
      })

      expect(result.current.state.canGoNext).toBe(true)
    })

    it('setCanPrevious overrides derived canGoPrevious', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      // At step 0, canGoPrevious is derived as false
      // But setCanPrevious can override it
      act(() => {
        result.current.setCanPrevious(true)
      })

      expect(result.current.state.canGoPrevious).toBe(true)
    })

    it('setCanNext does not affect other state', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(2)
      })

      const currentStep = result.current.state.currentStep

      act(() => {
        result.current.setCanNext(false)
      })

      expect(result.current.state.currentStep).toBe(currentStep)
      expect(result.current.state.isComplete).toBe(false)
    })
  })

  describe('rapid state changes', () => {
    it('handles rapid next/previous calls', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.nextStep()
        result.current.nextStep()
        result.current.nextStep()
        result.current.previousStep()
      })

      expect(result.current.state.currentStep).toBe(2)
    })

    it('handles multiple resets', () => {
      const { result } = renderHook(() => useWizard(), { wrapper })

      act(() => {
        result.current.setStep(3)
        result.current.setCanNext(false)
      })

      act(() => {
        result.current.reset()
      })

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

  describe('useWizard outside provider', () => {
    it('throws when used outside WizardProvider', () => {
      expect(() => {
        renderHook(() => useWizard())
      }).toThrow('useWizard must be used within a WizardProvider')
    })
  })
})
