import React, { useState, useEffect, useMemo } from 'react';
import { Box, Typography, Paper, List, ListItem, ListItemText, Divider, Button, Chip, IconButton } from '@mui/material';
import { Refresh, Download, ContentCopy } from '@mui/icons-material';

interface BundleInfo {
  name: string;
  size: number;
  sizeFormatted: string;
  path: string;
  type: 'js' | 'css' | 'other';
  chunks?: number;
}

interface ChunkInfo {
  name: string;
  size: number;
  sizeFormatted: string;
  modules: string[];
  parents: string[];
}

interface BundleAnalysis {
  totalSize: number;
  totalSizeFormatted: string;
  bundles: BundleInfo[];
  chunks: ChunkInfo[];
  recommendations: string[];
  potentialSavings: number;
  potentialSavingsFormatted: string;
}

export const BundleAnalyzer: React.FC = () => {
  const [analysis, setAnalysis] = useState<BundleAnalysis | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [showDetails, setShowDetails] = useState(false);

  // Simulate bundle analysis (in real app, this would fetch from build stats)
  const analyzeBundle = async (): Promise<BundleAnalysis> => {
    // Simulate API call delay
    await new Promise(resolve => setTimeout(resolve, 1500));

    // Mock bundle data
    const bundles: BundleInfo[] = [
      {
        name: 'main',
        size: 1850000,
        sizeFormatted: '1.85 MB',
        path: '/static/js/main.js',
        type: 'js',
        chunks: 5
      },
      {
        name: 'vendor',
        size: 3200000,
        sizeFormatted: '3.2 MB',
        path: '/static/js/vendor.js',
        type: 'js',
        chunks: 12
      },
      {
        name: 'collections',
        size: 450000,
        sizeFormatted: '450 KB',
        path: '/static/js/collections.js',
        type: 'js',
        chunks: 2
      },
      {
        name: 'main.css',
        size: 185000,
        sizeFormatted: '185 KB',
        path: '/static/css/main.css',
        type: 'css'
      },
      {
        name: 'components',
        size: 620000,
        sizeFormatted: '620 KB',
        path: '/static/js/components.js',
        type: 'js',
        chunks: 3
      }
    ];

    const chunks: ChunkInfo[] = [
      {
        name: 'collection-templates',
        size: 280000,
        sizeFormatted: '280 KB',
        modules: ['CollectionTemplates', 'TemplatePreview', 'TemplateCard'],
        parents: ['collections', 'main']
      },
      {
        name: 'advanced-search',
        size: 195000,
        sizeFormatted: '195 KB',
        modules: ['AdvancedSearch', 'RuleBuilder', 'SearchFilters'],
        parents: ['collections', 'main']
      },
      {
        name: 'collection-automation',
        size: 245000,
        sizeFormatted: '245 KB',
        modules: ['CollectionAutomation', 'WorkflowEditor', 'RuleEngine'],
        parents: ['collections', 'main']
      },
      {
        name: 'external-integrations',
        size: 180000,
        sizeFormatted: '180 KB',
        modules: ['ExternalIntegrations', 'IntegrationCard', 'SyncSettings'],
        parents: ['collections', 'main']
      }
    ];

    const totalSize = bundles.reduce((sum, bundle) => sum + bundle.size, 0);
    const recommendations = [
      'Consider lazy loading CollectionTemplates component (saves ~280KB)',
      'ExternalIntegrations can be loaded on-demand (saves ~180KB)',
      'Advanced search functionality could be code-split (saves ~195KB)',
      'Collection automation features have low usage, consider lazy loading (saves ~245KB)',
      'Vendor bundle contains unused libraries, consider tree shaking',
      'CSS bundle can be optimized by removing unused styles',
      'Large images in components should be compressed or moved to CDN'
    ];

    return {
      totalSize,
      totalSizeFormatted: `${(totalSize / 1024 / 1024).toFixed(2)} MB`,
      bundles,
      chunks,
      recommendations,
      potentialSavings: 900000,
      potentialSavingsFormatted: '900 KB'
    };
  };

  useEffect(() => {
    loadAnalysis();
  }, []);

  const loadAnalysis = async () => {
    setIsLoading(true);
    try {
      const data = await analyzeBundle();
      setAnalysis(data);
    } catch (error) {
      console.error('Failed to analyze bundle:', error);
    } finally {
      setIsLoading(false);
    }
  };

  // Calculate bundle statistics
  const bundleStats = useMemo(() => {
    if (!analysis) return null;

    const jsBundles = analysis.bundles.filter(b => b.type === 'js');
    const cssBundles = analysis.bundles.filter(b => b.type === 'css');
    
    return {
      totalBundles: analysis.bundles.length,
      jsBundles: jsBundles.length,
      cssBundles: cssBundles.length,
      largestBundle: analysis.bundles.reduce((largest, bundle) => 
        bundle.size > largest.size ? bundle : largest
      ),
      averageBundleSize: analysis.bundles.reduce((sum, b) => sum + b.size, 0) / analysis.bundles.length
    };
  }, [analysis]);

  const handleOptimizeBundle = () => {
    // In a real app, this would trigger bundle optimization
    console.log('Triggering bundle optimization...');
    // This could trigger a new build with optimizations
  };

  const handleExportReport = () => {
    if (!analysis) return;

    const report = {
      timestamp: new Date().toISOString(),
      analysis
    };

    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `bundle-analysis-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleCopyStats = () => {
    if (!analysis) return;

    const stats = `
Bundle Analysis Report
=====================
Total Size: ${analysis.totalSizeFormatted}
Potential Savings: ${analysis.potentialSavingsFormatted}
Total Bundles: ${bundleStats?.totalBundles}
JS Bundles: ${bundleStats?.jsBundles}
CSS Bundles: ${bundleStats?.cssBundles}
Largest Bundle: ${bundleStats?.largestBundle?.name} (${bundleStats?.largestBundle?.sizeFormatted})
Average Bundle Size: ${(bundleStats?.averageBundleSize! / 1024).toFixed(2)} KB
    `.trim();

    navigator.clipboard.writeText(stats);
  };

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight={400}>
        <Typography>Analyzing bundle...</Typography>
      </Box>
    );
  }

  if (!analysis) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight={400}>
        <Typography>Failed to load bundle analysis</Typography>
      </Box>
    );
  }

  return (
    <Box p={3}>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h4">Bundle Analysis</Typography>
        <Box>
          <IconButton onClick={loadAnalysis} title="Refresh">
            <Refresh />
          </IconButton>
          <IconButton onClick={handleCopyStats} title="Copy Stats">
            <ContentCopy />
          </IconButton>
          <IconButton onClick={handleExportReport} title="Export Report">
            <Download />
          </IconButton>
        </Box>
      </Box>

      {/* Summary Cards */}
      <Box display="flex" gap={3} mb={4} flexWrap="wrap">
        <Paper sx={{ p: 3, minWidth: 200 }}>
          <Typography variant="h6" color="primary">
            {analysis.totalSizeFormatted}
          </Typography>
          <Typography variant="body2">Total Bundle Size</Typography>
        </Paper>
        
        <Paper sx={{ p: 3, minWidth: 200 }}>
          <Typography variant="h6" color="secondary">
            {analysis.potentialSavingsFormatted}
          </Typography>
          <Typography variant="body2">Potential Savings</Typography>
        </Paper>
        
        <Paper sx={{ p: 3, minWidth: 200 }}>
          <Typography variant="h6" color="info.main">
            {bundleStats?.totalBundles}
          </Typography>
          <Typography variant="body2">Total Bundles</Typography>
        </Paper>
        
        <Paper sx={{ p: 3, minWidth: 200 }}>
          <Typography variant="h6" color="warning.main">
            {bundleStats?.jsBundles}
          </Typography>
          <Typography variant="body2">JS Bundles</Typography>
        </Paper>
      </Box>

      {/* Bundle Details */}
      <Paper sx={{ mb: 3 }}>
        <Box p={2} borderBottom="1px solid #eee">
          <Box display="flex" justifyContent="space-between" alignItems="center">
            <Typography variant="h6">Bundle Breakdown</Typography>
            <Button onClick={() => setShowDetails(!showDetails)}>
              {showDetails ? 'Hide Details' : 'Show Details'}
            </Button>
          </Box>
        </Box>
        
        <List>
          {analysis.bundles.map((bundle, index) => (
            <React.Fragment key={bundle.name}>
              <ListItem>
                <ListItemText
                  primary={
                    <Box display="flex" justifyContent="space-between" alignItems="center">
                      <Typography variant="subtitle1">{bundle.name}</Typography>
                      <Box>
                        <Typography variant="body2" color="textSecondary">
                          {bundle.sizeFormatted}
                        </Typography>
                        <Chip 
                          size="small" 
                          label={bundle.type.toUpperCase()}
                          color={bundle.type === 'js' ? 'primary' : 'secondary'}
                          sx={{ ml: 1 }}
                        />
                      </Box>
                    </Box>
                  }
                  secondary={
                    <Box mt={1}>
                      <Typography variant="body2" color="textSecondary">
                        Path: {bundle.path}
                      </Typography>
                      {bundle.chunks && (
                        <Typography variant="body2" color="textSecondary">
                          Chunks: {bundle.chunks}
                        </Typography>
                      )}
                    </Box>
                  }
                />
              </ListItem>
              {index < analysis.bundles.length - 1 && <Divider />}
            </React.Fragment>
          ))}
        </List>
      </Paper>

      {/* Optimization Recommendations */}
      <Paper sx={{ mb: 3 }}>
        <Box p={2} borderBottom="1px solid #eee">
          <Typography variant="h6">Optimization Recommendations</Typography>
        </Box>
        <List>
          {analysis.recommendations.map((recommendation, index) => (
            <ListItem key={index}>
              <ListItemText
                primary={
                  <Box display="flex" alignItems="center">
                    <Typography variant="body1">{recommendation}</Typography>
                  </Box>
                }
              />
            </ListItem>
          ))}
        </List>
      </Paper>

      {/* Action Buttons */}
      <Box display="flex" gap={2} justifyContent="center">
        <Button
          variant="contained"
          color="primary"
          onClick={handleOptimizeBundle}
          startIcon={<Refresh />}
        >
          Optimize Bundle
        </Button>
        <Button
          variant="outlined"
          onClick={handleExportReport}
          startIcon={<Download />}
        >
          Export Report
        </Button>
      </Box>

      {/* Chunk Details (when shown) */}
      {showDetails && (
        <Paper sx={{ mt: 3 }}>
          <Box p={2} borderBottom="1px solid #eee">
            <Typography variant="h6">Chunk Analysis</Typography>
          </Box>
          <List>
            {analysis.chunks.map((chunk, index) => (
              <React.Fragment key={chunk.name}>
                <ListItem>
                  <ListItemText
                    primary={
                      <Box display="flex" justifyContent="space-between" alignItems="center">
                        <Typography variant="subtitle1">{chunk.name}</Typography>
                        <Typography variant="body2" color="textSecondary">
                          {chunk.sizeFormatted}
                        </Typography>
                      </Box>
                    }
                    secondary={
                      <Box mt={1}>
                        <Typography variant="body2" color="textSecondary">
                          Modules: {chunk.modules.join(', ')}
                        </Typography>
                        <Typography variant="body2" color="textSecondary">
                          Parents: {chunk.parents.join(', ')}
                        </Typography>
                      </Box>
                    }
                  />
                </ListItem>
                {index < analysis.chunks.length - 1 && <Divider />}
              </React.Fragment>
            ))}
          </List>
        </Paper>
      )}
    </Box>
  );
};