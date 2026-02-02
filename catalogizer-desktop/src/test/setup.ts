import '@testing-library/jest-dom'
import { cleanup } from '@testing-library/react'
import { afterEach } from 'vitest'

// Mock Tauri API
const mockInvoke = vi.fn()

vi.mock('@tauri-apps/api/core', () => ({
  invoke: mockInvoke,
}))

vi.mock('@tauri-apps/plugin-shell', () => ({
  open: vi.fn(),
}))

// Global test utilities
;(globalThis as any).mockInvoke = mockInvoke

// Reset mocks and cleanup before each test
beforeEach(() => {
  vi.clearAllMocks()
})

// Cleanup after each test
afterEach(() => {
  cleanup()
})
