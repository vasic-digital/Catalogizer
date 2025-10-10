package services

import (
	"catalog-api/internal/config"
	"catalog-api/internal/models"
	"database/sql"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type CatalogService struct {
	db     *sql.DB
	config *config.Config
	logger *zap.Logger
}

func NewCatalogService(cfg *config.Config, logger *zap.Logger) *CatalogService {
	return &CatalogService{
		config: cfg,
		logger: logger,
	}
}

func (s *CatalogService) SetDB(db *sql.DB) {
	s.db = db
}

func (s *CatalogService) ListPath(path string, sortBy string, sortOrder string, limit, offset int) ([]models.FileInfo, error) {
	query := `
		SELECT id, name, path, is_directory, size, last_modified, hash, extension, mime_type, parent_id, smb_root, created_at, updated_at
		FROM files
		WHERE parent_id = (SELECT id FROM files WHERE path = ? LIMIT 1)
	`

	args := []interface{}{path}

	// Add sorting
	switch sortBy {
	case "name":
		query += " ORDER BY name"
	case "size":
		query += " ORDER BY size"
	case "modified":
		query += " ORDER BY last_modified"
	default:
		query += " ORDER BY is_directory DESC, name"
	}

	if sortOrder == "desc" {
		query += " DESC"
	} else {
		query += " ASC"
	}

	// Add pagination
	query += " LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []models.FileInfo
	for rows.Next() {
		var file models.FileInfo
		err := rows.Scan(
			&file.ID, &file.Name, &file.Path, &file.IsDirectory, &file.Size,
			&file.LastModified, &file.Hash, &file.Extension, &file.MimeType,
			&file.ParentID, &file.SmbRoot, &file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *CatalogService) GetFileInfo(id int64) (*models.FileInfo, error) {
	query := `
		SELECT id, name, path, is_directory, size, last_modified, hash, extension, mime_type, parent_id, smb_root, created_at, updated_at
		FROM files
		WHERE id = ?
	`

	var file models.FileInfo
	err := s.db.QueryRow(query, id).Scan(
		&file.ID, &file.Name, &file.Path, &file.IsDirectory, &file.Size,
		&file.LastModified, &file.Hash, &file.Extension, &file.MimeType,
		&file.ParentID, &file.SmbRoot, &file.CreatedAt, &file.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	return &file, nil
}

func (s *CatalogService) SearchFiles(req *models.SearchRequest) ([]models.FileInfo, int64, error) {
	baseQuery := `
		SELECT id, name, path, is_directory, size, last_modified, hash, extension, mime_type, parent_id, smb_root, created_at, updated_at
		FROM files
		WHERE 1=1
	`

	countQuery := `
		SELECT COUNT(*)
		FROM files
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}

	// Add search conditions
	if req.Query != "" {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+req.Query+"%")
	}

	if req.Path != "" {
		conditions = append(conditions, "path LIKE ?")
		args = append(args, req.Path+"%")
	}

	if req.Extension != "" {
		conditions = append(conditions, "extension = ?")
		args = append(args, req.Extension)
	}

	if req.MimeType != "" {
		conditions = append(conditions, "mime_type = ?")
		args = append(args, req.MimeType)
	}

	if req.MinSize != nil {
		conditions = append(conditions, "size >= ?")
		args = append(args, *req.MinSize)
	}

	if req.MaxSize != nil {
		conditions = append(conditions, "size <= ?")
		args = append(args, *req.MaxSize)
	}

	if len(req.SmbRoots) > 0 {
		placeholders := strings.Repeat("?,", len(req.SmbRoots))
		placeholders = placeholders[:len(placeholders)-1]
		conditions = append(conditions, fmt.Sprintf("smb_root IN (%s)", placeholders))
		for _, root := range req.SmbRoots {
			args = append(args, root)
		}
	}

	if req.IsDirectory != nil {
		conditions = append(conditions, "is_directory = ?")
		args = append(args, *req.IsDirectory)
	}

	// Build final queries
	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " AND " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int64
	err := s.db.QueryRow(countQuery+whereClause, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count results: %w", err)
	}

	// Add sorting and pagination to main query
	query := baseQuery + whereClause

	// Add sorting
	switch req.SortBy {
	case "name":
		query += " ORDER BY name"
	case "size":
		query += " ORDER BY size"
	case "modified":
		query += " ORDER BY last_modified"
	default:
		query += " ORDER BY is_directory DESC, name"
	}

	if req.SortOrder == "desc" {
		query += " DESC"
	} else {
		query += " ASC"
	}

	// Add pagination
	if req.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, req.Limit)
	}
	if req.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, req.Offset)
	}

	// Execute main query
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search files: %w", err)
	}
	defer rows.Close()

	var files []models.FileInfo
	for rows.Next() {
		var file models.FileInfo
		err := rows.Scan(
			&file.ID, &file.Name, &file.Path, &file.IsDirectory, &file.Size,
			&file.LastModified, &file.Hash, &file.Extension, &file.MimeType,
			&file.ParentID, &file.SmbRoot, &file.CreatedAt, &file.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file: %w", err)
		}
		files = append(files, file)
	}

	return files, total, nil
}

func (s *CatalogService) GetDirectoriesBySize(smbRoot string, limit int) ([]models.DirectoryStats, error) {
	query := `
		WITH RECURSIVE dir_sizes AS (
			SELECT
				id, path, name, is_directory,
				CASE WHEN is_directory THEN 0 ELSE size END as file_size,
				CASE WHEN is_directory THEN 0 ELSE 1 END as file_count
			FROM files
			WHERE smb_root = ? AND is_directory = true

			UNION ALL

			SELECT
				f.id, f.path, f.name, f.is_directory,
				CASE WHEN f.is_directory THEN 0 ELSE f.size END,
				CASE WHEN f.is_directory THEN 0 ELSE 1 END
			FROM files f
			JOIN dir_sizes ds ON f.parent_id = ds.id
			WHERE f.smb_root = ?
		)
		SELECT
			path,
			SUM(file_size) as total_size,
			SUM(file_count) as file_count,
			COUNT(CASE WHEN is_directory THEN 1 END) as directory_count
		FROM dir_sizes
		WHERE is_directory = true
		GROUP BY path
		ORDER BY total_size DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, smbRoot, smbRoot, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get directories by size: %w", err)
	}
	defer rows.Close()

	var stats []models.DirectoryStats
	for rows.Next() {
		var stat models.DirectoryStats
		err := rows.Scan(&stat.Path, &stat.TotalSize, &stat.FileCount, &stat.DirectoryCount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan directory stats: %w", err)
		}
		stats = append(stats, stat)
	}

	return stats, nil
}

func (s *CatalogService) GetDuplicateGroups(smbRoot string, minCount int, limit int) ([]models.DuplicateGroup, error) {
	query := `
		SELECT
			hash, size, COUNT(*) as count
		FROM files
		WHERE hash IS NOT NULL
			AND is_directory = false
			AND smb_root = ?
		GROUP BY hash, size
		HAVING COUNT(*) >= ?
		ORDER BY COUNT(*) DESC, size DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, smbRoot, minCount, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get duplicate groups: %w", err)
	}
	defer rows.Close()

	var groups []models.DuplicateGroup
	for rows.Next() {
		var group models.DuplicateGroup
		err := rows.Scan(&group.Hash, &group.Size, &group.Count)
		if err != nil {
			return nil, fmt.Errorf("failed to scan duplicate group: %w", err)
		}

		// Get files in this duplicate group
		filesQuery := `
			SELECT id, name, path, is_directory, size, last_modified, hash, extension, mime_type, parent_id, smb_root, created_at, updated_at
			FROM files
			WHERE hash = ? AND size = ? AND smb_root = ?
			ORDER BY path
		`

		fileRows, err := s.db.Query(filesQuery, group.Hash, group.Size, smbRoot)
		if err != nil {
			s.logger.Error("Failed to get files for duplicate group", zap.Error(err))
			continue
		}

		for fileRows.Next() {
			var file models.FileInfo
			err := fileRows.Scan(
				&file.ID, &file.Name, &file.Path, &file.IsDirectory, &file.Size,
				&file.LastModified, &file.Hash, &file.Extension, &file.MimeType,
				&file.ParentID, &file.SmbRoot, &file.CreatedAt, &file.UpdatedAt,
			)
			if err != nil {
				s.logger.Error("Failed to scan duplicate file", zap.Error(err))
				continue
			}
			group.Files = append(group.Files, file)
		}
		fileRows.Close()

		group.TotalSize = group.Size * int64(group.Count)
		groups = append(groups, group)
	}

	return groups, nil
}

func (s *CatalogService) GetSMBRoots() ([]string, error) {
	query := `SELECT DISTINCT smb_root FROM files ORDER BY smb_root`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMB roots: %w", err)
	}
	defer rows.Close()

	var roots []string
	for rows.Next() {
		var root string
		if err := rows.Scan(&root); err != nil {
			return nil, fmt.Errorf("failed to scan SMB root: %w", err)
		}
		roots = append(roots, root)
	}

	return roots, nil
}