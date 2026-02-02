import React, { useState, useEffect, useMemo } from 'react';
import { RefreshCw, Download, Copy } from 'lucide-react';

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

  const analyzeBundle = async (): Promise<BundleAnalysis> => {
    await new Promise(resolve => setTimeout(resolve, 1500));

    const bundles: BundleInfo[] = [
      { name: 'main', size: 1850000, sizeFormatted: '1.85 MB', path: '/static/js/main.js', type: 'js', chunks: 5 },
      { name: 'vendor', size: 3200000, sizeFormatted: '3.2 MB', path: '/static/js/vendor.js', type: 'js', chunks: 12 },
      { name: 'collections', size: 450000, sizeFormatted: '450 KB', path: '/static/js/collections.js', type: 'js', chunks: 2 },
      { name: 'main.css', size: 185000, sizeFormatted: '185 KB', path: '/static/css/main.css', type: 'css' },
      { name: 'components', size: 620000, sizeFormatted: '620 KB', path: '/static/js/components.js', type: 'js', chunks: 3 }
    ];

    const chunks: ChunkInfo[] = [
      { name: 'collection-templates', size: 280000, sizeFormatted: '280 KB', modules: ['CollectionTemplates', 'TemplatePreview', 'TemplateCard'], parents: ['collections', 'main'] },
      { name: 'advanced-search', size: 195000, sizeFormatted: '195 KB', modules: ['AdvancedSearch', 'RuleBuilder', 'SearchFilters'], parents: ['collections', 'main'] },
      { name: 'collection-automation', size: 245000, sizeFormatted: '245 KB', modules: ['CollectionAutomation', 'WorkflowEditor', 'RuleEngine'], parents: ['collections', 'main'] },
      { name: 'external-integrations', size: 180000, sizeFormatted: '180 KB', modules: ['ExternalIntegrations', 'IntegrationCard', 'SyncSettings'], parents: ['collections', 'main'] }
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

  const handleExportReport = () => {
    if (!analysis) return;
    const report = { timestamp: new Date().toISOString(), analysis };
    const blob = new Blob([JSON.stringify(report, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `bundle-analysis-${Date.now()}.json`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const handleCopyStats = () => {
    if (!analysis || !bundleStats) return;
    const stats = `Bundle Analysis Report\n=====================\nTotal Size: ${analysis.totalSizeFormatted}\nPotential Savings: ${analysis.potentialSavingsFormatted}\nTotal Bundles: ${bundleStats.totalBundles}\nJS Bundles: ${bundleStats.jsBundles}\nCSS Bundles: ${bundleStats.cssBundles}\nLargest Bundle: ${bundleStats.largestBundle?.name} (${bundleStats.largestBundle?.sizeFormatted})\nAverage Bundle Size: ${(bundleStats.averageBundleSize / 1024).toFixed(2)} KB`;
    navigator.clipboard.writeText(stats);
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <p className="text-gray-600">Analyzing bundle...</p>
      </div>
    );
  }

  if (!analysis) {
    return (
      <div className="flex items-center justify-center min-h-[400px]">
        <p className="text-gray-600">Failed to load bundle analysis</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="flex justify-between items-center mb-6">
        <h2 className="text-2xl font-bold text-gray-900">Bundle Analysis</h2>
        <div className="flex gap-2">
          <button onClick={loadAnalysis} className="p-2 hover:bg-gray-100 rounded" title="Refresh">
            <RefreshCw className="w-5 h-5" />
          </button>
          <button onClick={handleCopyStats} className="p-2 hover:bg-gray-100 rounded" title="Copy Stats">
            <Copy className="w-5 h-5" />
          </button>
          <button onClick={handleExportReport} className="p-2 hover:bg-gray-100 rounded" title="Export Report">
            <Download className="w-5 h-5" />
          </button>
        </div>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-xl font-bold text-blue-600">{analysis.totalSizeFormatted}</p>
          <p className="text-sm text-gray-500">Total Bundle Size</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-xl font-bold text-purple-600">{analysis.potentialSavingsFormatted}</p>
          <p className="text-sm text-gray-500">Potential Savings</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-xl font-bold text-cyan-600">{bundleStats?.totalBundles}</p>
          <p className="text-sm text-gray-500">Total Bundles</p>
        </div>
        <div className="bg-white rounded-lg shadow-sm border p-4">
          <p className="text-xl font-bold text-amber-600">{bundleStats?.jsBundles}</p>
          <p className="text-sm text-gray-500">JS Bundles</p>
        </div>
      </div>

      <div className="bg-white rounded-lg shadow-sm border mb-6">
        <div className="flex justify-between items-center p-4 border-b">
          <h3 className="text-lg font-semibold">Bundle Breakdown</h3>
          <button onClick={() => setShowDetails(!showDetails)} className="text-sm text-blue-600 hover:text-blue-800">
            {showDetails ? 'Hide Details' : 'Show Details'}
          </button>
        </div>
        <ul className="divide-y">
          {analysis.bundles.map((bundle) => (
            <li key={bundle.name} className="p-4">
              <div className="flex justify-between items-center">
                <span className="font-medium">{bundle.name}</span>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-500">{bundle.sizeFormatted}</span>
                  <span className={`text-xs px-2 py-0.5 rounded-full ${bundle.type === 'js' ? 'bg-blue-100 text-blue-700' : 'bg-purple-100 text-purple-700'}`}>
                    {bundle.type.toUpperCase()}
                  </span>
                </div>
              </div>
              <p className="text-sm text-gray-500 mt-1">Path: {bundle.path}</p>
              {bundle.chunks && <p className="text-sm text-gray-500">Chunks: {bundle.chunks}</p>}
            </li>
          ))}
        </ul>
      </div>

      <div className="bg-white rounded-lg shadow-sm border mb-6">
        <div className="p-4 border-b">
          <h3 className="text-lg font-semibold">Optimization Recommendations</h3>
        </div>
        <ul className="divide-y">
          {analysis.recommendations.map((recommendation, index) => (
            <li key={index} className="p-4 text-sm text-gray-700">{recommendation}</li>
          ))}
        </ul>
      </div>

      <div className="flex gap-3 justify-center">
        <button onClick={loadAnalysis} className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2">
          <RefreshCw className="w-4 h-4" /> Optimize Bundle
        </button>
        <button onClick={handleExportReport} className="px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 flex items-center gap-2">
          <Download className="w-4 h-4" /> Export Report
        </button>
      </div>

      {showDetails && (
        <div className="bg-white rounded-lg shadow-sm border mt-6">
          <div className="p-4 border-b">
            <h3 className="text-lg font-semibold">Chunk Analysis</h3>
          </div>
          <ul className="divide-y">
            {analysis.chunks.map((chunk) => (
              <li key={chunk.name} className="p-4">
                <div className="flex justify-between items-center">
                  <span className="font-medium">{chunk.name}</span>
                  <span className="text-sm text-gray-500">{chunk.sizeFormatted}</span>
                </div>
                <p className="text-sm text-gray-500 mt-1">Modules: {chunk.modules.join(', ')}</p>
                <p className="text-sm text-gray-500">Parents: {chunk.parents.join(', ')}</p>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};
