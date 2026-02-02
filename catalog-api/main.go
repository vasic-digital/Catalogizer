package main

import (
	root_config "catalogizer/config"
	"catalogizer/database"
	root_handlers "catalogizer/handlers"
	"catalogizer/internal/auth"
	internal_config "catalogizer/internal/config"
	"catalogizer/internal/handlers"
	"catalogizer/internal/middleware"
	"catalogizer/internal/services"
	root_middleware "catalogizer/middleware"
	root_repository "catalogizer/repository"
	root_services "catalogizer/services"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mutecomm/go-sqlcipher"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
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

	// Initialize database
	// Use the Database field as the path for SQLite
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "./data/catalogizer.db" // Default path
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize database connection wrapper
	databaseDB, err := database.NewConnection(&root_config.DatabaseConfig{
		Path: dbPath,
	})
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Run database migrations
	ctx := context.Background()
	log.Println("Running database migrations...")
	if err := databaseDB.RunMigrations(ctx); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}
	log.Println("Database migrations completed successfully")

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
	catalogService.SetDB(db)
	smbService := services.NewSMBService(internalCfg, logger)
	smbDiscoveryService := services.NewSMBDiscoveryService(logger)
	
	// Initialize services needed for recommendations
	mediaRecognitionService := services.NewMediaRecognitionService(db, logger, nil, nil, "", "", "", "", "", "")
	duplicateDetectionService := services.NewDuplicateDetectionService(db, logger, nil)
	fileRepository := root_repository.NewFileRepository(databaseDB)
	recommendationService := services.NewRecommendationService(
		mediaRecognitionService,
		duplicateDetectionService,
		fileRepository,
		db,
	)


	// Initialize repositories
	userRepo := root_repository.NewUserRepository(db)
	conversionRepo := root_repository.NewConversionRepository(db)
	analyticsRepo := root_repository.NewAnalyticsRepository(db)
	configurationRepo := root_repository.NewConfigurationRepository(db)
	errorReportingRepo := root_repository.NewErrorReportingRepository(db)
	crashReportingRepo := root_repository.NewCrashReportingRepository(db)
	logManagementRepo := root_repository.NewLogManagementRepository(db)
	favoritesRepo := root_repository.NewFavoritesRepository(db)

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
	internalAuthService := auth.NewAuthService(db, jwtSecret, logger)
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
	
	// Initialize subtitle service
	// Use SQL-based cache service for now
	cacheService := services.NewCacheService(db, logger)
	subtitleService := services.NewSubtitleService(db, logger, cacheService)

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

	// Stats handler
	statsRepo := root_repository.NewStatsRepository(databaseDB)
	statsHandler := root_handlers.NewStatsHandler(fileRepository, statsRepo)

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
	router.Use(middleware.Logger(logger))
	router.Use(middleware.ErrorHandler())
	router.Use(root_middleware.RequestID())
	router.Use(root_middleware.InputValidation(root_middleware.DefaultInputValidationConfig()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

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
		api.GET("/storage/roots", copyHandler.GetStorageRoots)

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
	if err := db.Close(); err != nil {
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
