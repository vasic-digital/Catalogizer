package main

import (
	root_config "catalogizer/config"
	"catalogizer/database"
	root_handlers "catalogizer/handlers"
	root_middleware "catalogizer/middleware"
	root_repository "catalogizer/repository"
	root_services "catalogizer/services"
	internal_config "catalogizer/internal/config"
	"catalogizer/internal/handlers"
	"catalogizer/internal/middleware"
	"catalogizer/internal/services"
	"context"
	"database/sql"
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
	defer db.Close()

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
	
	// Initialize repositories
	userRepo := root_repository.NewUserRepository(db)
	conversionRepo := root_repository.NewConversionRepository(db)
	
	// Initialize authentication and conversion services
	jwtSecret := cfg.Auth.JWTSecret
	if jwtSecret == "" {
		// Generate a default JWT secret if not configured
		jwtSecret = "default-secret-change-in-production"
		log.Println("WARNING: Using default JWT secret. Please set JWTSecret in config for production.")
	}
	authService := root_services.NewAuthService(userRepo, jwtSecret)
	conversionService := root_services.NewConversionService(conversionRepo, userRepo, authService)

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

	// Initialize handlers
	catalogHandler := handlers.NewCatalogHandler(catalogService, smbService, logger)
	downloadHandler := handlers.NewDownloadHandler(catalogService, smbService, cfg.Catalog.TempDir, cfg.Catalog.MaxArchiveSize, cfg.Catalog.DownloadChunkSize, logger)
	copyHandler := handlers.NewCopyHandler(catalogService, smbService, cfg.Catalog.TempDir, logger)
	smbDiscoveryHandler := handlers.NewSMBDiscoveryHandler(smbDiscoveryService, logger)
	conversionHandler := root_handlers.NewConversionHandler(conversionService, authService)
	authHandler := root_handlers.NewAuthHandler(authService)
	
	// Initialize JWT middleware
	jwtMiddleware := root_middleware.NewJWTMiddleware(jwtSecret)

	// Initialize Redis rate limiters
	var authRateLimiter, defaultRateLimiter gin.HandlerFunc
	if redisClient != nil {
		// Use Redis-based distributed rate limiting
		authRateLimiter = root_middleware.RedisRateLimit(root_middleware.AuthRedisRateLimiterConfig(redisClient))
		defaultRateLimiter = root_middleware.RedisRateLimit(root_middleware.DefaultRedisRateLimiterConfig(redisClient))
	} else {
		// Fall back to in-memory rate limiting
		authRateLimiter = root_middleware.AdvancedRateLimit(root_middleware.AuthRateLimiterConfig())
		defaultRateLimiter = root_middleware.IPRateLimit(100, 200)
	}

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
	api.Use(defaultRateLimiter) // Apply general rate limiting to API
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
		api.GET("/storage/list/*path", copyHandler.ListStoragePath)
		api.GET("/storage/roots", copyHandler.GetStorageRoots)

		// Statistics and sorting
		api.GET("/stats/directories/by-size", catalogHandler.GetDirectoriesBySize)
		api.GET("/stats/duplicates/count", catalogHandler.GetDuplicatesCount)

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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
