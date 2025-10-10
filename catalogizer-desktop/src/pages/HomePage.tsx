import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { Play, Clock, Star, TrendingUp, Calendar } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { MediaItem } from "../types";
import { apiService } from "../services/apiService";

export default function HomePage() {
  const [greeting, setGreeting] = useState("");

  // Get stats
  const { data: stats } = useQuery({
    queryKey: ["media-stats"],
    queryFn: () => apiService.getMediaStats(),
  });

  // Get recently added items
  const { data: recentItems } = useQuery({
    queryKey: ["recent-items"],
    queryFn: () =>
      apiService.searchMedia({
        sort_by: "created_at",
        sort_order: "desc",
        limit: 12,
      }),
  });

  // Get continue watching
  const { data: continueWatching } = useQuery({
    queryKey: ["continue-watching"],
    queryFn: () =>
      apiService.searchMedia({
        sort_by: "last_watched",
        sort_order: "desc",
        limit: 8,
      }),
  });

  // Get trending/popular items
  const { data: trending } = useQuery({
    queryKey: ["trending"],
    queryFn: () =>
      apiService.searchMedia({
        sort_by: "rating",
        sort_order: "desc",
        limit: 8,
      }),
  });

  useEffect(() => {
    const hour = new Date().getHours();
    if (hour < 12) setGreeting("Good morning");
    else if (hour < 18) setGreeting("Good afternoon");
    else setGreeting("Good evening");
  }, []);

  return (
    <div className="p-6 space-y-8">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-foreground mb-2">
          {greeting}! Welcome back
        </h1>
        <p className="text-muted-foreground">
          Here's what's new in your media library
        </p>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <StatCard
            icon={<TrendingUp className="h-5 w-5" />}
            title="Total Items"
            value={stats.total_items.toLocaleString()}
            description="Media files"
          />
          <StatCard
            icon={<Calendar className="h-5 w-5" />}
            title="Recent Additions"
            value={stats.recent_additions.toString()}
            description="This week"
          />
          <StatCard
            icon={<Star className="h-5 w-5" />}
            title="Movies"
            value={(stats.by_type.movie || 0).toString()}
            description="In collection"
          />
          <StatCard
            icon={<Play className="h-5 w-5" />}
            title="TV Shows"
            value={(stats.by_type.tv_show || 0).toString()}
            description="Series available"
          />
        </div>
      )}

      {/* Continue Watching */}
      {continueWatching?.items && continueWatching.items.length > 0 && (
        <MediaSection
          title="Continue Watching"
          items={continueWatching.items.filter(item =>
            item.watch_progress && item.watch_progress > 0 && item.watch_progress < 0.9
          )}
          showProgress
        />
      )}

      {/* Recently Added */}
      {recentItems?.items && recentItems.items.length > 0 && (
        <MediaSection
          title="Recently Added"
          items={recentItems.items}
        />
      )}

      {/* Trending */}
      {trending?.items && trending.items.length > 0 && (
        <MediaSection
          title="Highly Rated"
          items={trending.items}
        />
      )}
    </div>
  );
}

interface StatCardProps {
  icon: React.ReactNode;
  title: string;
  value: string;
  description: string;
}

function StatCard({ icon, title, value, description }: StatCardProps) {
  return (
    <div className="bg-card border border-border rounded-lg p-4">
      <div className="flex items-center gap-2 mb-2">
        <div className="text-primary">{icon}</div>
        <h3 className="font-medium text-foreground">{title}</h3>
      </div>
      <div className="text-2xl font-bold text-foreground mb-1">{value}</div>
      <p className="text-sm text-muted-foreground">{description}</p>
    </div>
  );
}

interface MediaSectionProps {
  title: string;
  items: MediaItem[];
  showProgress?: boolean;
}

function MediaSection({ title, items, showProgress }: MediaSectionProps) {
  if (items.length === 0) return null;

  return (
    <div>
      <h2 className="text-xl font-semibold text-foreground mb-4">{title}</h2>
      <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
        {items.map((item) => (
          <MediaCard key={item.id} item={item} showProgress={showProgress} />
        ))}
      </div>
    </div>
  );
}

interface MediaCardProps {
  item: MediaItem;
  showProgress?: boolean;
}

function MediaCard({ item, showProgress }: MediaCardProps) {
  const posterUrl = item.external_metadata?.[0]?.poster_url || item.cover_image;
  const progress = item.watch_progress || 0;

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

        {showProgress && progress > 0 && (
          <div className="absolute bottom-0 left-0 right-0 bg-black/60 p-2">
            <div className="w-full bg-muted h-1 rounded-full overflow-hidden">
              <div
                className="h-full bg-primary rounded-full transition-all"
                style={{ width: `${progress * 100}%` }}
              />
            </div>
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
        {showProgress && progress > 0 && (
          <div className="flex items-center gap-1 mt-1">
            <Clock className="h-3 w-3 text-muted-foreground" />
            <span className="text-xs text-muted-foreground">
              {Math.round(progress * 100)}% watched
            </span>
          </div>
        )}
      </div>
    </Link>
  );
}