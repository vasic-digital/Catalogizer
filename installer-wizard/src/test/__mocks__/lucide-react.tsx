// Mock all lucide-react icons as simple span components for testing
import React from 'react'

const createMockIcon = (name: string) => {
  const MockIcon = (props: any) => <span data-testid={`icon-${name}`} />
  MockIcon.displayName = name
  return MockIcon
}

// Export all icons used across the wizard step components
export const Settings = createMockIcon('Settings')
export const Folder = createMockIcon('Folder')
export const FolderOpen = createMockIcon('FolderOpen')
export const CheckCircle = createMockIcon('CheckCircle')
export const CheckCircle2 = createMockIcon('CheckCircle2')
export const AlertCircle = createMockIcon('AlertCircle')
export const AlertTriangle = createMockIcon('AlertTriangle')
export const Eye = createMockIcon('Eye')
export const EyeOff = createMockIcon('EyeOff')
export const TestTube = createMockIcon('TestTube')
export const Loader2 = createMockIcon('Loader2')
export const Plus = createMockIcon('Plus')
export const Trash2 = createMockIcon('Trash2')
export const Server = createMockIcon('Server')
export const FileText = createMockIcon('FileText')
export const HardDrive = createMockIcon('HardDrive')
export const Globe = createMockIcon('Globe')
export const Search = createMockIcon('Search')
export const Wifi = createMockIcon('Wifi')
export const RefreshCw = createMockIcon('RefreshCw')
export const Monitor = createMockIcon('Monitor')
export const Network = createMockIcon('Network')
export const Download = createMockIcon('Download')
export const Upload = createMockIcon('Upload')
export const ExternalLink = createMockIcon('ExternalLink')
export const Save = createMockIcon('Save')
export const Edit3 = createMockIcon('Edit3')
