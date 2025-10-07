import React, { createContext, useContext, useReducer, ReactNode } from 'react'
import { Configuration, ConfigurationAccess, ConfigurationSource, SMBConnectionConfig, FTPConnectionConfig, NFSConnectionConfig, WebDAVConnectionConfig, LocalConnectionConfig } from '../types'

interface ConfigurationState {
  configuration: Configuration
  currentSMBConfig: SMBConnectionConfig | null
  currentFTPConfig: FTPConnectionConfig | null
  currentNFSConfig: NFSConnectionConfig | null
  currentWebDAVConfig: WebDAVConnectionConfig | null
  currentLocalConfig: LocalConnectionConfig | null
  selectedProtocol: string | null
  selectedHosts: string[]
  isLoading: boolean
  error: string | null
  hasUnsavedChanges: boolean
}

interface ConfigurationAction {
  type:
    | 'SET_CONFIGURATION'
    | 'ADD_ACCESS'
    | 'UPDATE_ACCESS'
    | 'REMOVE_ACCESS'
    | 'ADD_SOURCE'
    | 'UPDATE_SOURCE'
    | 'REMOVE_SOURCE'
    | 'SET_CURRENT_SMB_CONFIG'
    | 'SET_CURRENT_FTP_CONFIG'
    | 'SET_CURRENT_NFS_CONFIG'
    | 'SET_CURRENT_WEBDAV_CONFIG'
    | 'SET_CURRENT_LOCAL_CONFIG'
    | 'SET_SELECTED_PROTOCOL'
    | 'SET_SELECTED_HOSTS'
    | 'SET_LOADING'
    | 'SET_ERROR'
    | 'CLEAR_ERROR'
    | 'SET_UNSAVED_CHANGES'
    | 'RESET'
  payload?: any
}

interface ConfigurationContextType {
  state: ConfigurationState
  dispatch: React.Dispatch<ConfigurationAction>
  setConfiguration: (config: Configuration) => void
  addAccess: (access: ConfigurationAccess) => void
  updateAccess: (index: number, access: ConfigurationAccess) => void
  removeAccess: (index: number) => void
  addSource: (source: ConfigurationSource) => void
  updateSource: (index: number, source: ConfigurationSource) => void
  removeSource: (index: number) => void
  setCurrentSMBConfig: (config: SMBConnectionConfig | null) => void
  setCurrentFTPConfig: (config: FTPConnectionConfig | null) => void
  setCurrentNFSConfig: (config: NFSConnectionConfig | null) => void
  setCurrentWebDAVConfig: (config: WebDAVConnectionConfig | null) => void
  setCurrentLocalConfig: (config: LocalConnectionConfig | null) => void
  setSelectedProtocol: (protocol: string | null) => void
  setSelectedHosts: (hosts: string[]) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  setUnsavedChanges: (hasChanges: boolean) => void
  reset: () => void
}

const initialState: ConfigurationState = {
  configuration: {
    accesses: [],
    sources: [],
  },
  currentSMBConfig: null,
  currentFTPConfig: null,
  currentNFSConfig: null,
  currentWebDAVConfig: null,
  currentLocalConfig: null,
  selectedProtocol: null,
  selectedHosts: [],
  isLoading: false,
  error: null,
  hasUnsavedChanges: false,
}

function configurationReducer(state: ConfigurationState, action: ConfigurationAction): ConfigurationState {
  switch (action.type) {
    case 'SET_CONFIGURATION':
      return {
        ...state,
        configuration: action.payload,
        hasUnsavedChanges: false,
      }
    case 'ADD_ACCESS':
      return {
        ...state,
        configuration: {
          ...state.configuration,
          accesses: [...state.configuration.accesses, action.payload],
        },
        hasUnsavedChanges: true,
      }
    case 'UPDATE_ACCESS':
      const updatedAccesses = [...state.configuration.accesses]
      updatedAccesses[action.payload.index] = action.payload.access
      return {
        ...state,
        configuration: {
          ...state.configuration,
          accesses: updatedAccesses,
        },
        hasUnsavedChanges: true,
      }
    case 'REMOVE_ACCESS':
      return {
        ...state,
        configuration: {
          ...state.configuration,
          accesses: state.configuration.accesses.filter((_, index) => index !== action.payload),
        },
        hasUnsavedChanges: true,
      }
    case 'ADD_SOURCE':
      return {
        ...state,
        configuration: {
          ...state.configuration,
          sources: [...state.configuration.sources, action.payload],
        },
        hasUnsavedChanges: true,
      }
    case 'UPDATE_SOURCE':
      const updatedSources = [...state.configuration.sources]
      updatedSources[action.payload.index] = action.payload.source
      return {
        ...state,
        configuration: {
          ...state.configuration,
          sources: updatedSources,
        },
        hasUnsavedChanges: true,
      }
    case 'REMOVE_SOURCE':
      return {
        ...state,
        configuration: {
          ...state.configuration,
          sources: state.configuration.sources.filter((_, index) => index !== action.payload),
        },
        hasUnsavedChanges: true,
      }
    case 'SET_CURRENT_SMB_CONFIG':
      return {
        ...state,
        currentSMBConfig: action.payload,
      }
    case 'SET_CURRENT_FTP_CONFIG':
      return {
        ...state,
        currentFTPConfig: action.payload,
      }
    case 'SET_CURRENT_NFS_CONFIG':
      return {
        ...state,
        currentNFSConfig: action.payload,
      }
    case 'SET_CURRENT_WEBDAV_CONFIG':
      return {
        ...state,
        currentWebDAVConfig: action.payload,
      }
    case 'SET_CURRENT_LOCAL_CONFIG':
      return {
        ...state,
        currentLocalConfig: action.payload,
      }
    case 'SET_SELECTED_PROTOCOL':
      return {
        ...state,
        selectedProtocol: action.payload,
      }
    case 'SET_SELECTED_HOSTS':
      return {
        ...state,
        selectedHosts: action.payload,
      }
    case 'SET_LOADING':
      return {
        ...state,
        isLoading: action.payload,
      }
    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
      }
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      }
    case 'SET_UNSAVED_CHANGES':
      return {
        ...state,
        hasUnsavedChanges: action.payload,
      }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

