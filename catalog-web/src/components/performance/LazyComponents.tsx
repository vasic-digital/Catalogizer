import React, { lazy, Suspense } from 'react';
import { Loader2 } from 'lucide-react';

// Lazy loaded components for better performance
const CollectionTemplates = lazy(() => import('../collections/CollectionTemplates'));
const AdvancedSearch = lazy(() => import('../collections/AdvancedSearch'));
const CollectionAutomation = lazy(() => import('../collections/CollectionAutomation'));
const ExternalIntegrations = lazy(() => import('../collections/ExternalIntegrations'));
const SmartCollectionBuilder = lazy(() => import('../collections/SmartCollectionBuilder').then(m => ({ default: m.SmartCollectionBuilder })));
const CollectionAnalytics = lazy(() => import('../collections/CollectionAnalytics').then(m => ({ default: m.CollectionAnalytics })));
const BulkOperations = lazy(() => import('../collections/BulkOperations'));

interface LazyComponentProps {
  componentName: string;
  fallback?: React.ReactNode;
  children?: React.ReactNode;
}

export const ComponentLoader: React.FC<LazyComponentProps> = ({
  componentName,
  fallback = (
    <div className="flex items-center justify-center min-h-[200px]">
      <Loader2 className="w-8 h-8 animate-spin text-blue-600" />
    </div>
  ),
  children
}) => (
  <Suspense fallback={fallback}>
    {children}
  </Suspense>
);

// Preload specific components for better UX
export const preloadComponent = (componentName: string) => {
  switch (componentName) {
    case 'CollectionTemplates':
      import('../collections/CollectionTemplates');
      break;
    case 'AdvancedSearch':
      import('../collections/AdvancedSearch');
      break;
    case 'CollectionAutomation':
      import('../collections/CollectionAutomation');
      break;
    case 'ExternalIntegrations':
      import('../collections/ExternalIntegrations');
      break;
    case 'SmartCollectionBuilder':
      import('../collections/SmartCollectionBuilder');
      break;
    case 'CollectionAnalytics':
      import('../collections/CollectionAnalytics');
      break;
    case 'BulkOperations':
      import('../collections/BulkOperations');
      break;
    default:
      console.warn(`Unknown component: ${componentName}`);
  }
};

// Export lazy-loaded components
export {
  CollectionTemplates,
  AdvancedSearch,
  CollectionAutomation,
  ExternalIntegrations,
  SmartCollectionBuilder,
  CollectionAnalytics,
  BulkOperations
};

// Component usage example:
// <ComponentLoader componentName="CollectionTemplates">
//   <CollectionTemplates {...props} />
// </ComponentLoader>