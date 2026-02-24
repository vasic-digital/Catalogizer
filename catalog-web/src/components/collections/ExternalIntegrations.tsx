import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Globe,
  Cloud,
  Settings,
  Plus,
  Trash2,
  Edit,
  CheckCircle,
  AlertCircle,
  XCircle,
  Clock,
  RefreshCw,
  Download,
  Upload,
  Link,
  Unlink,
  Key,
  Shield,
  Zap,
  Database,
  FolderSync,
  Share2,
  Play,
  Pause,
  Info,
  ExternalLink,
  TestTube,
  Activity
} from 'lucide-react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Badge } from '../ui/Badge';
import { Switch } from '../ui/Switch';
import { toast } from 'react-hot-toast';

// Types
interface ExternalIntegration {
  id: string;
  name: string;
  provider: string;
  type: 'storage' | 'metadata' | 'analytics' | 'automation' | 'sharing';
  status: 'connected' | 'disconnected' | 'error' | 'connecting';
  description: string;
  config: IntegrationConfig;
  syncSettings: SyncSettings;
  statistics: IntegrationStats;
  lastSync?: string;
  createdAt: string;
  enabled: boolean;
}

interface IntegrationConfig {
  apiKey?: string;
  apiSecret?: string;
  endpoint?: string;
  username?: string;
  password?: string;
  webhookUrl?: string;
  customFields?: Record<string, any>;
}

interface SyncSettings {
  enabled: boolean;
  frequency: 'realtime' | 'hourly' | 'daily' | 'weekly' | 'manual';
  direction: 'import' | 'export' | 'bidirectional';
  filters: SyncFilter[];
  mapping?: Record<string, string>;
}

interface SyncFilter {
  field: string;
  operator: string;
  value: any;
}

interface IntegrationStats {
  totalSyncs: number;
  successfulSyncs: number;
  failedSyncs: number;
  lastSyncStatus: 'success' | 'error' | 'pending';
  itemsProcessed: number;
  lastSyncDuration?: number;
  bytesTransferred?: number;
}

const INTEGRATION_TYPES = [
  { value: 'storage', label: 'Storage', description: 'Cloud storage services', icon: 'Database' },
  { value: 'metadata', label: 'Metadata', description: 'Metadata and information services', icon: 'Tag' },
  { value: 'analytics', label: 'Analytics', description: 'Analytics and reporting services', icon: 'BarChart3' },
  { value: 'automation', label: 'Automation', description: 'Automation and workflow services', icon: 'Zap' },
  { value: 'sharing', label: 'Sharing', description: 'File sharing and social services', icon: 'Share' }
];

const INTEGRATION_EXAMPLES = [
  {
    name: 'Google Drive',
    type: 'storage',
    description: 'Access files from Google Drive',
    features: ['File access', 'Metadata sync', 'Real-time updates'],
    setupUrl: 'https://console.developers.google.com/'
  },
  {
    name: 'TMDB',
    type: 'metadata',
    description: 'Fetch movie and TV metadata from The Movie Database',
    features: ['Movie metadata', 'TV metadata', 'Image posters'],
    setupUrl: 'https://www.themoviedb.org/settings/api'
  },
  {
    name: 'Plex',
    type: 'analytics',
    description: 'Sync with Plex media server',
    features: ['Library sync', 'Watch status', 'Analytics'],
    setupUrl: 'https://www.plex.tv/'
  },
  {
    name: 'Discord',
    type: 'sharing',
    description: 'Share collections to Discord channels',
    features: ['Channel notifications', 'File sharing', 'Status updates'],
    setupUrl: 'https://discord.com/developers/applications/'
  }
];

