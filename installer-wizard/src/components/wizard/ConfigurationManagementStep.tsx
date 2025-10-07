import { useState, useEffect } from 'react'
import { Button } from '../ui/Button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/Card'
import { useWizard } from '../../contexts/WizardContext'
import { useConfiguration } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'
import { Configuration, ConfigurationAccess, ConfigurationSource } from '../../types'
import {
  FileText,
  FolderOpen,
  Save,
  Upload,
  Download,
  Plus,
  Trash2,
  Edit3,
  CheckCircle,
  AlertCircle,
  Loader2
} from 'lucide-react'

export default function ConfigurationManagementStep() {
  const { setCanNext } = useWizard()
  const { setConfiguration } = useConfiguration()
  const [isLoading, setIsLoading] = useState(false)
  const [message, setMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [generatedConfig, setGeneratedConfig] = useState<Configuration | null>(null)

  useEffect(() => {
    // Always allow proceeding from this step
    setCanNext(true)
    generateConfigFromSources()
  }, [setCanNext])

  const generateConfigFromSources = () => {
    // Generate configuration from all configured sources
    // This should be updated to collect configs from all protocol steps
    const mockConfigs = [
      {
        protocol: 'smb',
        name: 'Media Server',
        host: '192.168.1.100',
        port: 445,
        share_name: 'shared',
        username: 'user',
        password: 'password',
        domain: 'WORKGROUP',
        path: '/media',
        enabled: true,
      },
      {
        protocol: 'ftp',
        name: 'FTP Server',
        host: 'ftp.example.com',
        port: 21,
        username: 'ftpuser',
        password: 'ftppass',
        path: '/',
        enabled: true,
      },
      {
        protocol: 'nfs',
        name: 'NFS Share',
        host: 'nfs.example.com',
        path: '/export/data',
        mount_point: '/mnt/nfs',
        options: 'vers=3',
        enabled: true,
      },
      {
        protocol: 'webdav',
        name: 'WebDAV Server',
        url: 'https://webdav.example.com/remote.php/dav/files/user/',
        username: 'webdavuser',
        password: 'webdavpass',
        path: '/',
        enabled: true,
      },
      {
        protocol: 'local',
        name: 'Local Storage',
        base_path: '/tmp/catalog-data',
        enabled: true,
      }
    ]

    if (mockConfigs.length > 0) {
      const accesses: ConfigurationAccess[] = mockConfigs
        .filter(config => config.protocol !== 'local')
        .map((config) => ({
          name: `${config.protocol}_${config.username || 'local'}`,
          type: 'credentials',
          account: config.username || 'local',
          secret: config.password || '',
        }))

      const sources: ConfigurationSource[] = mockConfigs.map((config) => {
        let url = ''
        switch (config.protocol) {
          case 'smb':
            url = `smb://${config.host}:${config.port}/${config.share_name}${config.path || ''}`
            break
          case 'ftp':
            url = `ftp://${config.host}:${config.port}${config.path || '/'}`
            break
          case 'nfs':
            url = `nfs://${config.host}${config.path}`
            break
          case 'webdav':
            url = `${config.url}${config.path || '/'}`
            break
          case 'local':
            url = `file://${config.base_path}`
            break
        }
        return {
          type: config.protocol === 'smb' ? 'samba' : config.protocol,
          url,
          access: config.protocol === 'local' ? 'local' : `${config.protocol}_${config.username}`,
        }
      })

      const configuration: Configuration = {
        accesses,
        sources,
      }

      setGeneratedConfig(configuration)
      setConfiguration(configuration)
    }
  }

  const handleLoadConfiguration = async () => {
    setIsLoading(true)
    setMessage(null)

    try {
      const config = await TauriService.openConfigurationFile()
      if (config) {
        if (TauriService.validateConfiguration(config)) {
          setConfiguration(config)
          setGeneratedConfig(config)
          setMessage({
            type: 'success',
            text: 'Configuration loaded successfully'
          })
        } else {
          setMessage({
            type: 'error',
            text: 'Invalid configuration file format'
          })
        }
      }
    } catch (error) {
      setMessage({
        type: 'error',
        text: `Failed to load configuration: ${error instanceof Error ? error.message : 'Unknown error'}`
      })
    } finally {
      setIsLoading(false)
    }
  }

  const handleSaveConfiguration = async () => {
    if (!generatedConfig) {
      setMessage({
        type: 'error',
        text: 'No configuration to save'
      })
      return
    }

    setIsLoading(true)
    setMessage(null)

    try {
      const success = await TauriService.saveConfigurationFile(generatedConfig)
      if (success) {
        setMessage({
          type: 'success',
          text: 'Configuration saved successfully'
        })
      }
    } catch (error) {
      setMessage({
        type: 'error',
        text: `Failed to save configuration: ${error instanceof Error ? error.message : 'Unknown error'}`
      })
    } finally {
      setIsLoading(false)
    }
  }

  const addAccess = () => {
    if (!generatedConfig) return

    const newAccess: ConfigurationAccess = {
      name: 'new_user',
      type: 'credentials',
      account: 'username',
      secret: 'password',
    }

    const updatedConfig = {
      ...generatedConfig,
      accesses: [...generatedConfig.accesses, newAccess]
    }

    setGeneratedConfig(updatedConfig)
    setConfiguration(updatedConfig)
  }

  const removeAccess = (index: number) => {
    if (!generatedConfig) return

    const updatedConfig = {
      ...generatedConfig,
      accesses: generatedConfig.accesses.filter((_, i) => i !== index)
    }

    setGeneratedConfig(updatedConfig)
    setConfiguration(updatedConfig)
  }

  const addSource = () => {
    if (!generatedConfig) return

    const newSource: ConfigurationSource = {
      type: 'samba',
      url: 'smb://host/share',
      access: generatedConfig.accesses[0]?.name || 'username',
    }

    const updatedConfig = {
      ...generatedConfig,
      sources: [...generatedConfig.sources, newSource]
    }

    setGeneratedConfig(updatedConfig)
    setConfiguration(updatedConfig)
  }

  const removeSource = (index: number) => {
    if (!generatedConfig) return

    const updatedConfig = {
      ...generatedConfig,
      sources: generatedConfig.sources.filter((_, i) => i !== index)
    }

    setGeneratedConfig(updatedConfig)
    setConfiguration(updatedConfig)
  }

  return (
    <div className="space-y-6">
      <div className="text-center space-y-4">
        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
          <FileText className="h-8 w-8 text-blue-600" />
        </div>
        <h2 className="text-xl font-bold text-gray-900">Configuration Management</h2>
        <p className="text-gray-600">
          Manage your Catalogizer configuration file
        </p>
      </div>

      {/* Message Display */}
      {message && (
        <Card className={message.type === 'success' ? 'border-green-200 bg-green-50' : 'border-red-200 bg-red-50'}>
          <CardContent className="pt-6">
            <div className={`flex items-center gap-2 ${message.type === 'success' ? 'text-green-800' : 'text-red-800'}`}>
              {message.type === 'success' ? (
                <CheckCircle className="h-5 w-5" />
              ) : (
                <AlertCircle className="h-5 w-5" />
              )}
              <span className="font-medium">{message.text}</span>
            </div>
          </CardContent>
        </Card>
      )}

      {/* File Operations */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <FileText className="h-5 w-5" />
            Configuration File Operations
          </CardTitle>
          <CardDescription>
            Load an existing configuration file or save your current configuration
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <Button
              variant="outline"
              onClick={handleLoadConfiguration}
              disabled={isLoading}
              className="flex items-center gap-2 h-12"
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <FolderOpen className="h-4 w-4" />
              )}
              Load Configuration
            </Button>

            <Button
              onClick={handleSaveConfiguration}
              disabled={isLoading || !generatedConfig}
              className="flex items-center gap-2 h-12"
            >
              {isLoading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <Save className="h-4 w-4" />
              )}
              Save Configuration
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Configuration Editor */}
      {generatedConfig && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {/* Access Credentials */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <Upload className="h-5 w-5" />
                  Access Credentials ({generatedConfig.accesses.length})
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={addAccess}
                  className="flex items-center gap-2"
                >
                  <Plus className="h-4 w-4" />
                  Add
                </Button>
              </CardTitle>
               <CardDescription>
                 Manage authentication credentials for all sources
               </CardDescription>
            </CardHeader>
            <CardContent>
              {generatedConfig.accesses.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <Upload className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium">No credentials configured</p>
                  <p className="text-sm">Add credentials for your SMB sources</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {generatedConfig.accesses.map((access, index) => (
                    <div
                      key={index}
                      className="p-4 border rounded-lg hover:border-gray-300 transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="font-medium">{access.name}</div>
                          <div className="text-sm text-gray-500">
                            Type: {access.type} • Account: {access.account}
                          </div>
                          <div className="text-xs text-gray-400">
                            Secret: {'*'.repeat(access.secret.length)}
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-red-600 hover:text-red-700"
                            onClick={() => removeAccess(index)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Sources */}
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span className="flex items-center gap-2">
                  <Download className="h-5 w-5" />
                  Media Sources ({generatedConfig.sources.length})
                </span>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={addSource}
                  className="flex items-center gap-2"
                >
                  <Plus className="h-4 w-4" />
                  Add
                </Button>
              </CardTitle>
               <CardDescription>
                 Manage all media source configurations
               </CardDescription>
            </CardHeader>
            <CardContent>
              {generatedConfig.sources.length === 0 ? (
                <div className="text-center py-8 text-gray-500">
                  <Download className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p className="text-lg font-medium">No sources configured</p>
                  <p className="text-sm">Add SMB sources for your media</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {generatedConfig.sources.map((source, index) => (
                    <div
                      key={index}
                      className="p-4 border rounded-lg hover:border-gray-300 transition-colors"
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex-1">
                          <div className="font-medium">{source.type}</div>
                          <div className="text-sm text-gray-500 break-all">
                            URL: {source.url}
                          </div>
                          <div className="text-xs text-gray-400">
                            Access: {source.access}
                          </div>
                        </div>
                        <div className="flex items-center gap-2">
                          <Button
                            variant="outline"
                            size="sm"
                            className="text-red-600 hover:text-red-700"
                            onClick={() => removeSource(index)}
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </div>
      )}

      {/* Configuration Preview */}
      {generatedConfig && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Edit3 className="h-5 w-5" />
              Configuration Preview
            </CardTitle>
            <CardDescription>
              JSON representation of your configuration
            </CardDescription>
          </CardHeader>
          <CardContent>
            <pre className="bg-gray-50 p-4 rounded-lg text-sm overflow-auto max-h-64">
              {JSON.stringify(generatedConfig, null, 2)}
            </pre>
          </CardContent>
        </Card>
      )}

      {generatedConfig && (
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="flex items-center gap-2 text-blue-800">
            <CheckCircle className="h-4 w-4" />
            <span className="font-medium">Configuration Ready</span>
          </div>
          <p className="text-sm text-blue-700 mt-1">
            Your configuration is ready. Click "Next" to review the summary.
          </p>
        </div>
      )}
    </div>
  )
}