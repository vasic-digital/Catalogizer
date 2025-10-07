import React, { createContext, useContext, useReducer, ReactNode } from 'react'
import { WizardState } from '../types'

interface WizardAction {
  type: 'NEXT_STEP' | 'PREVIOUS_STEP' | 'SET_STEP' | 'SET_CAN_NEXT' | 'SET_CAN_PREVIOUS' | 'SET_TOTAL_STEPS' | 'RESET'
  payload?: any
}

interface WizardContextType {
  state: WizardState
  dispatch: React.Dispatch<WizardAction>
  nextStep: () => void
  previousStep: () => void
  setStep: (step: number) => void
  setCanNext: (canNext: boolean) => void
  setCanPrevious: (canPrevious: boolean) => void
  setTotalSteps: (total: number) => void
  reset: () => void
}

const initialState: WizardState = {
  currentStep: 0,
  totalSteps: 5, // Base steps: welcome, protocol, config, manage, summary
  canGoNext: true,
  canGoPrevious: false,
  isComplete: false,
}

function wizardReducer(state: WizardState, action: WizardAction): WizardState {
  switch (action.type) {
    case 'NEXT_STEP':
      const nextStep = Math.min(state.currentStep + 1, state.totalSteps - 1)
      return {
        ...state,
        currentStep: nextStep,
        canGoPrevious: nextStep > 0,
        isComplete: nextStep === state.totalSteps - 1,
      }
    case 'PREVIOUS_STEP':
      const prevStep = Math.max(state.currentStep - 1, 0)
      return {
        ...state,
        currentStep: prevStep,
        canGoPrevious: prevStep > 0,
        isComplete: false,
      }
    case 'SET_STEP':
      const targetStep = Math.max(0, Math.min(action.payload, state.totalSteps - 1))
      return {
        ...state,
        currentStep: targetStep,
        canGoPrevious: targetStep > 0,
        isComplete: targetStep === state.totalSteps - 1,
      }
    case 'SET_CAN_NEXT':
      return {
        ...state,
        canGoNext: action.payload,
      }
    case 'SET_CAN_PREVIOUS':
      return {
        ...state,
        canGoPrevious: action.payload,
      }
    case 'SET_TOTAL_STEPS':
      return {
        ...state,
        totalSteps: action.payload,
      }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

const WizardContext = createContext<WizardContextType | undefined>(undefined)

export function WizardProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(wizardReducer, initialState)

  const nextStep = () => dispatch({ type: 'NEXT_STEP' })
  const previousStep = () => dispatch({ type: 'PREVIOUS_STEP' })
  const setStep = (step: number) => dispatch({ type: 'SET_STEP', payload: step })
  const setCanNext = (canNext: boolean) => dispatch({ type: 'SET_CAN_NEXT', payload: canNext })
  const setCanPrevious = (canPrevious: boolean) => dispatch({ type: 'SET_CAN_PREVIOUS', payload: canPrevious })
  const setTotalSteps = (total: number) => dispatch({ type: 'SET_TOTAL_STEPS', payload: total })
  const reset = () => dispatch({ type: 'RESET' })

  const value: WizardContextType = {
    state,
    dispatch,
    nextStep,
    previousStep,
    setStep,
    setCanNext,
    setCanPrevious,
    setTotalSteps,
    reset,
  }

  return <WizardContext.Provider value={value}>{children}</WizardContext.Provider>
}

export function useWizard() {
  const context = useContext(WizardContext)
  if (context === undefined) {
    throw new Error('useWizard must be used within a WizardProvider')
  }
  return context
}