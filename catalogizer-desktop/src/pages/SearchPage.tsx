import { useState, useEffect, useCallback } from "react";
import { Search, Film, Music, Image, FileText, Loader2, Star, Clock } from "lucide-react";
import { apiService } from "../services/apiService";
import { useNavigate } from "react-router-dom";
import type { MediaItem, MediaSearchResponse } from "../types";

const MEDIA_TYPE_FILTERS = [
  { value: "", label: "All", icon: Search },
  { value: "movie", label: "Movies", icon: Film },
  { value: "music", label: "Music", icon: Music },
  { value: "image", label: "Images", icon: Image },
  { value: "document", label: "Documents", icon: FileText },
];

export default function SearchPage() {
  const navigate = useNavigate();
  const [query, setQuery] = useState("");
  const [mediaType, setMediaType] = useState("");
  const [results, setResults] = useState<MediaItem[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);

  const performSearch = useCallback(async (searchQuery: string, type: string) => {
    if (!searchQuery.trim()) {
      setResults([]);
      setTotal(0);
      setHasSearched(false);
      return;
    }

    setIsLoading(true);
    setError(null);
    setHasSearched(true);

    try {
      const response: MediaSearchResponse = await apiService.searchMedia({
        query: searchQuery.trim(),
        media_type: type || undefined,
        limit: 50,
        offset: 0,
        sort_by: "updated_at",
        sort_order: "desc",
      });
      setResults(response.items || []);
      setTotal(response.total || 0);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Search failed");
      setResults([]);
      setTotal(0);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Debounced search
  useEffect(() => {
    const timer = setTimeout(() => {
      performSearch(query, mediaType);
    }, 300);
    return () => clearTimeout(timer);
  }, [query, mediaType, performSearch]);

  const formatDuration = (seconds?: number) => {
    if (!seconds) return "";
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    return h > 0 ? `${h}h ${m}m` : `${m}m`;
  };

  const formatSize = (bytes?: number) => {
    if (!bytes) return "";
    const gb = bytes / (1024 * 1024 * 1024);
    if (gb >= 1) return `${gb.toFixed(1)} GB`;
    const mb = bytes / (1024 * 1024);
    return `${mb.toFixed(0)} MB`;
  };

  return (
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-3xl font-bold text-foreground mb-2">Search</h1>
        <p className="text-muted-foreground">
          Search across your entire media library
        </p>
      </div>

      <div className="max-w-2xl">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search for movies, TV shows, music..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-3 border border-input bg-background rounded-lg text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent text-lg"
          />
        </div>

        {/* Media type filters */}
        <div className="flex gap-2 mt-3">
          {MEDIA_TYPE_FILTERS.map(({ value, label, icon: Icon }) => (
            <button
              key={value}
              onClick={() => setMediaType(value)}
              className={`flex items-center gap-1.5 px-3 py-1.5 rounded-md text-sm transition-colors ${
                mediaType === value
                  ? "bg-primary text-primary-foreground"
                  : "bg-secondary text-secondary-foreground hover:bg-secondary/80"
              }`}
            >
              <Icon className="h-3.5 w-3.5" />
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      )}

      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-800 dark:text-red-300 rounded-md p-4">
          {error}
        </div>
      )}

      {/* Results */}
      {!isLoading && hasSearched && (
        <div>
          <p className="text-sm text-muted-foreground mb-4">
            {total} result{total !== 1 ? "s" : ""} found
          </p>

          {results.length > 0 ? (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {results.map((item) => (
                <button
                  key={item.id}
                  onClick={() => navigate(`/media/${item.id}`)}
                  className="text-left bg-card border border-border rounded-lg p-4 hover:shadow-md transition-shadow"
                >
                  <div className="flex gap-3">
                    {item.cover_image ? (
                      <img
                        src={item.cover_image}
                        alt={item.title}
                        className="w-16 h-20 object-cover rounded"
                      />
                    ) : (
                      <div className="w-16 h-20 bg-muted rounded flex items-center justify-center">
                        <Film className="h-6 w-6 text-muted-foreground" />
                      </div>
                    )}
                    <div className="flex-1 min-w-0">
                      <h3 className="font-medium text-foreground truncate">
                        {item.title}
                      </h3>
                      <p className="text-sm text-muted-foreground">
                        {item.media_type}
                        {item.year ? ` (${item.year})` : ""}
                      </p>
                      <div className="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
                        {item.rating != null && (
                          <span className="flex items-center gap-0.5">
                            <Star className="h-3 w-3" />
                            {item.rating.toFixed(1)}
                          </span>
                        )}
                        {item.duration != null && (
                          <span className="flex items-center gap-0.5">
                            <Clock className="h-3 w-3" />
                            {formatDuration(item.duration)}
                          </span>
                        )}
                        {item.file_size != null && (
                          <span>{formatSize(item.file_size)}</span>
                        )}
                      </div>
                      {item.quality && (
                        <span className="inline-block mt-1 px-1.5 py-0.5 text-xs bg-secondary text-secondary-foreground rounded">
                          {item.quality}
                        </span>
                      )}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          ) : (
            <div className="text-center py-12 text-muted-foreground">
              No results found for "{query}"
            </div>
          )}
        </div>
      )}

      {/* Empty state */}
      {!isLoading && !hasSearched && (
        <div className="text-center py-12 text-muted-foreground">
          Enter a search term to find media in your library
        </div>
      )}
    </div>
  );
}
