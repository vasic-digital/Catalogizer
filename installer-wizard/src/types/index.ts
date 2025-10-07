// Network and SMB types
export interface NetworkHost {
  ip: string
  hostname?: string
  mac_address?: string
  vendor?: string
  open_ports: number[]
  smb_shares: string[]
}

export interface SMBShare {
  host: string
  share_name: string
  path: string
  writable: boolean
  description?: string
}

export interface FileEntry {
  name: string
  path: string
  is_directory: boolean
  size?: number
  modified?: string
}

// Configuration types (matching the JSON format)
export interface ConfigurationAccess {
  name: string
  type: string
  account: string
  secret: string
}

export interface ConfigurationSource {
  type: string
  url: string
  access: string
}

export interface Configuration {
  accesses: ConfigurationAccess[]
  sources: ConfigurationSource[]
}

// Wizard state types
export interface WizardState {
  currentStep: number
  totalSteps: number
  canGoNext: boolean
  canGoPrevious: boolean
  isComplete: boolean
}

// SMB Connection Configuration
export interface SMBConnectionConfig {
  name: string
  host: string
  port: number
  share_name: string
  username: string
  password: string
  domain?: string
  path?: string
  enabled: boolean
}

// FTP Connection Configuration
export interface FTPConnectionConfig {
  name: string
  host: string
  port: number
  username: string
  password: string
  path?: string
  enabled: boolean
}

// NFS Connection Configuration
export interface NFSConnectionConfig {
  name: string
  host: string
  path: string
  mount_point: string
  options?: string
  enabled: boolean
}

// WebDAV Connection Configuration
export interface WebDAVConnectionConfig {
  name: string
  url: string
  username: string
  password: string
  path?: string
  enabled: boolean
}

// Local Connection Configuration
export interface LocalConnectionConfig {
  name: string
  base_path: string
  enabled: boolean
}

// Form validation errors
export interface ValidationError {
  field: string
  message: string
}

// API Response types
export interface ApiResponse<T> {
  success: boolean
  data?: T
  error?: string
  message?: string
}

// UI Component props
export interface StepProps {
  onNext: () => void
  onPrevious: () => void
  canNext: boolean
  canPrevious: boolean
}

export interface WizardStep {
  title: string
  description: string
  path: string
  component: React.ComponentType<StepProps>
}

// Configuration validation
export interface ConfigValidationResult {
  isValid: boolean
  errors: ValidationError[]
  warnings: string[]
}