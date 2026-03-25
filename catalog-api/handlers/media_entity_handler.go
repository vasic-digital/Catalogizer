package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"catalogizer/database"
	"catalogizer/internal/media/models"
	"catalogizer/repository"
	"catalogizer/utils"

	"github.com/gin-gonic/gin"
)

// MediaEntityHandler handles entity-level media browsing endpoints.
type MediaEntityHandler struct {
	itemRepo     *repository.MediaItemRepository
	fileRepo     *repository.MediaFileRepository
	extMetaRepo  *repository.ExternalMetadataRepository
	userMetaRepo *repository.UserMetadataRepository
	db           *database.DB
}

// NewMediaEntityHandler creates a new media entity handler.
func NewMediaEntityHandler(
	itemRepo *repository.MediaItemRepository,
	fileRepo *repository.MediaFileRepository,
	extMetaRepo *repository.ExternalMetadataRepository,
	userMetaRepo *repository.UserMetadataRepository,
	dbArgs ...*database.DB,
) *MediaEntityHandler {
	h := &MediaEntityHandler{
		itemRepo:     itemRepo,
		fileRepo:     fileRepo,
		extMetaRepo:  extMetaRepo,
		userMetaRepo: userMetaRepo,
	}
	if len(dbArgs) > 0 && dbArgs[0] != nil {
		h.db = dbArgs[0]
		fmt.Println("[MediaEntityHandler] Database connected for metadata enrichment")
	} else {
		fmt.Println("[MediaEntityHandler] WARNING: No database provided, enrichment disabled")
	}
	return h
}

// ListEntities handles GET /api/v1/entities — list entities with filters and pagination.
func (h *MediaEntityHandler) ListEntities(c *gin.Context) {
	ctx := c.Request.Context()

	query := c.Query("query")
	mediaType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "24")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit <= 0 || limit > 200 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}

	var mediaTypeIDs []int64
	if mediaType != "" {
		_, typeID, err := h.itemRepo.GetMediaTypeByName(ctx, mediaType)
		if err != nil {
			utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media type", err)
			return
		}
		mediaTypeIDs = append(mediaTypeIDs, typeID)
	}

	if query == "" {
		query = "%"
	}

	items, total, err := h.itemRepo.Search(ctx, query, mediaTypeIDs, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list entities", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEntity handles GET /api/v1/entities/:id — get entity with details.
func (h *MediaEntityHandler) GetEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	// Get file count
	fileCount, _ := h.fileRepo.CountByItem(ctx, id)

	// Get children count
	children, _ := h.itemRepo.GetChildren(ctx, id)
	childrenCount := len(children)

	// Get external metadata
	extMeta, _ := h.extMetaRepo.GetByItem(ctx, id)

	// Get media type name
	types, _ := h.itemRepo.GetMediaTypes(ctx)
	typeName := ""
	for _, mt := range types {
		if mt.ID == item.MediaTypeID {
			typeName = mt.Name
			break
		}
	}

	result := entityDetailJSON(item, typeName, fileCount, int64(childrenCount), extMeta)
	c.JSON(http.StatusOK, result)
}

// GetEntityChildren handles GET /api/v1/entities/:id/children.
func (h *MediaEntityHandler) GetEntityChildren(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.itemRepo.GetByParent(ctx, id, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get children", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// GetEntityFiles handles GET /api/v1/entities/:id/files.
func (h *MediaEntityHandler) GetEntityFiles(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get files", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
		"total": len(files),
	})
}

// GetEntityMetadata handles GET /api/v1/entities/:id/metadata.
func (h *MediaEntityHandler) GetEntityMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	metadata, err := h.extMetaRepo.GetByItem(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get metadata", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"metadata": metadata,
	})
}

// GetEntityDuplicates handles GET /api/v1/entities/:id/duplicates.
func (h *MediaEntityHandler) GetEntityDuplicates(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	dups, err := h.itemRepo.GetDuplicates(ctx, item.Title, item.MediaTypeID, item.Year)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to find duplicates", err)
		return
	}

	// Exclude self
	var filtered []*models.MediaItem
	for _, d := range dups {
		if d.ID != id {
			filtered = append(filtered, d)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"duplicates": itemsToJSON(filtered),
		"total":      len(filtered),
	})
}

