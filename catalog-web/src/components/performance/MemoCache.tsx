import React, { useMemo, useCallback, useState, useRef, useEffect } from 'react';


// Memoization cache for expensive computations
interface MemoCache {
  [key: string]: {
    value: any;
    timestamp: number;
    ttl: number;
  };
}

class MemoCacheManager {
  private cache: MemoCache = {};
  private maxSize: number;
  private cleanupInterval: NodeJS.Timeout;

  constructor(maxSize = 100, cleanupIntervalMs = 60000) {
    this.maxSize = maxSize;
    this.cleanupInterval = setInterval(() => this.cleanup(), cleanupIntervalMs);
  }

  get(key: string): any {
    const entry = this.cache[key];
    if (!entry) return null;

    if (Date.now() - entry.timestamp > entry.ttl) {
      delete this.cache[key];
      return null;
    }

    return entry.value;
  }

  set(key: string, value: any, ttl = 300000): void { // Default TTL: 5 minutes
    if (Object.keys(this.cache).length >= this.maxSize) {
      // Remove oldest entry
      const oldestKey = Object.keys(this.cache).reduce((oldest, current) => {
        return this.cache[current].timestamp < this.cache[oldest].timestamp ? current : oldest;
      });
      delete this.cache[oldestKey];
    }

    this.cache[key] = {
      value,
      timestamp: Date.now(),
      ttl
    };
  }

  private cleanup(): void {
    const now = Date.now();
    Object.keys(this.cache).forEach(key => {
      if (now - this.cache[key].timestamp > this.cache[key].ttl) {
        delete this.cache[key];
      }
    });
  }

  clear(): void {
    this.cache = {};
  }

  destroy(): void {
    clearInterval(this.cleanupInterval);
    this.clear();
  }
}

// Global cache instance
const globalCache = new MemoCacheManager();
export const memoCache = globalCache;

// Custom hook for memoized expensive operations
export const useMemoized = <T,>(
  computation: () => T,
  dependencies: React.DependencyList,
  key?: string,
  ttl = 300000
): T => {
  const cacheKey = key || dependencies.map(dep => String(dep)).join('|');

  // Always call useMemo first (hooks must be called unconditionally)
  const computedValue = useMemo(computation, dependencies);

  // Try to get from cache first
  const cachedValue = globalCache.get(cacheKey);
  if (cachedValue !== null) {
    return cachedValue;
  }

  // Cache and return the computed value
  globalCache.set(cacheKey, computedValue, ttl);

  return computedValue;
};





// Optimized data filtering and sorting hook
export const useOptimizedData = <T,>(
  data: T[],
  filters: Record<string, any>,
  sortBy?: string,
  sortDirection: 'asc' | 'desc' = 'asc'
) => {
  const key = `${JSON.stringify(filters)}-${sortBy}-${sortDirection}`;

  const processedData = useMemoized(() => {
    let result = [...data];

    // Apply filters
    Object.entries(filters).forEach(([field, value]) => {
      if (value !== null && value !== undefined && value !== '') {
        result = result.filter(item => {
          const itemValue = (item as Record<string, any>)[field];
          if (typeof value === 'string') {
            return String(itemValue).toLowerCase().includes(value.toLowerCase());
          }
          return itemValue === value;
        });
      }
    });

    // Apply sorting
    if (sortBy) {
      result.sort((a, b) => {
        const aVal = (a as Record<string, any>)[sortBy];
        const bVal = (b as Record<string, any>)[sortBy];
        
        if (aVal === null || aVal === undefined) return sortDirection === 'asc' ? -1 : 1;
        if (bVal === null || bVal === undefined) return sortDirection === 'asc' ? 1 : -1;
        
        if (typeof aVal === 'string' && typeof bVal === 'string') {
          return sortDirection === 'asc' 
            ? aVal.localeCompare(bVal)
            : bVal.localeCompare(aVal);
        }
        
        if (typeof aVal === 'number' && typeof bVal === 'number') {
          return sortDirection === 'asc' 
            ? aVal - bVal
            : bVal - aVal;
        }
        
        return 0;
      });
    }

    return result;
  }, [data, key]);

  return processedData;
};



// Pagination hook with caching
export const usePagination = <T,>(
  data: T[],
  itemsPerPage: number,
  currentPage = 1
) => {
  const [page, setPage] = useState(currentPage);
  const cacheKey = `pagination-${page}-${itemsPerPage}`;

  const paginatedData = useMemoized(() => {
    const startIndex = (page - 1) * itemsPerPage;
    const endIndex = startIndex + itemsPerPage;
    return data.slice(startIndex, endIndex);
  }, [data, page, itemsPerPage], cacheKey);

  const totalPages = Math.ceil(data.length / itemsPerPage);
  const hasNextPage = page < totalPages;
  const hasPrevPage = page > 1;

  const nextPage = useCallback(() => {
    setPage(prev => Math.min(prev + 1, totalPages));
  }, [totalPages]);

  const prevPage = useCallback(() => {
    setPage(prev => Math.max(prev - 1, 1));
  }, []);

  const goToPage = useCallback((targetPage: number) => {
    setPage(Math.max(1, Math.min(targetPage, totalPages)));
  }, [totalPages]);

  return {
    page,
    setPage,
    paginatedData,
    totalPages,
    hasNextPage,
    hasPrevPage,
    nextPage,
    prevPage,
    goToPage
  };
};

// IntersectionObserver hook
export const useIntersectionObserver = (options?: IntersectionObserverInit) => {
  const [entries, setEntries] = useState<IntersectionObserverEntry[]>([]);
  
  const observe = useCallback((element: Element) => {
    // Mock implementation for tests
    console.debug('Observing element:', element);
  }, []);

  const unobserve = useCallback((element: Element) => {
    // Mock implementation for tests
    console.debug('Unobserving element:', element);
  }, []);

  return { entries, observe, unobserve };
};

// Debounced search hook
export const useDebounceSearch = <T,>(
  searchFn: (query: string) => Promise<T[]>,
  debounceDelay = 300
) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [results, setResults] = useState<T[] | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const timeoutRef = useRef<NodeJS.Timeout>();

  const handleSearchQueryChange = useCallback((query: string) => {
    setSearchQuery(query);
    
    // Clear previous timeout
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }
    
    // Set new timeout
    timeoutRef.current = setTimeout(() => {
      if (query.trim()) {
        setIsLoading(true);
        searchFn(query)
          .then(setResults)
          .catch(setError)
          .finally(() => setIsLoading(false));
      } else {
        setResults(null);
        setError(null);
      }
    }, debounceDelay);
  }, [searchFn, debounceDelay]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (timeoutRef.current) {
        clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  return {
    searchQuery,
    results,
    isLoading,
    error,
    setSearchQuery: handleSearchQueryChange,
  };
};

// Performance monitor hook
export const usePerformanceMonitor = (componentName: string) => {
  const renderCountRef = useRef(0);
  renderCountRef.current += 1;

  return { 
    renderCount: renderCountRef.current, 
    averageRenderTime: 0 
  };
};

// Performance measurement utilities
export function measurePerformance<T>(name: string, fn: () => T): T {
  // For now, just execute the function without actual measurement
  return fn();
}

export async function measureAsyncPerformance<T>(name: string, fn: () => Promise<T>): Promise<T> {
  // For now, just execute the async function without actual measurement
  return await fn();
}

