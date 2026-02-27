import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  Bot,
  Play,
  Plus,
  Trash2,
  Edit,
  Clock,
  Zap,
  AlertCircle,
  CheckCircle,
  Calendar,
  Settings,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  TestTube,
  Activity
} from 'lucide-react';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Badge } from '../ui/Badge';
import { Switch } from '../ui/Switch';
import { toast } from 'react-hot-toast';

// Types
interface AutomationRule {
  id: string;
  name: string;
  description: string;
  enabled: boolean;
  trigger: {
    type: 'schedule' | 'event' | 'manual';
    schedule?: string; // cron expression
    eventType?: string;
    conditions?: RuleCondition[];
  };
  actions: AutomationAction[];
  createdAt: string;
  lastRun?: string;
  nextRun?: string;
  runCount: number;
  successCount: number;
  errorCount: number;
  status: 'idle' | 'running' | 'completed' | 'error';
}

interface AutomationAction {
  id: string;
  type: 'add_to_collection' | 'remove_from_collection' | 'tag_files' | 'move_files' | 'copy_files' | 'notify' | 'run_script';
  parameters: Record<string, unknown>;
  order: number;
}

interface RuleCondition {
  field: string;
  operator: string;
  value: unknown;
  logic?: 'AND' | 'OR';
}

const TRIGGER_TYPES = [
  { value: 'schedule', label: 'Schedule', description: 'Run on a time-based schedule' },
  { value: 'event', label: 'Event', description: 'Run when specific events occur' },
  { value: 'manual', label: 'Manual', description: 'Run manually' }
];

const ACTION_TYPES = [
  { value: 'add_to_collection', label: 'Add to Collection', description: 'Add files to a collection' },
  { value: 'remove_from_collection', label: 'Remove from Collection', description: 'Remove files from a collection' },
  { value: 'tag_files', label: 'Tag Files', description: 'Apply tags to files' },
  { value: 'move_files', label: 'Move Files', description: 'Move files to another location' },
  { value: 'copy_files', label: 'Copy Files', description: 'Copy files to another location' },
  { value: 'notify', label: 'Send Notification', description: 'Send a notification' },
  { value: 'run_script', label: 'Run Script', description: 'Execute a custom script' }
];

