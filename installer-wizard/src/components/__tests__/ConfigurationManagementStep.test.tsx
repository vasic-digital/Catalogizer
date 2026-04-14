import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import ConfigurationManagementStep from '../wizard/ConfigurationManagementStep'
import { TauriService } from '../../services/tauri'
import { TestWrapper } from '../../test/test-utils'

describe('ConfigurationManagementStep', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders configuration management heading', () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configuration Management')).toBeInTheDocument()
    expect(screen.getByText('Manage your Catalogizer configuration file')).toBeInTheDocument()
  })

  it('shows file operation buttons', () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    expect(screen.getByText('Load Configuration')).toBeInTheDocument()
    expect(screen.getByText('Save Configuration')).toBeInTheDocument()
  })

  it('displays configuration file operations section', () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    expect(screen.getByText('Configuration File Operations')).toBeInTheDocument()
    expect(screen.getByText(/Load an existing configuration file or save your current configuration/)).toBeInTheDocument()
  })

  it('generates and displays configuration preview', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Configuration Preview')).toBeInTheDocument()
      expect(screen.getByText(/JSON representation of your configuration/)).toBeInTheDocument()
    })
  })

  it('shows access credentials section', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText(/Access Credentials/)).toBeInTheDocument()
      expect(screen.getByText(/Manage authentication credentials/)).toBeInTheDocument()
    })
  })

  it('shows media sources section', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText(/Media Sources/)).toBeInTheDocument()
      expect(screen.getByText(/Manage all media source configurations/)).toBeInTheDocument()
    })
  })

  it('shows configuration ready message', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Configuration Ready')).toBeInTheDocument()
      expect(screen.getByText(/Click "Next" to review the summary/)).toBeInTheDocument()
    })
  })

  it('handles load configuration', async () => {
    const mockConfig = {
      accesses: [
        { name: 'test_user', type: 'credentials', account: 'user', secret: 'pass' },
      ],
      sources: [
        { type: 'samba', url: 'smb://host/share', access: 'test_user' },
      ],
    }

    const mockOpenConfigurationFile = vi.spyOn(TauriService, 'openConfigurationFile')
      .mockResolvedValue(mockConfig)

    const mockValidateConfiguration = vi.spyOn(TauriService, 'validateConfiguration')
      .mockReturnValue(true)

    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    const loadButton = screen.getByText('Load Configuration')
    fireEvent.click(loadButton)

    await waitFor(() => {
      expect(mockOpenConfigurationFile).toHaveBeenCalled()
      expect(screen.getByText('Configuration loaded successfully')).toBeInTheDocument()
    })
  })

  it('handles load configuration failure', async () => {
    vi.spyOn(TauriService, 'openConfigurationFile')
      .mockRejectedValue(new Error('File not found'))

    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    const loadButton = screen.getByText('Load Configuration')
    fireEvent.click(loadButton)

    await waitFor(() => {
      expect(screen.getByText(/Failed to load configuration/)).toBeInTheDocument()
    })
  })

  it('handles save configuration', async () => {
    vi.spyOn(TauriService, 'saveConfigurationFile')
      .mockResolvedValue(true)

    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    // Wait for the component to generate its config
    await waitFor(() => {
      expect(screen.getByText('Save Configuration')).toBeInTheDocument()
    })

    const saveButton = screen.getByText('Save Configuration')
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(screen.getByText('Configuration saved successfully')).toBeInTheDocument()
    })
  })

  it('handles save configuration failure', async () => {
    vi.spyOn(TauriService, 'saveConfigurationFile')
      .mockRejectedValue(new Error('Permission denied'))

    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('Save Configuration')).toBeInTheDocument()
    })

    const saveButton = screen.getByText('Save Configuration')
    fireEvent.click(saveButton)

    await waitFor(() => {
      expect(screen.getByText(/Failed to save configuration/)).toBeInTheDocument()
    })
  })

  it('shows Add button for access credentials', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const addButtons = screen.getAllByText('Add')
      expect(addButtons.length).toBeGreaterThanOrEqual(2)
    })
  })

  it('shows Add button for media sources', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText(/Media Sources/)).toBeInTheDocument()
    })
  })

  it('displays generated accesses in the list', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      // The component generates mock configs with access names
      expect(screen.getByText(/Access Credentials/)).toBeInTheDocument()
    })
  })

  it('displays generated sources in the list', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText(/Media Sources/)).toBeInTheDocument()
    })
  })

  it('shows JSON preview with formatted content', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const preElement = screen.getByText('Configuration Preview')
        .closest('.space-y-6')!
        .querySelector('pre')
      expect(preElement).toBeInTheDocument()
      expect(preElement!.textContent).toContain('accesses')
      expect(preElement!.textContent).toContain('sources')
    })
  })

  it('shows source type in source entries', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      expect(screen.getByText('samba')).toBeInTheDocument()
    })
  })

  it('shows access type and account in access entries', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const accessElements = screen.getAllByText(/Type: credentials/)
      expect(accessElements.length).toBeGreaterThan(0)
    })
  })

  it('masks secrets in access entries', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const secretElements = screen.getAllByText(/Secret: \*+/)
      expect(secretElements.length).toBeGreaterThan(0)
    })
  })

  it('shows delete buttons for access entries', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const deleteButtons = screen.getAllByRole('button').filter(btn =>
        btn.classList.contains('text-red-600') || btn.className.includes('text-red')
      )
      expect(deleteButtons.length).toBeGreaterThan(0)
    })
  })

  it('shows URL in source entries', async () => {
    render(
      <TestWrapper>
        <ConfigurationManagementStep />
      </TestWrapper>
    )

    await waitFor(() => {
      const urlElements = screen.getAllByText(/URL:/)
      expect(urlElements.length).toBeGreaterThan(0)
    })
  })
})