// GetEntityTypes handles GET /api/v1/entities/types — list media types with counts.
func (h *MediaEntityHandler) GetEntityTypes(c *gin.Context) {
	ctx := c.Request.Context()

	types, err := h.itemRepo.GetMediaTypes(ctx)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to get media types", err)
		return
	}

	counts, _ := h.itemRepo.CountByType(ctx)

	var result []gin.H
	for _, mt := range types {
		count := int64(0)
		if c, ok := counts[mt.Name]; ok {
			count = c
		}
		result = append(result, gin.H{
			"id":          mt.ID,
			"name":        mt.Name,
			"description": mt.Description,
			"count":       count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"types": result,
	})
}

// BrowseByType handles GET /api/v1/entities/browse/:type.
func (h *MediaEntityHandler) BrowseByType(c *gin.Context) {
	ctx := c.Request.Context()

	typeName := c.Param("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "24"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 200 {
		limit = 24
	}
	if offset < 0 {
		offset = 0
	}

	_, typeID, err := h.itemRepo.GetMediaTypeByName(ctx, typeName)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid media type", err)
		return
	}

	items, total, err := h.itemRepo.GetByType(ctx, typeID, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to browse type", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":  itemsToJSON(items),
		"total":  total,
		"type":   typeName,
		"limit":  limit,
		"offset": offset,
	})
}

