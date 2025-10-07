import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/Card'
import { useWizard } from '../../contexts/WizardContext'
import { useConfiguration } from '../../contexts/ConfigurationContext'
import { ConfigurationAccess, ConfigurationSource } from '../../types'
import { TauriService } from '../../services/tauri'
import { SMBConnectionConfig } from '../../types'
import {
  Settings,
  Eye,
  EyeOff,
  TestTube,
  CheckCircle,
  AlertCircle,
  Loader2,
  Folder,
  Plus,
  Trash2
} from 'lucide-react'

const smbConfigSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  host: z.string().min(1, 'Host is required'),
  port: z.number().min(1).max(65535).default(445),
  share_name: z.string().min(1, 'Share name is required'),
  username: z.string().min(1, 'Username is required'),
  password: z.string().min(1, 'Password is required'),
  domain: z.string().optional(),
  path: z.string().optional(),
  enabled: z.boolean().default(true),
})

type SMBConfigForm = z.infer<typeof smbConfigSchema>

export default function SMBConfigurationStep() {
  const { setCanNext } = useWizard()
  const { state: configState, addAccess, addSource } = useConfiguration()
  const [showPassword, setShowPassword] = useState(false)
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [isTestingConnection, setIsTestingConnection] = useState(false)
  const [smbConfigs, setSmbConfigs] = useState<SMBConnectionConfig[]>([])
  const [editingIndex, setEditingIndex] = useState<number | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    reset,
  } = useForm<SMBConfigForm>({
    resolver: zodResolver(smbConfigSchema),
    defaultValues: {
      port: 445,
      enabled: true,
    }
  })

  const watchedValues = watch()

  useEffect(() => {
    // Can proceed if we have at least one valid SMB configuration
    setCanNext(smbConfigs.length > 0)
  }, [smbConfigs, setCanNext])

  useEffect(() => {
    // Pre-populate with selected hosts from network scan
    if (configState.selectedHosts.length > 0 && smbConfigs.length === 0) {
      const defaultConfigs = configState.selectedHosts.map((host, index) => ({
        name: `SMB Source ${index + 1}`,
        host: host.ip,
        port: 445,
        share_name: '',
        username: '',
        password: '',
        domain: '',
        path: '',
        enabled: true,
      }))
      setSmbConfigs(defaultConfigs)
      if (defaultConfigs.length > 0) {
        startEditing(0, defaultConfigs[0])
      }
    }
  }, [configState.selectedHosts, smbConfigs.length])

  const startEditing = (index: number, config: SMBConnectionConfig) => {
    setEditingIndex(index)
    reset(config)
    setTestResult(null)
  }

  const handleTestConnection = async () => {
    const values = watchedValues
    if (!values.host || !values.share_name || !values.username || !values.password) {
      setTestResult({
        success: false,
        message: 'Please fill in all required fields before testing'
      })
      return
    }

    setIsTestingConnection(true)
    setTestResult(null)

    try {
      const success = await TauriService.testSMBConnection(
        values.host,
        values.share_name,
        values.username,
        values.password,
        values.domain
      )

      setTestResult({
        success,
        message: success
          ? 'Connection successful!'
          : 'Connection failed. Please check your credentials and network connectivity.'
      })
    } catch (error) {
      setTestResult({
        success: false,
        message: `Connection test failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      })
    } finally {
      setIsTestingConnection(false)
    }
  }

  const onSubmit = (data: SMBConfigForm) => {
    // Create access entry
    const accessName = `${data.name.toLowerCase().replace(/\s+/g, '_')}_access`
    const access: ConfigurationAccess = {
      name: accessName,
      type: 'credentials',
      account: data.username,
      secret: data.password,
    }

    // Create source entry
    const url = `smb://${data.host}:${data.port}/${data.share_name}${data.path ? data.path : ''}`
    const source: ConfigurationSource = {
      type: 'samba',
      url,
      access: accessName,
    }

    // Add to global configuration
    addAccess(access)
    addSource(source)

    if (editingIndex !== null) {
      // Update existing config
      const updatedConfigs = [...smbConfigs]
      updatedConfigs[editingIndex] = data as SMBConnectionConfig
      setSmbConfigs(updatedConfigs)
    } else {
      // Add new config
      setSmbConfigs([...smbConfigs, data as SMBConnectionConfig])
    }

    // Reset form for next entry
    setEditingIndex(null)
    reset({
      name: '',
      host: '',
      port: 445,
      share_name: '',
      username: '',
      password: '',
      domain: '',
      path: '',
      enabled: true,
    })
    setTestResult(null)
  }

  const addNewConfig = () => {
    setEditingIndex(null)
    reset({
      name: '',
      host: '',
      port: 445,
      share_name: '',
      username: '',
      password: '',
      domain: '',
      path: '',
      enabled: true,
    })
    setTestResult(null)
  }

  const removeConfig = (index: number) => {
    const updatedConfigs = smbConfigs.filter((_, i) => i !== index)
    setSmbConfigs(updatedConfigs)
    if (editingIndex === index) {
      setEditingIndex(null)
      reset()
      setTestResult(null)
    }
  }

  return (
    <div className="space-y-6">
      <div className="text-center space-y-4">
        <div className="mx-auto w-16 h-16 bg-blue-100 rounded-full flex items-center justify-center">
          <Settings className="h-8 w-8 text-blue-600" />
        </div>
        <h2 className="text-xl font-bold text-gray-900">SMB Configuration</h2>
        <p className="text-gray-600">
          Configure SMB connections for your selected devices
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Configuration Form */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5" />
              {editingIndex !== null ? 'Edit Configuration' : 'Add Configuration'}
            </CardTitle>
            <CardDescription>
              Enter the SMB connection details
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Configuration Name</label>
                <Input
                  {...register('name')}
                  placeholder="e.g., Media Server"
                  className={errors.name ? 'border-red-500' : ''}
                />
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name.message}</p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Host/IP Address</label>
                  <Input
                    {...register('host')}
                    placeholder="192.168.1.100"
                    className={errors.host ? 'border-red-500' : ''}
                  />
                  {errors.host && (
                    <p className="text-red-500 text-sm mt-1">{errors.host.message}</p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Port</label>
                  <Input
                    type="number"
                    {...register('port', { valueAsNumber: true })}
                    className={errors.port ? 'border-red-500' : ''}
                  />
                  {errors.port && (
                    <p className="text-red-500 text-sm mt-1">{errors.port.message}</p>
                  )}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Share Name</label>
                <Input
                  {...register('share_name')}
                  placeholder="shared"
                  className={errors.share_name ? 'border-red-500' : ''}
                />
                {errors.share_name && (
                  <p className="text-red-500 text-sm mt-1">{errors.share_name.message}</p>
                )}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Username</label>
                  <Input
                    {...register('username')}
                    placeholder="username"
                    className={errors.username ? 'border-red-500' : ''}
                  />
                  {errors.username && (
                    <p className="text-red-500 text-sm mt-1">{errors.username.message}</p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Domain (optional)</label>
                  <Input
                    {...register('domain')}
                    placeholder="WORKGROUP"
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Password</label>
                <div className="relative">
                  <Input
                    type={showPassword ? 'text' : 'password'}
                    {...register('password')}
                    placeholder="password"
                    className={`pr-10 ${errors.password ? 'border-red-500' : ''}`}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 transform -translate-y-1/2 text-gray-500"
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
                {errors.password && (
                  <p className="text-red-500 text-sm mt-1">{errors.password.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Path (optional)</label>
                <Input
                  {...register('path')}
                  placeholder="/media/movies"
                />
              </div>

              {/* Test Connection */}
              <div className="space-y-3">
                <Button
                  type="button"
                  variant="outline"
                  onClick={handleTestConnection}
                  disabled={isTestingConnection}
                  className="w-full flex items-center gap-2"
                >
                  {isTestingConnection ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <TestTube className="h-4 w-4" />
                  )}
                  {isTestingConnection ? 'Testing...' : 'Test Connection'}
                </Button>

                {testResult && (
                  <div className={`p-3 rounded-lg flex items-center gap-2 ${
                    testResult.success
                      ? 'bg-green-50 border border-green-200 text-green-800'
                      : 'bg-red-50 border border-red-200 text-red-800'
                  }`}>
                    {testResult.success ? (
                      <CheckCircle className="h-4 w-4" />
                    ) : (
                      <AlertCircle className="h-4 w-4" />
                    )}
                    <span className="text-sm">{testResult.message}</span>
                  </div>
                )}
              </div>

              <div className="flex gap-3">
                <Button type="submit" className="flex-1">
                  {editingIndex !== null ? 'Update Configuration' : 'Add Configuration'}
                </Button>
                {editingIndex !== null && (
                  <Button type="button" variant="outline" onClick={addNewConfig}>
                    Cancel
                  </Button>
                )}
              </div>
            </form>
          </CardContent>
        </Card>

        {/* Configuration List */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span className="flex items-center gap-2">
                <Folder className="h-5 w-5" />
                Configured Sources ({smbConfigs.length})
              </span>
              <Button
                variant="outline"
                size="sm"
                onClick={addNewConfig}
                className="flex items-center gap-2"
              >
                <Plus className="h-4 w-4" />
                Add New
              </Button>
            </CardTitle>
            <CardDescription>
              Manage your SMB source configurations
            </CardDescription>
          </CardHeader>
          <CardContent>
            {smbConfigs.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Folder className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p className="text-lg font-medium">No configurations yet</p>
                <p className="text-sm">
                  Add your first SMB configuration to get started
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {smbConfigs.map((config, index) => (
                  <div
                    key={index}
                    className={`p-4 border rounded-lg transition-colors ${
                      editingIndex === index
                        ? 'border-blue-500 bg-blue-50'
                        : 'border-gray-200 hover:border-gray-300'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex-1">
                        <div className="font-medium">{config.name}</div>
                        <div className="text-sm text-gray-500">
                          {config.host}:{config.port} → {config.share_name}
                          {config.path && ` (${config.path})`}
                        </div>
                        <div className="text-xs text-gray-400">
                          User: {config.username}
                          {config.domain && ` • Domain: ${config.domain}`}
                        </div>
                      </div>
                      <div className="flex items-center gap-2">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => startEditing(index, config)}
                        >
                          Edit
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => removeConfig(index)}
                          className="text-red-600 hover:text-red-700"
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

      {smbConfigs.length > 0 && (
        <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center gap-2 text-green-800">
            <CheckCircle className="h-4 w-4" />
            <span className="font-medium">
              {smbConfigs.length} SMB source(s) configured
            </span>
          </div>
          <p className="text-sm text-green-700 mt-1">
            Click "Next" to manage your configuration file
          </p>
        </div>
      )}
    </div>
  )
}