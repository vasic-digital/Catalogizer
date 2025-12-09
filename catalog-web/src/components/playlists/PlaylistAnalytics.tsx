import React from 'react';
import { motion } from 'framer-motion';
import {
  BarChart3,
  Play,
  Clock,
  Users,
  TrendingUp,
  Calendar,
  Music,
  Film,
  Image,
  FileText
} from 'lucide-react';
import { PlaylistAnalytics as PlaylistAnalyticsType } from '../../types/playlists';
import { Card } from '../ui/Card';

interface PlaylistAnalyticsProps {
  analytics: PlaylistAnalyticsType;
  className?: string;
}

const MEDIA_TYPE_ICONS = {
  music: Music,
  video: Film,
  image: Image,
  document: FileText,
};

export const PlaylistAnalytics: React.FC<PlaylistAnalyticsProps> = ({
  analytics,
  className = ''
}) => {
  const formatNumber = (num: number) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatDuration = (seconds: number) => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    if (hours > 0) {
      return `${hours}h ${minutes}m`;
    }
    return `${minutes}m ${secs}s`;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Overview Stats */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
        >
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
                <Play className="w-5 h-5 text-blue-600 dark:text-blue-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Total Plays</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {formatNumber(analytics.total_plays)}
                </p>
              </div>
            </div>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
        >
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg">
                <Users className="w-5 h-5 text-green-600 dark:text-green-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Unique Viewers</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {formatNumber(analytics.unique_viewers)}
                </p>
              </div>
            </div>
          </Card>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
        >
          <Card className="p-4">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-orange-100 dark:bg-orange-900/30 rounded-lg">
                <TrendingUp className="w-5 h-5 text-orange-600 dark:text-orange-400" />
              </div>
              <div>
                <p className="text-sm text-gray-600 dark:text-gray-400">Completion Rate</p>
                <p className="text-2xl font-bold text-gray-900 dark:text-white">
                  {Math.round(analytics.average_completion_rate)}%
                </p>
              </div>
            </div>
          </Card>
        </motion.div>
      </div>

      {/* Popular Items */}
      {analytics.popular_items && analytics.popular_items.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
        >
          <Card className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Popular Items
            </h3>
            <div className="space-y-3">
              {analytics.popular_items.slice(0, 5).map((item, index) => (
                <div key={index} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800 rounded-lg">
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-gray-900 dark:text-white truncate">
                      {item.title}
                    </p>
                    <p className="text-sm text-gray-600 dark:text-gray-400">
                      {formatNumber(item.play_count)} plays
                    </p>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </motion.div>
      )}

      {/* Viewing Stats */}
      {analytics.viewing_stats && analytics.viewing_stats.length > 0 && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.7 }}
        >
          <Card className="p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
              Viewing Trends
            </h3>
            <div className="space-y-3">
              {analytics.viewing_stats.slice(0, 7).map((stat, index) => (
                <div key={index} className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Calendar className="w-4 h-4 text-gray-600 dark:text-gray-400" />
                    <span className="text-gray-900 dark:text-white">
                      {formatDate(stat.date)}
                    </span>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <p className="text-sm text-gray-600 dark:text-gray-400">Plays</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {formatNumber(stat.plays)}
                      </p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm text-gray-600 dark:text-gray-400">Viewers</p>
                      <p className="font-medium text-gray-900 dark:text-white">
                        {formatNumber(stat.viewers)}
                      </p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </Card>
        </motion.div>
      )}
    </div>
  );
};