const CollectionAutomation: React.FC = () => {
  const [rules, setRules] = useState<AutomationRule[]>([]);
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [editingRule, setEditingRule] = useState<AutomationRule | null>(null);
  const [testingRule, setTestingRule] = useState<string | null>(null);
  const [expandedRule, setExpandedRule] = useState<string | null>(null);
  const [filter, setFilter] = useState<'all' | 'enabled' | 'disabled'>('all');
  const [sortBy, setSortBy] = useState<'name' | 'created' | 'lastRun' | 'status'>('created');
  const [searchQuery, setSearchQuery] = useState('');

  // Load existing automation rules
  const loadRules = useCallback(() => {
    const mockRules: AutomationRule[] = [
      {
        id: '1',
        name: 'Auto-Tag New Movies',
        description: 'Automatically tag new movie files with genre and year',
        enabled: true,
        trigger: {
          type: 'event',
          eventType: 'file_added',
          conditions: [
            { field: 'file_type', operator: 'equals', value: 'video' },
            { field: 'file_path', operator: 'contains', value: '/movies/', logic: 'AND' }
          ]
        },
        actions: [
          {
            id: '1',
            type: 'tag_files',
            parameters: {
              tags: ['movie'],
              autoDetect: true,
              overwriteExisting: false
            },
            order: 1
          },
          {
            id: '2',
            type: 'notify',
            parameters: {
              message: 'New movie tagged: {filename}',
              channels: ['email', 'webhook']
            },
            order: 2
          }
        ],
        createdAt: '2024-01-15T10:30:00Z',
        lastRun: '2024-01-20T14:25:00Z',
        nextRun: '2024-01-21T14:25:00Z',
        runCount: 15,
        successCount: 14,
        errorCount: 1,
        status: 'idle'
      },
      {
        id: '2',
        name: 'Weekly Collection Cleanup',
        description: 'Remove duplicate and broken files from collections',
        enabled: true,
        trigger: {
          type: 'schedule',
          schedule: '0 2 * * 0' // Sunday at 2 AM
        },
        actions: [
          {
            id: '1',
            type: 'remove_from_collection',
            parameters: {
              condition: 'duplicate_or_broken',
              backup: true,
              notify: true
            },
            order: 1
          }
        ],
        createdAt: '2024-01-10T09:00:00Z',
        lastRun: '2024-01-21T02:00:00Z',
        nextRun: '2024-01-28T02:00:00Z',
        runCount: 4,
        successCount: 4,
        errorCount: 0,
        status: 'completed'
      },
      {
        id: '3',
        name: 'Sync to External Drive',
        description: 'Sync new photos to external backup drive',
        enabled: false,
        trigger: {
          type: 'event',
          eventType: 'file_added',
          conditions: [
            { field: 'file_type', operator: 'in', value: ['image', 'raw'] },
            { field: 'file_size', operator: 'greater_than', value: 1024000, logic: 'AND' }
          ]
        },
        actions: [
          {
            id: '1',
            type: 'copy_files',
            parameters: {
              destination: '/backup/photos/',
              preserveStructure: true,
              verify: true
            },
            order: 1
          }
        ],
        createdAt: '2024-01-18T16:45:00Z',
        runCount: 0,
        successCount: 0,
        errorCount: 0,
        status: 'idle'
      }
    ];
    setRules(mockRules);
  }, []);

  useEffect(() => {
    loadRules();
  }, [loadRules]);

  // Filter and sort rules
  const filteredAndSortedRules = React.useMemo(() => {
    let filtered = [...rules];

    // Apply filter
    if (filter === 'enabled') {
      filtered = filtered.filter(rule => rule.enabled);
    } else if (filter === 'disabled') {
      filtered = filtered.filter(rule => !rule.enabled);
    }

    // Apply search
    if (searchQuery) {
      filtered = filtered.filter(rule =>
        rule.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        rule.description.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    // Apply sorting
    filtered.sort((a, b) => {
      switch (sortBy) {
        case 'name':
          return a.name.localeCompare(b.name);
        case 'created':
          return new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime();
        case 'lastRun':
          if (!a.lastRun) return 1;
          if (!b.lastRun) return -1;
          return new Date(b.lastRun).getTime() - new Date(a.lastRun).getTime();
        case 'status':
          return a.status.localeCompare(b.status);
        default:
          return 0;
      }
    });

    return filtered;
  }, [rules, filter, sortBy, searchQuery]);

  // Toggle rule enabled/disabled
  const toggleRuleStatus = async (ruleId: string) => {
    setRules(prev => prev.map(rule => 
      rule.id === ruleId ? { ...rule, enabled: !rule.enabled } : rule
    ));
    toast.success('Rule status updated');
  };

  // Test rule execution
  const testRule = async (ruleId: string) => {
    setTestingRule(ruleId);
    
    // Simulate rule testing
    setTimeout(() => {
      setTestingRule(null);
      toast.success('Rule test completed successfully');
    }, 3000);
  };

  // Run rule manually
  const runRule = async (ruleId: string) => {
    setRules(prev => prev.map(rule => 
      rule.id === ruleId ? { ...rule, status: 'running' } : rule
    ));

    // Simulate rule execution
    setTimeout(() => {
      setRules(prev => prev.map(rule => 
        rule.id === ruleId ? { 
          ...rule, 
          status: 'completed',
          lastRun: new Date().toISOString(),
          runCount: rule.runCount + 1,
          successCount: rule.successCount + 1
        } : rule
      ));
      toast.success('Rule executed successfully');
    }, 2000);
  };

  // Delete rule
  const deleteRule = async (ruleId: string) => {
    setRules(prev => prev.filter(rule => rule.id !== ruleId));
    toast.success('Rule deleted successfully');
  };

  // Get status icon and color
  const getStatusInfo = (status: string) => {
    switch (status) {
      case 'running':
        return { icon: RefreshCw, color: 'text-blue-500', bg: 'bg-blue-50' };
      case 'completed':
        return { icon: CheckCircle, color: 'text-green-500', bg: 'bg-green-50' };
      case 'error':
        return { icon: AlertCircle, color: 'text-red-500', bg: 'bg-red-50' };
      default:
        return { icon: Clock, color: 'text-gray-500', bg: 'bg-gray-50' };
    }
  };

  // Get trigger type icon
  const getTriggerIcon = (type: string) => {
    switch (type) {
      case 'schedule':
        return Calendar;
      case 'event':
        return Zap;
      case 'manual':
        return Play;
      default:
        return Settings;
    }
  };

  // Format next run time
  const formatNextRun = (rule: AutomationRule) => {
    if (rule.trigger.type === 'manual') {
      return 'Manual trigger';
    }
    if (!rule.nextRun) {
      return 'Not scheduled';
    }
    const nextRun = new Date(rule.nextRun);
    const now = new Date();
    const diff = nextRun.getTime() - now.getTime();
    
    if (diff < 0) return 'Overdue';
    if (diff < 3600000) return `${Math.floor(diff / 60000)} minutes`;
    if (diff < 86400000) return `${Math.floor(diff / 3600000)} hours`;
    return `${Math.floor(diff / 86400000)} days`;
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="text-lg font-semibold text-gray-900">Automation Rules</h3>
          <p className="text-sm text-gray-500 mt-1">
            Create workflows to automatically manage your collections
          </p>
        </div>
        <Button
          onClick={() => setIsCreateModalOpen(true)}
          className="flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          Create Rule
        </Button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Rules</p>
              <p className="text-2xl font-bold text-gray-900">{rules.length}</p>
            </div>
            <Bot className="w-8 h-8 text-blue-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Active</p>
              <p className="text-2xl font-bold text-green-600">
                {rules.filter(r => r.enabled).length}
              </p>
            </div>
            <Play className="w-8 h-8 text-green-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Total Runs</p>
              <p className="text-2xl font-bold text-blue-600">
                {rules.reduce((sum, r) => sum + r.runCount, 0)}
              </p>
            </div>
            <Activity className="w-8 h-8 text-blue-500" />
          </div>
        </div>
        
        <div className="bg-white rounded-lg border p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Success Rate</p>
              <p className="text-2xl font-bold text-gray-900">
                {rules.length > 0 
                  ? Math.round((rules.reduce((sum, r) => sum + r.successCount, 0) / 
                      rules.reduce((sum, r) => sum + r.runCount, 0)) * 100)
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
          {(['all', 'enabled', 'disabled'] as const).map((filterOption) => (
            <Button
              key={filterOption}
              variant={filter === filterOption ? 'default' : 'outline'}
              size="sm"
              onClick={() => setFilter(filterOption)}
            >
              {filterOption.charAt(0).toUpperCase() + filterOption.slice(1)}
            </Button>
          ))}
        </div>
        
        <div className="flex gap-2 items-center w-full sm:w-auto">
          <Input
            placeholder="Search rules..."
            value={searchQuery}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
            className="w-full sm:w-64"
          />
          
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as any)}
            className="px-3 py-2 border rounded-lg bg-white"
          >
            <option value="created">Created</option>
            <option value="name">Name</option>
            <option value="lastRun">Last Run</option>
            <option value="status">Status</option>
          </select>
        </div>
      </div>

      {/* Rules List */}
      <div className="space-y-4">
        <AnimatePresence>
          {filteredAndSortedRules.map((rule) => {
            const StatusIcon = getStatusInfo(rule.status).icon;
            const TriggerIcon = getTriggerIcon(rule.trigger.type);
            const isExpanded = expandedRule === rule.id;
            
            return (
              <motion.div
                key={rule.id}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                className="bg-white rounded-lg border hover:shadow-md transition-shadow"
              >
                {/* Rule Header */}
                <div className="p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <h4 className="font-semibold text-gray-900">{rule.name}</h4>
                        <Badge variant={rule.enabled ? 'default' : 'secondary'}>
                          {rule.enabled ? 'Enabled' : 'Disabled'}
                        </Badge>
                        <div className={`flex items-center gap-1 px-2 py-1 rounded-full text-xs ${
                          getStatusInfo(rule.status).bg
                        } ${getStatusInfo(rule.status).color}`}>
                          <StatusIcon className="w-3 h-3" />
                          {rule.status}
                        </div>
                      </div>
                      
                      <p className="text-sm text-gray-600 mb-3">{rule.description}</p>
                      
                      <div className="flex flex-wrap items-center gap-4 text-xs text-gray-500">
                        <div className="flex items-center gap-1">
                          <TriggerIcon className="w-3 h-3" />
                          {rule.trigger.type === 'schedule' ? `Scheduled: ${rule.trigger.schedule}` :
                           rule.trigger.type === 'event' ? `Event: ${rule.trigger.eventType}` : 'Manual'}
                        </div>
                        
                        <div className="flex items-center gap-1">
                          <Clock className="w-3 h-3" />
                          Next: {formatNextRun(rule)}
                        </div>
                        
                        <div className="flex items-center gap-1">
                          <Activity className="w-3 h-3" />
                          Runs: {rule.runCount} ({rule.successCount} success)
                        </div>
                        
                        {rule.lastRun && (
                          <div className="flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            Last: {new Date(rule.lastRun).toLocaleDateString()}
                          </div>
                        )}
                      </div>
                    </div>
                    
                    <div className="flex items-center gap-2 ml-4">
                      <Switch
                        checked={rule.enabled}
                        onCheckedChange={() => toggleRuleStatus(rule.id)}
                      />
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setExpandedRule(isExpanded ? null : rule.id)}
                      >
                        {isExpanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => testRule(rule.id)}
                        disabled={testingRule === rule.id}
                      >
                        {testingRule === rule.id ? (
                          <RefreshCw className="w-4 h-4 animate-spin" />
                        ) : (
                          <TestTube className="w-4 h-4" />
                        )}
                      </Button>
                      
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => runRule(rule.id)}
                        disabled={rule.status === 'running'}
                      >
                        {rule.status === 'running' ? (
                          <RefreshCw className="w-4 h-4 animate-spin" />
                        ) : (
                          <Play className="w-4 h-4" />
                        )}
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setEditingRule(rule)}
                      >
                        <Edit className="w-4 h-4" />
                      </Button>
                      
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => deleteRule(rule.id)}
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
                        {/* Trigger Details */}
                        <div>
                          <h5 className="font-medium text-sm text-gray-700 mb-2">Trigger</h5>
                          <div className="bg-white rounded p-3 text-sm">
                            <div className="flex items-center gap-2 mb-2">
                              <TriggerIcon className="w-4 h-4 text-blue-500" />
                              <span className="font-medium capitalize">{rule.trigger.type}</span>
                            </div>
                            {rule.trigger.schedule && (
                              <p className="text-gray-600">Schedule: {rule.trigger.schedule}</p>
                            )}
                            {rule.trigger.eventType && (
                              <p className="text-gray-600">Event: {rule.trigger.eventType}</p>
                            )}
                            {rule.trigger.conditions && rule.trigger.conditions.length > 0 && (
                              <div className="mt-2">
                                <p className="text-xs text-gray-500 mb-1">Conditions:</p>
                                {rule.trigger.conditions.map((condition, index) => (
                                  <div key={index} className="text-xs text-gray-600 ml-4">
                                    {condition.field} {condition.operator} {String(condition.value)}
                                    {condition.logic ? <span className="ml-2 text-blue-600">{condition.logic}</span> : null}
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        </div>
                        
                        {/* Actions */}
                        <div>
                          <h5 className="font-medium text-sm text-gray-700 mb-2">Actions</h5>
                          <div className="space-y-2">
                            {rule.actions.map((action, index) => (
                              <div key={action.id} className="bg-white rounded p-3 text-sm">
                                <div className="flex items-center gap-2 mb-2">
                                  <span className="font-medium capitalize">
                                    {action.type.replace('_', ' ')}
                                  </span>
                                  <Badge variant="outline" className="text-xs">
                                    Step {action.order}
                                  </Badge>
                                </div>
                                <div className="text-xs text-gray-600">
                                  {Object.entries(action.parameters).map(([key, value]) => (
                                    <div key={key} className="ml-4">
                                      {key}: {typeof value === 'object' ? JSON.stringify(value) : String(value)}
                                    </div>
                                  ))}
                                </div>
                              </div>
                            ))}
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
        
        {filteredAndSortedRules.length === 0 && (
          <div className="text-center py-12 bg-white rounded-lg border">
            <Bot className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No automation rules found</h3>
            <p className="text-gray-500 mb-4">
              Create your first automation rule to start managing collections automatically
            </p>
            <Button onClick={() => setIsCreateModalOpen(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create Your First Rule
            </Button>
          </div>
        )}
      </div>

      {/* Create Modal */}
      {isCreateModalOpen && (
        <AutomationRuleModal
          onSave={(rule) => {
            const newRule: AutomationRule = {
              id: Date.now().toString(),
              name: rule.name || 'Untitled Rule',
              description: rule.description || '',
              enabled: rule.enabled ?? true,
              trigger: rule.trigger || { type: 'manual' },
              actions: rule.actions || [],
              createdAt: new Date().toISOString(),
              runCount: 0,
              successCount: 0,
              errorCount: 0,
              status: 'idle',
            }
            setRules(prev => [...prev, newRule])
            setIsCreateModalOpen(false)
            toast.success('Automation rule created')
          }}
          onClose={() => setIsCreateModalOpen(false)}
        />
      )}

      {/* Edit Modal */}
      {editingRule && (
        <AutomationRuleModal
          rule={editingRule}
          onSave={(updated) => {
            setRules(prev => prev.map(r => r.id === editingRule.id ? { ...r, ...updated } : r))
            setEditingRule(null)
            toast.success('Automation rule updated')
          }}
          onClose={() => setEditingRule(null)}
        />
      )}
    </div>
  )
}

// Automation Rule Create/Edit Modal
const AutomationRuleModal: React.FC<{
  rule?: AutomationRule
  onSave: (rule: Partial<AutomationRule>) => void
  onClose: () => void
}> = ({ rule, onSave, onClose }) => {
  const [name, setName] = useState(rule?.name || '')
  const [description, setDescription] = useState(rule?.description || '')
  const [enabled, setEnabled] = useState(rule?.enabled ?? true)
  const [triggerType, setTriggerType] = useState<'schedule' | 'event' | 'manual'>(rule?.trigger?.type || 'manual')
  const [schedule, setSchedule] = useState(rule?.trigger?.schedule || '0 0 * * *')
  const [eventType, setEventType] = useState(rule?.trigger?.eventType || 'file_added')
  const [actions, setActions] = useState<AutomationAction[]>(rule?.actions || [])

  const addAction = () => {
    setActions(prev => [...prev, {
      id: Date.now().toString(),
      type: 'add_to_collection',
      parameters: {},
      order: prev.length + 1,
    }])
  }

  const removeAction = (id: string) => {
    setActions(prev => prev.filter(a => a.id !== id))
  }

  const updateActionType = (id: string, type: AutomationAction['type']) => {
    setActions(prev => prev.map(a => a.id === id ? { ...a, type, parameters: {} } : a))
  }

  const handleSubmit = () => {
    if (!name.trim()) {
      toast.error('Rule name is required')
      return
    }
    onSave({
      name,
      description,
      enabled,
      trigger: {
        type: triggerType,
        ...(triggerType === 'schedule' && { schedule }),
        ...(triggerType === 'event' && { eventType }),
      },
      actions,
    })
  }

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg max-w-4xl w-full max-h-[90vh] overflow-y-auto p-6">
        <h3 className="text-xl font-bold mb-4">{rule ? 'Edit' : 'Create'} Automation Rule</h3>

        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Rule Name</label>
            <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="Enter rule name" />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Description</label>
            <Input value={description} onChange={(e) => setDescription(e.target.value)} placeholder="Describe what this rule does" />
          </div>

          <div className="flex items-center gap-3">
            <Switch checked={enabled} onCheckedChange={setEnabled} />
            <span className="text-sm text-gray-700 dark:text-gray-300">Enabled</span>
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">Trigger Type</label>
            <div className="grid grid-cols-3 gap-3">
              {TRIGGER_TYPES.map(t => (
                <button
                  key={t.value}
                  onClick={() => setTriggerType(t.value as typeof triggerType)}
                  className={`p-3 rounded-lg border text-left transition-colors ${
                    triggerType === t.value
                      ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div className="font-medium text-sm">{t.label}</div>
                  <div className="text-xs text-gray-500 mt-1">{t.description}</div>
                </button>
              ))}
            </div>
          </div>

          {triggerType === 'schedule' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Cron Schedule</label>
              <Input value={schedule} onChange={(e) => setSchedule(e.target.value)} placeholder="0 0 * * *" />
              <p className="text-xs text-gray-500 mt-1">Standard cron format: minute hour day month weekday</p>
            </div>
          )}

          {triggerType === 'event' && (
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">Event Type</label>
              <select
                value={eventType}
                onChange={(e) => setEventType(e.target.value)}
                className="w-full border rounded-lg p-2 text-sm"
              >
                <option value="file_added">File Added</option>
                <option value="file_removed">File Removed</option>
                <option value="file_modified">File Modified</option>
                <option value="scan_complete">Scan Complete</option>
                <option value="collection_updated">Collection Updated</option>
              </select>
            </div>
          )}

          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-sm font-medium text-gray-700 dark:text-gray-300">Actions</label>
              <Button variant="outline" onClick={addAction}>
                <Plus className="w-3 h-3 mr-1" />
                Add Action
              </Button>
            </div>
            <div className="space-y-2">
              {actions.map((action, idx) => (
                <div key={action.id} className="flex items-center gap-2 p-3 border rounded-lg">
                  <span className="text-xs text-gray-500 w-6">{idx + 1}.</span>
                  <select
                    value={action.type}
                    onChange={(e) => updateActionType(action.id, e.target.value as AutomationAction['type'])}
                    className="flex-1 border rounded p-2 text-sm"
                  >
                    {ACTION_TYPES.map(a => (
                      <option key={a.value} value={a.value}>{a.label}</option>
                    ))}
                  </select>
                  <Button variant="outline" onClick={() => removeAction(action.id)}>
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>
              ))}
              {actions.length === 0 && (
                <p className="text-sm text-gray-500 text-center py-4">No actions added. Click &quot;Add Action&quot; above.</p>
              )}
            </div>
          </div>
        </div>

        <div className="flex justify-end gap-2 mt-6 pt-4 border-t">
          <Button variant="outline" onClick={onClose}>Cancel</Button>
          <Button onClick={handleSubmit}>{rule ? 'Save Changes' : 'Create Rule'}</Button>
        </div>
      </div>
    </div>
  )
}


export default CollectionAutomation;