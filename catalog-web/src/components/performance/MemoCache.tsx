import React, { useMemo, useCallback, useState, useRef, useEffect } from 'react';
import debounce from 'lodash/debounce';

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

// Custom hook for memoized expensive operations
export const useMemoized = <T,>(
  computation: () => T,
  dependencies: React.DependencyList,
  key?: string,
  ttl = 300000
): T => {
  const cacheKey = key || dependencies.map(dep => String(dep)).join('|');
  
  // Try to get from cache first
  const cachedValue = globalCache.get(cacheKey);
  if (cachedValue !== null) {
    return cachedValue;
  }

  // Compute and cache
  const value = useMemo(computation, dependencies);
  globalCache.set(cacheKey, value, ttl);
  
  return value;
};

// Debounced search hook
export const useDebounceSearch = <T,>(
  searchFunction: (query: string) => Promise<T>,
  delay = 300
) => {
  const [searchQuery, setSearchQuery] = useState('');
  const [results, setResults] = useState<T | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);

  const debouncedSearch = useMemo(
    () => debounce(async (query: string) => {
      if (!query.trim()) {
        setResults(null);
        return;
      }

      setIsLoading(true);
      setError(null);

      try {
        const result = await searchFunction(query);
        setResults(result);
      } catch (err) {
        setError(err as Error);
      } finally {
        setIsLoading(false);
      }
    }, delay),
    [searchFunction, delay]
  );

  useEffect(() => {
    debouncedSearch(searchQuery);
    return () => debouncedSearch.cancel();
  }, [searchQuery, debouncedSearch]);

  return {
    searchQuery,
    setSearchQuery,
    results,
    isLoading,
    error
  };
};

// Component to monitor and optimize render performance
export const usePerformanceMonitor = (componentName: string) => {
  const renderCount = useRef(0);
  const renderTimes = useRef<number[]>([]);
  const lastRenderTime = useRef(Date.now());

  useEffect(() => {
    renderCount.current++;
    const now = Date.now();
    const renderTime = now - lastRenderTime.current;
    renderTimes.current.push(renderTime);
    lastRenderTime.current = now;

    // Keep only last 10 renders
    if (renderTimes.current.length > 10) {
      renderTimes.current.shift();
    }

    // Log performance warnings
    if (renderTime > 16) { // 16ms = 60fps threshold
      console.warn(`${componentName}: Slow render detected (${renderTime}ms)`);
    }

    if (renderCount.current % 10 === 0) {
      const avgRenderTime = renderTimes.current.reduce((a, b) => a + b, 0) / renderTimes.current.length;
      console.log(`${componentName}: Render #${renderCount.current}, Avg render time: ${avgRenderTime.toFixed(2)}ms`);
    }
  });

  return {
    renderCount: renderCount.current,
    averageRenderTime: renderTimes.current.length > 0 
      ? renderTimes.current.reduce((a, b) => a + b, 0) / renderTimes.current.length
      : 0
  };
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
          const itemValue = item[field];
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
        const aVal = a[sortBy];
        const bVal = b[sortBy];
        
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

// Intersection Observer hook for lazy loading
export const useIntersectionObserver = (
  options: IntersectionObserverInit = {}
) => {
  const [entries, setEntries] = useState<IntersectionObserverEntry[]>([]);
  const observer = useRef<IntersectionObserver>();

  const observe = useCallback((element: Element) => {
    if (observer.current) {
      observer.current.observe(element);
    }
  }, []);

  const unobserve = useCallback((element: Element) => {
    if (observer.current) {
      observer.current.unobserve(element);
    }
  }, []);

  useEffect(() => {
    if (typeof IntersectionObserver === 'undefined') {
      return;
    }

    observer.current = new IntersectionObserver((entries) => {
      setEntries(entries);
    }, options);

    return () => {
      if (observer.current) {
        observer.current.disconnect();
      }
    };
  }, [options]);

  return { entries, observe, unobserve };
};

// Pagination hook with caching
export const usePagination = <T,>(
  data: T[],
  itemsPerPage: number,
  currentPage: number = 1
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

// Export cache manager for direct usage if needed
export { globalCache as memoCache };

// Performance monitoring utility
export const measurePerformance = <T,>(
  name: string,
  fn: () => T
): T => {
  const start = performance.now();
  const result = fn();
  const end = performance.now();
  console.log(`${name}: ${(end - start).toFixed(2)}ms`);
  return result;
};

// Async performance monitoring
export const measureAsyncPerformance = async <T,>(
  name: string,
  fn: () => Promise<T>
): Promise<T> => {
  const start = performance.now();
  const result = await fn();
  const end = performance.now();
  console.log(`${name}: ${(end - start).toFixed(2)}ms`);
  return result;
};