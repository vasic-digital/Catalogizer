import React, { useMemo, useCallback, useState, useRef, useEffect } from 'react';
import { Loader2, Search, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react';

interface VirtualListProps {
  items: any[];
  itemHeight: number;
  height: number;
  renderItem: (props: any) => React.ReactNode;
  overscanCount?: number;
  itemKey?: (index: number, data: any) => string;
}

interface VirtualizedTableProps {
  data: any[];
  columns: Array<{
    key: string;
    label: string;
    width: number;
    render?: (value: any, row: any) => React.ReactNode;
  }>;
  height?: number;
  rowHeight?: number;
  searchable?: boolean;
  sortable?: boolean;
}

interface InfiniteScrollProps {
  items: any[];
  hasNextPage: boolean;
  isNextPageLoading: boolean;
  loadNextPage: () => void;
  renderItem: (item: any, index: number) => React.ReactNode;
  threshold?: number;
}

// Generic virtualized list component
export const VirtualList: React.FC<VirtualListProps> = ({
  items,
  itemHeight,
  height,
  renderItem,
  overscanCount = 5,
  itemKey
}) => {
  const [startIndex, setStartIndex] = useState(0);
  const [endIndex, setEndIndex] = useState(Math.ceil(height / itemHeight) + overscanCount);
  const containerRef = useRef<HTMLDivElement>(null);

  const visibleItems = useMemo(() => {
    return items.slice(startIndex, endIndex);
  }, [items, startIndex, endIndex]);

  const handleScroll = useCallback((e: React.UIEvent<HTMLDivElement>) => {
    const scrollTop = e.currentTarget.scrollTop;
    const newStartIndex = Math.floor(scrollTop / itemHeight);
    const newEndIndex = newStartIndex + Math.ceil(height / itemHeight) + overscanCount;
    
    setStartIndex(Math.max(0, newStartIndex));
    setEndIndex(Math.min(items.length, newEndIndex));
  }, [itemHeight, height, overscanCount, items.length]);

  return (
    <div
      ref={containerRef}
      style={{ height, overflow: 'auto' }}
      onScroll={handleScroll}
      className="virtual-list-container"
    >
      <div style={{ height: items.length * itemHeight, position: 'relative' }}>
        {visibleItems.map((item, index) => {
          const actualIndex = startIndex + index;
          const key = itemKey ? itemKey(actualIndex, item) : `item-${actualIndex}`;
          
          return (
            <div
              key={key}
              style={{
                position: 'absolute',
                top: actualIndex * itemHeight,
                left: 0,
                right: 0,
                height: itemHeight
              }}
            >
              {renderItem({ index: actualIndex, data: item })}
            </div>
          );
        })}
      </div>
    </div>
  );
};

// Virtualized table component
export const VirtualizedTable: React.FC<VirtualizedTableProps> = ({
  data,
  columns,
  height = 400,
  rowHeight = 50,
  searchable = false,
  sortable = false
}) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [sortConfig, setSortConfig] = useState<{ key: string; direction: 'asc' | 'desc' } | null>(null);

  // Filter and sort data
  const processedData = useMemo(() => {
    let result = [...data];

    // Apply search filter
    if (searchTerm) {
      result = result.filter(item =>
        Object.values(item).some(value =>
          String(value).toLowerCase().includes(searchTerm.toLowerCase())
        )
      );
    }

    // Apply sorting
    if (sortConfig) {
      result.sort((a, b) => {
        const aValue = a[sortConfig.key];
        const bValue = b[sortConfig.key];

        if (aValue < bValue) return sortConfig.direction === 'asc' ? -1 : 1;
        if (aValue > bValue) return sortConfig.direction === 'asc' ? 1 : -1;
        return 0;
      });
    }

    return result;
  }, [data, searchTerm, sortConfig]);

  const handleSort = (key: string) => {
    setSortConfig(current => {
      if (!current || current.key !== key) {
        return { key, direction: 'asc' };
      }
      if (current.direction === 'asc') {
        return { key, direction: 'desc' };
      }
      return null;
    });
  };

  const renderRow = useCallback(({ index, data: rowData }) => (
    <div className="flex items-center border-b border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800">
      {columns.map((column, colIndex) => (
        <div
          key={column.key}
          className="px-4 py-2 text-sm text-gray-900 dark:text-white"
          style={{ width: column.width, flexShrink: 0 }}
        >
          {column.render ? column.render(rowData[column.key], rowData) : rowData[column.key]}
        </div>
      ))}
    </div>
  ), [columns]);

  return (
    <div className="virtualized-table">
      {/* Search Bar */}
      {searchable && (
        <div className="mb-4 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
          />
        </div>
      )}

      {/* Table Header */}
      <div className="flex items-center border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
        {columns.map((column) => (
          <div
            key={column.key}
            className="px-4 py-2 text-sm font-semibold text-gray-900 dark:text-white flex items-center"
            style={{ width: column.width, flexShrink: 0 }}
          >
            {column.label}
            {sortable && (
              <button
                onClick={() => handleSort(column.key)}
                className="ml-2 p-1 hover:bg-gray-200 dark:hover:bg-gray-700 rounded"
              >
                {sortConfig?.key === column.key ? (
                  sortConfig.direction === 'asc' ? (
                    <ArrowUp className="w-3 h-3" />
                  ) : (
                    <ArrowDown className="w-3 h-3" />
                  )
                ) : (
                  <ArrowUpDown className="w-3 h-3" />
                )}
              </button>
            )}
          </div>
        ))}
      </div>

      {/* Virtual List for Table Rows */}
      <VirtualList
        items={processedData}
        itemHeight={rowHeight}
        height={height}
        renderItem={renderRow}
      />
    </div>
  );
};

// Infinite scroll component
export const InfiniteScroll: React.FC<InfiniteScrollProps> = ({
  items,
  hasNextPage,
  isNextPageLoading,
  loadNextPage,
  renderItem,
  threshold = 0.8
}) => {
  const [isIntersecting, setIsIntersecting] = useState(false);
  const loadMoreRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const observer = new IntersectionObserver(
      ([entry]) => {
        setIsIntersecting(entry.isIntersecting);
      },
      { threshold }
    );

    if (loadMoreRef.current) {
      observer.observe(loadMoreRef.current);
    }

    return () => {
      if (loadMoreRef.current) {
        observer.unobserve(loadMoreRef.current);
      }
    };
  }, [threshold]);

  useEffect(() => {
    if (isIntersecting && hasNextPage && !isNextPageLoading) {
      loadNextPage();
    }
  }, [isIntersecting, hasNextPage, isNextPageLoading, loadNextPage]);

  return (
    <div className="infinite-scroll-container">
      {items.map((item, index) => renderItem(item, index))}
      
      {/* Loading indicator */}
      {isNextPageLoading && (
        <div className="flex justify-center py-4">
          <Loader2 className="w-6 h-6 animate-spin text-blue-600" />
        </div>
      )}
      
      {/* Intersection observer target */}
      <div ref={loadMoreRef} className="h-1" />
    </div>
  );
};