const ExternalIntegrations: React.FC = () => {
  const [integrations, setIntegrations] = useState<ExternalIntegration[]>([]);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingIntegration, setEditingIntegration] = useState<ExternalIntegration | null>(null);
  const [testingIntegration, setTestingIntegration] = useState<string | null>(null);
  const [expandedIntegration, setExpandedIntegration] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'connected' | 'disconnected'>('all');
  const [typeFilter, setTypeFilter] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState('');

  // Load existing integrations
  const loadIntegrations = useCallback(() => {
    const mockIntegrations: ExternalIntegration[] = [
      {
        id: '1',
        name: 'Google Drive Backup',
        provider: 'Google Drive',
        type: 'storage',
        status: 'connected',
        description: 'Backup collection metadata and files to Google Drive',
        config: {
          apiKey: '••••••••••••••••',
          endpoint: 'https://www.googleapis.com/drive/v3',
          webhookUrl: 'https://catalogizer.app/webhooks/google-drive'
        },
        syncSettings: {
          enabled: true,
          frequency: 'daily',
          direction: 'export',
          filters: [
            { field: 'collection_type', operator: 'equals', value: 'movies' }
          ]
        },
        statistics: {
          totalSyncs: 45,
          successfulSyncs: 43,
          failedSyncs: 2,
          lastSyncStatus: 'success',
          itemsProcessed: 1250,
          lastSyncDuration: 180,
          bytesTransferred: 2048576000
        },
        lastSync: '2024-01-21T02:30:00Z',
        createdAt: '2024-01-10T10:00:00Z',
        enabled: true
      },
      {
        id: '2',
        name: 'TMDB Metadata',
        provider: 'The Movie Database',
        type: 'metadata',
        status: 'connected',
        description: 'Fetch movie and TV show metadata from TMDB',
        config: {
          apiKey: '••••••••••••••••',
          endpoint: 'https://api.themoviedb.org/3'
        },
        syncSettings: {
          enabled: true,
          frequency: 'realtime',
          direction: 'import',
          filters: [
            { field: 'media_type', operator: 'in', value: ['movie', 'tv'] }
          ],
          mapping: {
            'title': 'original_title',
            'overview': 'overview',
            'release_date': 'release_date'
          }
        },
        statistics: {
          totalSyncs: 1240,
          successfulSyncs: 1235,
          failedSyncs: 5,
          lastSyncStatus: 'success',
          itemsProcessed: 3420,
          lastSyncDuration: 45
        },
        lastSync: '2024-01-21T14:25:00Z',
        createdAt: '2024-01-05T09:00:00Z',
        enabled: true
      },
      {
        id: '3',
        name: 'Plex Media Server',
        provider: 'Plex',
        type: 'sharing',
        status: 'disconnected',
        description: 'Share collections with Plex Media Server',
        config: {
          endpoint: 'http://192.168.1.100:32400',
          username: 'admin',
          password: '••••••••••••••••'
        },
        syncSettings: {
          enabled: false,
          frequency: 'manual',
          direction: 'bidirectional',
          filters: []
        },
        statistics: {
          totalSyncs: 0,
          successfulSyncs: 0,
          failedSyncs: 0,
          lastSyncStatus: 'pending',
          itemsProcessed: 0
        },
        createdAt: '2024-01-18T16:45:00Z',
        enabled: false
      },
      {
        id: '4',
        name: 'Discord Notifications',
        provider: 'Discord',
        type: 'automation',
        status: 'connected',
        description: 'Send notifications to Discord channels',
        config: {
          webhookUrl: 'https://discord.com/api/webhooks/••••••••••••••••'
        },
        syncSettings: {
          enabled: true,
          frequency: 'realtime',
          direction: 'export',
          filters: [
            { field: 'event_type', operator: 'in', value: ['new_movie', 'collection_update'] }
          ]
        },
        statistics: {
          totalSyncs: 67,
          successfulSyncs: 67,
          failedSyncs: 0,
          lastSyncStatus: 'success',
          itemsProcessed: 67
        },
        lastSync: '2024-01-21T15:30:00Z',
        createdAt: '2024-01-12T14:20:00Z',
        enabled: true
      }
    ];
    setIntegrations(mockIntegrations);
  }, []);

  useEffect(() => {
    loadIntegrations();
  }, [loadIntegrations]);

  // Filter integrations
  const filteredIntegrations = React.useMemo(() => {
    let filtered = [...integrations];

    // Apply connection status filter
    if (filter === 'connected') {
      filtered = filtered.filter(integration => integration.status === 'connected');
    } else if (filter === 'disconnected') {
      filtered = filtered.filter(integration => integration.status !== 'connected');
    }

    // Apply type filter
    if (typeFilter !== 'all') {
      filtered = filtered.filter(integration => integration.type === typeFilter);
    }

    // Apply search
    if (searchQuery) {
      filtered = filtered.filter(integration =>
        integration.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        integration.provider.toLowerCase().includes(searchQuery.toLowerCase()) ||
        integration.description.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    return filtered;
  }, [integrations, filter, typeFilter, searchQuery]);

  // Toggle integration enabled/disabled
  const toggleIntegrationStatus = async (integrationId: string) => {
    setIntegrations(prev => prev.map(integration => 
      integration.id === integrationId 
        ? { ...integration, enabled: !integration.enabled }
        : integration
    ));
    toast.success('Integration status updated');
  };

  // Test integration connection
  const testIntegration = async (integrationId: string) => {
    setTestingIntegration(integrationId);
    
    // Simulate connection test
    setTimeout(() => {
      setTestingIntegration(null);
      setIntegrations(prev => prev.map(integration => 
        integration.id === integrationId 
          ? { 
              ...integration, 
              status: Math.random() > 0.2 ? 'connected' : 'error',
              lastSync: new Date().toISOString()
            }
          : integration
      ));
      toast.success('Connection test completed');
    }, 2000);
  };

  // Manual sync
  const syncIntegration = async (integrationId: string) => {
    const integration = integrations.find(i => i.id === integrationId);
    if (!integration || integration.status !== 'connected') return;

    // Update status to show syncing
    setIntegrations(prev => prev.map(i => 
      i.id === integrationId 
        ? { 
            ...i, 
            lastSync: new Date().toISOString(),
            statistics: {
              ...i.statistics,
              lastSyncStatus: 'pending'
            }
          }
        : i
    ));

    // Simulate sync
    setTimeout(() => {
      setIntegrations(prev => prev.map(i => 
        i.id === integrationId 
          ? { 
              ...i,
              statistics: {
                ...i.statistics,
                totalSyncs: i.statistics.totalSyncs + 1,
                successfulSyncs: i.statistics.successfulSyncs + 1,
                lastSyncStatus: 'success',
                itemsProcessed: i.statistics.itemsProcessed + Math.floor(Math.random() * 50)
              }
            }
          : i
      ));
      toast.success('Sync completed successfully');
    }, 3000);
  };

  // Delete integration
  const deleteIntegration = async (integrationId: string) => {
    setIntegrations(prev => prev.filter(integration => integration.id !== integrationId));
    toast.success('Integration deleted successfully');
  };

  // Get status icon and color
  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'connected':
        return { icon: CheckCircle, color: 'text-green-500', bg: 'bg-green-50' };
      case 'disconnected':
        return { icon: XCircle, color: 'text-red-500', bg: 'bg-red-50' };
      case 'error':
        return { icon: AlertCircle, color: 'text-red-500', bg: 'bg-red-50' };
      case 'connecting':
        return { icon: RefreshCw, color: 'text-blue-500', bg: 'bg-blue-50' };
      default:
        return { icon: Clock, color: 'text-gray-500', bg: 'bg-gray-50' };
    }
  };

  // Get type icon
  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'storage':
        return Cloud;
      case 'metadata':
        return Database;
      case 'analytics':
        return Activity;
      case 'automation':
        return Zap;
      case 'sharing':
        return Share2;
      default:
        return Globe;
    }
  };

  // Format bytes
  const formatBytes = (bytes?: number) => {
    if (!bytes) return 'N/A';
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${sizes[i]}`;
  };

  // Format duration
  const formatDuration = (seconds?: number) => {
    if (!seconds) return 'N/A';
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">External Integrations</h3>
          <p className="text-sm text-gray-500 mt-1">
            Connect with external services to extend functionality
          </p>
        </div>
        <Button
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Add Integration
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Integrations</p>
              <p className="text-2xl font-bold text-gray-900">{integrations.length}</p>
            </div>
            <Globe className="w-8 h-8 text-blue-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Connected</p>
              <p className="text-2xl font-bold text-green-600">
                {integrations.filter(i => i.status === 'connected').length}
              </p>
            </div>
            <Link className="w-8 h-8 text-green-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Active Syncs</p>
              <p className="text-2xl font-bold text-blue-600">
                {integrations.filter(i => i.syncSettings.enabled).length}
              </p>
            </div>
            <FolderSync className="w-8 h-8 text-blue-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Success Rate</p>
              <p className="text-2xl font-bold text-gray-900">
                {integrations.length > 0 && integrations.some(i => i.statistics.totalSyncs > 0)
                  ? Math.round(
                      (integrations.reduce((sum, i) => sum + i.statistics.successfulSyncs, 0) / 
                       integrations.reduce((sum, i) => sum + i.statistics.totalSyncs, 0)) * 100
                    )
                  : 0}%
            </p>
            </div>
            <CheckCircle className="w-8 h-8 text-green-500" />
          </div>
        </div>
      </div>

      {/* Filters and Search */}
      <div className="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between bg-white rounded-lg border p-4">
        <div className="flex flex-wrap gap-2">
          {(['all', 'connected', 'disconnected'] as const).map((filterOption) => (
            <Button
              key={filterOption}
              variant={filter === filterOption ? 'default' : 'outline'}
              size="sm"
              onClick={() => setFilter(filterOption)}
            >
              {filterOption.charAt(0).toUpperCase() + filterOption.slice(1)}
            </Button>
          ))}
          
          <select
            value={typeFilter}
            onChange={(e) => setTypeFilter(e.target.value)}
            className="px-3 py-1 text-sm border rounded-lg bg-white"
          >
            <option value="all">All Types</option>
            <option value="storage">Storage</option>
            <option value="metadata">Metadata</option>
            <option value="analytics">Analytics</option>
            <option value="automation">Automation</option>
            <option value="sharing">Sharing</option>
          </select>
        </div>
        
        <Input
          placeholder="Search integrations..."
          value={searchQuery}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
          className="w-full sm:w-64"
        />
      </div>

      {/* Integrations List */}
      <div className="space-y-4">
        <AnimatePresence>
          {filteredIntegrations.map((integration) => {
            const StatusIcon = getStatusInfo(integration.status).icon;
            const TypeIcon = getTypeIcon(integration.type);
            const isExpanded = expandedIntegration === integration.id;
            
            return (
              <motion.div
                key={integration.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="bg-white rounded-lg border hover:shadow-md transition-shadow"
              >
                {/* Integration Header */}
                <div className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h4 className="font-semibold text-gray-900">{integration.name}</h4>
                        <Badge variant="outline">{integration.provider}</Badge>
                        <Badge variant="secondary">{integration.type}</Badge>
                        <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs ${
                          getStatusInfo(integration.status).bg
                        } ${getStatusInfo(integration.status).color}`}>
                          <StatusIcon className="w-3 h-3" />
                          {integration.status}
                        </div>
                      </div>
                      
                      <p className="text-sm text-gray-600 mb-3">{integration.description}</p>
                      
                      <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500">
                        <div className="flex items-center gap-1">
                          <TypeIcon className="w-3 h-3" />
                          {integration.type}
                        </div>
                        
                        <div className="flex items-center gap-1">
                          <FolderSync className="w-3 h-3" />
                          {integration.syncSettings.enabled ? `${integration.syncSettings.frequency} sync` : 'Sync disabled'}
                        </div>
                        
                        {integration.lastSync && (
                          <div className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            Last sync: {new Date(integration.lastSync).toLocaleDateString()}
                          </div>
                        )}
                        
                        <div className="flex items-center gap-1">
                          <Activity className="w-3 h-3" />
                          {integration.statistics.itemsProcessed} items processed
                        </div>
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-2 ml-4">
                      <Switch
                        checked={integration.enabled}
                        onCheckedChange={() => toggleIntegrationStatus(integration.id)}
                      />
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setExpandedIntegration(isExpanded ? null : integration.id)}
                      >
                        <Info className="w-4 h-4" />
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => testIntegration(integration.id)}
                        disabled={testingIntegration === integration.id}
                      >
                        {testingIntegration === integration.id ? (
                          <RefreshCw className="w-4 h-4 animate-spin" />
                        ) : (
                          <TestTube className="w-4 h-4" />
                        )}
                      </Button>
                      
                      {integration.status === 'connected' && (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={() => syncIntegration(integration.id)}
                        >
                          <RefreshCw className="w-4 h-4" />
                        </Button>
                      )}
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditingIntegration(integration)}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteIntegration(integration.id)}
                      >
                        <Trash2 className="w-4 h-4" />
                      </Button>
                    </div>
                  </div>
                </div>
                
                {/* Expanded Content */}
                <AnimatePresence>
                  {isExpanded && (
                    <motion.div
                      initial={{ height: 0 }}
                      animate={{ height: 'auto' }}
                      exit={{ height: 0 }}
                      transition={{ duration: 0.2 }}
                      className="border-t overflow-hidden"
                    >
                      <div className="p-4 bg-gray-50 space-y-4">
                        {/* Configuration */}
                        <div>
                          <h5 className="font-medium text-sm text-gray-700 mb-2">Configuration</h5>
                          <div className="bg-white rounded p-3 text-sm space-y-2">
                            {integration.config.endpoint && (
                              <div className="flex items-center gap-2">
                                <ExternalLink className="w-3 h-3 text-gray-400" />
                                <span className="text-gray-600">Endpoint: </span>
                                <code className="bg-gray-100 px-2 py-1 rounded text-xs">
                                  {integration.config.endpoint}
                                </code>
                              </div>
                            )}
                            {integration.config.webhookUrl && (
                              <div className="flex items-center gap-2">
                                <Link className="w-3 h-3 text-gray-400" />
                                <span className="text-gray-600">Webhook: </span>
                                <code className="bg-gray-100 px-2 py-1 rounded text-xs truncate max-w-xs">
                                  {integration.config.webhookUrl}
                                </code>
                              </div>
                            )}
                            {integration.config.apiKey && (
                              <div className="flex items-center gap-2">
                                <Key className="w-3 h-3 text-gray-400" />
                                <span className="text-gray-600">API Key: </span>
                                <code className="bg-gray-100 px-2 py-1 rounded text-xs">
                                  {integration.config.apiKey}
                                </code>
                              </div>
                            )}
                          </div>
                        </div>
                        
                        {/* Sync Settings */}
                        <div>
                          <h5 className="font-medium text-sm text-gray-700 mb-2">Sync Settings</h5>
                          <div className="bg-white rounded p-3 text-sm">
                            <div className="grid grid-cols-2 gap-4 text-xs">
                              <div>
                                <span className="text-gray-500">Enabled: </span>
                                <span className={integration.syncSettings.enabled ? 'text-green-600' : 'text-gray-600'}>
                                  {integration.syncSettings.enabled ? 'Yes' : 'No'}
                                </span>
                              </div>
                              <div>
                                <span className="text-gray-500">Frequency: </span>
                                <span className="text-gray-600">{integration.syncSettings.frequency}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Direction: </span>
                                <span className="text-gray-600">{integration.syncSettings.direction}</span>
                              </div>
                              <div>
                                <span className="text-gray-500">Filters: </span>
                                <span className="text-gray-600">{integration.syncSettings.filters.length}</span>
                              </div>
                            </div>
                          </div>
                        </div>
                        
                        {/* Statistics */}
                        <div>
                          <h5 className="font-medium text-sm text-gray-700 mb-2">Statistics</h5>
                          <div className="bg-white rounded p-3">
                            <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                              <div className="text-center">
                                <p className="text-2xl font-bold text-gray-900">{integration.statistics.totalSyncs}</p>
                                <p className="text-xs text-gray-500">Total Syncs</p>
                              </div>
                              <div className="text-center">
                                <p className="text-2xl font-bold text-green-600">{integration.statistics.successfulSyncs}</p>
                                <p className="text-xs text-gray-500">Successful</p>
                              </div>
                              <div className="text-center">
                                <p className="text-2xl font-bold text-red-600">{integration.statistics.failedSyncs}</p>
                                <p className="text-xs text-gray-500">Failed</p>
                              </div>
                              <div className="text-center">
                                <p className="text-2xl font-bold text-blue-600">{integration.statistics.itemsProcessed}</p>
                                <p className="text-xs text-gray-500">Items</p>
                              </div>
                            </div>
                            {integration.statistics.lastSyncDuration && (
                              <div className="mt-3 pt-3 border-t text-xs text-gray-600">
                                Last sync duration: {formatDuration(integration.statistics.lastSyncDuration)}
                                {integration.statistics.bytesTransferred && (
                                  <span className="ml-3">
                                    Data transferred: {formatBytes(integration.statistics.bytesTransferred)}
                                  </span>
                                )}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </motion.div>
            );
          })}
        </AnimatePresence>
        
        {filteredIntegrations.length === 0 && (
          <div className="text-center py-12 bg-white rounded-lg border">
            <Globe className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No integrations found</h3>
            <p className="text-gray-500 mb-4">
              Connect with external services to extend your collection management capabilities
            </p>
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Add Your First Integration
            </Button>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {isCreateModalOpen && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto p-6">
            <h3 className="text-xl font-bold mb-4">Add External Integration</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Select Service</label>
                <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
                  {INTEGRATION_EXAMPLES.map(example => (
                    <button
                      key={example.name}
                      onClick={() => {
                        const newIntegration: ExternalIntegration = {
                          id: Date.now().toString(),
                          name: example.name,
                          provider: example.name.toLowerCase().replace(/\s+/g, '_'),
                          type: example.type as ExternalIntegration['type'],
                          status: 'disconnected',
                          description: example.description,
                          config: {},
                          syncSettings: {
                            enabled: false,
                            frequency: 'daily',
                            direction: 'import',
                            filters: [],
                          },
                          statistics: {
                            totalSyncs: 0,
                            successfulSyncs: 0,
                            failedSyncs: 0,
                            lastSyncStatus: 'pending',
                            itemsProcessed: 0,
                          },
                          createdAt: new Date().toISOString(),
                          enabled: true,
                        }
                        setIntegrations(prev => [...prev, newIntegration])
                        setIsCreateModalOpen(false)
                        setEditingIntegration(newIntegration)
                        toast.success(`${example.name} integration added. Configure credentials to connect.`)
                      }}
                      className="p-4 border rounded-lg hover:border-blue-400 hover:bg-blue-50 dark:hover:bg-blue-900/20 text-left transition-colors"
                    >
                      <div className="font-medium text-sm">{example.name}</div>
                      <div className="text-xs text-gray-500 mt-1">{example.description}</div>
                      <div className="flex flex-wrap gap-1 mt-2">
                        {example.features.slice(0, 2).map(f => (
                          <span key={f} className="text-xs bg-gray-100 dark:bg-gray-700 px-1.5 py-0.5 rounded">{f}</span>
                        ))}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-6 pt-4 border-t">
              <Button variant="outline" onClick={() => setIsCreateModalOpen(false)}>Cancel</Button>
            </div>
          </div>
        </div>
      )}

      {/* Edit Modal */}
      {editingIntegration && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
          <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto p-6">
            <h3 className="text-xl font-bold mb-4">Edit Integration</h3>
            <p className="text-sm text-gray-500 mb-4">Configure {editingIntegration.name}</p>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Key</label>
                <Input
                  type="password"
                  placeholder="Enter API key"
                  defaultValue={editingIntegration.config.apiKey || ''}
                  onChange={(e) => {
                    setEditingIntegration(prev => prev ? { ...prev, config: { ...prev.config, apiKey: e.target.value } } : null)
                  }}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">API Endpoint (optional)</label>
                <Input
                  placeholder="https://api.example.com"
                  defaultValue={editingIntegration.config.endpoint || ''}
                  onChange={(e) => {
                    setEditingIntegration(prev => prev ? { ...prev, config: { ...prev.config, endpoint: e.target.value } } : null)
                  }}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Webhook URL (optional)</label>
                <Input
                  placeholder="https://hooks.example.com/callback"
                  defaultValue={editingIntegration.config.webhookUrl || ''}
                  onChange={(e) => {
                    setEditingIntegration(prev => prev ? { ...prev, config: { ...prev.config, webhookUrl: e.target.value } } : null)
                  }}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Sync Frequency</label>
                <select
                  className="w-full border rounded-lg p-2 text-sm"
                  defaultValue={editingIntegration.syncSettings.frequency}
                  onChange={(e) => {
                    setEditingIntegration(prev => prev ? { ...prev, syncSettings: { ...prev.syncSettings, frequency: e.target.value as SyncSettings['frequency'] } } : null)
                  }}
                >
                  <option value="realtime">Real-time</option>
                  <option value="hourly">Hourly</option>
                  <option value="daily">Daily</option>
                  <option value="weekly">Weekly</option>
                  <option value="manual">Manual only</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Sync Direction</label>
                <select
                  className="w-full border rounded-lg p-2 text-sm"
                  defaultValue={editingIntegration.syncSettings.direction}
                  onChange={(e) => {
                    setEditingIntegration(prev => prev ? { ...prev, syncSettings: { ...prev.syncSettings, direction: e.target.value as SyncSettings['direction'] } } : null)
                  }}
                >
                  <option value="import">Import only</option>
                  <option value="export">Export only</option>
                  <option value="bidirectional">Bidirectional</option>
                </select>
              </div>
            </div>

            <div className="flex justify-end gap-2 mt-6 pt-4 border-t">
              <Button variant="outline" onClick={() => setEditingIntegration(null)}>Cancel</Button>
              <Button onClick={() => {
                setIntegrations(prev => prev.map(i => i.id === editingIntegration.id ? editingIntegration : i))
                setEditingIntegration(null)
                toast.success('Integration updated')
              }}>Save Changes</Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ExternalIntegrations;