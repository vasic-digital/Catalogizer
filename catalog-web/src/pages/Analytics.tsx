import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/Card'
import { mediaApi } from '@/lib/mediaApi'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
  Area,
  AreaChart
} from 'recharts'
import {
  TrendingUp,
  Database,
  HardDrive,
  Star,
  Film,
  Music,
  Gamepad2,
  Monitor
} from 'lucide-react'
import { motion } from 'framer-motion'

const COLORS = [
  '#3B82F6', '#EF4444', '#10B981', '#F59E0B',
  '#8B5CF6', '#EC4899', '#14B8A6', '#F97316'
]

const StatCard: React.FC<{
  title: string
  value: string
  icon: React.ReactNode
  change?: string
  changeType?: 'positive' | 'negative'
  trend?: number[]
}> = ({ title, value, icon, change, changeType, trend }) => (
  <Card className="relative overflow-hidden">
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
      {trend && (
        <div className="mt-4 h-8">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={trend.map((value, index) => ({ value, index }))}>
              <Line
                type="monotone"
                dataKey="value"
                stroke={changeType === 'positive' ? '#10B981' : '#EF4444'}
                strokeWidth={2}
                dot={false}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>
      )}
    </CardContent>
  </Card>
)

export const Analytics: React.FC = () => {
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['media-stats'],
    queryFn: mediaApi.getMediaStats,
    staleTime: 1000 * 60 * 5,
  })

  const { data: recentMedia } = useQuery({
    queryKey: ['recent-media', 20],
    queryFn: () => mediaApi.getRecentMedia(20),
    staleTime: 1000 * 60 * 5,
  })

  // Transform stats data for charts
  const mediaTypeData = stats ? Object.entries(stats.by_type).map(([type, count]) => ({
    name: type.replace('_', ' ').toUpperCase(),
    value: count,
    count
  })) : []

  const qualityData = stats ? Object.entries(stats.by_quality).map(([quality, count]) => ({
    name: quality.toUpperCase(),
    value: count,
    count
  })) : []

  // Simulate growth trends (in a real app, this would come from the API)
  const growthTrend = Array.from({ length: 30 }, (_, i) => ({
    day: i + 1,
    items: Math.floor(Math.random() * 50) + (stats?.total_items || 0) - 1000 + i * 10,
    size: Math.floor(Math.random() * 100) + 500 + i * 20
  }))

  const weeklyAdditions = Array.from({ length: 7 }, (_, i) => ({
    day: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i],
    additions: Math.floor(Math.random() * 20) + 5
  }))

  if (statsLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-300 dark:bg-gray-600 rounded w-1/4 mb-4" />
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-32 bg-gray-300 dark:bg-gray-600 rounded-xl" />
            ))}
          </div>
        </div>
      </div>
    )
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
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-2">
            Analytics Dashboard
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Insights and statistics about your media collection
          </p>
        </div>

        {/* Key Metrics */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <StatCard
            title="Total Media Items"
            value={stats?.total_items?.toLocaleString() || '0'}
            icon={<Database className="h-8 w-8" />}
            change="+12% from last month"
            changeType="positive"
            trend={[45, 52, 48, 61, 67, 73, 82]}
          />
          <StatCard
            title="Storage Used"
            value={`${((stats?.total_size || 0) / (1024 ** 3)).toFixed(1)} GB`}
            icon={<HardDrive className="h-8 w-8" />}
            change="+8.2 GB this week"
            changeType="positive"
            trend={[120, 125, 128, 135, 142, 148, 155]}
          />
          <StatCard
            title="Recent Additions"
            value={stats?.recent_additions?.toString() || '0'}
            icon={<TrendingUp className="h-8 w-8" />}
            change="+5 from yesterday"
            changeType="positive"
            trend={[8, 12, 15, 11, 18, 22, 25]}
          />
          <StatCard
            title="Media Types"
            value={Object.keys(stats?.by_type || {}).length.toString()}
            icon={<Star className="h-8 w-8" />}
            change="2 new types detected"
            changeType="positive"
            trend={[6, 7, 7, 8, 8, 9, 10]}
          />
        </div>

        {/* Charts Row 1 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Media Types Distribution */}
          <Card>
            <CardHeader>
              <CardTitle>Media Types Distribution</CardTitle>
              <CardDescription>
                Breakdown of your media collection by type
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie
                      data={mediaTypeData}
                      cx="50%"
                      cy="50%"
                      labelLine={false}
                      label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                      outerRadius={80}
                      fill="#8884d8"
                      dataKey="value"
                    >
                      {mediaTypeData.map((entry, index) => (
                        <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip formatter={(value) => [value, 'Items']} />
                  </PieChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Quality Distribution */}
          <Card>
            <CardHeader>
              <CardTitle>Quality Distribution</CardTitle>
              <CardDescription>
                Media items by quality level
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={qualityData}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="name" />
                    <YAxis />
                    <Tooltip formatter={(value) => [value, 'Items']} />
                    <Bar dataKey="value" fill="#3B82F6" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Charts Row 2 */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
          {/* Growth Trend */}
          <Card>
            <CardHeader>
              <CardTitle>Collection Growth</CardTitle>
              <CardDescription>
                Media items and storage growth over time
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <AreaChart data={growthTrend}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="day" />
                    <YAxis />
                    <Tooltip />
                    <Area
                      type="monotone"
                      dataKey="items"
                      stackId="1"
                      stroke="#3B82F6"
                      fill="#3B82F6"
                      fillOpacity={0.6}
                    />
                  </AreaChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>

          {/* Weekly Additions */}
          <Card>
            <CardHeader>
              <CardTitle>Weekly Activity</CardTitle>
              <CardDescription>
                New media items added each day this week
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="h-80">
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={weeklyAdditions}>
                    <CartesianGrid strokeDasharray="3 3" />
                    <XAxis dataKey="day" />
                    <YAxis />
                    <Tooltip formatter={(value) => [value, 'Items Added']} />
                    <Bar dataKey="additions" fill="#10B981" />
                  </BarChart>
                </ResponsiveContainer>
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Recent Activity */}
        <Card>
          <CardHeader>
            <CardTitle>Recently Added Media</CardTitle>
            <CardDescription>
              Your latest additions to the collection
            </CardDescription>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              {recentMedia?.slice(0, 10).map((item) => {
                const getIcon = (type: string) => {
                  switch (type.toLowerCase()) {
                    case 'movie':
                    case 'tv_show':
                      return <Film className="h-4 w-4" />
                    case 'music':
                      return <Music className="h-4 w-4" />
                    case 'game':
                      return <Gamepad2 className="h-4 w-4" />
                    case 'software':
                      return <Monitor className="h-4 w-4" />
                    default:
                      return <Database className="h-4 w-4" />
                  }
                }

                return (
                  <div key={item.id} className="flex items-center space-x-4 p-3 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors">
                    <div className="text-blue-600 dark:text-blue-400">
                      {getIcon(item.media_type)}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-gray-900 dark:text-white">
                          {item.title}
                        </span>
                        {item.year && (
                          <span className="text-sm text-gray-500 dark:text-gray-400">
                            ({item.year})
                          </span>
                        )}
                      </div>
                      <div className="flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-400">
                        <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs dark:bg-blue-900 dark:text-blue-200">
                          {item.media_type.replace('_', ' ')}
                        </span>
                        {item.quality && (
                          <span>{item.quality.toUpperCase()}</span>
                        )}
                        {item.file_size && (
                          <span>{((item.file_size || 0) / (1024 ** 2)).toFixed(0)} MB</span>
                        )}
                      </div>
                    </div>
                    <div className="text-xs text-gray-500 dark:text-gray-400">
                      {new Date(item.created_at).toLocaleDateString()}
                    </div>
                  </div>
                )
              })}
            </div>
          </CardContent>
        </Card>
      </motion.div>
    </div>
  )
}