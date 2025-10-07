package main

import (
	"catalog-api/internal/config"
	"catalog-api/internal/handlers"
	"catalog-api/internal/middleware"
	"catalog-api/internal/services"
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
)

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
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	dbConn, err := database.NewConnection(cfg.Database)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer dbConn.Close()

	// Run database migrations
	if err := dbConn.RunMigrations(context.Background()); err != nil {
		log.Fatal("Failed to run database migrations:", err)
	}

	// Initialize services
	catalogService := services.NewCatalogService(cfg, logger)
	catalogService.SetDB(dbConn.DB)
	fileSystemService := services.NewFileSystemService(cfg, logger)

	// Initialize handlers
	catalogHandler := handlers.NewCatalogHandler(catalogService, fileSystemService, logger)
	downloadHandler := handlers.NewDownloadHandler(catalogService, fileSystemService, cfg.Catalog.TempDir, cfg.Catalog.MaxArchiveSize, cfg.Catalog.DownloadChunkSize, logger)
	copyHandler := handlers.NewCopyHandler(catalogService, fileSystemService, cfg.Catalog.TempDir, logger)

	// Setup Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.ErrorHandler())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "time": time.Now().UTC()})
	})

	// API routes
	api := router.Group("/api/v1")
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