import React, { useState, useEffect, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import {
  BarChart3,
  TrendingUp,
  TrendingDown,
  Play,
  Clock,
  Users,
  Share2,
  Download,
  Star,
  Eye,
  Filter,
  Calendar,
  PieChart,
  Activity,
  FileAudio,
  FileVideo,
  FileImage,
  FileText,
  RefreshCw,
  Download as DownloadIcon,
  Upload,
  Link,
  Mail,
  MessageCircle,
  X
} from 'lucide-react'
import { Button } from '../ui/Button'
import { Card } from '../ui/Card'
import { Select } from '../ui/Select'
import { Input } from '../ui/Input'
import { Switch } from '../ui/Switch'
import { SmartCollection } from '../../types/collections'
import { useCollection } from '../../hooks/useCollections'

interface CollectionAnalyticsProps {
  collection: SmartCollection
  onClose?: () => void
}

interface AnalyticsData {
  overview: {
    totalItems: number
    totalSize: string
    totalDuration: string
    averageRating: number
    playCount: number
    downloadCount: number
    shareCount: number
    viewCount: number
  }
  mediaBreakdown: {
    music: number
    video: number
    image: number
    document: number
  }
  activity: {
    daily: Array<{ date: string; plays: number; downloads: number; shares: number }>
    weekly: Array<{ week: string; plays: number; downloads: number; shares: number }>
    monthly: Array<{ month: string; plays: number; downloads: number; shares: number }>
  }
  topItems: Array<{
    id: string
    title: string
    type: string
    plays: number
    rating: number
    size: string
  }>
  sharingStats: {
    sharedByType: {
      link: number
      email: number
      social: number
      embed: number
    }
    sharePerformance: Array<{
      date: string
      shares: number
      clicks: number
      downloads: number
    }>
    topReferrers: Array<{
      source: string
      visits: number
      conversion: number
    }>
  }
  engagementMetrics: {
    completionRate: number
    averageWatchTime: string
    skipRate: number
    repeatViews: number
    comments: number
    likes: number
  }
}

const MEDIA_ICONS = {
  music: FileAudio,
  video: FileVideo,
  image: FileImage,
  document: FileText
}

const TIME_RANGES = [
  { value: '7d', label: 'Last 7 Days' },
  { value: '30d', label: 'Last 30 Days' },
  { value: '90d', label: 'Last 90 Days' },
  { value: '1y', label: 'Last Year' },
  { value: 'all', label: 'All Time' }
]

const CHART_TYPES = [
  { value: 'plays', label: 'Plays' },
  { value: 'downloads', label: 'Downloads' },
  { value: 'shares', label: 'Shares' },
  { value: 'views', label: 'Views' }
]

export const CollectionAnalytics: React.FC<CollectionAnalyticsProps> = ({
  collection,
  onClose
}) => {
  const [timeRange, setTimeRange] = useState('30d')
  const [chartType, setChartType] = useState('plays')
  const [isLoading, setIsLoading] = useState(false)
  const [showComparison, setShowComparison] = useState(false)
  const [activeTab, setActiveTab] = useState<'overview' | 'activity' | 'content' | 'sharing' | 'engagement'>('overview')
  
  const { collectionItems, isLoading: itemsLoading } = useCollection(collection?.id || '')

  // Use collection items from props if available
  const items = collectionItems || []

  // Mock analytics data generation
  const analyticsData = useMemo((): AnalyticsData => {
    const totalItems = items.length
    const totalSize = items.reduce((sum: any, item: any) => sum + (item.size || 0), 0)
    const averageRating = items.reduce((sum: any, item: any) => sum + (item.rating || 0), 0) / totalItems || 0
    
    // Generate activity data based on time range
    const generateActivityData = (period: 'daily' | 'weekly' | 'monthly'): any[] => {
      const now = new Date()
      const data: any[] = []
      let periods = 30
      
      if (timeRange === '7d') periods = period === 'daily' ? 7 : period === 'weekly' ? 1 : 1
      else if (timeRange === '30d') periods = period === 'daily' ? 30 : period === 'weekly' ? 4 : 1
      else if (timeRange === '90d') periods = period === 'daily' ? 90 : period === 'weekly' ? 12 : 3
      else if (timeRange === '1y') periods = period === 'daily' ? 365 : period === 'weekly' ? 52 : 12
      
      for (let i = periods - 1; i >= 0; i--) {
        const date = new Date(now)
        
        if (period === 'daily') {
          date.setDate(date.getDate() - i)
          data.push({
            date: date.toLocaleDateString(),
            plays: Math.floor(Math.random() * 100) + 20,
            downloads: Math.floor(Math.random() * 20) + 5,
            shares: Math.floor(Math.random() * 10) + 1
          })
        } else if (period === 'weekly') {
          date.setDate(date.getDate() - (i * 7))
          data.push({
            week: `Week ${i + 1}`,
            plays: Math.floor(Math.random() * 500) + 100,
            downloads: Math.floor(Math.random() * 50) + 10,
            shares: Math.floor(Math.random() * 30) + 5
          })
        } else {
          date.setMonth(date.getMonth() - i)
          data.push({
            month: date.toLocaleDateString('en', { month: 'short', year: 'numeric' }),
            plays: Math.floor(Math.random() * 2000) + 500,
            downloads: Math.floor(Math.random() * 200) + 50,
            shares: Math.floor(Math.random() * 100) + 20
          })
        }
      }
      
      return data
    }

    // Generate media breakdown
    const mediaTypes = items.reduce((acc: any, item: any) => {
      const type = item.media_type || 'document'
      acc[type] = (acc[type] || 0) + 1
      return acc
    }, {} as Record<string, number>)

    // Generate top items
    const topItems = items
      .filter((item: any) => item.rating || Math.random() > 0.5)
      .slice(0, 10)
      .map((item: any) => ({
        id: item.id,
        title: item.title || item.name || 'Unknown',
        type: item.media_type || 'document',
        plays: Math.floor(Math.random() * 1000) + 100,
        rating: item.rating || Math.random() * 5,
        size: formatFileSize(item.size || Math.random() * 1000000000)
      }))
      .sort((a: any, b: any) => b.plays - a.plays)

    return {
      overview: {
        totalItems,
        totalSize: formatFileSize(totalSize),
        totalDuration: formatDuration(items.reduce((sum: any, item: any) => sum + (item.duration || 0), 0)),
        averageRating,
        playCount: Math.floor(Math.random() * 10000) + 1000,
        downloadCount: Math.floor(Math.random() * 1000) + 100,
        shareCount: Math.floor(Math.random() * 500) + 50,
        viewCount: Math.floor(Math.random() * 20000) + 2000
      },
      mediaBreakdown: {
        music: mediaTypes.music || 0,
        video: mediaTypes.video || 0,
        image: mediaTypes.image || 0,
        document: mediaTypes.document || 0
      },
      activity: {
        daily: generateActivityData('daily'),
        weekly: generateActivityData('weekly'),
        monthly: generateActivityData('monthly')
      },
      topItems,
      sharingStats: {
        sharedByType: {
          link: Math.floor(Math.random() * 100) + 20,
          email: Math.floor(Math.random() * 50) + 10,
          social: Math.floor(Math.random() * 30) + 5,
          embed: Math.floor(Math.random() * 20) + 5
        },
        sharePerformance: Array.from({ length: 30 }, (_, i) => ({
          date: new Date(Date.now() - (29 - i) * 24 * 60 * 60 * 1000).toLocaleDateString(),
          shares: Math.floor(Math.random() * 20) + 5,
          clicks: Math.floor(Math.random() * 100) + 20,
          downloads: Math.floor(Math.random() * 50) + 10
        })),
        topReferrers: [
          { source: 'Direct', visits: 1000, conversion: 15.5 },
          { source: 'Google', visits: 500, conversion: 8.2 },
          { source: 'Facebook', visits: 300, conversion: 12.1 },
          { source: 'Twitter', visits: 200, conversion: 9.8 },
          { source: 'Email', visits: 150, conversion: 22.3 }
        ]
      },
      engagementMetrics: {
        completionRate: Math.random() * 50 + 50,
        averageWatchTime: formatDuration(Math.floor(Math.random() * 3600) + 300),
        skipRate: Math.random() * 30,
        repeatViews: Math.floor(Math.random() * 1000) + 200,
        comments: Math.floor(Math.random() * 100) + 20,
        likes: Math.floor(Math.random() * 500) + 100
      }
    }
  }, [items, timeRange])

  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)
    const secs = Math.floor(seconds % 60)
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`
    } else {
      return `${secs}s`
    }
  }

  const handleExportAnalytics = (format: 'csv' | 'json' | 'pdf') => {
    const data = {
      collection: collection.name,
      generated: new Date().toISOString(),
      timeRange,
      analytics: analyticsData
    }
    
    if (format === 'json') {
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `collection-analytics-${collection.name}-${Date.now()}.json`
      a.click()
      URL.revokeObjectURL(url)
    } else if (format === 'csv') {
      // Simple CSV export for overview stats
      const csv = `Metric,Value\nTotal Items,${analyticsData.overview.totalItems}\nTotal Size,${analyticsData.overview.totalSize}\nAverage Rating,${analyticsData.overview.averageRating.toFixed(2)}\nPlay Count,${analyticsData.overview.playCount}\nDownload Count,${analyticsData.overview.downloadCount}\nShare Count,${analyticsData.overview.shareCount}\nView Count,${analyticsData.overview.viewCount}`
      
      const blob = new Blob([csv], { type: 'text/csv' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `collection-analytics-${collection.name}-${Date.now()}.csv`
      a.click()
      URL.revokeObjectURL(url)
    }
  }

  const StatCard = ({ title, value, icon: Icon, change, changeType }: {
    title: string
    value: string | number
    icon: any
    change?: number
    changeType?: 'increase' | 'decrease'
  }) => (
    <motion.div
      whileHover={{ scale: 1.02 }}
      className="bg-white dark:bg-gray-800 rounded-lg p-6 shadow-sm border border-gray-200 dark:border-gray-700"
    >
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm text-gray-600 dark:text-gray-400 mb-1">{title}</p>
          <p className="text-2xl font-bold text-gray-900 dark:text-white">{value}</p>
          {change !== undefined && (
            <div className="flex items-center mt-2 text-sm">
              {changeType === 'increase' ? (
                <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
              ) : (
                <TrendingDown className="w-4 h-4 text-red-500 mr-1" />
              )}
              <span className={changeType === 'increase' ? 'text-green-500' : 'text-red-500'}>
                {change}%
              </span>
              <span className="text-gray-500 dark:text-gray-400 ml-1">vs last period</span>
            </div>
          )}
        </div>
        <div className="bg-blue-50 dark:bg-blue-900/20 p-3 rounded-lg">
          <Icon className="w-6 h-6 text-blue-600 dark:text-blue-400" />
        </div>
      </div>
    </motion.div>
  )

  const renderOverview = () => (
    <div className="space-y-6">
      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Items"
          value={analyticsData.overview.totalItems.toLocaleString()}
          icon={BarChart3}
          change={12.5}
          changeType="increase"
        />
        <StatCard
          title="Total Plays"
          value={analyticsData.overview.playCount.toLocaleString()}
          icon={Play}
          change={8.3}
          changeType="increase"
        />
        <StatCard
          title="Downloads"
          value={analyticsData.overview.downloadCount.toLocaleString()}
          icon={Download}
          change={-2.1}
          changeType="decrease"
        />
        <StatCard
          title="Shares"
          value={analyticsData.overview.shareCount.toLocaleString()}
          icon={Share2}
          change={15.7}
          changeType="increase"
        />
      </div>

      {/* Media Breakdown */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Media Breakdown</h3>
          <div className="space-y-3">
            {Object.entries(analyticsData.mediaBreakdown).map(([type, count]) => {
              const Icon = MEDIA_ICONS[type as keyof typeof MEDIA_ICONS]
              const total = Object.values(analyticsData.mediaBreakdown).reduce((a, b) => a + b, 0)
              const percentage = total > 0 ? (count / total) * 100 : 0
              
              return (
                <div key={type} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="bg-gray-100 dark:bg-gray-700 p-2 rounded">
                      <Icon className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    </div>
                    <div>
                      <p className="font-medium text-gray-900 dark:text-white capitalize">{type}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">{count} items</p>
                    </div>
                  </div>
                  <div className="text-right">
                    <p className="font-semibold text-gray-900 dark:text-white">{percentage.toFixed(1)}%</p>
                    <div className="w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2 mt-1">
                      <div 
                        className="bg-blue-500 h-2 rounded-full" 
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </Card>

        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Collection Stats</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Total Size</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.overview.totalSize}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Total Duration</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.overview.totalDuration}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Average Rating</span>
              <div className="flex items-center gap-2">
                <Star className="w-4 h-4 text-yellow-500 fill-current" />
                <span className="font-semibold text-gray-900 dark:text-white">
                  {analyticsData.overview.averageRating.toFixed(1)}
                </span>
              </div>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Total Views</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.overview.viewCount.toLocaleString()}
              </span>
            </div>
          </div>
        </Card>
      </div>

      {/* Top Items */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Top Items</h3>
        <div className="space-y-3">
          {analyticsData.topItems.slice(0, 5).map((item, index) => (
            <div key={item.id} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div className="flex items-center gap-3">
                <div className="w-8 h-8 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center text-sm font-semibold text-blue-600 dark:text-blue-400">
                  {index + 1}
                </div>
                <div>
                  <p className="font-medium text-gray-900 dark:text-white truncate max-w-xs">
                    {item.title}
                  </p>
                  <div className="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
                    {(() => {
                      const IconComponent = MEDIA_ICONS[item.type as keyof typeof MEDIA_ICONS] || FileText
                      return <IconComponent className="w-3 h-3" />
                    })()}
                    <span>{item.size}</span>
                  </div>
                </div>
              </div>
              <div className="text-right">
                <p className="font-semibold text-gray-900 dark:text-white">{item.plays}</p>
                <div className="flex items-center gap-1 text-sm">
                  <Star className="w-3 h-3 text-yellow-500 fill-current" />
                  <span className="text-gray-500 dark:text-gray-400">{item.rating.toFixed(1)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )

  const renderActivity = () => (
    <div className="space-y-6">
      {/* Activity Chart Controls */}
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Activity Trends</h3>
        <div className="flex items-center gap-4">
          <Select
            value={timeRange}
            onChange={setTimeRange}
            options={TIME_RANGES}
            className="w-40"
          />
          <Select
            value={chartType}
            onChange={setChartType}
            options={CHART_TYPES}
            className="w-32"
          />
        </div>
      </div>

      {/* Simple Activity Chart */}
      <Card className="p-6">
        <div className="space-y-4">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {['daily', 'weekly', 'monthly'].map((period) => (
              <div key={period} className="space-y-2">
                <h4 className="font-medium text-gray-900 dark:text-white capitalize">{period} Activity</h4>
                <div className="space-y-2">
                  {analyticsData.activity[period as keyof typeof analyticsData.activity].slice(0, 5).map((data, index) => (
                    <div key={index} className="flex justify-between text-sm">
                      <span className="text-gray-600 dark:text-gray-400">
                        {period === 'daily' ? (data as any).date : period === 'weekly' ? (data as any).week : (data as any).month}
                      </span>
                      <span className="font-medium text-gray-900 dark:text-white">
                        {data[chartType as keyof typeof data] || 0}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </Card>

      {/* Engagement Metrics */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Engagement Metrics</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
              {analyticsData.engagementMetrics.completionRate.toFixed(1)}%
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">Completion Rate</p>
          </div>
          <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <p className="text-2xl font-bold text-green-600 dark:text-green-400">
              {analyticsData.engagementMetrics.averageWatchTime}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">Avg Watch Time</p>
          </div>
          <div className="text-center p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
            <p className="text-2xl font-bold text-purple-600 dark:text-purple-400">
              {analyticsData.engagementMetrics.repeatViews.toLocaleString()}
            </p>
            <p className="text-sm text-gray-600 dark:text-gray-400">Repeat Views</p>
          </div>
        </div>
      </Card>
    </div>
  )

  const renderSharing = () => (
    <div className="space-y-6">
      {/* Share Performance */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Share Types</h3>
          <div className="space-y-3">
            {Object.entries(analyticsData.sharingStats.sharedByType).map(([type, count]) => {
              const total = Object.values(analyticsData.sharingStats.sharedByType).reduce((a, b) => a + b, 0)
              const percentage = total > 0 ? (count / total) * 100 : 0
              
              const iconMap = {
                link: Link,
                email: Mail,
                social: MessageCircle,
                embed: Upload
              }
              const Icon = iconMap[type as keyof typeof iconMap] || Share2
              
              return (
                <div key={type} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Icon className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    <span className="capitalize text-gray-900 dark:text-white">{type}</span>
                  </div>
                  <div className="flex items-center gap-3">
                    <div className="w-20 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                      <div 
                        className="bg-green-500 h-2 rounded-full" 
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                    <span className="text-sm font-medium text-gray-900 dark:text-white w-12 text-right">
                      {count}
                    </span>
                  </div>
                </div>
              )
            })}
          </div>
        </Card>

        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Top Referrers</h3>
          <div className="space-y-3">
            {analyticsData.sharingStats.topReferrers.map((referrer, index) => (
              <div key={referrer.source} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center text-xs font-semibold text-blue-600 dark:text-blue-400">
                    {index + 1}
                  </div>
                  <span className="font-medium text-gray-900 dark:text-white">{referrer.source}</span>
                </div>
                <div className="text-right">
                  <p className="font-semibold text-gray-900 dark:text-white">{referrer.visits}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">{referrer.conversion}% conv.</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </div>

      {/* Recent Share Activity */}
      <Card className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Recent Share Activity</h3>
        <div className="space-y-2">
          {analyticsData.sharingStats.sharePerformance.slice(0, 10).map((data, index) => (
            <div key={index} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div>
                <p className="font-medium text-gray-900 dark:text-white">{data.date}</p>
              </div>
              <div className="flex items-center gap-6 text-sm">
                <div className="flex items-center gap-1">
                  <Share2 className="w-3 h-3 text-gray-500" />
                  <span className="text-gray-600 dark:text-gray-400">{data.shares}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Eye className="w-3 h-3 text-gray-500" />
                  <span className="text-gray-600 dark:text-gray-400">{data.clicks}</span>
                </div>
                <div className="flex items-center gap-1">
                  <Download className="w-3 h-3 text-gray-500" />
                  <span className="text-gray-600 dark:text-gray-400">{data.downloads}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )

  const renderContent = () => (
    <div className="space-y-6">
      <Card className="p-6">
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Content Performance</h3>
        <div className="space-y-4">
          {analyticsData.topItems.map((item, index) => (
            <div key={item.id} className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <div className="flex items-center gap-4">
                <div className="w-10 h-10 bg-blue-100 dark:bg-blue-900 rounded-full flex items-center justify-center text-sm font-semibold text-blue-600 dark:text-blue-400">
                  {index + 1}
                </div>
                <div className="flex-1">
                  <p className="font-medium text-gray-900 dark:text-white mb-1">{item.title}</p>
                  <div className="flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
                    <span className="capitalize">{item.type}</span>
                    <span>{item.size}</span>
                    <div className="flex items-center gap-1">
                      <Star className="w-3 h-3 text-yellow-500 fill-current" />
                      <span>{item.rating.toFixed(1)}</span>
                    </div>
                  </div>
                </div>
              </div>
              <div className="text-right">
                <p className="text-xl font-bold text-gray-900 dark:text-white">{item.plays}</p>
                <p className="text-sm text-gray-500 dark:text-gray-400">plays</p>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  )

  const renderEngagement = () => (
    <div className="space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-medium text-gray-900 dark:text-white">Completion Rate</h3>
            <Activity className="w-5 h-5 text-blue-600 dark:text-blue-400" />
          </div>
          <div className="space-y-2">
            <div className="text-3xl font-bold text-blue-600 dark:text-blue-400">
              {analyticsData.engagementMetrics.completionRate.toFixed(1)}%
            </div>
            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
              <div 
                className="bg-blue-500 h-3 rounded-full" 
                style={{ width: `${analyticsData.engagementMetrics.completionRate}%` }}
              />
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-medium text-gray-900 dark:text-white">Skip Rate</h3>
            <TrendingDown className="w-5 h-5 text-red-600 dark:text-red-400" />
          </div>
          <div className="space-y-2">
            <div className="text-3xl font-bold text-red-600 dark:text-red-400">
              {analyticsData.engagementMetrics.skipRate.toFixed(1)}%
            </div>
            <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
              <div 
                className="bg-red-500 h-3 rounded-full" 
                style={{ width: `${analyticsData.engagementMetrics.skipRate}%` }}
              />
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <h3 className="font-medium text-gray-900 dark:text-white">Average Watch Time</h3>
            <Clock className="w-5 h-5 text-green-600 dark:text-green-400" />
          </div>
          <div className="space-y-2">
            <div className="text-3xl font-bold text-green-600 dark:text-green-400">
              {analyticsData.engagementMetrics.averageWatchTime}
            </div>
            <p className="text-sm text-gray-500 dark:text-gray-400">per session</p>
          </div>
        </Card>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">User Interactions</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Repeat Views</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.engagementMetrics.repeatViews.toLocaleString()}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Comments</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.engagementMetrics.comments}
              </span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-gray-600 dark:text-gray-400">Likes</span>
              <span className="font-semibold text-gray-900 dark:text-white">
                {analyticsData.engagementMetrics.likes}
              </span>
            </div>
          </div>
        </Card>

        <Card className="p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Engagement Quality</h3>
          <div className="space-y-4">
            <div className="text-center p-4 bg-green-50 dark:bg-green-900/20 rounded-lg">
              <p className="text-2xl font-bold text-green-600 dark:text-green-400 mb-1">
                {((analyticsData.engagementMetrics.completionRate + (100 - analyticsData.engagementMetrics.skipRate)) / 2).toFixed(1)}%
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">Overall Engagement Score</p>
            </div>
            <div className="text-center p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg">
              <p className="text-2xl font-bold text-blue-600 dark:text-blue-400 mb-1">
                {(analyticsData.engagementMetrics.repeatViews / analyticsData.overview.viewCount * 100).toFixed(1)}%
              </p>
              <p className="text-sm text-gray-600 dark:text-gray-400">Repeat View Rate</p>
            </div>
          </div>
        </Card>
      </div>
    </div>
  )

  const tabs = [
    { id: 'overview', label: 'Overview', icon: BarChart3 },
    { id: 'activity', label: 'Activity', icon: Activity },
    { id: 'content', label: 'Content', icon: FileVideo },
    { id: 'sharing', label: 'Sharing', icon: Share2 },
    { id: 'engagement', label: 'Engagement', icon: Users }
  ]

  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.95 }}
      animate={{ opacity: 1, scale: 1 }}
      exit={{ opacity: 0, scale: 0.95 }}
      className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
      onClick={onClose}
    >
      <motion.div
        initial={{ y: 20 }}
        animate={{ y: 0 }}
        className="bg-white dark:bg-gray-900 rounded-xl shadow-2xl max-w-6xl w-full max-h-[90vh] overflow-hidden"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div>
            <h2 className="text-2xl font-bold text-gray-900 dark:text-white">Collection Analytics</h2>
            <p className="text-gray-600 dark:text-gray-400">{collection.name}</p>
          </div>
          <div className="flex items-center gap-3">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowComparison(!showComparison)}
              className="flex items-center gap-2"
            >
              <RefreshCw className="w-4 h-4" />
              Compare
            </Button>
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleExportAnalytics('csv')}
              >
                <DownloadIcon className="w-4 h-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleExportAnalytics('json')}
              >
                <Upload className="w-4 h-4" />
              </Button>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={onClose}
            >
              <X className="w-4 h-4" />
            </Button>
          </div>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center px-6 overflow-x-auto">
            {tabs.map((tab) => {
              const Icon = tab.icon
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id as any)}
                  className={`flex items-center gap-2 px-4 py-3 border-b-2 transition-colors whitespace-nowrap ${
                    activeTab === tab.id
                      ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                      : 'border-transparent text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white'
                  }`}
                >
                  <Icon className="w-4 h-4" />
                  {tab.label}
                </button>
              )
            })}
          </div>
        </div>

        {/* Content */}
        <div className="overflow-y-auto p-6 max-h-[calc(90vh-140px)]">
          {itemsLoading ? (
            <div className="flex items-center justify-center h-64">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : (
            <>
              {activeTab === 'overview' && renderOverview()}
              {activeTab === 'activity' && renderActivity()}
              {activeTab === 'content' && renderContent()}
              {activeTab === 'sharing' && renderSharing()}
              {activeTab === 'engagement' && renderEngagement()}
            </>
          )}
        </div>
      </motion.div>
    </motion.div>
  )
}