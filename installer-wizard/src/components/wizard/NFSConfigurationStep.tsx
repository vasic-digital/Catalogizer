import { useState, useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Button } from '../ui/Button'
import { Input } from '../ui/Input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../ui/Card'
import { useWizard } from '../../contexts/WizardContext'
import { useConfiguration } from '../../contexts/ConfigurationContext'
import { TauriService } from '../../services/tauri'
import { NFSConnectionConfig } from '../../types'
import {
  Settings,
  TestTube,
  CheckCircle,
  AlertCircle,
  Loader2,
  Folder,
  Plus,
  Trash2
} from 'lucide-react'

const nfsConfigSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  host: z.string().min(1, 'Host is required'),
  path: z.string().min(1, 'Path is required'),
  mount_point: z.string().min(1, 'Mount point is required'),
  options: z.string().optional(),
  enabled: z.boolean().default(true),
})

type NFSConfigForm = z.infer<typeof nfsConfigSchema>

export default function NFSConfigurationStep() {
  const { setCanNext } = useWizard()
  const { state: configState } = useConfiguration()
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [isTestingConnection, setIsTestingConnection] = useState(false)
  const [nfsConfigs, setNfsConfigs] = useState<NFSConnectionConfig[]>([])
  const [editingIndex, setEditingIndex] = useState<number | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    reset,
  } = useForm<NFSConfigForm>({
    resolver: zodResolver(nfsConfigSchema),
    defaultValues: {
      options: 'vers=3',
      enabled: true,
    }
  })

  const watchedValues = watch()

  useEffect(() => {
    // Can proceed if we have at least one valid NFS configuration
    setCanNext(nfsConfigs.length > 0)
  }, [nfsConfigs, setCanNext])

  useEffect(() => {
    // Pre-populate with selected hosts from network scan if they have NFS ports (2049)
    if (configState.selectedHosts.length > 0 && nfsConfigs.length === 0) {
      const nfsHosts = configState.selectedHosts.filter(host => host.open_ports.includes(2049))
      const defaultConfigs = nfsHosts.map((host, index) => ({
        name: `NFS Server ${index + 1}`,
        host: host.ip,
        path: '/export/data',
        mount_point: `/mnt/nfs${index + 1}`,
        options: 'vers=3',
        enabled: true,
      }))
      setNfsConfigs(defaultConfigs)
      if (defaultConfigs.length > 0) {
        startEditing(0, defaultConfigs[0])
      }
    }
  }, [configState.selectedHosts, nfsConfigs.length])

  const startEditing = (index: number, config: NFSConnectionConfig) => {
    setEditingIndex(index)
    reset(config)
    setTestResult(null)
  }

  const handleTestConnection = async () => {
    const values = watchedValues
    if (!values.host || !values.path || !values.mount_point) {
      setTestResult({
        success: false,
        message: 'Please fill in all required fields before testing'
      })
      return
    }

    setIsTestingConnection(true)
    setTestResult(null)

    try {
      const success = await TauriService.testNFSConnection(
        values.host,
        values.path,
        values.mount_point,
        values.options
      )

      setTestResult({
        success,
        message: success
          ? 'Connection successful!'
          : 'Connection failed. Please check your network connectivity and NFS configuration.'
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

  const onSubmit = (data: NFSConfigForm) => {
    if (editingIndex !== null) {
      // Update existing config
      const updatedConfigs = [...nfsConfigs]
      updatedConfigs[editingIndex] = data as NFSConnectionConfig
      setNfsConfigs(updatedConfigs)
    } else {
      // Add new config
      setNfsConfigs([...nfsConfigs, data as NFSConnectionConfig])
    }

    // Reset form for next entry
    setEditingIndex(null)
    reset({
      name: '',
      host: '',
      path: '/export/data',
      mount_point: '/mnt/nfs',
      options: 'vers=3',
      enabled: true,
    })
    setTestResult(null)
  }

  const addNewConfig = () => {
    setEditingIndex(null)
    reset({
      name: '',
      host: '',
      path: '/export/data',
      mount_point: '/mnt/nfs',
      options: 'vers=3',
      enabled: true,
    })
    setTestResult(null)
  }

  const removeConfig = (index: number) => {
    const updatedConfigs = nfsConfigs.filter((_, i) => i !== index)
    setNfsConfigs(updatedConfigs)
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
        <h2 className="text-xl font-bold text-gray-900">NFS Configuration</h2>
        <p className="text-gray-600">
          Configure NFS connections for your selected devices
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
              Enter the NFS connection details
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Configuration Name</label>
                <Input
                  {...register('name')}
                  placeholder="e.g., Media NFS Server"
                  className={errors.name ? 'border-red-500' : ''}
                />
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name.message}</p>
                )}
              </div>

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

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-sm font-medium mb-1">Export Path</label>
                  <Input
                    {...register('path')}
                    placeholder="/export/data"
                    className={errors.path ? 'border-red-500' : ''}
                  />
                  {errors.path && (
                    <p className="text-red-500 text-sm mt-1">{errors.path.message}</p>
                  )}
                </div>

                <div>
                  <label className="block text-sm font-medium mb-1">Mount Point</label>
                  <Input
                    {...register('mount_point')}
                    placeholder="/mnt/nfs"
                    className={errors.mount_point ? 'border-red-500' : ''}
                  />
                  {errors.mount_point && (
                    <p className="text-red-500 text-sm mt-1">{errors.mount_point.message}</p>
                  )}
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Mount Options (optional)</label>
                <Input
                  {...register('options')}
                  placeholder="vers=3"
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
                Configured Sources ({nfsConfigs.length})
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
              Manage your NFS source configurations
            </CardDescription>
          </CardHeader>
          <CardContent>
            {nfsConfigs.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Folder className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p className="text-lg font-medium">No configurations yet</p>
                <p className="text-sm">
                  Add your first NFS configuration to get started
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {nfsConfigs.map((config, index) => (
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
                          {config.host}:{config.path} → {config.mount_point}
                        </div>
                        <div className="text-xs text-gray-400">
                          Options: {config.options || 'default'}
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

      {nfsConfigs.length > 0 && (
        <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center gap-2 text-green-800">
            <CheckCircle className="h-4 w-4" />
            <span className="font-medium">
              {nfsConfigs.length} NFS source(s) configured
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