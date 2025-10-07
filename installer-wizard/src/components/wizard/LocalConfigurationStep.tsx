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
import { LocalConnectionConfig } from '../../types'
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

const localConfigSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  base_path: z.string().min(1, 'Base path is required'),
  enabled: z.boolean().default(true),
})

type LocalConfigForm = z.infer<typeof localConfigSchema>

export default function LocalConfigurationStep() {
  const { setCanNext } = useWizard()
  const { } = useConfiguration()
  const [testResult, setTestResult] = useState<{ success: boolean; message: string } | null>(null)
  const [isTestingConnection, setIsTestingConnection] = useState(false)
  const [localConfigs, setLocalConfigs] = useState<LocalConnectionConfig[]>([])
  const [editingIndex, setEditingIndex] = useState<number | null>(null)

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    reset,
  } = useForm<LocalConfigForm>({
    resolver: zodResolver(localConfigSchema),
    defaultValues: {
      enabled: true,
    }
  })

  const watchedValues = watch()

  useEffect(() => {
    // Can proceed if we have at least one valid local configuration
    setCanNext(localConfigs.length > 0)
  }, [localConfigs, setCanNext])

  useEffect(() => {
    // Pre-populate with default local paths
    if (localConfigs.length === 0) {
      const defaultConfigs = [
        {
          name: 'Local Media',
          base_path: '/home/user/media',
          enabled: true,
        }
      ]
      setLocalConfigs(defaultConfigs)
      startEditing(0, defaultConfigs[0])
    }
  }, [localConfigs.length])

  const startEditing = (index: number, config: LocalConnectionConfig) => {
    setEditingIndex(index)
    reset(config)
    setTestResult(null)
  }

  const handleTestConnection = async () => {
    const values = watchedValues
    if (!values.base_path) {
      setTestResult({
        success: false,
        message: 'Please fill in the base path before testing'
      })
      return
    }

    setIsTestingConnection(true)
    setTestResult(null)

    try {
      const success = await TauriService.testLocalConnection(values.base_path)

      setTestResult({
        success,
        message: success
          ? 'Path accessible!'
          : 'Path not accessible. Please check permissions and path existence.'
      })
    } catch (error) {
      setTestResult({
        success: false,
        message: `Path test failed: ${error instanceof Error ? error.message : 'Unknown error'}`
      })
    } finally {
      setIsTestingConnection(false)
    }
  }

  const onSubmit = (data: LocalConfigForm) => {
    if (editingIndex !== null) {
      // Update existing config
      const updatedConfigs = [...localConfigs]
      updatedConfigs[editingIndex] = data as LocalConnectionConfig
      setLocalConfigs(updatedConfigs)
    } else {
      // Add new config
      setLocalConfigs([...localConfigs, data as LocalConnectionConfig])
    }

    // Reset form for next entry
    setEditingIndex(null)
    reset({
      name: '',
      base_path: '',
      enabled: true,
    })
    setTestResult(null)
  }

  const addNewConfig = () => {
    setEditingIndex(null)
    reset({
      name: '',
      base_path: '',
      enabled: true,
    })
    setTestResult(null)
  }

  const removeConfig = (index: number) => {
    const updatedConfigs = localConfigs.filter((_, i) => i !== index)
    setLocalConfigs(updatedConfigs)
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
        <h2 className="text-xl font-bold text-gray-900">Local Configuration</h2>
        <p className="text-gray-600">
          Configure local filesystem paths for your media
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
              Enter the local filesystem path details
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
              <div>
                <label className="block text-sm font-medium mb-1">Configuration Name</label>
                <Input
                  {...register('name')}
                  placeholder="e.g., Local Media Library"
                  className={errors.name ? 'border-red-500' : ''}
                />
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name.message}</p>
                )}
              </div>

              <div>
                <label className="block text-sm font-medium mb-1">Base Path</label>
                <Input
                  {...register('base_path')}
                  placeholder="/home/user/media"
                  className={errors.base_path ? 'border-red-500' : ''}
                />
                {errors.base_path && (
                  <p className="text-red-500 text-sm mt-1">{errors.base_path.message}</p>
                )}
              </div>

              {/* Test Path */}
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
                  {isTestingConnection ? 'Testing...' : 'Test Path'}
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
                Configured Sources ({localConfigs.length})
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
              Manage your local filesystem source configurations
            </CardDescription>
          </CardHeader>
          <CardContent>
            {localConfigs.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <Folder className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                <p className="text-lg font-medium">No configurations yet</p>
                <p className="text-sm">
                  Add your first local configuration to get started
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {localConfigs.map((config, index) => (
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
                        <div className="text-sm text-gray-500 break-all">
                          {config.base_path}
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

      {localConfigs.length > 0 && (
        <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="flex items-center gap-2 text-green-800">
            <CheckCircle className="h-4 w-4" />
            <span className="font-medium">
              {localConfigs.length} local source(s) configured
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