const ConfigurationContext = createContext<ConfigurationContextType | undefined>(undefined)

export function ConfigurationProvider({ children }: { children: ReactNode }) {
  const [state, dispatch] = useReducer(configurationReducer, initialState)

  const setConfiguration = (config: Configuration) =>
    dispatch({ type: 'SET_CONFIGURATION', payload: config })

  const addAccess = (access: ConfigurationAccess) =>
    dispatch({ type: 'ADD_ACCESS', payload: access })

  const updateAccess = (index: number, access: ConfigurationAccess) =>
    dispatch({ type: 'UPDATE_ACCESS', payload: { index, access } })

  const removeAccess = (index: number) =>
    dispatch({ type: 'REMOVE_ACCESS', payload: index })

  const addSource = (source: ConfigurationSource) =>
    dispatch({ type: 'ADD_SOURCE', payload: source })

  const updateSource = (index: number, source: ConfigurationSource) =>
    dispatch({ type: 'UPDATE_SOURCE', payload: { index, source } })

  const removeSource = (index: number) =>
    dispatch({ type: 'REMOVE_SOURCE', payload: index })

  const setCurrentSMBConfig = (config: SMBConnectionConfig | null) =>
    dispatch({ type: 'SET_CURRENT_SMB_CONFIG', payload: config })

  const setCurrentFTPConfig = (config: FTPConnectionConfig | null) =>
    dispatch({ type: 'SET_CURRENT_FTP_CONFIG', payload: config })

  const setCurrentNFSConfig = (config: NFSConnectionConfig | null) =>
    dispatch({ type: 'SET_CURRENT_NFS_CONFIG', payload: config })

  const setCurrentWebDAVConfig = (config: WebDAVConnectionConfig | null) =>
    dispatch({ type: 'SET_CURRENT_WEBDAV_CONFIG', payload: config })

  const setCurrentLocalConfig = (config: LocalConnectionConfig | null) =>
    dispatch({ type: 'SET_CURRENT_LOCAL_CONFIG', payload: config })

  const setSelectedProtocol = (protocol: string | null) =>
    dispatch({ type: 'SET_SELECTED_PROTOCOL', payload: protocol })

  const setSelectedHosts = (hosts: string[]) =>
    dispatch({ type: 'SET_SELECTED_HOSTS', payload: hosts })

  const setLoading = (loading: boolean) =>
    dispatch({ type: 'SET_LOADING', payload: loading })

  const setError = (error: string | null) =>
    dispatch({ type: 'SET_ERROR', payload: error })

  const clearError = () =>
    dispatch({ type: 'CLEAR_ERROR' })

  const setUnsavedChanges = (hasChanges: boolean) =>
    dispatch({ type: 'SET_UNSAVED_CHANGES', payload: hasChanges })

  const reset = () =>
    dispatch({ type: 'RESET' })

  const value: ConfigurationContextType = {
    state,
    dispatch,
    setConfiguration,
    addAccess,
    updateAccess,
    removeAccess,
    addSource,
    updateSource,
    removeSource,
    setCurrentSMBConfig,
    setCurrentFTPConfig,
    setCurrentNFSConfig,
    setCurrentWebDAVConfig,
    setCurrentLocalConfig,
    setSelectedProtocol,
    setSelectedHosts,
    setLoading,
    setError,
    clearError,
    setUnsavedChanges,
    reset,
  }

  return <ConfigurationContext.Provider value={value}>{children}</ConfigurationContext.Provider>
}

export function useConfiguration() {
  const context = useContext(ConfigurationContext)
  if (context === undefined) {
    throw new Error('useConfiguration must be used within a ConfigurationProvider')
  }
  return context
}