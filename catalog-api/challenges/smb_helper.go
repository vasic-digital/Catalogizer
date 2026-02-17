package challenges

import (
	"catalogizer/filesystem"
	"context"
	"fmt"
	"path/filepath"
	"strings"
)

// newSMBClient creates and connects an SMB client from endpoint config.
func newSMBClient(ctx context.Context, ep *Endpoint) (*filesystem.SmbClient, error) {
	cfg := &filesystem.SmbConfig{
		Host:     ep.Host,
		Port:     ep.Port,
		Share:    ep.Share,
		Username: ep.Username,
		Password: ep.Password,
		Domain:   ep.Domain,
	}
	client := filesystem.NewSmbClient(cfg)
	if err := client.Connect(ctx); err != nil {
		return nil, fmt.Errorf("connect to %s:%d/%s: %w", ep.Host, ep.Port, ep.Share, err)
	}
	return client, nil
}

// walkResult holds aggregated file scan results.
type walkResult struct {
	FileCount       int
	DirCount        int
	TotalSize       int64
	ExtensionsFound map[string]int
}

// walkDirectory walks an SMB directory up to maxDepth levels,
// counting files that match the given extensions set.
func walkDirectory(
	ctx context.Context,
	client *filesystem.SmbClient,
	basePath string,
	extensions map[string]bool,
	maxDepth int,
) (*walkResult, error) {
	result := &walkResult{
		ExtensionsFound: make(map[string]int),
	}
	return result, walkRecursive(ctx, client, basePath, extensions, maxDepth, 0, result)
}

func walkRecursive(
	ctx context.Context,
	client *filesystem.SmbClient,
	path string,
	extensions map[string]bool,
	maxDepth, currentDepth int,
	result *walkResult,
) error {
	if currentDepth > maxDepth {
		return nil
	}

	entries, err := client.ListDirectory(ctx, path)
	if err != nil {
		return fmt.Errorf("list directory %s: %w", path, err)
	}

	for _, entry := range entries {
		if entry.IsDir {
			result.DirCount++
			if currentDepth < maxDepth {
				childPath := path + "/" + entry.Name
				if err := walkRecursive(ctx, client, childPath, extensions, maxDepth, currentDepth+1, result); err != nil {
					return err
				}
			}
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name))
			if len(extensions) == 0 || extensions[ext] {
				result.FileCount++
				result.TotalSize += entry.Size
				result.ExtensionsFound[ext]++
			}
		}
	}

	return nil
}

// extensionSet converts a slice of extensions into a set for fast lookup.
func extensionSet(exts []string) map[string]bool {
	m := make(map[string]bool, len(exts))
	for _, e := range exts {
		m[e] = true
	}
	return m
}
