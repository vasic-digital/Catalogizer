import { useParams, useNavigate } from "react-router-dom";
import { ArrowLeft, Play, Download, Heart, Star, Calendar, HardDrive } from "lucide-react";
import { useQuery } from "@tanstack/react-query";
import { apiService } from "../services/apiService";

export default function MediaDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: media, isLoading } = useQuery({
    queryKey: ["media", id],
    queryFn: () => apiService.getMediaById(Number(id)),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div className="p-6">
        <div className="animate-pulse space-y-4">
          <div className="h-8 bg-muted rounded w-1/4"></div>
          <div className="h-64 bg-muted rounded"></div>
          <div className="h-4 bg-muted rounded w-3/4"></div>
          <div className="h-4 bg-muted rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  if (!media) {
    return (
      <div className="p-6 text-center">
        <div className="text-muted-foreground">Media not found</div>
      </div>
    );
  }

  const posterUrl = media.external_metadata?.[0]?.poster_url || media.cover_image;

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <div className="p-6 border-b border-border">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-muted-foreground hover:text-foreground transition-colors"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </button>
      </div>

      {/* Content */}
      <div className="p-6">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Poster */}
          <div className="lg:col-span-1">
            <div className="aspect-[2/3] bg-muted rounded-lg overflow-hidden">
              {posterUrl ? (
                <img
                  src={posterUrl}
                  alt={media.title}
                  className="w-full h-full object-cover"
                />
              ) : (
                <div className="w-full h-full flex items-center justify-center text-muted-foreground">
                  <Play className="h-16 w-16" />
                </div>
              )}
            </div>
          </div>

          {/* Details */}
          <div className="lg:col-span-2 space-y-6">
            <div>
              <h1 className="text-4xl font-bold text-foreground mb-2">
                {media.title}
              </h1>
              <div className="flex items-center gap-4 text-muted-foreground">
                {media.year && (
                  <div className="flex items-center gap-1">
                    <Calendar className="h-4 w-4" />
                    <span>{media.year}</span>
                  </div>
                )}
                {media.rating && (
                  <div className="flex items-center gap-1">
                    <Star className="h-4 w-4 text-yellow-500 fill-current" />
                    <span>{media.rating.toFixed(1)}</span>
                  </div>
                )}
                <span className="capitalize bg-muted px-2 py-1 rounded text-xs">
                  {media.media_type.replace("_", " ")}
                </span>
              </div>
            </div>

            {/* Description */}
            {media.description && (
              <div>
                <h3 className="font-semibold text-foreground mb-2">Description</h3>
                <p className="text-muted-foreground leading-relaxed">
                  {media.description}
                </p>
              </div>
            )}

            {/* Technical Info */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {media.file_size && (
                <div className="flex items-center gap-2">
                  <HardDrive className="h-4 w-4 text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">
                    File Size: {(media.file_size / (1024 * 1024 * 1024)).toFixed(2)} GB
                  </span>
                </div>
              )}
              {media.quality && (
                <div className="text-sm text-muted-foreground">
                  Quality: {media.quality}
                </div>
              )}
            </div>

            {/* Actions */}
            <div className="flex items-center gap-4">
              <button className="flex items-center gap-2 bg-primary text-primary-foreground px-6 py-3 rounded-lg hover:bg-primary/90 transition-colors">
                <Play className="h-5 w-5" />
                Play
              </button>
              <button className="flex items-center gap-2 bg-secondary text-secondary-foreground px-6 py-3 rounded-lg hover:bg-secondary/80 transition-colors">
                <Download className="h-5 w-5" />
                Download
              </button>
              <button className="flex items-center gap-2 bg-secondary text-secondary-foreground px-4 py-3 rounded-lg hover:bg-secondary/80 transition-colors">
                <Heart className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}