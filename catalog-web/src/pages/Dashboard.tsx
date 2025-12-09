import { useState, useEffect } from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { useQuery } from '@tanstack/react-query'
import { DashboardStats } from '@/components/dashboard/DashboardStats'
import { MediaDistributionChart } from '@/components/dashboard/MediaDistributionChart'
import { ActivityFeed } from '@/components/dashboard/ActivityFeed'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { mediaApi } from '@/lib/mediaApi'
import {
  Film,
  Upload,
  Search,
  Settings,
  Activity,
  HardDrive,
  Zap,
  Clock
} from 'lucide-react'
import { motion } from 'framer-motion'
import toast from 'react-hot-toast'
import type { MediaStats, UserStats, QuickAction } from '@/types/dashboard'

const QuickActions: React.FC = () => {
  const handleUploadMedia = () => {
    // Navigate to upload page or open upload modal
    toast.success('Opening upload interface...')
  }

  const handleScanLibrary = () => {
    // Trigger library scan
    toast.promise(
      mediaApi.analyzeDirectory('/'),
      {
        loading: 'Scanning library...',
        success: 'Library scan started',
        error: 'Failed to start scan'
      }
    )
  }

  const handleSearchMedia = () => {
    // Navigate to media browser with focus on search
    toast.success('Opening search interface...')
  }

  const handleSettings = () => {
    // Navigate to settings
    toast.success('Opening settings...')
  }

  const quickActions: QuickAction[] = [
    {
      id: 'upload',
      title: 'Upload Media',
      description: 'Add new media to your library',
      icon: Upload,
      action: handleUploadMedia,
      variant: 'default'
    },
    {
      id: 'scan',
      title: 'Scan Library',
      description: 'Update media library with new files',
      icon: Activity,
      action: handleScanLibrary,
      variant: 'outline'
    },
    {
      id: 'search',
      title: 'Search',
      description: 'Find specific media quickly',
      icon: Search,
      action: handleSearchMedia,
      variant: 'outline'
    },
    {
      id: 'settings',
      title: 'Settings',
      description: 'Configure system preferences',
      icon: Settings,
      action: handleSettings,
      variant: 'outline'
    }
  ]

  return (
    <Card>
      <CardHeader>
        <CardTitle>Quick Actions</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          {quickActions.map(action => {
            const Icon = action.icon
            return (
              <Button
                key={action.id}
                variant={action.variant}
                onClick={action.action}
                className="h-auto p-4 flex flex-col items-center space-y-2"
              >
                <Icon className="w-6 h-6" />
                <span className="text-sm font-medium">{action.title}</span>
              </Button>
            )
          })}
        </div>
      </CardContent>
    </Card>
  )
}

const SystemStatus: React.FC = () => {
  const [status, setStatus] = useState({
    cpu: 45,
    memory: 62,
    disk: 78,
    network: true,
    uptime: '5d 12h 34m'
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          <Activity className="h-5 w-5" />
          System Status
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="space-y-4">
          {/* CPU Usage */}
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>CPU Usage</span>
              <span>{status.cpu}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full ${
                  status.cpu > 80 ? 'bg-red-500' : 
                  status.cpu > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${status.cpu}%` }}
              />
            </div>
          </div>

          {/* Memory Usage */}
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Memory Usage</span>
              <span>{status.memory}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full ${
                  status.memory > 80 ? 'bg-red-500' : 
                  status.memory > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${status.memory}%` }}
              />
            </div>
          </div>

          {/* Disk Usage */}
          <div className="space-y-2">
            <div className="flex justify-between text-sm">
              <span>Disk Usage</span>
              <span>{status.disk}%</span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className={`h-2 rounded-full ${
                  status.disk > 80 ? 'bg-red-500' : 
                  status.disk > 60 ? 'bg-yellow-500' : 'bg-green-500'
                }`}
                style={{ width: `${status.disk}%` }}
              />
            </div>
          </div>

          {/* Additional Status */}
          <div className="flex justify-between text-sm pt-2 border-t">
            <span>Network</span>
            <span className={status.network ? 'text-green-600' : 'text-red-600'}>
              {status.network ? 'Online' : 'Offline'}
            </span>
          </div>
          <div className="flex justify-between text-sm">
            <span>Uptime</span>
            <span>{status.uptime}</span>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}

export const Dashboard: React.FC = () => {
  const { user } = useAuth()

  // Fetch media statistics
  const { 
    data: mediaStats, 
    isLoading: mediaLoading, 
    error: mediaError 
  } = useQuery(
    ['media-stats'],
    () => mediaApi.getMediaStats(),
    {
      refetchInterval: 30000, // Refresh every 30 seconds
      staleTime: 10000
    }
  )

  // Fetch user statistics (mock for now)
  const userStats: UserStats = {
    active_users: 3,
    total_users: 12,
    sessions_today: 24,
    avg_session_duration: 45
  }

  // Handle errors
  useEffect(() => {
    if (mediaError) {
      toast.error('Failed to load media statistics')
    }
  }, [mediaError])

  return (
    <div className="space-y-6">
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        className="flex items-center justify-between"
      >
        <div>
          <h1 className="text-3xl font-bold text-gray-900">
            Welcome back, {user?.username || 'User'}!
          </h1>
          <p className="text-gray-600">
            Here's what's happening with your media library today.
          </p>
        </div>
        
        <div className="flex items-center space-x-2">
          <Button variant="outline">
            <Clock className="w-4 h-4 mr-2" />
            Last updated: {new Date().toLocaleTimeString()}
          </Button>
        </div>
      </motion.div>

      {/* Main Stats Grid */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.1 }}
      >
        <DashboardStats
          mediaStats={mediaStats}
          userStats={userStats}
          loading={mediaLoading}
        />
      </motion.div>

      {/* Charts and Activity Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Media Distribution Chart */}
        <motion.div
          initial={{ opacity: 0, x: -20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
        >
          <MediaDistributionChart
            data={mediaStats?.by_type}
            loading={mediaLoading}
          />
        </motion.div>

        {/* System Status */}
        <motion.div
          initial={{ opacity: 0, x: 20 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ delay: 0.2 }}
        >
          <SystemStatus />
        </motion.div>
      </div>

      {/* Activity Feed and Quick Actions */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Activity Feed */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="lg:col-span-2"
        >
          <ActivityFeed limit={8} />
        </motion.div>

        {/* Quick Actions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
        >
          <QuickActions />
        </motion.div>
      </div>
    </div>
  )
}