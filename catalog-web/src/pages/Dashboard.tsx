import React from 'react'
import { useAuth } from '@/contexts/AuthContext'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import {
  Database,
  Film,
  Music,
  Gamepad2,
  Monitor,
  BookOpen,
  TrendingUp,
  Users,
  Activity,
  HardDrive
} from 'lucide-react'
import { motion } from 'framer-motion'

const StatCard: React.FC<{
  title: string
  value: string
  icon: React.ReactNode
  change?: string
  changeType?: 'positive' | 'negative'
}> = ({ title, value, icon, change, changeType }) => (
  <Card className="hover:shadow-lg transition-shadow duration-200">
    <CardContent className="p-6">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600 dark:text-gray-400">{title}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{value}</p>
          {change && (
            <p className={`text-sm ${
              changeType === 'positive' ? 'text-green-600 dark:text-green-400' : 'text-red-600 dark:text-red-400'
            }`}>
              {change}
            </p>
          )}
        </div>
        <div className="text-blue-600 dark:text-blue-400">
          {icon}
        </div>
      </div>
    </CardContent>
  </Card>
)

const QuickActionCard: React.FC<{
  title: string
  description: string
  icon: React.ReactNode
  onClick: () => void
}> = ({ title, description, icon, onClick }) => (
  <Card className="hover:shadow-lg transition-all duration-200 cursor-pointer hover:scale-105" onClick={onClick}>
    <CardContent className="p-6">
      <div className="flex items-center space-x-4">
        <div className="text-blue-600 dark:text-blue-400">
          {icon}
        </div>
        <div>
          <h3 className="font-semibold text-gray-900 dark:text-white">{title}</h3>
          <p className="text-sm text-gray-600 dark:text-gray-400">{description}</p>
        </div>
      </div>
    </CardContent>
  </Card>
)

export const Dashboard: React.FC = () => {
  const { user } = useAuth()

  const stats = [
    {
      title: 'Total Media Items',
      value: '1,234',
      icon: <Database className="h-8 w-8" />,
      change: '+12% from last month',
      changeType: 'positive' as const,
    },
    {
      title: 'Movies',
      value: '456',
      icon: <Film className="h-8 w-8" />,
      change: '+8% from last month',
      changeType: 'positive' as const,
    },
    {
      title: 'Music Albums',
      value: '789',
      icon: <Music className="h-8 w-8" />,
      change: '+15% from last month',
      changeType: 'positive' as const,
    },
    {
      title: 'Games',
      value: '123',
      icon: <Gamepad2 className="h-8 w-8" />,
      change: '+5% from last month',
      changeType: 'positive' as const,
    },
  ]

  const quickActions = [
    {
      title: 'Browse Media',
      description: 'Explore your media collection',
      icon: <Database className="h-6 w-6" />,
      onClick: () => console.log('Browse media'),
    },
    {
      title: 'View Analytics',
      description: 'See detailed statistics',
      icon: <TrendingUp className="h-6 w-6" />,
      onClick: () => console.log('View analytics'),
    },
    {
      title: 'System Health',
      description: 'Check system status',
      icon: <Activity className="h-6 w-6" />,
      onClick: () => console.log('System health'),
    },
    {
      title: 'Storage Usage',
      description: 'Monitor disk usage',
      icon: <HardDrive className="h-6 w-6" />,
      onClick: () => console.log('Storage usage'),
    },
  ]

  if (user?.role === 'admin') {
    quickActions.push({
      title: 'User Management',
      description: 'Manage system users',
      icon: <Users className="h-6 w-6" />,
      onClick: () => console.log('User management'),
    })
  }

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Welcome back, {user?.first_name || user?.username}!
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-2">
            Here&apos;s what&apos;s happening with your media collection today.
          </p>
        </div>

        {/* Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {stats.map((stat, index) => (
            <motion.div
              key={stat.title}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <StatCard {...stat} />
            </motion.div>
          ))}
        </div>

        {/* Quick Actions */}
        <div className="mb-8">
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
            Quick Actions
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {quickActions.map((action, index) => (
              <motion.div
                key={action.title}
                initial={{ opacity: 0, x: -20 }}
                animate={{ opacity: 1, x: 0 }}
                transition={{ duration: 0.5, delay: index * 0.1 }}
              >
                <QuickActionCard {...action} />
              </motion.div>
            ))}
          </div>
        </div>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
            <CardDescription>
              Latest changes in your media collection
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {[
                {
                  type: 'Movie',
                  title: 'The Matrix (1999)',
                  action: 'Added to collection',
                  time: '2 hours ago',
                  icon: <Film className="h-4 w-4" />,
                },
                {
                  type: 'Album',
                  title: 'Dark Side of the Moon',
                  action: 'Metadata updated',
                  time: '4 hours ago',
                  icon: <Music className="h-4 w-4" />,
                },
                {
                  type: 'Game',
                  title: 'Cyberpunk 2077',
                  action: 'Quality analysis completed',
                  time: '6 hours ago',
                  icon: <Gamepad2 className="h-4 w-4" />,
                },
                {
                  type: 'Software',
                  title: 'Adobe Photoshop 2024',
                  action: 'New version detected',
                  time: '1 day ago',
                  icon: <Monitor className="h-4 w-4" />,
                },
              ].map((activity, index) => (
                <div key={index} className="flex items-center space-x-4 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                  <div className="text-blue-600 dark:text-blue-400">
                    {activity.icon}
                  </div>
                  <div className="flex-1">
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-medium text-gray-900 dark:text-white">
                        {activity.title}
                      </span>
                      <span className="text-xs px-2 py-1 bg-blue-100 text-blue-800 rounded-full dark:bg-blue-900 dark:text-blue-200">
                        {activity.type}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {activity.action}
                    </p>
                  </div>
                  <span className="text-xs text-gray-500 dark:text-gray-400">
                    {activity.time}
                  </span>
                </div>
              ))}
            </div>
            <div className="mt-6 text-center">
              <Button variant="outline">
                View All Activity
              </Button>
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}