package main

import (
	"catalogizer/challenges"
	root_config "catalogizer/config"
	"catalogizer/database"
	"catalogizer/filesystem"
	root_handlers "catalogizer/handlers"
	"catalogizer/internal/auth"
	internal_config "catalogizer/internal/config"
	"catalogizer/internal/handlers"
	"catalogizer/internal/metrics"
	"catalogizer/internal/middleware"
	"catalogizer/internal/services"
	root_middleware "catalogizer/middleware"
	root_repository "catalogizer/repository"
	root_services "catalogizer/services"
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"digital.vasic.assets/pkg/defaults"
	"digital.vasic.assets/pkg/event"
	"digital.vasic.assets/pkg/manager"
	"digital.vasic.assets/pkg/resolver"
	asset_store "digital.vasic.assets/pkg/store"
	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"golang.org/x/crypto/bcrypt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Version information injected via ldflags at build time
var (
	Version     = "dev"
	BuildNumber = "0"
	BuildDate   = "unknown"
)

// atoi converts string to int with default fallback
func atoi(s string) int {
	if i, err := strconv.Atoi(s); err == nil {
		return i
	}
	return 8080 // default port
}

// @title Catalog API
// @version 2.0
// @description REST API for browsing and searching multi-protocol file catalog (SMB, FTP, NFS, WebDAV, Local)
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Parse command line flags
	testMode := flag.Bool("test-mode", false, "Run in test mode with additional logging")
	flag.Parse()

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	if *testMode {
		logger.Info("Running in test mode")
	}

	// Load configuration
	cfg, err := root_config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Override sensitive config with environment variables (security best practice)
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.Auth.JWTSecret = jwtSecret
	}
	if adminUser := os.Getenv("ADMIN_USERNAME"); adminUser != "" {
		cfg.Auth.AdminUsername = adminUser
	}
	if adminPass := os.Getenv("ADMIN_PASSWORD"); adminPass != "" {
		cfg.Auth.AdminPassword = adminPass
	}
	if port := os.Getenv("PORT"); port != "" {
		cfg.Server.Port = atoi(port) // Use helper function
	}
	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		gin.SetMode(ginMode)
	}

	// Apply DATABASE_* env overrides before creating connection
	if dbType := os.Getenv("DATABASE_TYPE"); dbType != "" {
		cfg.Database.Type = dbType
	}
	if dbHost := os.Getenv("DATABASE_HOST"); dbHost != "" {
		cfg.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DATABASE_PORT"); dbPort != "" {
		cfg.Database.Port = atoi(dbPort)
	}
	if dbName := os.Getenv("DATABASE_NAME"); dbName != "" {
		cfg.Database.Name = dbName
	}
	if dbUser := os.Getenv("DATABASE_USER"); dbUser != "" {
		cfg.Database.User = dbUser
	}
	if dbPass := os.Getenv("DATABASE_PASSWORD"); dbPass != "" {
		cfg.Database.Password = dbPass
	}
	if dbSSL := os.Getenv("DATABASE_SSL_MODE"); dbSSL != "" {
		cfg.Database.SSLMode = dbSSL
	}

	// Default SQLite path if not set
	if cfg.Database.Path == "" {
		cfg.Database.Path = "./data/catalogizer.db"
	}
	// Default SSLMode
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}

	// Initialize single database connection
	databaseDB, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	log.Printf("Database connected: %s", databaseDB.DatabaseType())

	// Run database migrations
	ctx := context.Background()
	log.Println("Running database migrations...")
	if err := databaseDB.RunMigrations(ctx); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}
	log.Println("Database migrations completed successfully")

	// Seed default admin user if none exists
	if err := seedDefaultAdmin(databaseDB, cfg.Auth.AdminUsername, cfg.Auth.AdminPassword); err != nil {
		log.Printf("Warning: failed to seed admin user: %v", err)
	}

	// Initialize services
	// Convert config to internal format
	internalCfg := &internal_config.Config{
		Server: internal_config.ServerConfig{
			Host:         cfg.Server.Host,
			Port:         fmt.Sprintf("%d", cfg.Server.Port),
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
			EnableCORS:   cfg.Server.EnableCORS,
			EnableHTTPS:  cfg.Server.EnableHTTPS,
		},
		Database: internal_config.DatabaseConfig{
			Database: cfg.Database.Path,
		},
		Catalog: internal_config.CatalogConfig{
			TempDir:           cfg.Catalog.TempDir,
			MaxArchiveSize:    cfg.Catalog.MaxArchiveSize,
			DownloadChunkSize: cfg.Catalog.DownloadChunkSize,
		},
	}

	catalogService := services.NewCatalogService(internalCfg, logger)
	catalogService.SetDB(databaseDB)
	smbService := services.NewSMBService(internalCfg, logger)
	smbDiscoveryService := services.NewSMBDiscoveryService(logger)

	// Initialize services needed for recommendations
	mediaRecognitionService := services.NewMediaRecognitionService(databaseDB, logger, nil, nil, "", "", "", "", "", "")
	duplicateDetectionService := services.NewDuplicateDetectionService(databaseDB, logger, nil)
	fileRepository := root_repository.NewFileRepository(databaseDB)
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		fileRepository,
		databaseDB,
	)

	// Initialize repositories
	userRepo := root_repository.NewUserRepository(databaseDB)
	conversionRepo := root_repository.NewConversionRepository(databaseDB)
	analyticsRepo := root_repository.NewAnalyticsRepository(databaseDB)
	configurationRepo := root_repository.NewConfigurationRepository(databaseDB)
	errorReportingRepo := root_repository.NewErrorReportingRepository(databaseDB)
	crashReportingRepo := root_repository.NewCrashReportingRepository(databaseDB)
	logManagementRepo := root_repository.NewLogManagementRepository(databaseDB)
	favoritesRepo := root_repository.NewFavoritesRepository(databaseDB)

	// Initialize authentication and conversion services
	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" || jwtSecret == "change-this-secret-in-production" {
		// Generate a cryptographically secure random secret at startup
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			log.Fatal("FATAL: Failed to generate JWT secret: ", err)
		}
		jwtSecret = hex.EncodeToString(secretBytes)
		log.Println("WARNING: No JWT secret configured. Generated ephemeral secret. Set Auth.JWTSecret in config for persistent sessions across restarts.")
	}
	authService := root_services.NewAuthService(userRepo, jwtSecret)
	conversionService := root_services.NewConversionService(conversionRepo, userRepo, authService)
	analyticsService := root_services.NewAnalyticsService(analyticsRepo)
	reportingService := root_services.NewReportingService(analyticsRepo, userRepo)
	configurationService := root_services.NewConfigurationService(configurationRepo, "./config.json")
	errorReportingService := root_services.NewErrorReportingService(errorReportingRepo, crashReportingRepo)
	logManagementService := root_services.NewLogManagementService(logManagementRepo)
	favoritesService := root_services.NewFavoritesService(favoritesRepo, authService)

	// Initialize internal auth service and middleware for rate limiting
	internalAuthService := auth.NewAuthService(databaseDB, jwtSecret, logger)
	authMiddleware := auth.NewAuthMiddleware(internalAuthService, logger)

	// Initialize Redis client for distributed rate limiting
	redisClient := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Test Redis connection
	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Printf("Warning: Redis connection failed (%v), falling back to in-memory rate limiting", err)
		redisClient = nil
	} else {
		log.Println("Redis connected successfully for distributed rate limiting")
	}
	
	// Initialize challenge service
	challengeService := root_services.NewChallengeService(
		filepath.Join(".", "data", "challenge_results"),
	)
	challenges.RegisterAll(challengeService)

	// Initialize media entity repositories
	mediaItemRepo := root_repository.NewMediaItemRepository(databaseDB)
	mediaFileRepo := root_repository.NewMediaFileRepository(databaseDB)
	extMetaRepo := root_repository.NewExternalMetadataRepository(databaseDB)
	userMetaRepo := root_repository.NewUserMetadataRepository(databaseDB)
	dirAnalysisRepo := root_repository.NewDirectoryAnalysisRepository(databaseDB)

	// Initialize universal scanner for file system scanning
	clientFactory := filesystem.NewDefaultClientFactory()
	universalScanner := services.NewUniversalScanner(databaseDB, logger, nil, clientFactory)
	if err := universalScanner.Start(); err != nil {
		log.Fatalf("Failed to start universal scanner: %v", err)
	}
	defer universalScanner.Stop()

	// Initialize aggregation service and hook into scanner
	aggregationService := services.NewAggregationService(databaseDB, logger, mediaItemRepo, mediaFileRepo, dirAnalysisRepo, extMetaRepo)
	universalScanner.SetAggregationService(aggregationService)

	// Initialize subtitle service
	// Use SQL-based cache service for now
	cacheService := services.NewCacheService(databaseDB, logger)
	subtitleService := services.NewSubtitleService(databaseDB, logger, cacheService)

	// Initialize handlers
	catalogHandler := handlers.NewCatalogHandler(catalogService, smbService, logger)
	downloadHandler := handlers.NewDownloadHandler(catalogService, smbService, cfg.Catalog.TempDir, cfg.Catalog.MaxArchiveSize, cfg.Catalog.DownloadChunkSize, logger)
	copyHandler := handlers.NewCopyHandler(catalogService, smbService, cfg.Catalog.TempDir, logger)
	smbDiscoveryHandler := handlers.NewSMBDiscoveryHandler(smbDiscoveryService, logger)
	conversionHandler := root_handlers.NewConversionHandler(conversionService, authService)
	authHandler := root_handlers.NewAuthHandler(authService)
	androidTVMediaHandler := root_handlers.NewAndroidTVMediaHandler(databaseDB)
	
	// Simple recommendation handler for testing
	simpleRecHandler := root_handlers.NewSimpleRecommendationHandler()
	
	// Recommendation handler
	recommendationHandler := root_handlers.NewRecommendationHandler(recommendationService)
	
	// Subtitle handler
	subtitleHandler := root_handlers.NewSubtitleHandler(subtitleService, logger)

	// Challenge handler
	challengeHandler := root_handlers.NewChallengeHandler(challengeService)

	// Stats handler
	statsRepo := root_repository.NewStatsRepository(databaseDB)
	statsHandler := root_handlers.NewStatsHandler(fileRepository, statsRepo)

	// Media browse handler (wires /media/search and /media/stats to the database)
	mediaBrowseHandler := root_handlers.NewMediaBrowseHandler(fileRepository, statsRepo, databaseDB)

	// WebSocket handler for real-time updates
	wsHandler := root_handlers.NewWebSocketHandler()

	// Initialize asset management system
	assetRepo := root_repository.NewAssetRepository(databaseDB)
	assetStore, err := asset_store.NewFileStore(filepath.Join(".", "cache", "assets"))
	if err != nil {
		log.Printf("Warning: failed to create asset store: %v", err)
	}
	assetEventBus := event.NewInMemoryBus()
	assetResolver := resolver.NewChain(
		services.NewCachedFileResolver(filepath.Join(".", "cache", "cover_art"), 1),
		services.NewExternalMetadataResolver(databaseDB, 2),
		services.NewLocalScanResolver(4),
	)
	assetManager := manager.New(
		manager.WithStore(assetStore),
		manager.WithResolver(assetResolver),
		manager.WithEventBus(assetEventBus),
		manager.WithDefaults(defaults.NewEmbeddedProvider()),
		manager.WithWorkers(4),
	)
	defer assetManager.Stop()
	assetHandler := root_handlers.NewAssetHandler(assetManager, assetRepo)

	// Bridge asset events to WebSocket clients
	assetEventBus.Subscribe(func(evt event.Event) {
		if evt.Type == event.AssetReady || evt.Type == event.AssetFailed {
			wsHandler.BroadcastToClients(map[string]interface{}{
				"type":        "asset_update",
				"action":      string(evt.Type),
				"asset_id":    string(evt.AssetID),
				"asset_type":  string(evt.AssetType),
				"entity_type": evt.Metadata["entity_type"],
				"entity_id":   evt.Metadata["entity_id"],
			})
		}
	})

	// Media entity handler for structured media browsing
	mediaEntityHandler := root_handlers.NewMediaEntityHandler(mediaItemRepo, mediaFileRepo, extMetaRepo, userMetaRepo)

	// Scan handler for storage roots and scan operations
	scanHandler := root_handlers.NewScanHandler(universalScanner, databaseDB)

	// Create service adapters to bridge interface differences between services and handlers
	authAdapter := &root_handlers.AuthServiceAdapter{Inner: authService}
	configAdapter := &root_handlers.ConfigurationServiceAdapter{Inner: configurationService}
	errorAdapter := &root_handlers.ErrorReportingServiceAdapter{Inner: errorReportingService}
	logAdapter := &root_handlers.LogManagementServiceAdapter{Inner: logManagementService}

	// User management, role, configuration, error reporting, and log management handlers
	userHandler := root_handlers.NewUserHandler(userRepo, authAdapter)
	roleHandler := root_handlers.NewRoleHandler(userRepo, authAdapter)
	configurationHandler := root_handlers.NewConfigurationHandler(configAdapter, authAdapter)
	errorReportingHandler := root_handlers.NewErrorReportingHandler(errorAdapter, authAdapter)
	logManagementHandler := root_handlers.NewLogManagementHandler(logAdapter, authAdapter)

	// Analytics and reporting services used by stats handler via repositories
	_ = analyticsService
	_ = reportingService
	_ = favoritesService

	// Initialize JWT middleware
	jwtMiddleware := root_middleware.NewJWTMiddleware(jwtSecret)

	// Initialize rate limiters using internal auth middleware
	authRateLimiter := authMiddleware.RateLimitByUser(5, "1m")      // 5 requests per minute for auth
	defaultRateLimiter := authMiddleware.RateLimitByUser(100, "1m") // 100 requests per minute default

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(root_middleware.CORS())
	router.Use(metrics.GinMiddleware())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.ErrorHandler())
	router.Use(root_middleware.RequestID())
	router.Use(root_middleware.InputValidation(root_middleware.DefaultInputValidationConfig()))

	// Start runtime metrics collector (goroutines, memory)
	metrics.StartRuntimeCollector(15 * time.Second)

	// Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":       "healthy",
			"time":         time.Now().UTC(),
			"version":      Version,
			"build_number": BuildNumber,
			"build_date":   BuildDate,
		})
	})

	// WebSocket endpoint (auth via query parameter, not header)
	router.GET("/ws", wsHandler.HandleConnection)

	// Asset serving (public — no auth needed for serving images)
	router.GET("/api/v1/assets/:id", assetHandler.ServeAsset)

	// Authentication routes (no auth required)
	authGroup := router.Group("/api/v1/auth")
	authGroup.Use(authRateLimiter) // Apply strict rate limiting to auth endpoints
	{
		authGroup.POST("/login", authHandler.LoginGin)
		authGroup.POST("/register", func(c *gin.Context) {
			authHandler.RegisterGin(c, userRepo)
		})
		authGroup.POST("/refresh", authHandler.RefreshTokenGin)
		authGroup.POST("/logout", authHandler.LogoutGin)
		authGroup.GET("/me", jwtMiddleware.RequireAuth(), authHandler.GetCurrentUserGin)
		authGroup.GET("/status", authHandler.GetAuthStatusGin)
		authGroup.GET("/permissions", jwtMiddleware.RequireAuth(), authHandler.GetPermissionsGin)
		authGroup.GET("/profile", jwtMiddleware.RequireAuth(), authHandler.GetCurrentUserGin)
	}

	// API routes
	api := router.Group("/api/v1")
	api.Use(jwtMiddleware.RequireAuth()) // Apply auth middleware to all API routes
	api.Use(defaultRateLimiter)          // Apply general rate limiting to API
	{
		// Catalog browsing endpoints
		api.GET("/catalog", catalogHandler.ListRoot)
		api.GET("/catalog/*path", catalogHandler.ListPath)
		api.GET("/catalog-info/*path", catalogHandler.GetFileInfo)

		// Search endpoints
		api.GET("/search", catalogHandler.Search)
		api.GET("/search/duplicates", catalogHandler.SearchDuplicates)

		// Download endpoints
		api.GET("/download/file/:id", downloadHandler.DownloadFile)
		api.GET("/download/directory/*path", downloadHandler.DownloadDirectory)
		api.POST("/download/archive", downloadHandler.DownloadArchive)

		// File operations
		api.POST("/copy/storage", copyHandler.CopyToStorage)
		api.POST("/copy/local", copyHandler.CopyToLocal)
		api.POST("/copy/upload", copyHandler.CopyFromLocal)

		// Media browsing endpoints (must be before :id to prevent route conflict)
		api.GET("/media/search", mediaBrowseHandler.SearchMedia)
		api.GET("/media/stats", mediaBrowseHandler.GetMediaStats)

		// Media operations
		api.GET("/media/:id", androidTVMediaHandler.GetMediaByID)
		api.PUT("/media/:id/progress", androidTVMediaHandler.UpdateWatchProgress)
		api.PUT("/media/:id/favorite", androidTVMediaHandler.UpdateFavoriteStatus)
		
		// Test recommendation endpoints
		api.GET("/recommendations/test", simpleRecHandler.GetSimpleRecommendation)
		api.GET("/recommendations/error", simpleRecHandler.GetTest)
		
		// Recommendation endpoints
		recGroup := api.Group("/recommendations")
		{
			recGroup.GET("/similar/:media_id", recommendationHandler.GetSimilarItems)
			recGroup.GET("/trending", recommendationHandler.GetTrendingItems)
			recGroup.GET("/personalized/:user_id", recommendationHandler.GetPersonalizedRecommendations)
		}
		
		// Subtitle endpoints
		subGroup := api.Group("/subtitles")
		{
			subGroup.GET("/search", subtitleHandler.SearchSubtitles)
			subGroup.POST("/download", subtitleHandler.DownloadSubtitle)
			subGroup.GET("/media/:media_id", subtitleHandler.GetSubtitles)
			subGroup.GET("/:subtitle_id/verify-sync/:media_id", subtitleHandler.VerifySubtitleSync)
			subGroup.POST("/translate", subtitleHandler.TranslateSubtitle)
			subGroup.POST("/upload", subtitleHandler.UploadSubtitle)
			subGroup.GET("/languages", subtitleHandler.GetSupportedLanguages)
			subGroup.GET("/providers", subtitleHandler.GetSupportedProviders)
		}
		api.GET("/storage/list/*path", copyHandler.ListStoragePath)
		api.GET("/storage/roots", scanHandler.GetStorageRoots)
		api.POST("/storage/roots", scanHandler.CreateStorageRoot)

		// Statistics and sorting
		api.GET("/stats/directories/by-size", catalogHandler.GetDirectoriesBySize)
		api.GET("/stats/duplicates/count", catalogHandler.GetDuplicatesCount)

		// Advanced statistics endpoints
		statsGroup := api.Group("/stats")
		{
			statsGroup.GET("/overall", statsHandler.GetOverallStats)
			statsGroup.GET("/smb/:smb_root", statsHandler.GetSmbRootStats)
			statsGroup.GET("/filetypes", statsHandler.GetFileTypeStats)
			statsGroup.GET("/sizes", statsHandler.GetSizeDistribution)
			statsGroup.GET("/duplicates", statsHandler.GetDuplicateStats)
			statsGroup.GET("/duplicates/groups", statsHandler.GetTopDuplicateGroups)
			statsGroup.GET("/access", statsHandler.GetAccessPatterns)
			statsGroup.GET("/growth", statsHandler.GetGrowthTrends)
			statsGroup.GET("/scans", statsHandler.GetScanHistory)
		}

		// SMB Discovery endpoints
		smbGroup := api.Group("/smb")
		{
			smbGroup.POST("/discover", smbDiscoveryHandler.DiscoverShares)
			smbGroup.GET("/discover", smbDiscoveryHandler.DiscoverSharesGET)
			smbGroup.POST("/test", smbDiscoveryHandler.TestConnection)
			smbGroup.GET("/test", smbDiscoveryHandler.TestConnectionGET)
			smbGroup.POST("/browse", smbDiscoveryHandler.BrowseShare)
		}

		// Scan endpoints
		scanGroup := api.Group("/scans")
		{
			scanGroup.POST("", scanHandler.QueueScan)
			scanGroup.GET("", scanHandler.ListScans)
			scanGroup.GET("/:job_id", scanHandler.GetScanStatus)
		}

		// Conversion endpoints
		conversionGroup := api.Group("/conversion")
		{
			conversionGroup.POST("/jobs", conversionHandler.CreateJob)
			conversionGroup.GET("/jobs", conversionHandler.ListJobs)
			conversionGroup.GET("/jobs/:id", conversionHandler.GetJob)
			conversionGroup.POST("/jobs/:id/cancel", conversionHandler.CancelJob)
			conversionGroup.GET("/formats", conversionHandler.GetSupportedFormats)
		}

		// User management endpoints
		wrap := root_handlers.WrapHTTPHandler
		usersGroup := api.Group("/users")
		{
			usersGroup.POST("", wrap(userHandler.CreateUser))
			usersGroup.GET("", wrap(userHandler.ListUsers))
			usersGroup.GET("/:id", wrap(userHandler.GetUser))
			usersGroup.PUT("/:id", wrap(userHandler.UpdateUser))
			usersGroup.DELETE("/:id", wrap(userHandler.DeleteUser))
			usersGroup.POST("/:id/reset-password", wrap(userHandler.ResetPassword))
			usersGroup.POST("/:id/lock", wrap(userHandler.LockAccount))
			usersGroup.POST("/:id/unlock", wrap(userHandler.UnlockAccount))
		}

		// Role management endpoints
		rolesGroup := api.Group("/roles")
		{
			rolesGroup.POST("", wrap(roleHandler.CreateRole))
			rolesGroup.GET("", wrap(roleHandler.ListRoles))
			rolesGroup.GET("/:id", wrap(roleHandler.GetRole))
			rolesGroup.PUT("/:id", wrap(roleHandler.UpdateRole))
			rolesGroup.DELETE("/:id", wrap(roleHandler.DeleteRole))
			rolesGroup.GET("/permissions", wrap(roleHandler.GetPermissions))
		}

		// Configuration endpoints
		configGroup := api.Group("/configuration")
		{
			configGroup.GET("", wrap(configurationHandler.GetConfiguration))
			configGroup.POST("/test", wrap(configurationHandler.TestConfiguration))
			configGroup.GET("/status", wrap(configurationHandler.GetSystemStatus))
			configGroup.GET("/wizard/step/:step_id", wrap(configurationHandler.GetWizardStep))
			configGroup.POST("/wizard/step/:step_id/validate", wrap(configurationHandler.ValidateWizardStep))
			configGroup.POST("/wizard/step/:step_id/save", wrap(configurationHandler.SaveWizardProgress))
			configGroup.GET("/wizard/progress", wrap(configurationHandler.GetWizardProgress))
			configGroup.POST("/wizard/complete", wrap(configurationHandler.CompleteWizard))
		}

		// Error reporting endpoints
		errorsGroup := api.Group("/errors")
		{
			errorsGroup.POST("/report", wrap(errorReportingHandler.ReportError))
			errorsGroup.POST("/crash", wrap(errorReportingHandler.ReportCrash))
			errorsGroup.GET("/reports", wrap(errorReportingHandler.ListErrorReports))
			errorsGroup.GET("/reports/:id", wrap(errorReportingHandler.GetErrorReport))
			errorsGroup.PUT("/reports/:id/status", wrap(errorReportingHandler.UpdateErrorStatus))
			errorsGroup.GET("/crashes", wrap(errorReportingHandler.ListCrashReports))
			errorsGroup.GET("/crashes/:id", wrap(errorReportingHandler.GetCrashReport))
			errorsGroup.PUT("/crashes/:id/status", wrap(errorReportingHandler.UpdateCrashStatus))
			errorsGroup.GET("/statistics", wrap(errorReportingHandler.GetErrorStatistics))
			errorsGroup.GET("/crash-statistics", wrap(errorReportingHandler.GetCrashStatistics))
			errorsGroup.GET("/health", wrap(errorReportingHandler.GetSystemHealth))
		}

		// Log management endpoints
		logsGroup := api.Group("/logs")
		{
			logsGroup.POST("/collect", wrap(logManagementHandler.CreateLogCollection))
			logsGroup.GET("/collections", wrap(logManagementHandler.ListLogCollections))
			logsGroup.GET("/collections/:id", wrap(logManagementHandler.GetLogCollection))
			logsGroup.GET("/collections/:id/entries", wrap(logManagementHandler.GetLogEntries))
			logsGroup.POST("/collections/:id/export", wrap(logManagementHandler.ExportLogs))
			logsGroup.GET("/collections/:id/analyze", wrap(logManagementHandler.AnalyzeLogs))
			logsGroup.POST("/share", wrap(logManagementHandler.CreateLogShare))
			logsGroup.GET("/share/:token", wrap(logManagementHandler.GetLogShare))
			logsGroup.DELETE("/share/:id", wrap(logManagementHandler.RevokeLogShare))
			logsGroup.GET("/stream", wrap(logManagementHandler.StreamLogs))
			logsGroup.GET("/statistics", wrap(logManagementHandler.GetLogStatistics))
		}

		// Asset management endpoints (authenticated)
		assetsGroup := api.Group("/assets")
		{
			assetsGroup.POST("/request", assetHandler.RequestAsset)
			assetsGroup.GET("/by-entity/:type/:id", assetHandler.GetByEntity)
		}

		// Media entity endpoints (structured media browsing)
		entityGroup := api.Group("/entities")
		{
			entityGroup.GET("", mediaEntityHandler.ListEntities)
			entityGroup.GET("/types", mediaEntityHandler.GetEntityTypes)
			entityGroup.GET("/stats", mediaEntityHandler.GetEntityStats)
			entityGroup.GET("/duplicates", mediaEntityHandler.ListDuplicateGroups)
			entityGroup.GET("/browse/:type", mediaEntityHandler.BrowseByType)
			entityGroup.GET("/:id", mediaEntityHandler.GetEntity)
			entityGroup.GET("/:id/children", mediaEntityHandler.GetEntityChildren)
			entityGroup.GET("/:id/files", mediaEntityHandler.GetEntityFiles)
			entityGroup.GET("/:id/metadata", mediaEntityHandler.GetEntityMetadata)
			entityGroup.GET("/:id/duplicates", mediaEntityHandler.GetEntityDuplicates)
			entityGroup.GET("/:id/stream", mediaEntityHandler.StreamEntity)
			entityGroup.GET("/:id/download", mediaEntityHandler.DownloadEntity)
			entityGroup.POST("/:id/metadata/refresh", mediaEntityHandler.RefreshEntityMetadata)
			entityGroup.PUT("/:id/user-metadata", mediaEntityHandler.UpdateUserMetadata)
		}

		// Challenge endpoints
		challengeGroup := api.Group("/challenges")
		{
			challengeGroup.GET("", challengeHandler.ListChallenges)
			challengeGroup.GET("/:id", challengeHandler.GetChallenge)
			challengeGroup.POST("/:id/run", challengeHandler.RunChallenge)
			challengeGroup.POST("/run", challengeHandler.RunAll)
			challengeGroup.POST("/run/category/:category", challengeHandler.RunByCategory)
			challengeGroup.GET("/results", challengeHandler.GetResults)
		}
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeout) * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting catalog API server", zap.String("address", cfg.GetServerAddress()))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop runtime metrics collector
	metrics.StopRuntimeCollector()

	// Shutdown HTTP server (stops accepting new connections, waits for in-flight requests)
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error", zap.Error(err))
	}

	// Close Redis connection if available
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("Redis connection close error", zap.Error(err))
		} else {
			logger.Info("Redis connection closed")
		}
	}

	// Close database connection
	if err := databaseDB.Close(); err != nil {
		logger.Error("Database close error", zap.Error(err))
	} else {
		logger.Info("Database connection closed")
	}

	logger.Info("Server exited cleanly")
}

