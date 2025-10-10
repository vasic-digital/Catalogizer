import { useState } from "react";
import { Link } from "react-router-dom";
import { Grid, List, Search, Play, Star } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { MediaItem, MediaSearchRequest, MediaType, SortOption, SortOrder } from "../types";
import { apiService } from "../services/apiService";
import { cn } from "../utils/cn";

export default function LibraryPage() {
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedType, setSelectedType] = useState<MediaType | "">("");
  const [sortBy, setSortBy] = useState<SortOption>("updated_at");
  const [sortOrder, setSortOrder] = useState<SortOrder>("desc");

  const searchRequest: MediaSearchRequest = {
    query: searchQuery || undefined,
    media_type: selectedType || undefined,
    sort_by: sortBy,
    sort_order: sortOrder,
    limit: 50,
  };

  const { data: mediaResponse, isLoading } = useQuery({
    queryKey: ["library", searchRequest],
    queryFn: () => apiService.searchMedia(searchRequest),
  });

  const mediaTypes: Array<{ value: MediaType | ""; label: string }> = [
    { value: "", label: "All Types" },
    { value: "movie", label: "Movies" },
    { value: "tv_show", label: "TV Shows" },
    { value: "music", label: "Music" },
    { value: "documentary", label: "Documentaries" },
    { value: "anime", label: "Anime" },
  ];

  const sortOptions: Array<{ value: SortOption; label: string }> = [
    { value: "updated_at", label: "Recently Updated" },
    { value: "created_at", label: "Recently Added" },
    { value: "title", label: "Title" },
    { value: "year", label: "Year" },
    { value: "rating", label: "Rating" },
    { value: "file_size", label: "File Size" },
  ];

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-foreground mb-2">Library</h1>
        <p className="text-muted-foreground">
          Browse and manage your media collection
        </p>
      </div>

      {/* Filters and Controls */}
      <div className="bg-card border border-border rounded-lg p-4">
        <div className="flex flex-col lg:flex-row gap-4">
          {/* Search */}
          <div className="flex-1">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
              <input
                type="text"
                placeholder="Search your library..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-10 pr-4 py-2 border border-input bg-background rounded-md text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
              />
            </div>
          </div>

          {/* Type Filter */}
          <select
            value={selectedType}
            onChange={(e) => setSelectedType(e.target.value as MediaType | "")}
            className="px-3 py-2 border border-input bg-background rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            {mediaTypes.map((type) => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </select>

          {/* Sort */}
          <select
            value={sortBy}
            onChange={(e) => setSortBy(e.target.value as SortOption)}
            className="px-3 py-2 border border-input bg-background rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            {sortOptions.map((option) => (
              <option key={option.value} value={option.value}>
                {option.label}
              </option>
            ))}
          </select>

          <select
            value={sortOrder}
            onChange={(e) => setSortOrder(e.target.value as SortOrder)}
            className="px-3 py-2 border border-input bg-background rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent"
          >
            <option value="desc">Descending</option>
            <option value="asc">Ascending</option>
          </select>

          {/* View Toggle */}
          <div className="flex items-center border border-input rounded-md">
            <button
              onClick={() => setViewMode("grid")}
              className={cn(
                "p-2 rounded-l-md transition-colors",
                viewMode === "grid"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <Grid className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode("list")}
              className={cn(
                "p-2 rounded-r-md transition-colors",
                viewMode === "list"
                  ? "bg-primary text-primary-foreground"
                  : "text-muted-foreground hover:text-foreground"
              )}
            >
              <List className="h-4 w-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Results */}
      {isLoading ? (
        <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
          {Array.from({ length: 12 }).map((_, i) => (
            <div key={i} className="aspect-[2/3] bg-muted rounded-lg animate-pulse" />
          ))}
        </div>
      ) : (
        <div className="space-y-4">
          {mediaResponse?.items && mediaResponse.items.length > 0 ? (
            <>
              <div className="text-sm text-muted-foreground">
                Showing {mediaResponse.items.length} of {mediaResponse.total} items
              </div>

              {viewMode === "grid" ? (
                <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
                  {mediaResponse.items.map((item) => (
                    <MediaGridCard key={item.id} item={item} />
                  ))}
                </div>
              ) : (
                <div className="space-y-2">
                  {mediaResponse.items.map((item) => (
                    <MediaListItem key={item.id} item={item} />
                  ))}
                </div>
              )}
            </>
          ) : (
            <div className="text-center py-12">
              <div className="text-muted-foreground mb-2">No media found</div>
              <div className="text-sm text-muted-foreground">
                Try adjusting your search or filters
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

function MediaGridCard({ item }: { item: MediaItem }) {
  const posterUrl = item.external_metadata?.[0]?.poster_url || item.cover_image;

  return (
    <Link
      to={`/media/${item.id}`}
      className="group block bg-card border border-border rounded-lg overflow-hidden hover:border-primary/50 transition-colors"
    >
      <div className="aspect-[2/3] bg-muted relative overflow-hidden">
        {posterUrl ? (
          <img
            src={posterUrl}
            alt={item.title}
            className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-200"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-muted-foreground">
            <Play className="h-8 w-8" />
          </div>
        )}
      </div>

      <div className="p-3">
        <h3 className="font-medium text-foreground truncate group-hover:text-primary transition-colors">
          {item.title}
        </h3>
        <div className="flex items-center gap-2 mt-1">
          {item.year && (
            <span className="text-sm text-muted-foreground">{item.year}</span>
          )}
          {item.rating && (
            <div className="flex items-center gap-1">
              <Star className="h-3 w-3 text-yellow-500 fill-current" />
              <span className="text-sm text-muted-foreground">
                {item.rating.toFixed(1)}
              </span>
            </div>
          )}
        </div>
      </div>
    </Link>
  );
}

function MediaListItem({ item }: { item: MediaItem }) {
  const posterUrl = item.external_metadata?.[0]?.poster_url || item.cover_image;

  return (
    <Link
      to={`/media/${item.id}`}
      className="flex items-center gap-4 p-4 bg-card border border-border rounded-lg hover:border-primary/50 transition-colors group"
    >
      <div className="w-16 h-24 bg-muted rounded overflow-hidden flex-shrink-0">
        {posterUrl ? (
          <img
            src={posterUrl}
            alt={item.title}
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-muted-foreground">
            <Play className="h-6 w-6" />
          </div>
        )}
      </div>

      <div className="flex-1 min-w-0">
        <h3 className="font-medium text-foreground truncate group-hover:text-primary transition-colors">
          {item.title}
        </h3>
        <div className="flex items-center gap-4 mt-1 text-sm text-muted-foreground">
          {item.year && <span>{item.year}</span>}
          <span className="capitalize">{item.media_type.replace("_", " ")}</span>
          {item.rating && (
            <div className="flex items-center gap-1">
              <Star className="h-3 w-3 text-yellow-500 fill-current" />
              <span>{item.rating.toFixed(1)}</span>
            </div>
          )}
        </div>
        {item.description && (
          <p className="text-sm text-muted-foreground mt-1 line-clamp-2">
            {item.description}
          </p>
        )}
      </div>
    </Link>
  );
}