package models

import (
	"time"
)

// File represents a file record from the catalog database
type File struct {
	ID               int64      `json:"id" db:"id"`
	StorageRootID    int64      `json:"storage_root_id" db:"storage_root_id"`
	StorageRootName  string     `json:"storage_root_name" db:"storage_root_name"`
	Path             string     `json:"path" db:"path"`
	Name             string     `json:"name" db:"name"`
	Extension        *string    `json:"extension" db:"extension"`
	MimeType         *string    `json:"mime_type" db:"mime_type"`
	FileType         *string    `json:"file_type" db:"file_type"`
	Size             int64      `json:"size" db:"size"`
	IsDirectory      bool       `json:"is_directory" db:"is_directory"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	ModifiedAt       time.Time  `json:"modified_at" db:"modified_at"`
	AccessedAt       *time.Time `json:"accessed_at" db:"accessed_at"`
	Deleted          bool       `json:"deleted" db:"deleted"`
	DeletedAt        *time.Time `json:"deleted_at" db:"deleted_at"`
	LastScanAt       time.Time  `json:"last_scan_at" db:"last_scan_at"`
	LastVerifiedAt   *time.Time `json:"last_verified_at" db:"last_verified_at"`
	MD5              *string    `json:"md5" db:"md5"`
	SHA256           *string    `json:"sha256" db:"sha256"`
	SHA1             *string    `json:"sha1" db:"sha1"`
	BLAKE3           *string    `json:"blake3" db:"blake3"`
	QuickHash        *string    `json:"quick_hash" db:"quick_hash"`
	IsDuplicate      bool       `json:"is_duplicate" db:"is_duplicate"`
	DuplicateGroupID *int64     `json:"duplicate_group_id" db:"duplicate_group_id"`
	ParentID         *int64     `json:"parent_id" db:"parent_id"`
}

// StorageRoot represents a storage root configuration for any protocol
type StorageRoot struct {
	ID                       int64      `json:"id" db:"id"`
	Name                     string     `json:"name" db:"name"`
	Protocol                 string     `json:"protocol" db:"protocol"` // smb, ftp, nfs, webdav, local
	Host                     *string    `json:"host,omitempty" db:"host"`
	Port                     *int       `json:"port,omitempty" db:"port"`
	Path                     *string    `json:"path,omitempty" db:"path"` // share for SMB, path for FTP/NFS/WebDAV, base_path for local
	Username                 *string    `json:"username,omitempty" db:"username"`
	Password                 *string    `json:"password,omitempty" db:"password"`
	Domain                   *string    `json:"domain,omitempty" db:"domain"`           // SMB specific
	MountPoint               *string    `json:"mount_point,omitempty" db:"mount_point"` // NFS specific
	Options                  *string    `json:"options,omitempty" db:"options"`         // NFS/WebDAV specific
	URL                      *string    `json:"url,omitempty" db:"url"`                 // WebDAV specific
	Enabled                  bool       `json:"enabled" db:"enabled"`
	MaxDepth                 int        `json:"max_depth" db:"max_depth"`
	EnableDuplicateDetection bool       `json:"enable_duplicate_detection" db:"enable_duplicate_detection"`
	EnableMetadataExtraction bool       `json:"enable_metadata_extraction" db:"enable_metadata_extraction"`
	IncludePatterns          *string    `json:"include_patterns" db:"include_patterns"`
	ExcludePatterns          *string    `json:"exclude_patterns" db:"exclude_patterns"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
	LastScanAt               *time.Time `json:"last_scan_at" db:"last_scan_at"`
}

