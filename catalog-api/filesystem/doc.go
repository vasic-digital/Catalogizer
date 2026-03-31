// Package filesystem provides a unified multi-protocol filesystem abstraction for accessing
// remote and local storage.
//
// It defines the UnifiedClient interface and a factory that creates per-protocol clients
// for SMB, FTP, NFS, WebDAV, and local filesystems. New protocols are supported by
// implementing the UnifiedClient interface and registering them with the factory. This
// package enables the catalog-api to scan, browse, and stream media files across diverse
// storage backends through a single consistent API.
package filesystem