// NopCacheService is a no-operation cache service for when Redis is not available
type NopCacheService struct{}

func (n *NopCacheService) Get(ctx context.Context, key string, dest interface{}) (bool, error) {
	return false, nil
}

func (n *NopCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

// seedDefaultAdmin creates a default admin user if none exists in the database.
// Uses the same password hashing scheme as services.AuthService (bcrypt(password + salt)).
func seedDefaultAdmin(db *database.DB, username, password string) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE role_id = 1").Scan(&count)
	if err != nil {
		return fmt.Errorf("check admin count: %w", err)
	}
	if count > 0 {
		return nil // admin already exists
	}

	// Generate salt
	saltBytes := make([]byte, 16)
	if _, err := rand.Read(saltBytes); err != nil {
		return fmt.Errorf("generate salt: %w", err)
	}
	salt := hex.EncodeToString(saltBytes)

	// Hash password with salt (same as services.AuthService.hashPassword)
	hash, err := bcryptHash([]byte(password + salt))
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = db.Exec(
		`INSERT INTO users (username, email, password_hash, salt, role_id, first_name, last_name, display_name, is_active)
		 VALUES (?, ?, ?, ?, 1, 'System', 'Administrator', 'Admin', ?)`,
		username, username+"@catalogizer.local", string(hash), salt, 1,
	)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}

	log.Printf("Default admin user '%s' created", username)
	return nil
}

// bcryptHash wraps bcrypt.GenerateFromPassword.
func bcryptHash(data []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(data, bcrypt.DefaultCost)
}