// RefreshEntityMetadata handles POST /api/v1/entities/:id/metadata/refresh.
// Finds cover images in the entity's directory and stores as external metadata.
func (h *MediaEntityHandler) RefreshEntityMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	if h.db == nil {
		c.JSON(http.StatusAccepted, gin.H{"message": "Metadata refresh queued", "entity_id": id})
		return
	}

	// Find the directory path for this entity via directory_analyses
	var dirPath string
	err = h.db.QueryRowContext(ctx,
		`SELECT da.directory_path FROM directory_analyses da
		 WHERE da.media_item_id = ? LIMIT 1`, id).Scan(&dirPath)
	if err != nil || dirPath == "" {
		// Fallback: try media_files junction
		err = h.db.QueryRowContext(ctx,
			`SELECT f.directory_path FROM files f
			 JOIN media_files mf ON mf.file_id = f.id
			 WHERE mf.media_item_id = ? LIMIT 1`, id).Scan(&dirPath)
	}
	if err != nil || dirPath == "" {
		c.JSON(http.StatusOK, gin.H{"message": "No directory found", "entity_id": id})
		return
	}

	parentDir := dirPath

	// Search for cover images: directory_path stores full path like "Dir/Sub/cover.jpg"
	// So we look for files where the path starts with the entity dir and ends with a cover name
	coverNames := []string{"cover.jpg", "folder.jpg", "poster.jpg", "cover.png", "folder.png", "poster.png"}
	var coverFileID int64
	var coverTitle string
	for _, name := range coverNames {
		err := h.db.QueryRowContext(ctx,
			`SELECT id, title FROM files
			 WHERE directory_path LIKE ? LIMIT 1`,
			parentDir+"/%"+name).Scan(&coverFileID, &coverTitle)
		if err == nil && coverFileID > 0 {
			break
		}
		coverFileID = 0
	}

	if coverFileID > 0 {
		coverURL := fmt.Sprintf("/api/v1/download/file/%d", coverFileID)
		meta := &models.ExternalMetadata{
			MediaItemID: id,
			Provider:    "local_scan",
			ExternalID:  fmt.Sprintf("local:%d", id),
			CoverURL:    &coverURL,
		}
		_ = h.extMetaRepo.Upsert(ctx, meta)
		c.JSON(http.StatusOK, gin.H{
			"message": "Cover art found (local)", "entity_id": id, "cover_url": coverURL,
		})
		return
	}

	// Fallback: try TMDB for movies/TV
	entity, err := h.itemRepo.GetByID(ctx, id)
	if err == nil && entity != nil {
		year := 0
		if entity.Year != nil {
			year = *entity.Year
		}
		tmdbResult := fetchTMDBMetadata(entity.Title, year, int(entity.MediaTypeID))
		if tmdbResult != nil {
			meta := &models.ExternalMetadata{
				MediaItemID: id,
				Provider:    "tmdb",
				ExternalID:  fmt.Sprintf("tmdb:%d", tmdbResult.tmdbID),
				CoverURL:    &tmdbResult.posterURL,
				Data:        tmdbResult.overview,
				Rating:      tmdbResult.rating,
			}
			_ = h.extMetaRepo.Upsert(ctx, meta)

			// Also update the entity description/rating if empty
			if (entity.Description == nil || *entity.Description == "") && tmdbResult.overview != "" {
				_, _ = h.db.ExecContext(ctx,
					`UPDATE media_items SET description = ? WHERE id = ? AND (description IS NULL OR description = '')`,
					tmdbResult.overview, id)
			}
			if (entity.Rating == nil || *entity.Rating == 0) && tmdbResult.rating != nil {
				_, _ = h.db.ExecContext(ctx,
					`UPDATE media_items SET rating = ? WHERE id = ? AND (rating IS NULL OR rating = 0)`,
					*tmdbResult.rating, id)
			}

			c.JSON(http.StatusOK, gin.H{
				"message": "Metadata enriched from TMDB", "entity_id": id,
				"cover_url": tmdbResult.posterURL, "title": tmdbResult.title,
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "No metadata sources available", "entity_id": id, "directory": parentDir,
	})
}

// tmdbSearchResult holds parsed TMDB API response data.
type tmdbSearchResult struct {
	tmdbID    int
	title     string
	overview  string
	posterURL string
	rating    *float64
	year      string
}

// fetchTMDBMetadata searches TMDB for a movie/TV show and returns poster + metadata.
func fetchTMDBMetadata(title string, year int, mediaTypeID int) *tmdbSearchResult {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return nil
	}

	// Determine search type
	searchType := "movie"
	if mediaTypeID == 2 { // tv_show
		searchType = "tv"
	}

	// Build search URL
	searchURL := fmt.Sprintf("https://api.themoviedb.org/3/search/%s?api_key=%s&query=%s",
		searchType, apiKey, url.QueryEscape(title))
	if year > 0 {
		if searchType == "movie" {
			searchURL += fmt.Sprintf("&year=%d", year)
		} else {
			searchURL += fmt.Sprintf("&first_air_date_year=%d", year)
		}
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(searchURL)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var searchResp struct {
		Results []struct {
			ID          int     `json:"id"`
			Title       string  `json:"title"`
			Name        string  `json:"name"`
			Overview    string  `json:"overview"`
			PosterPath  string  `json:"poster_path"`
			VoteAverage float64 `json:"vote_average"`
			ReleaseDate string  `json:"release_date"`
			FirstAirDate string `json:"first_air_date"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &searchResp); err != nil || len(searchResp.Results) == 0 {
		return nil
	}

	best := searchResp.Results[0]
	resultTitle := best.Title
	if resultTitle == "" {
		resultTitle = best.Name
	}

	var posterURL string
	if best.PosterPath != "" {
		posterURL = "https://image.tmdb.org/t/p/w500" + best.PosterPath
	}

	rating := best.VoteAverage
	return &tmdbSearchResult{
		tmdbID:    best.ID,
		title:     resultTitle,
		overview:  best.Overview,
		posterURL: posterURL,
		rating:    &rating,
		year:      best.ReleaseDate,
	}
}

// EnrichAllEntities handles POST /api/v1/entities/enrich — batch metadata enrichment.
// Scans all entities for local cover art and stores results.
func (h *MediaEntityHandler) EnrichAllEntities(c *gin.Context) {
	ctx := c.Request.Context()

	if h.db == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database not available"})
		return
	}

	// Get all entities that don't have external metadata yet
	rows, err := h.db.QueryContext(ctx,
		`SELECT mi.id, da.directory_path
		 FROM media_items mi
		 JOIN directory_analyses da ON da.media_item_id = mi.id
		 LEFT JOIN external_metadata em ON em.media_item_id = mi.id
		 WHERE em.id IS NULL
		 GROUP BY mi.id
		 LIMIT 500`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("query: %v", err)})
		return
	}
	defer rows.Close()

	type entityDir struct {
		ID      int64
		DirPath string
	}
	var entities []entityDir
	for rows.Next() {
		var ed entityDir
		if err := rows.Scan(&ed.ID, &ed.DirPath); err != nil {
			continue
		}
		entities = append(entities, ed)
	}

	enriched := 0
	tmdbEnriched := 0
	localEnriched := 0
	coverNames := []string{"cover.jpg", "folder.jpg", "poster.jpg", "cover.png"}

	for _, ent := range entities {
		// Try local cover first
		parentDir := ent.DirPath
		found := false
		for _, name := range coverNames {
			var coverFileID int64
			err := h.db.QueryRowContext(ctx,
				`SELECT id FROM files WHERE directory_path LIKE ? LIMIT 1`,
				parentDir+"/%"+name).Scan(&coverFileID)
			if err == nil && coverFileID > 0 {
				coverURL := fmt.Sprintf("/api/v1/download/file/%d", coverFileID)
				meta := &models.ExternalMetadata{
					MediaItemID: ent.ID,
					Provider:    "local_scan",
					ExternalID:  fmt.Sprintf("local:%d", ent.ID),
					CoverURL:    &coverURL,
				}
				_ = h.extMetaRepo.Upsert(ctx, meta)
				localEnriched++
				enriched++
				found = true
				break
			}
		}

		// Try TMDB if no local cover
		if !found {
			// Get entity details for TMDB search
			var title string
			var year int
			var mediaTypeID int
			err := h.db.QueryRowContext(ctx,
				`SELECT title, COALESCE(year, 0), media_type_id FROM media_items WHERE id = ?`,
				ent.ID).Scan(&title, &year, &mediaTypeID)
			if err == nil && title != "" {
				result := fetchTMDBMetadata(title, year, mediaTypeID)
				if result != nil && result.posterURL != "" {
					meta := &models.ExternalMetadata{
						MediaItemID: ent.ID,
						Provider:    "tmdb",
						ExternalID:  fmt.Sprintf("tmdb:%d", result.tmdbID),
						CoverURL:    &result.posterURL,
						Data:        result.overview,
						Rating:      result.rating,
					}
					_ = h.extMetaRepo.Upsert(ctx, meta)
					// Update entity description
					if result.overview != "" {
						_, _ = h.db.ExecContext(ctx,
							`UPDATE media_items SET description = ? WHERE id = ? AND (description IS NULL OR description = '')`,
							result.overview, ent.ID)
					}
					if result.rating != nil && *result.rating > 0 {
						_, _ = h.db.ExecContext(ctx,
							`UPDATE media_items SET rating = ? WHERE id = ? AND (rating IS NULL OR rating = 0)`,
							*result.rating, ent.ID)
					}
					tmdbEnriched++
					enriched++
					// Rate limit TMDB API (40 requests per 10 seconds)
					time.Sleep(250 * time.Millisecond)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":        "Batch enrichment complete",
		"scanned":        len(entities),
		"enriched":       enriched,
		"local_covers":   localEnriched,
		"tmdb_enriched":  tmdbEnriched,
	})
}

// UpdateUserMetadata handles PUT /api/v1/entities/:id/user-metadata.
func (h *MediaEntityHandler) UpdateUserMetadata(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	var req struct {
		UserRating    *float64 `json:"user_rating"`
		WatchedStatus *string  `json:"watched_status"`
		Favorite      *bool    `json:"favorite"`
		IsFavorite    *bool    `json:"is_favorite"`
		PersonalNotes *string  `json:"personal_notes"`
		Tags          []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Extract user ID from JWT context (default to 1 for now)
	userID := int64(1)
	if uid, exists := c.Get("user_id"); exists {
		if uidInt, ok := uid.(int64); ok {
			userID = uidInt
		}
	}

	um := &models.UserMetadata{
		MediaItemID:   id,
		UserID:        userID,
		UserRating:    req.UserRating,
		WatchedStatus: req.WatchedStatus,
		PersonalNotes: req.PersonalNotes,
		Tags:          req.Tags,
	}
	// Determine favorite value from either field
	if req.IsFavorite != nil {
		um.Favorite = *req.IsFavorite
	} else if req.Favorite != nil {
		um.Favorite = *req.Favorite
	}

	if err := h.userMetaRepo.Upsert(ctx, um); err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to update user metadata", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "User metadata updated",
		"is_favorite": um.Favorite,
	})
}

// GetEntityStats handles GET /api/v1/entities/stats.
func (h *MediaEntityHandler) GetEntityStats(c *gin.Context) {
	ctx := c.Request.Context()

	totalCount, _ := h.itemRepo.Count(ctx)
	countByType, _ := h.itemRepo.CountByType(ctx)

	c.JSON(http.StatusOK, gin.H{
		"total_entities": totalCount,
		"by_type":        countByType,
	})
}

// StreamEntity handles GET /api/v1/entities/:id/stream — returns streaming info for primary file.
func (h *MediaEntityHandler) StreamEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available for streaming", err)
		return
	}

	// Find primary file or use first one
	primary := files[0]
	for _, f := range files {
		if f.IsPrimary {
			primary = f
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":  id,
		"file_id":    primary.FileID,
		"stream_url": fmt.Sprintf("/api/v1/download/file/%d", primary.FileID),
	})
}

// DownloadEntity handles GET /api/v1/entities/:id/download — returns download info.
func (h *MediaEntityHandler) DownloadEntity(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	fileIDStr := c.Query("file_id")

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available for download", err)
		return
	}

	target := files[0]
	if fileIDStr != "" {
		fileID, _ := strconv.ParseInt(fileIDStr, 10, 64)
		for _, f := range files {
			if f.FileID == fileID {
				target = f
				break
			}
		}
	} else {
		for _, f := range files {
			if f.IsPrimary {
				target = f
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":    id,
		"file_id":      target.FileID,
		"download_url": fmt.Sprintf("/api/v1/download/file/%d", target.FileID),
		"total_files":  len(files),
	})
}

// GetInstallInfo handles GET /api/v1/entities/:id/install-info — software installation details.
func (h *MediaEntityHandler) GetInstallInfo(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Invalid entity ID", err)
		return
	}

	item, err := h.itemRepo.GetByID(ctx, id)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusNotFound, "Entity not found", err)
		return
	}

	// Verify this is a software entity
	types, _ := h.itemRepo.GetMediaTypes(ctx)
	typeName := ""
	for _, mt := range types {
		if mt.ID == item.MediaTypeID {
			typeName = mt.Name
			break
		}
	}
	if typeName != "software" {
		utils.SendErrorResponse(c, http.StatusBadRequest, "Install info is only available for software entities", nil)
		return
	}

	files, err := h.fileRepo.GetFilesByItem(ctx, id)
	if err != nil || len(files) == 0 {
		utils.SendErrorResponse(c, http.StatusNotFound, "No files available", err)
		return
	}

	primary := files[0]
	for _, f := range files {
		if f.IsPrimary {
			primary = f
			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"entity_id":    id,
		"title":        item.Title,
		"file_id":      primary.FileID,
		"download_url": fmt.Sprintf("/api/v1/entities/%d/download", id),
		"total_files":  len(files),
	})
}

// ListDuplicateGroups handles GET /api/v1/entities/duplicates — global duplicate listing.
func (h *MediaEntityHandler) ListDuplicateGroups(c *gin.Context) {
	ctx := c.Request.Context()

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	groups, total, err := h.itemRepo.ListDuplicateGroups(ctx, limit, offset)
	if err != nil {
		utils.SendErrorResponse(c, http.StatusInternalServerError, "Failed to list duplicates", err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"groups": groups,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// --- JSON helpers ---

func itemsToJSON(items []*models.MediaItem) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	result := make([]gin.H, 0, len(items))
	for _, item := range items {
		result = append(result, itemToJSON(item))
	}
	return result
}

func itemToJSON(item *models.MediaItem) gin.H {
	h := gin.H{
		"id":             item.ID,
		"media_type_id":  item.MediaTypeID,
		"title":          item.Title,
		"status":         item.Status,
		"first_detected": item.FirstDetected,
		"last_updated":   item.LastUpdated,
	}
	if item.OriginalTitle != nil {
		h["original_title"] = *item.OriginalTitle
	}
	if item.Year != nil {
		h["year"] = *item.Year
	}
	if item.Description != nil {
		h["description"] = *item.Description
	}
	if len(item.Genre) > 0 {
		h["genre"] = item.Genre
	}
	if item.Director != nil {
		h["director"] = *item.Director
	}
	if item.Rating != nil {
		h["rating"] = *item.Rating
	}
	if item.Runtime != nil {
		h["runtime"] = *item.Runtime
	}
	if item.Language != nil {
		h["language"] = *item.Language
	}
	if item.ParentID != nil {
		h["parent_id"] = *item.ParentID
	}
	if item.SeasonNumber != nil {
		h["season_number"] = *item.SeasonNumber
	}
	if item.EpisodeNumber != nil {
		h["episode_number"] = *item.EpisodeNumber
	}
	if item.TrackNumber != nil {
		h["track_number"] = *item.TrackNumber
	}
	return h
}

func entityDetailJSON(item *models.MediaItem, typeName string, fileCount, childrenCount int64, extMeta []*models.ExternalMetadata) gin.H {
	h := itemToJSON(item)
	h["media_type"] = typeName
	h["file_count"] = fileCount
	h["children_count"] = childrenCount
	if extMeta != nil {
		h["external_metadata"] = extMeta
	} else {
		h["external_metadata"] = []interface{}{}
	}
	return h
}

func strPtr(s string) *string {
	return &s
}