// SmbRoot represents an SMB root configuration (deprecated - use StorageRoot)
type SmbRoot struct {
	ID                       int64      `json:"id" db:"id"`
	Name                     string     `json:"name" db:"name"`
	Host                     string     `json:"host" db:"host"`
	Port                     int        `json:"port" db:"port"`
	Share                    string     `json:"share" db:"share"`
	Username                 string     `json:"username" db:"username"`
	Domain                   *string    `json:"domain" db:"domain"`
	Enabled                  bool       `json:"enabled" db:"enabled"`
	MaxDepth                 int        `json:"max_depth" db:"max_depth"`
	EnableDuplicateDetection bool       `json:"enable_duplicate_detection" db:"enable_duplicate_detection"`
	EnableMetadataExtraction bool       `json:"enable_metadata_extraction" db:"enable_metadata_extraction"`
	IncludePatterns          *string    `json:"include_patterns" db:"include_patterns"`
	ExcludePatterns          *string    `json:"exclude_patterns" db:"exclude_patterns"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
	LastScanAt               *time.Time `json:"last_scan_at" db:"last_scan_at"`
}

// FileMetadata represents file metadata
type FileMetadata struct {
	ID       int64  `json:"id" db:"id"`
	FileID   int64  `json:"file_id" db:"file_id"`
	Key      string `json:"key" db:"key"`
	Value    string `json:"value" db:"value"`
	DataType string `json:"data_type" db:"data_type"`
}

// DuplicateGroup represents a group of duplicate files
type DuplicateGroup struct {
	ID        int64     `json:"id" db:"id"`
	FileCount int       `json:"file_count" db:"file_count"`
	TotalSize int64     `json:"total_size" db:"total_size"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// VirtualPath represents a virtual file system path
type VirtualPath struct {
	ID         int64     `json:"id" db:"id"`
	Path       string    `json:"path" db:"path"`
	TargetType string    `json:"target_type" db:"target_type"`
	TargetID   int64     `json:"target_id" db:"target_id"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// ScanHistory represents scan operation history
type ScanHistory struct {
	ID             int64      `json:"id" db:"id"`
	StorageRootID  int64      `json:"storage_root_id" db:"storage_root_id"`
	ScanType       string     `json:"scan_type" db:"scan_type"`
	Status         string     `json:"status" db:"status"`
	StartTime      time.Time  `json:"start_time" db:"start_time"`
	EndTime        *time.Time `json:"end_time" db:"end_time"`
	FilesProcessed int        `json:"files_processed" db:"files_processed"`
	FilesAdded     int        `json:"files_added" db:"files_added"`
	FilesUpdated   int        `json:"files_updated" db:"files_updated"`
	FilesDeleted   int        `json:"files_deleted" db:"files_deleted"`
	ErrorCount     int        `json:"error_count" db:"error_count"`
	ErrorMessage   *string    `json:"error_message" db:"error_message"`
}

// FileWithMetadata represents a file with its metadata
type FileWithMetadata struct {
	File
	Metadata []FileMetadata `json:"metadata,omitempty"`
}

// DirectoryInfo represents directory information with statistics
type DirectoryInfo struct {
	Path            string    `json:"path" db:"path"`
	Name            string    `json:"name" db:"name"`
	StorageRootName string    `json:"storage_root_name" db:"storage_root_name"`
	FileCount       int       `json:"file_count" db:"file_count"`
	DirectoryCount  int       `json:"directory_count" db:"directory_count"`
	TotalSize       int64     `json:"total_size" db:"total_size"`
	DuplicateCount  int       `json:"duplicate_count" db:"duplicate_count"`
	ModifiedAt      time.Time `json:"modified_at" db:"modified_at"`
}

// SearchFilter represents search filter criteria
type SearchFilter struct {
	Query              string     `json:"query,omitempty"`
	Path               string     `json:"path,omitempty"`
	Name               string     `json:"name,omitempty"`
	Extension          string     `json:"extension,omitempty"`
	FileType           string     `json:"file_type,omitempty"`
	MimeType           string     `json:"mime_type,omitempty"`
	StorageRoots       []string   `json:"storage_roots,omitempty"`
	MinSize            *int64     `json:"min_size,omitempty"`
	MaxSize            *int64     `json:"max_size,omitempty"`
	ModifiedAfter      *time.Time `json:"modified_after,omitempty"`
	ModifiedBefore     *time.Time `json:"modified_before,omitempty"`
	IncludeDeleted     bool       `json:"include_deleted"`
	OnlyDuplicates     bool       `json:"only_duplicates"`
	ExcludeDuplicates  bool       `json:"exclude_duplicates"`
	IncludeDirectories bool       `json:"include_directories"`
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string `json:"field"` // name, size, modified_at, created_at, path, extension
	Order string `json:"order"` // asc, desc
}

// PaginationOptions represents pagination options
type PaginationOptions struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// SearchResult represents search results with pagination
type SearchResult struct {
	Files      []FileWithMetadata `json:"files"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"total_pages"`
}

// StressTest represents a stress test configuration
type StressTest struct {
	ID              int64                `json:"id" db:"id"`
	UserID          int                  `json:"user_id" db:"user_id"`
	Name            string               `json:"name" db:"name"`
	Description     string               `json:"description" db:"description"`
	Type            string               `json:"type" db:"type"`
	Status          string               `json:"status" db:"status"`
	Scenarios       []StressTestScenario `json:"scenarios" db:"scenarios"`
	Configuration   StressTestConfig     `json:"configuration" db:"configuration"`
	ConcurrentUsers int                  `json:"concurrent_users" db:"concurrent_users"`
	Duration        time.Duration        `json:"duration" db:"duration"`
	DurationSeconds int                  `json:"duration_seconds" db:"duration_seconds"`
	RampUpTime      int                  `json:"ramp_up_time" db:"ramp_up_time"`
	RequestTimeout  int                  `json:"request_timeout" db:"request_timeout"`
	RequestDelay    int                  `json:"request_delay" db:"request_delay"`
	StartedAt       *time.Time           `json:"started_at" db:"started_at"`
	CompletedAt     *time.Time           `json:"completed_at" db:"completed_at"`
	CreatedBy       string               `json:"created_by" db:"created_by"`
	CreatedAt       time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time            `json:"updated_at" db:"updated_at"`
}

// StressTestScenario represents a test scenario
type StressTestScenario struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Actions     []StressTestAction `json:"actions"`
	URL         string             `json:"url"`
	Method      string             `json:"method"`
	Weight      int                `json:"weight"`
	RequestBody *string            `json:"request_body,omitempty"`
	Headers     map[string]string  `json:"headers,omitempty"`
}

// StressTestAction represents an action in a scenario
type StressTestAction struct {
	Type   string                 `json:"type"`
	Target string                 `json:"target"`
	Params map[string]interface{} `json:"params"`
	Delay  int                    `json:"delay"`
}

// StressTestConfig represents test configuration
type StressTestConfig struct {
	Timeout       int  `json:"timeout"`
	RetryCount    int  `json:"retry_count"`
	EnableMetrics bool `json:"enable_metrics"`
	EnableLogging bool `json:"enable_logging"`
}

// StressTestExecution represents a test execution
type StressTestExecution struct {
	ID           int64                      `json:"id" db:"id"`
	StressTestID int64                      `json:"stress_test_id" db:"stress_test_id"`
	Status       string                     `json:"status" db:"status"`
	StartTime    time.Time                  `json:"start_time" db:"start_time"`
	EndTime      *time.Time                 `json:"end_time" db:"end_time"`
	StartedAt    *time.Time                 `json:"started_at" db:"started_at"`
	CompletedAt  *time.Time                 `json:"completed_at" db:"completed_at"`
	Results      StressTestExecutionResults `json:"results" db:"results"`
	Metrics      map[string]interface{}     `json:"metrics" db:"metrics"`
	ErrorMessage *string                    `json:"error_message" db:"error_message"`
}

// StressTestExecutionResults represents execution results
type StressTestExecutionResults struct {
	TotalRequests       int64                  `json:"total_requests"`
	SuccessfulRequests  int64                  `json:"successful_requests"`
	FailedRequests      int64                  `json:"failed_requests"`
	AverageResponseTime float64                `json:"average_response_time"`
	MinResponseTime     float64                `json:"min_response_time"`
	MaxResponseTime     float64                `json:"max_response_time"`
	ErrorRate           float64                `json:"error_rate"`
	Throughput          float64                `json:"throughput"`
	Metrics             map[string]interface{} `json:"metrics"`
}

// StressTestStatistics represents test statistics
type StressTestStatistics struct {
	StressTestID         int64          `json:"stress_test_id" db:"stress_test_id"`
	TotalExecutions      int            `json:"total_executions"`
	TotalTests           int            `json:"total_tests"`
	SuccessfulRuns       int            `json:"successful_runs"`
	FailedRuns           int            `json:"failed_runs"`
	AverageDuration      float64        `json:"average_duration"`
	AvgExecutionDuration float64        `json:"avg_execution_duration"`
	LastExecutionTime    *time.Time     `json:"last_execution_time"`
	TestsByStatus        map[string]int `json:"tests_by_status"`
}

// StressTestResult represents the result of a stress test execution
type StressTestResult struct {
	TestID            int64             `json:"test_id"`
	Status            string            `json:"status"`
	StartTime         time.Time         `json:"start_time"`
	EndTime           *time.Time        `json:"end_time"`
	CompletedAt       *time.Time        `json:"completed_at"`
	Duration          time.Duration     `json:"duration"`
	Metrics           StressTestMetrics `json:"metrics"`
	Errors            []string          `json:"errors"`
	Recommendations   []string          `json:"recommendations"`
	Summary           string            `json:"summary"`
	ErrorMessage      *string           `json:"error_message"`
	TotalRequests     int64             `json:"total_requests"`
	SuccessfulReqs    int64             `json:"successful_reqs"`
	FailedRequests    int64             `json:"failed_requests"`
	RequestsPerSecond float64           `json:"requests_per_second"`
	AvgResponseTime   float64           `json:"avg_response_time"`
	MinResponseTime   float64           `json:"min_response_time"`
	MaxResponseTime   float64           `json:"max_response_time"`
	ErrorRate         float64           `json:"error_rate"`
	StatusCodeDist    map[string]int    `json:"status_code_dist"`
	ErrorDistribution map[string]int    `json:"error_distribution"`
}

// StressTestMetrics represents performance metrics from a stress test
type StressTestMetrics struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	RequestsPerSecond   float64       `json:"requests_per_second"`
	ErrorRate           float64       `json:"error_rate"`
	Throughput          float64       `json:"throughput"`
}

// TestScenario represents a test scenario for stress testing
type TestScenario struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	Weight      int               `json:"weight"` // For weighted random selection
}

// StressTestStatus represents the status of a stress test
type StressTestStatus struct {
	TestID            int64         `json:"test_id"`
	Status            string        `json:"status"`
	Message           string        `json:"message"`
	Progress          float64       `json:"progress"`
	StartTime         time.Time     `json:"start_time"`
	UpdateTime        time.Time     `json:"update_time"`
	CreatedAt         time.Time     `json:"created_at"`
	StartedAt         time.Time     `json:"started_at"`
	IsRunning         bool          `json:"is_running"`
	ElapsedTime       time.Duration `json:"elapsed_time"`
	TotalRequests     int64         `json:"total_requests"`
	SuccessfulReqs    int64         `json:"successful_reqs"`
	FailedRequests    int64         `json:"failed_requests"`
	RequestsPerSecond float64       `json:"requests_per_second"`
	ErrorRate         float64       `json:"error_rate"`
}

// LoadTestReport represents a comprehensive load test report
type LoadTestReport struct {
	Test            *StressTest       `json:"test"`
	Result          *StressTestResult `json:"result"`
	GeneratedAt     time.Time         `json:"generated_at"`
	TestID          int64             `json:"test_id"`
	TestName        string            `json:"test_name"`
	StartTime       time.Time         `json:"start_time"`
	EndTime         time.Time         `json:"end_time"`
	Duration        time.Duration     `json:"duration"`
	Status          string            `json:"status"`
	Summary         string            `json:"summary"`
	Metrics         StressTestMetrics `json:"metrics"`
	Recommendations []string          `json:"recommendations"`
	Charts          []ChartData       `json:"charts"`
}

// ChartData represents data for charts in reports
type ChartData struct {
	Name   string    `json:"name"`
	Type   string    `json:"type"` // line, bar, pie
	Labels []string  `json:"labels"`
	Data   []float64 `json:"data"`
}

// SystemLoadMetrics represents system load metrics
type SystemLoadMetrics struct {
	CPUUsage    float64   `json:"cpu_usage"`
	MemoryUsage float64   `json:"memory_usage"`
	DiskUsage   float64   `json:"disk_usage"`
	NetworkIO   float64   `json:"network_io"`
	Timestamp   time.Time `json:"timestamp"`
	ActiveTests int       `json:"active_tests"`
}

// MediaItem represents a media item for testing
type MediaItem struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Path   string `json:"path"`
	Size   int64  `json:"size"`
}

// Stress test status constants
const (
	StressTestStatusPending   = "pending"
	StressTestStatusRunning   = "running"
	StressTestStatusCompleted = "completed"
	StressTestStatusFailed    = "failed"
	StressTestStatusTimeout   = "timeout"
	StressTestStatusCancelled = "cancelled"
)

// WizardSession represents a configuration wizard session
type WizardSession struct {
	SessionID     string                 `json:"session_id" db:"session_id"`
	UserID        int                    `json:"user_id" db:"user_id"`
	CurrentStep   int                    `json:"current_step" db:"current_step"`
	TotalSteps    int                    `json:"total_steps" db:"total_steps"`
	StepData      map[string]interface{} `json:"step_data" db:"step_data"`
	Configuration map[string]interface{} `json:"configuration" db:"configuration"`
	StartedAt     time.Time              `json:"started_at" db:"started_at"`
	LastActivity  time.Time              `json:"last_activity" db:"last_activity"`
	IsCompleted   bool                   `json:"is_completed" db:"is_completed"`
	ConfigType    string                 `json:"config_type" db:"config_type"`
}

// ConfigurationProfile represents a saved configuration profile
type ConfigurationProfile struct {
	ProfileID     string                 `json:"profile_id" db:"profile_id"`
	Name          string                 `json:"name" db:"name"`
	Description   string                 `json:"description" db:"description"`
	UserID        int                    `json:"user_id" db:"user_id"`
	Configuration map[string]interface{} `json:"configuration" db:"configuration"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" db:"updated_at"`
	IsActive      bool                   `json:"is_active" db:"is_active"`
	Tags          []string               `json:"tags" db:"tags"`
}
